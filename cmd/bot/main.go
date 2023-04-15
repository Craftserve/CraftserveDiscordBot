package main

import (
	"csrvbot/commands"
	"csrvbot/internal/repos"
	"csrvbot/internal/services"
	"csrvbot/listeners"
	"csrvbot/pkg"
	"csrvbot/pkg/database"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
	"log"
	"os"
)

type Config struct {
	MysqlConfig        []database.MySQLConfiguration `json:"mysql_config"`
	GiveawayCron       string                        `json:"cron_line"`
	GiveawayTimeString string                        `json:"giveaway_time_string"`
	SystemToken        string                        `json:"system_token"`
	CsrvSecret         string                        `json:"csrv_secret"`
}

var BotConfig Config

func init() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Panic(err)
	}

	err = json.NewDecoder(configFile).Decode(&BotConfig)
	if err != nil {
		log.Panic("init#Decoder.Decode(&BotConfig)", err)
	}
}

func main() {
	ctx := pkg.CreateContext()
	db := database.NewProvider()

	err := db.InitMySQLDatabases(BotConfig.MysqlConfig)
	if err != nil {
		panic(err)
	}

	dbMap, err := db.GetMySQLDatabase("main")
	if err != nil {
		panic(err)
	}

	var giveawayRepo = repos.NewGiveawayRepo(dbMap)
	var serverRepo = repos.NewServerRepo(dbMap)
	var userRepo = repos.NewUserRepo(dbMap)

	err = db.CreateTablesIfNotExists()
	if err != nil {
		panic(err)
	}

	var csrvClient = services.NewCsrvClient(BotConfig.CsrvSecret)
	var githubClient = services.NewGithubClient()
	var giveawayService = services.NewGiveawayService(csrvClient, serverRepo, giveawayRepo)
	var helperService = services.NewHelperService(serverRepo, giveawayRepo, userRepo)

	session, err := discordgo.New("Bot " + BotConfig.SystemToken)
	if err != nil {
		panic(err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers

	var giveawayCommand = commands.NewGiveawayCommand(giveawayRepo, BotConfig.GiveawayTimeString)
	var thxCommand = commands.NewThxCommand(giveawayRepo, userRepo, serverRepo, BotConfig.GiveawayTimeString)
	var thxmeCommand = commands.NewThxmeCommand(giveawayRepo, userRepo, serverRepo, BotConfig.GiveawayTimeString)
	var csrvbotCommand = commands.NewCsrvbotCommand(BotConfig.GiveawayTimeString, serverRepo, giveawayRepo, userRepo, csrvClient, giveawayService, helperService)
	var docCommand = commands.NewDocCommand(githubClient)
	var resendCommand = commands.NewResendCommand(giveawayRepo)
	var interactionCreateListener = listeners.NewInteractionCreateListener(giveawayCommand, thxCommand, thxmeCommand, csrvbotCommand, docCommand, resendCommand, BotConfig.GiveawayTimeString, giveawayRepo, serverRepo, helperService)
	var guildCreateListener = listeners.NewGuildCreateListener(giveawayRepo, serverRepo, userRepo, giveawayService, helperService)
	var guildMemberAddListener = listeners.NewGuildMemberAddListener(userRepo)
	var guildMemberUpdateListener = listeners.NewGuildMemberUpdateListener(userRepo)
	session.AddHandler(interactionCreateListener.Handle)
	session.AddHandler(guildCreateListener.Handle)
	session.AddHandler(guildMemberAddListener.Handle)
	session.AddHandler(guildMemberUpdateListener.Handle)

	err = session.Open()
	if err != nil {
		panic(err)
	}

	log.Println("Bot logged in as", session.State.User)

	giveawayCommand.Register(session)
	thxCommand.Register(session)
	thxmeCommand.Register(session)
	csrvbotCommand.Register(session)
	docCommand.Register(session)
	resendCommand.Register(session)

	c := cron.New()
	_ = c.AddFunc(BotConfig.GiveawayCron, func() {
		giveawayService.FinishGiveaways(ctx, session)
	})
	c.Start()

	stop := make(chan os.Signal, 1)
	<-stop
	log.Println("Shutting down...")
	err = session.Close()
	if err != nil {
		log.Panicln("Could not close session", err)
	}
}
