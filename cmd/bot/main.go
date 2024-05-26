package main

import (
	"csrvbot/commands"
	"csrvbot/internal/repos"
	"csrvbot/internal/services"
	"csrvbot/listeners"
	"csrvbot/pkg"
	"csrvbot/pkg/database"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	"github.com/robfig/cron"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	MysqlConfig  []database.MySQLConfiguration `json:"mysql_config"`
	SentryConfig struct {
		DSN     string `json:"dsn"`
		Release string `json:"release"`
		Debug   bool   `json:"debug"`
	} `json:"sentry_config"`
	CraftserveUrl             string `json:"craftserve_url"`
	ThxGiveawayCron           string `json:"thx_giveaway_cron_line"`
	ThxGiveawayTimeString     string `json:"thx_giveaway_time_string"`
	MessageGiveawayCron       string `json:"message_giveaway_cron_line"`
	UnconditionalGiveawayCron string `json:"unconditional_giveaway_cron_line"`
	ConditionalGiveawayCron   string `json:"conditional_giveaway_cron_line"`
	SystemToken               string `json:"system_token"`
	CsrvSecret                string `json:"csrv_secret"`
	RegisterCommands          bool   `json:"register_commands"`
	Environment               string `json:"environment"` // development or production
	LevelPrefix               string `json:"level_prefix"`
}

var BotConfig Config

func init() {
	ctx := pkg.CreateContext()
	logger.ConfigureLogger()
	log := logger.GetLoggerFromContext(ctx)
	log.Debug("Opening config.json")
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Panic(err)
	}

	log.Debug("Decoding config.json")
	err = json.NewDecoder(configFile).Decode(&BotConfig)
	if err != nil {
		log.Panic("init#Decoder.Decode(&BotConfig)", err)
	}
}

func main() {
	ctx := pkg.CreateContext()
	log := logger.GetLoggerFromContext(ctx)

	log.Debugf("Setting discord level prefix to [%s]", BotConfig.LevelPrefix)
	discord.LevelPrefix = BotConfig.LevelPrefix

	if BotConfig.Environment == "development" {
		log.Warn("Running in developer mode!")
	}

	log.Debug("Initializing Sentry")
	initSentry(BotConfig.SentryConfig.DSN, BotConfig.Environment, BotConfig.SentryConfig.Release, BotConfig.SentryConfig.Debug)
	defer sentry.Flush(2 * time.Second)

	db := database.NewProvider()
	log.Debug("Initializing MySQL databases")
	err := db.InitMySQLDatabases(ctx, BotConfig.MysqlConfig)
	if err != nil {
		log.Panic(err)
	}

	log.Debug("Getting MySQL database")
	dbMap, err := db.GetMySQLDatabase("main")
	if err != nil {
		log.Panic(err)
	}

	var giveawayRepo = repos.NewGiveawayRepo(dbMap)
	var messageGiveawayRepo = repos.NewMessageGiveawayRepo(dbMap)
	var unconditionalGiveawayRepo = repos.NewUnconditionalGiveawayRepo(dbMap)
	var conditionalGiveawayRepo = repos.NewConditionalGiveawayRepo(dbMap)
	var serverRepo = repos.NewServerRepo(dbMap)
	var userRepo = repos.NewUserRepo(dbMap)

	log.Debug("Creating tables...")
	err = db.CreateTablesIfNotExists()
	if err != nil {
		log.Panic(err)
	}

	var csrvClient = services.NewCsrvClient(BotConfig.CsrvSecret, BotConfig.Environment)
	var githubClient = services.NewGithubClient()
	var giveawayService = services.NewGiveawayService(csrvClient, BotConfig.CraftserveUrl, serverRepo, giveawayRepo, messageGiveawayRepo, unconditionalGiveawayRepo, conditionalGiveawayRepo)
	var helperService = services.NewHelperService(serverRepo, giveawayRepo, userRepo)
	var savedRoleService = services.NewSavedRoleService(userRepo)

	log.Debug("Initializing discordgo session")
	session, err := discordgo.New("Bot " + BotConfig.SystemToken)
	if err != nil {
		log.Panic(err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers
	log.Debugf("Running with intents: Guilds, GuildMessages, GuildMembers (%v)", session.Identify.Intents)

	var giveawayCommand = commands.NewGiveawayCommand(giveawayRepo, BotConfig.ThxGiveawayTimeString, BotConfig.CraftserveUrl)
	var thxCommand = commands.NewThxCommand(giveawayRepo, userRepo, serverRepo, BotConfig.ThxGiveawayTimeString, BotConfig.CraftserveUrl)
	var thxmeCommand = commands.NewThxmeCommand(giveawayRepo, userRepo, serverRepo, BotConfig.ThxGiveawayTimeString)
	var csrvbotCommand = commands.NewCsrvbotCommand(BotConfig.CraftserveUrl, BotConfig.ThxGiveawayTimeString, serverRepo, giveawayRepo, userRepo, csrvClient, giveawayService, helperService)
	var docCommand = commands.NewDocCommand(githubClient)
	var resendCommand = commands.NewResendCommand(giveawayRepo, messageGiveawayRepo, BotConfig.CraftserveUrl)
	var interactionCreateListener = listeners.NewInteractionCreateListener(giveawayCommand, thxCommand, thxmeCommand, csrvbotCommand, docCommand, resendCommand, BotConfig.ThxGiveawayTimeString, BotConfig.CraftserveUrl, giveawayRepo, messageGiveawayRepo, serverRepo, helperService, unconditionalGiveawayRepo, conditionalGiveawayRepo)
	var guildCreateListener = listeners.NewGuildCreateListener(giveawayRepo, serverRepo, giveawayService, helperService, savedRoleService)
	var guildMemberAddListener = listeners.NewGuildMemberAddListener(userRepo)
	var guildMemberUpdateListener = listeners.NewGuildMemberUpdateListener(userRepo, savedRoleService)
	var messageCreateListener = listeners.NewMessageCreateListener(messageGiveawayRepo)
	session.AddHandler(interactionCreateListener.Handle)
	session.AddHandler(guildCreateListener.Handle)
	session.AddHandler(guildMemberAddListener.Handle)
	session.AddHandler(guildMemberUpdateListener.Handle)
	session.AddHandler(messageCreateListener.Handle)

	log.Debug("Opening discordgo session")
	err = session.Open()
	if err != nil {
		log.Panic(err)
	}

	log.WithField("username", session.State.User).Info("Bot logged in")

	if BotConfig.RegisterCommands {
		log.Debug("Registering commands")
		giveawayCommand.Register(ctx, session)
		thxCommand.Register(ctx, session)
		thxmeCommand.Register(ctx, session)
		csrvbotCommand.Register(ctx, session)
		docCommand.Register(ctx, session)
		resendCommand.Register(ctx, session)
	} else {
		log.Debug("Skipping command registration")
	}

	log.Debugf("Creating cron jobs: %s | %s | %s | %s", BotConfig.ThxGiveawayCron, BotConfig.MessageGiveawayCron, BotConfig.UnconditionalGiveawayCron, BotConfig.ConditionalGiveawayCron)
	c := cron.New()
	err = c.AddFunc(BotConfig.ThxGiveawayCron, func() {
		giveawayService.FinishGiveaways(ctx, session)
	})
	if err != nil {
		log.Errorf("Could not set thx giveaway cron job: %v", err)
	}

	err = c.AddFunc(BotConfig.MessageGiveawayCron, func() {
		giveawayService.FinishMessageGiveaways(ctx, session)
	})
	if err != nil {
		log.Errorf("Could not set message giveaway cron job: %v", err)
	}

	err = c.AddFunc(BotConfig.UnconditionalGiveawayCron, func() {
		giveawayService.FinishUnconditionalGiveaways(ctx, session)
	})
	if err != nil {
		log.Errorf("Could not set unconditional giveaway cron job: %v", err)
	}

	err = c.AddFunc(BotConfig.ConditionalGiveawayCron, func() {
		giveawayService.FinishConditionalGiveaways(ctx, session)
	})
	if err != nil {
		log.Errorf("Could not set conditional giveaway cron job: %v", err)
	}
	c.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("Shutting down...")
	err = session.Close()
	if err != nil {
		log.Panic("Could not close session", err)
	}
}

func initSentry(dsn, environment, release string, debug bool) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Release:          release,
		Debug:            debug,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		logger.Logger.WithError(err).Error("Could not initialize sentry")
	}
}
