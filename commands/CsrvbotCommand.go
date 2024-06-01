package commands

import (
	"context"
	"csrvbot/internal/repos"
	"csrvbot/internal/services"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"reflect"
	"strconv"
	"strings"
)

type CsrvbotCommand struct {
	Name                     string
	Description              string
	DMPermission             bool
	DefaultMemberPermissions int64
	Zero                     float64
	GiveawayHours            string
	CraftserveUrl            string
	ServerRepo               repos.ServerRepo
	GiveawayRepo             repos.GiveawayRepo
	UserRepo                 repos.UserRepo
	CsrvClient               services.CsrvClient
	GiveawayService          services.GiveawayService
	HelperService            services.HelperService
}

const (
	// CsrvbotCommand Subcommands
	SettingSubcommand           = "settings"
	DeleteSubcommand            = "delete"
	StartSubcommand             = "start"
	BlacklistSubcommand         = "blacklist"
	UnblacklistSubcommand       = "unblacklist"
	HelperBlacklistSubcommand   = "helperblacklist"
	HelperUnblacklistSubcommand = "helperunblacklist"

	// SettingSubcommand Subcommands
	GiveawayChannelSubcommand              = "giveawaychannel"
	ThxInfoChannelSubcommand               = "thxinfochannel"
	AdminRoleSubcommand                    = "adminrole"
	HelperRoleSubcommand                   = "helperrole"
	HelperThxAmountSubcommand              = "helperthxamount"
	WinnerCountSubcommand                  = "winnercount"
	UnconditionalGiveawayChannelSubcommand = "unconditionalgiveawaychannel"
	UnconditionalWinnerCountSubcommand     = "unconditionalwinnercount"
	ConditionalGiveawayChannelSubcommand   = "conditionalgiveawaychannel"
	ConditionalWinnerCountSubcommand       = "conditionalwinnercount"
	ConditionalGiveawayLevelsSubcommand    = "conditionalgiveawaylevels"
)

func NewCsrvbotCommand(craftserveUrl, giveawayHours string, serverRepo *repos.ServerRepo, giveawayRepo *repos.GiveawayRepo, userRepo *repos.UserRepo, csrvClient *services.CsrvClient, giveawayService *services.GiveawayService, helperService *services.HelperService) CsrvbotCommand {
	return CsrvbotCommand{
		Name:                     "csrvbot",
		Description:              "Komendy konfiguracyjne i administracyjne",
		DMPermission:             false,
		DefaultMemberPermissions: discordgo.PermissionManageMessages,
		Zero:                     0.0,
		GiveawayHours:            giveawayHours,
		CraftserveUrl:            craftserveUrl,
		ServerRepo:               *serverRepo,
		GiveawayRepo:             *giveawayRepo,
		UserRepo:                 *userRepo,
		CsrvClient:               *csrvClient,
		GiveawayService:          *giveawayService,
		HelperService:            *helperService,
	}
}

func (h CsrvbotCommand) Register(ctx context.Context, s *discordgo.Session) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(h.Name)
	log.Debug("Registering command")
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:                     h.Name,
		Description:              h.Description,
		DMPermission:             &h.DMPermission,
		DefaultMemberPermissions: &h.DefaultMemberPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        SettingSubcommand,
				Description: "Konfiguracja giveawayów",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        GiveawayChannelSubcommand,
						Description: "Kanał na którym jest prezentowany zwycięzca giveawaya",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type: discordgo.ApplicationCommandOptionChannel,
								ChannelTypes: []discordgo.ChannelType{
									discordgo.ChannelTypeGuildText,
								},
								Name:        "channel",
								Description: "Kanał na którym jest prezentowany zwycięzca giveawaya",
								Required:    true,
							},
						},
					},
					{
						Name:        ThxInfoChannelSubcommand,
						Description: "Kanał na którym są wysyłane wszystkie thxy do rozpatrzenia",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type: discordgo.ApplicationCommandOptionChannel,
								ChannelTypes: []discordgo.ChannelType{
									discordgo.ChannelTypeGuildText,
								},
								Name:        "channel",
								Description: "Kanał na którym są wysyłane wszystkie thxy do rozpatrzenia",
								Required:    true,
							},
						},
					},
					{
						Name:        AdminRoleSubcommand,
						Description: "Rola, która ma dostęp do akceptowania thx i komend administracyjnych",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionRole,
								Name:        "role",
								Description: "Rola, która ma dostęp do akceptowania thx i komend administracyjnych",
								Required:    true,
							},
						},
					},
					{
						Name:        HelperRoleSubcommand,
						Description: "Rola którą dostanie użytkownik, gdy osiągnie daną ilość thx",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionRole,
								Name:        "role",
								Description: "Rola, która ma dostęp do akceptowania thx i komend administracyjnych",
								Required:    true,
							},
						},
					},
					{
						Name:        HelperThxAmountSubcommand,
						Description: "Ilość wymaganych thx do uzyskania roli helpera",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "amount",
								Description: "Ilość wymaganych thx do uzyskania roli helpera",
								Required:    true,
								MinValue:    &h.Zero,
							},
						},
					},
					{
						Name:        WinnerCountSubcommand,
						Description: "Ilość wybieranych zwycięzców w giveawayu z wiadomości",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "amount",
								Description: "Ilość wybieranych zwycięzców w giveawayu z wiadomości",
								Required:    true,
								MinValue:    &h.Zero,
							},
						},
					},
					{
						Name:        UnconditionalGiveawayChannelSubcommand,
						Description: "Kanał na którym odbywa się bezwarunkowy giveaway",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type: discordgo.ApplicationCommandOptionChannel,
								ChannelTypes: []discordgo.ChannelType{
									discordgo.ChannelTypeGuildText,
								},
								Name:        "channel",
								Description: "Kanał na którym odbywa się bezwarunkowy giveaway",
								Required:    true,
							},
						},
					},
					{
						Name:        UnconditionalWinnerCountSubcommand,
						Description: "Ilość wybieranych zwycięzców w bezwarunkowym giveawayu",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "amount",
								Description: "Ilość wybieranych zwycięzców w bezwarunkowym giveawayu",
								Required:    true,
								MinValue:    &h.Zero,
							},
						},
					},
					{
						Name:        ConditionalGiveawayChannelSubcommand,
						Description: "Kanał na którym odbywa się warunkowy giveaway",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type: discordgo.ApplicationCommandOptionChannel,
								ChannelTypes: []discordgo.ChannelType{
									discordgo.ChannelTypeGuildText,
								},
								Name:        "channel",
								Description: "Kanał na którym odbywa się warunkowy giveaway",
								Required:    true,
							},
						},
					},
					{
						Name:        ConditionalWinnerCountSubcommand,
						Description: "Ilość wybieranych zwycięzców w warunkowym giveawayu",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "amount",
								Description: "Ilość wybieranych zwycięzców w warunkowym giveawayu",
								Required:    true,
								MinValue:    &h.Zero,
							},
						},
					},
					{
						Name:        ConditionalGiveawayLevelsSubcommand,
						Description: "Progi poziomów dla warunkowego giveawayu",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "levels",
								Description: "Progi poziomów dla warunkowego giveawayu (np. 5, 10, 15)",
								Required:    true,
							},
						},
					},
				},
				Type: discordgo.ApplicationCommandOptionSubCommandGroup,
			},
			{
				Name:        DeleteSubcommand,
				Description: "Usuwa użytkownika z obecnego giveawaya",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Użytkownik, który ma zostać usunięty",
						Required:    true,
					},
				},
			},
			{
				Name:        StartSubcommand,
				Description: "Rozstrzyga obecny giveaway",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "type",
						Description: "Typ giveawaya",
						Required:    true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{
								Name:  "thx-giveaway",
								Value: "thx",
							},
							{
								Name:  "message-giveaway",
								Value: "message",
							},
							{
								Name:  "unconditional-giveaway",
								Value: "unconditional",
							},
							{
								Name:  "conditional-giveaway",
								Value: "conditional",
							},
						},
					},
				},
			},
			{
				Name:        BlacklistSubcommand,
				Description: "Dodaje użytkownika do blacklisty możliwości udziału w giveawayu",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Użytkownik, który ma zostać dodany",
						Required:    true,
					},
				},
			},
			{
				Name:        UnblacklistSubcommand,
				Description: "Usuwa użytkownika z blacklisty możliwości udziału w giveawayu",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Użytkownik, który ma zostać usunięty",
						Required:    true,
					},
				},
			},
			{
				Name:        HelperBlacklistSubcommand,
				Description: "Dodaje użytkownika do blacklisty możliwości posiadania rangi helpera",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Użytkownik, który ma zostać dodany",
						Required:    true,
					},
				},
			},
			{
				Name:        HelperUnblacklistSubcommand,
				Description: "Usuwa użytkownika z blacklisty możliwości posiadania rangi helpera",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Użytkownik, który ma zostać usunięty",
						Required:    true,
					},
				},
			},
		},
	})
	if err != nil {
		log.WithError(err).Error("Could not register command")
	}
}

func (h CsrvbotCommand) Handle(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	member := i.Member
	adminRole, err := h.ServerRepo.GetAdminRoleForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("Could not get admin role")
	}

	if !discord.HasAdminPermissions(ctx, s, member, adminRole, i.GuildID) {
		log.Debug("User is not an admin")
		discord.RespondWithMessage(ctx, s, i, "Nie masz uprawnień do tej komendy")
		return
	}

	if len(i.ApplicationCommandData().Options) == 0 {
		log.WithError(err).Error("No options provided")
		return
	}
	log = log.WithSubcommand(i.ApplicationCommandData().Options[0].Name)
	ctx = logger.ContextWithLogger(ctx, log)

	switch i.ApplicationCommandData().Options[0].Name {
	case SettingSubcommand:
		h.handleSettings(ctx, s, i)
	case DeleteSubcommand:
		h.handleDelete(ctx, s, i)
	case StartSubcommand:
		h.handleStart(ctx, s, i)
	case BlacklistSubcommand:
		h.handleBlacklist(ctx, s, i)
	case UnblacklistSubcommand:
		h.handleUnblacklist(ctx, s, i)
	case HelperBlacklistSubcommand:
		h.handleHelperBlacklist(ctx, s, i)
	case HelperUnblacklistSubcommand:
		h.handleHelperUnblacklist(ctx, s, i)
	}
}

func (h CsrvbotCommand) handleSettings(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Options[0].Options[0].Name {
	case GiveawayChannelSubcommand:
		h.handleGiveawayChannelSet(ctx, s, i)
	case ThxInfoChannelSubcommand:
		h.handleThxInfoChannelSet(ctx, s, i)
	case AdminRoleSubcommand:
		h.handleAdminRoleSet(ctx, s, i)
	case HelperRoleSubcommand:
		h.handleHelperRoleSet(ctx, s, i)
	case HelperThxAmountSubcommand:
		h.handleHelperThxAmountSet(ctx, s, i)
	case WinnerCountSubcommand:
		h.handleWinnerCountSet(ctx, s, i)
	case UnconditionalGiveawayChannelSubcommand:
		h.handleUnconditionalGiveawayChannelSet(ctx, s, i)
	case UnconditionalWinnerCountSubcommand:
		h.handleUnconditionalWinnersCountSet(ctx, s, i)
	case ConditionalGiveawayChannelSubcommand:
		h.handleConditionalGiveawayChannelSet(ctx, s, i)
	case ConditionalWinnerCountSubcommand:
		h.handleConditionalWinnersCountSet(ctx, s, i)
	case ConditionalGiveawayLevelsSubcommand:
		h.handleConditionalGiveawayLevelsSet(ctx, s, i)
	}
}

func (h CsrvbotCommand) handleStart(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	giveawayType := i.ApplicationCommandData().Options[0].Options[0].StringValue()

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleStart s.Guild")
		return
	}

	switch giveawayType {
	case "thx":
		log.Debug("Starting thx giveaway")
		go func() {
			h.GiveawayService.FinishGiveaway(ctx, s, guild.ID)
		}()
	case "message":
		serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
		if err != nil {
			log.WithError(err).Error("handleStart h.ServerRepo.GetServerConfigForGuild")
			return
		}
		if serverConfig.MessageGiveawayWinners == 0 {
			log.Debug("No winners set for message giveaway")
			discord.RespondWithMessage(ctx, s, i, "Nie ustawiono liczby zwycięzców")
			return
		}
		log.Debug("Starting message giveaway")
		go h.GiveawayService.FinishMessageGiveaway(ctx, s, guild.ID)
	case "unconditional":
		log.Debug("Starting unconditional giveaway")
		//go h.GiveawayService.FinishUnconditionalGiveaway(ctx, s, guild.ID)
		go h.GiveawayService.FinishJoinableGiveaway(ctx, s, guild.ID, false)
	case "conditional":
		log.Debug("Starting conditional giveaway")
		//go h.GiveawayService.FinishConditionalGiveaway(ctx, s, guild.ID)
		go h.GiveawayService.FinishJoinableGiveaway(ctx, s, guild.ID, true)
	}
	discord.RespondWithMessage(ctx, s, i, "Podjęto próbę rozstrzygnięcia giveawayu")
}

func (h CsrvbotCommand) handleDelete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleDelete h.GiveawayRepo.GetGiveawayForGuild")
		return
	}
	participants, err := h.GiveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.WithError(err).Error("handleDelete h.GiveawayRepo.GetParticipantsForGiveaway")
		return
	}
	selectedUser := i.ApplicationCommandData().Options[0].Options[0].UserValue(s)
	discord.RespondWithMessage(ctx, s, i, "Podjęto próbę usunięcia użytkownika z giveawayu")
	log.Debug("Removing all entries from database for participant ", selectedUser.ID)
	err = h.GiveawayRepo.RemoveAllParticipantEntries(ctx, giveaway.Id, selectedUser.ID)
	if err != nil {
		log.WithError(err).Error("handleDelete h.GiveawayRepo.RemoveAllParticipantEntries")
		return
	}
	log.Infof("%s removed all entries for user %s", i.Member.User.Username, selectedUser.Username)
	participantNames, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.WithError(err).Error("handleDelete h.GiveawayRepo.GetParticipantNamesForGiveaway")
		return
	}
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleDelete h.ServerRepo.GetServerConfigForGuild")
		return
	}
	for _, participant := range participants {
		if participant.UserId != selectedUser.ID {
			return
		}
		log.WithMessage(participant.MessageId).Debug("Updating thx embed after entry deletion for participant ", participant.UserId)
		embed := discord.ConstructThxEmbed(h.CraftserveUrl, participantNames, h.GiveawayHours, participant.UserId, "", "reject")
		_, err = s.ChannelMessageEditEmbed(participant.ChannelId, participant.MessageId, embed)
		if err != nil {
			log.WithError(err).Error("handleDelete s.ChannelMessageEditEmbed")
			return
		}
		thxNotification, err := h.GiveawayRepo.GetThxNotification(ctx, participant.MessageId)
		if err != nil {
			log.WithError(err).Error("handleDelete h.GiveawayRepo.GetThxNotification")
			return
		}
		log.Debug("Updating thx notification message after entry deletion for participant ", participant.UserId)
		_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, participant.MessageId, participant.UserId, "", "reject", h.CraftserveUrl)
		if err != nil {
			log.WithError(err).Error("handleDelete discord.NotifyThxOnThxInfoChannel")
			return
		}

	}
}

func (h CsrvbotCommand) handleBlacklist(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	selectedUser := i.ApplicationCommandData().Options[0].Options[0].UserValue(s)
	if selectedUser.Bot {
		log.Debug("User is a bot")
		discord.RespondWithMessage(ctx, s, i, "Nie możesz dodać bota do blacklisty")
		return
	}
	isUserBlacklisted, err := h.UserRepo.IsUserBlacklisted(ctx, selectedUser.ID, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleBlacklist h.UserRepo.IsUserBlacklisted")
		return
	}
	if isUserBlacklisted {
		log.Debug("User is already blacklisted")
		discord.RespondWithMessage(ctx, s, i, "Użytkownik jest już na blackliście")
		return
	}

	log.Debug("Adding user to blacklist")
	err = h.UserRepo.AddBlacklistForUser(ctx, selectedUser.ID, i.GuildID, i.Member.User.ID)
	if err != nil {
		log.WithError(err).Error("handleBlacklist h.UserRepo.AddBlacklistForUser")
		discord.RespondWithMessage(ctx, s, i, "Nie udało się dodać użytkownika do blacklisty")
		return
	}
	log.Infof("%s blacklisted %s", i.Member.User.Username, selectedUser.Username)
	discord.RespondWithMessage(ctx, s, i, "Dodano użytkownika do blacklisty")
}

func (h CsrvbotCommand) handleUnblacklist(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	selectedUser := i.ApplicationCommandData().Options[0].Options[0].UserValue(s)

	isUserBlacklisted, err := h.UserRepo.IsUserBlacklisted(ctx, selectedUser.ID, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleUnblacklist h.UserRepo.IsUserBlacklisted", err)
		return
	}
	if !isUserBlacklisted {
		log.Debug("User is not blacklisted")
		discord.RespondWithMessage(ctx, s, i, "Użytkownik nie jest na blackliście")
		return
	}

	log.Debug("Removing user from blacklist")
	err = h.UserRepo.RemoveBlacklistForUser(ctx, selectedUser.ID, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleUnblacklist h.UserRepo.RemoveBlacklistForUser", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się usunąć użytkownika z blacklisty")
		return
	}
	log.Infof("%s unblacklisted %s", i.Member.User.Username, selectedUser.Username)
	discord.RespondWithMessage(ctx, s, i, "Usunięto użytkownika z blacklisty")
}

func (h CsrvbotCommand) handleHelperBlacklist(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	selectedUser := i.ApplicationCommandData().Options[0].Options[0].UserValue(s)
	if selectedUser.Bot {
		log.Debug("User is a bot")
		discord.RespondWithMessage(ctx, s, i, "Nie możesz dodać bota do helper-blacklisty")
		return
	}
	isUserHelperBlacklisted, err := h.UserRepo.IsUserHelperBlacklisted(ctx, selectedUser.ID, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleHelperBlacklist h.UserRepo.IsUserHelperBlacklisted", err)
		return
	}
	if isUserHelperBlacklisted {
		log.Debug("User is already helper-blacklisted")
		discord.RespondWithMessage(ctx, s, i, "Użytkownik jest już na helper-blackliście")
		return
	}

	log.Debug("Adding user to helper-blacklist")
	err = h.UserRepo.AddHelperBlacklistForUser(ctx, selectedUser.ID, i.GuildID, i.Member.User.ID)
	if err != nil {
		log.WithError(err).Error("handleHelperBlacklist h.UserRepo.AddHelperBlacklistForUser", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się dodać użytkownika do helper-blacklisty")
		return
	}
	log.Infof("%s helper-blacklisted %s", i.Member.User.Username, selectedUser.Username)
	discord.RespondWithMessage(ctx, s, i, "Użytkownik został zablokowany z możliwości zostania pomocnym")
	log.Debug("Checking if user should be removed from helpers")
	h.HelperService.CheckHelper(ctx, s, i.GuildID, selectedUser.ID)
}

func (h CsrvbotCommand) handleHelperUnblacklist(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	selectedUser := i.ApplicationCommandData().Options[0].Options[0].UserValue(s)

	isUserHelperBlacklisted, err := h.UserRepo.IsUserHelperBlacklisted(ctx, selectedUser.ID, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleHelperUnblacklist h.UserRepo.IsUserHelperBlacklisted", err)
		return
	}
	if !isUserHelperBlacklisted {
		log.Debug("User is not helper-blacklisted")
		discord.RespondWithMessage(ctx, s, i, "Użytkownik nie jest na helper-blackliście")
		return
	}

	log.Debug("Removing user from helper-blacklist")
	err = h.UserRepo.RemoveHelperBlacklistForUser(ctx, selectedUser.ID, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleHelperUnblacklist h.UserRepo.RemoveHelperBlacklistForUser", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się usunąć użytkownika z helper-blacklisty")
		return
	}
	log.Infof("%s helper-unblacklisted %s", i.Member.User.Username, selectedUser.Username)
	discord.RespondWithMessage(ctx, s, i, "Użytkownik został usunięty z helper-blacklisty")
	log.Debug("Checking if user should be re-added to helpers")
	h.HelperService.CheckHelper(ctx, s, i.GuildID, selectedUser.ID)
}

func (h CsrvbotCommand) handleGiveawayChannelSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	channelId := i.ApplicationCommandData().Options[0].Options[0].Options[0].ChannelValue(s).ID
	channel, err := s.Channel(channelId)
	if err != nil {
		log.WithError(err).Error("handleGiveawayChannelSet s.Channel", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleGiveawayChannelSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}
	if serverConfig.MainChannel == channelId {
		log.Debug("Giveaway channel is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Kanał do ogłoszeń giveawaya jest już ustawiony na "+channel.Mention())
		return
	}
	serverConfig.MainChannel = channelId
	log.Debug("Updating server config with new giveaway channel")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleGiveawayChannelSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}
	log.Infof("%s set giveaway channel to %s (%s)", i.Member.User.Username, channel.Name, channel.ID)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono kanał do ogłoszeń giveawaya na "+channel.Mention())
}

func (h CsrvbotCommand) handleThxInfoChannelSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	channelId := i.ApplicationCommandData().Options[0].Options[0].Options[0].ChannelValue(s).ID
	channel, err := s.Channel(channelId)
	if err != nil {
		log.WithError(err).Error("handleThxInfoChannelSet s.Channel", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleThxInfoChannelSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}
	if serverConfig.ThxInfoChannel == channelId {
		log.Debug("Thx info channel is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Kanał do powiadomień o thx jest już ustawiony na "+channel.Mention())
		return
	}
	serverConfig.ThxInfoChannel = channelId
	log.Debug("Updating server config with new thx info channel")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleThxInfoChannelSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}
	log.Infof("%s set thx info channel to %s (%s)", i.Member.User.Username, channel.Name, channel.ID)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono kanał do powiadomień o thx na "+channel.Mention())
}

func (h CsrvbotCommand) handleAdminRoleSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	roleId := i.ApplicationCommandData().Options[0].Options[0].Options[0].RoleValue(s, i.GuildID).ID
	role, err := s.State.Role(i.GuildID, roleId)
	if err != nil {
		log.WithError(err).Error("handleAdminRoleSet s.State.Role", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić roli")
		return
	}
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleAdminRoleSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić roli")
		return
	}
	if serverConfig.AdminRoleId == roleId {
		log.Debug("Admin role is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Rola admina jest już ustawiona na "+role.Name)
		return
	}
	serverConfig.AdminRoleId = roleId
	log.Debug("Updating server config with new admin role")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleAdminRoleSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić roli")
		return
	}
	log.Infof("%s set admin role to %s (%s)", i.Member.User.Username, role.Name, role.ID)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono rolę admina na "+role.Name)
}

func (h CsrvbotCommand) handleHelperRoleSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	roleId := i.ApplicationCommandData().Options[0].Options[0].Options[0].RoleValue(s, i.GuildID).ID
	role, err := s.State.Role(i.GuildID, roleId)
	if err != nil {
		log.WithError(err).Error("handleHelperRoleSet s.State.Role", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić roli")
		return
	}
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleHelperRoleSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić roli")
		return
	}
	if serverConfig.HelperRoleId == roleId {
		log.Debug("Helper role is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Rola helpera jest już ustawiona na "+role.Name)
		return
	}
	serverConfig.HelperRoleId = roleId
	log.Debug("Updating server config with new helper role")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleHelperRoleSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić roli")
		return
	}
	log.Infof("%s set helper role to %s (%s)", i.Member.User.Username, role.Name, role.ID)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono rolę helpera na "+role.Name)
}

func (h CsrvbotCommand) handleHelperThxAmountSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	amount := i.ApplicationCommandData().Options[0].Options[0].Options[0].UintValue()
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleHelperThxAmountSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić ilości thx")
		return
	}
	if serverConfig.HelperRoleThxesNeeded == int(amount) {
		log.Debug("Thx amount is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Ilość thx potrzebna do uzyskania rangi helpera jest już ustawiona na "+strconv.FormatUint(amount, 10))
		return
	}
	serverConfig.HelperRoleThxesNeeded = int(amount)
	log.Debug("Updating server config with new helper thx amount")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleHelperThxAmountSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić ilości thx")
		return
	}
	log.Infof("%s set helper thx amount to %d", i.Member.User.Username, amount)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono ilość thx potrzebną do uzyskania rangi helpera na "+strconv.FormatUint(amount, 10))
	log.Debug("Checking helpers after helper thx amount set")
	h.HelperService.CheckHelpers(ctx, s, i.GuildID)
}

func (h CsrvbotCommand) handleWinnerCountSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	amount := i.ApplicationCommandData().Options[0].Options[0].Options[0].UintValue()
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleWinnerCountSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić ilości wybieranych osób")
		return
	}
	if amount > 10 {
		log.Debug("Winner count is too high")
		discord.RespondWithMessage(ctx, s, i, "Nie można ustawić ilości wybieranych osób na więcej niż 10")
		return
	}
	if serverConfig.MessageGiveawayWinners == int(amount) {
		log.Debug("Winner count is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Ilość wybieranych osób jest już ustawiona na "+strconv.FormatUint(amount, 10))
		return
	}
	serverConfig.MessageGiveawayWinners = int(amount)
	log.Debug("Updating server config with new winner count")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleWinnerCountSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić ilości wybieranych osób")
		return
	}
	log.Infof("%s set winnercount to %d", i.Member.User.Username, amount)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono ilość wybieranych osób w giveawayu z wiadomości na "+strconv.FormatUint(amount, 10))
}

func (h CsrvbotCommand) handleUnconditionalGiveawayChannelSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	channelId := i.ApplicationCommandData().Options[0].Options[0].Options[0].ChannelValue(s).ID
	channel, err := s.Channel(channelId)
	if err != nil {
		log.WithError(err).Error("handleUnconditionalGiveawayChannelSet s.Channel", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleUnconditionalGiveawayChannelSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}
	if serverConfig.UnconditionalGiveawayChannel == channelId {
		log.Debug("Unconditional giveaway channel is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Kanał do bezwarunkowych giveawayów jest już ustawiony na "+channel.Mention())
		return
	}

	serverConfig.UnconditionalGiveawayChannel = channelId
	log.Debug("Updating server config with new unconditional giveaway channel")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleUnconditionalGiveawayChannelSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}

	log.Infof("%s set unconditional giveaway channel to %s (%s)", i.Member.User.Username, channel.Name, channel.ID)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono kanał do bezwarunkowych giveawayów na "+channel.Mention())
}

func (h CsrvbotCommand) handleUnconditionalWinnersCountSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	amount := i.ApplicationCommandData().Options[0].Options[0].Options[0].UintValue()
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleUnconditionalWinnersCountSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić ilości wybieranych osób")

		return
	}

	if amount > 10 {
		log.Debug("Winner count is too high")
		discord.RespondWithMessage(ctx, s, i, "Nie można ustawić ilości wybieranych osób na więcej niż 10")

		return
	}

	if serverConfig.UnconditionalGiveawayWinners == int(amount) {
		log.Debug("Winner count is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Ilość wybieranych osób jest już ustawiona na "+strconv.FormatUint(amount, 10))

		return
	}

	serverConfig.UnconditionalGiveawayWinners = int(amount)
	log.Debug("Updating server config with new winner count")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleUnconditionalWinnersCountSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić ilości wybieranych osób")

		return
	}

	log.Infof("%s set unconditional winnercount to %d", i.Member.User.Username, amount)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono ilość wybieranych osób w bezwarunkowym giveawayu na "+strconv.FormatUint(amount, 10))
}

func (h CsrvbotCommand) handleConditionalGiveawayChannelSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	channelId := i.ApplicationCommandData().Options[0].Options[0].Options[0].ChannelValue(s).ID
	channel, err := s.Channel(channelId)
	if err != nil {
		log.WithError(err).Error("handleConditionalGiveawayChannelSet s.Channel", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleConditionalGiveawayChannelSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}

	if serverConfig.ConditionalGiveawayChannel == channelId {
		log.Debug("Conditional giveaway channel is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Kanał do warunkowych giveawayów jest już ustawiony na "+channel.Mention())
		return
	}

	serverConfig.ConditionalGiveawayChannel = channelId
	log.Debug("Updating server config with new conditional giveaway channel")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleConditionalGiveawayChannelSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić kanału")
		return
	}

	log.Infof("%s set conditional giveaway channel to %s (%s)", i.Member.User.Username, channel.Name, channel.ID)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono kanał do warunkowych giveawayów na "+channel.Mention())
}

func (h CsrvbotCommand) handleConditionalWinnersCountSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	amount := i.ApplicationCommandData().Options[0].Options[0].Options[0].UintValue()
	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleConditionalWinnersCountSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić ilości wybieranych osób")
		return
	}

	if amount > 10 {
		log.Debug("Winner count is too high")
		discord.RespondWithMessage(ctx, s, i, "Nie można ustawić ilości wybieranych osób na więcej niż 10")
		return
	}

	if serverConfig.ConditionalGiveawayWinners == int(amount) {
		log.Debug("Winner count is the same as current")
		discord.RespondWithMessage(ctx, s, i, "Ilość wybieranych osób jest już ustawiona na "+strconv.FormatUint(amount, 10))
		return
	}

	serverConfig.ConditionalGiveawayWinners = int(amount)
	log.Debug("Updating server config with new winner count")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleConditionalWinnersCountSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić ilości wybieranych osób")
		return
	}

	log.Infof("%s set conditional winnercount to %d", i.Member.User.Username, amount)
	discord.RespondWithMessage(ctx, s, i, "Ustawiono ilość wybieranych osób w warunkowym giveawayu na "+strconv.FormatUint(amount, 10))
}

func (h CsrvbotCommand) handleConditionalGiveawayLevelsSet(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	levelsStrings := strings.Split(i.ApplicationCommandData().Options[0].Options[0].Options[0].StringValue(), ",")

	levels := make([]int, 0)
	for index, levelString := range levelsStrings {
		levelsStrings[index] = strings.TrimSpace(levelString)

		level, err := strconv.Atoi(strings.TrimSpace(levelString))
		if err != nil {
			log.WithError(err).Error("handleConditionalGiveawayLevelsSet strconv.Atoi", err)
			discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić poziomów, sprawdź czy są one poprawne")
			return
		}

		levels = append(levels, level)
	}

	currentLevels, err := h.ServerRepo.GetConditionalGiveawayLevels(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleConditionalGiveawayLevelsSet h.ServerRepo.GetConditionalGiveawayLevels", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić poziomów")
		return
	}

	if reflect.DeepEqual(levels, currentLevels) {
		log.Debug("Levels are the same as current")
		discord.RespondWithMessage(ctx, s, i, "Poziomy są już ustawione na "+strings.Join(levelsStrings, ", "))
		return
	}

	valid, err := discord.ValidateLevels(ctx, s, i.GuildID, levels)
	if err != nil {
		log.WithError(err).Error("handleConditionalGiveawayLevelsSet discord.ValidateLevels", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić poziomów, sprawdź czy są one poprawne")
		return
	}

	if !valid {
		log.Debug("Levels are not valid")
		discord.RespondWithMessage(ctx, s, i, "Podane poziomy nie są poprawne z powodu braku odpowiadającej roli")
		return
	}

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("handleConditionalGiveawayLevelsSet h.ServerRepo.GetServerConfigForGuild", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić poziomów")
		return
	}

	jsonLevels, err := json.Marshal(levels)
	if err != nil {
		log.WithError(err).Error("handleConditionalGiveawayLevelsSet json.Marshal", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić poziomów")
		return
	}

	serverConfig.ConditionalGiveawayLevels = jsonLevels
	log.Debug("Updating server config with new levels")
	err = h.ServerRepo.UpdateServerConfig(ctx, &serverConfig)
	if err != nil {
		log.WithError(err).Error("handleConditionalGiveawayLevelsSet h.ServerRepo.UpdateServerConfig", err)
		discord.RespondWithMessage(ctx, s, i, "Nie udało się ustawić poziomów")
		return
	}

	log.Infof("%s set conditional giveaway levels to %s", i.Member.User.Username, strings.Join(levelsStrings, ", "))
	discord.RespondWithMessage(ctx, s, i, "Ustawiono poziomy giveawayu warunkowego na "+strings.Join(levelsStrings, ", "))
}
