package commands

import (
	"csrvbot/internal/repos"
	"csrvbot/pkg"
	"csrvbot/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"log"
)

type ThxmeCommand struct {
	Name          string
	Description   string
	DMPermission  bool
	GiveawayHours string
	GiveawayRepo  repos.GiveawayRepo
	UserRepo      repos.UserRepo
	ServerRepo    repos.ServerRepo
}

func NewThxmeCommand(giveawayRepo *repos.GiveawayRepo, userRepo *repos.UserRepo, serverRepo *repos.ServerRepo, giveawayHours string) ThxmeCommand {
	return ThxmeCommand{
		Name:          "thxme",
		Description:   "Poproszenie użytkownika o podziękowanie",
		DMPermission:  false,
		GiveawayRepo:  *giveawayRepo,
		UserRepo:      *userRepo,
		ServerRepo:    *serverRepo,
		GiveawayHours: giveawayHours,
	}
}

func (h ThxmeCommand) Register(s *discordgo.Session) {
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		Description:  h.Description,
		DMPermission: &h.DMPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Użytkownik, którego chcesz poprosić o podziękowanie",
				Required:    true,
			},
		},
	})
	if err != nil {
		log.Println("Could not register command", err)
	}

	_, err = s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		DMPermission: &h.DMPermission,
		Type:         discordgo.MessageApplicationCommand,
	})
	if err != nil {
		log.Println("Could not register context command", err)
	}

	_, err = s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		DMPermission: &h.DMPermission,
		Type:         discordgo.UserApplicationCommand,
	})
	if err != nil {
		log.Println("Could not register context command", err)
	}
}

func (h ThxmeCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := pkg.CreateContext()
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.Println("("+i.GuildID+") handleThxmeCommand#session.Guild", err)
		return
	}
	var selectedUser *discordgo.User
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		if len(data.Resolved.Messages) != 0 {
			selectedUser = data.Resolved.Messages[data.TargetID].Author
		} else if len(data.Resolved.Users) != 0 {
			selectedUser = data.Resolved.Users[data.TargetID]
		} else {
			log.Printf("(%s) handleThxCommand# could not get selectedUser", i.GuildID)
			return
		}
	} else {
		selectedUser = data.Options[0].UserValue(s)
	}
	author := i.Member.User
	if author.ID == selectedUser.ID {
		discord.RespondWithMessage(s, i, "Nie można poprosić o podziękowanie samego siebie!")
		return
	}
	if selectedUser.Bot {
		discord.RespondWithMessage(s, i, "Nie można prosić o podziękowanie bota!")
		return
	}
	isUserBlacklisted, err := h.UserRepo.IsUserBlacklisted(ctx, i.GuildID, selectedUser.ID)
	if err != nil {
		log.Println("("+i.GuildID+") handleThxmeCommand#UserRepo.IsUserBlacklisted", err)
		return
	}
	if isUserBlacklisted {
		discord.RespondWithMessage(s, i, "Nie możesz poprosić o podziękowanie, gdyż jesteś na czarnej liście!")
		return
	}

	discord.RespondWithMessage(s, i, selectedUser.Mention()+", czy chcesz podziękować użytkownikowi "+author.Mention()+"?")

	response, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.Println("("+i.GuildID+") Could not fetch a response of interaction ("+i.ID+")", err)
		return
	}

	err = h.GiveawayRepo.InsertParticipantCandidate(ctx, i.GuildID, guild.Name, author.ID, author.Username, selectedUser.ID, selectedUser.Username, i.ChannelID, response.ID)
	if err != nil {
		log.Println("("+i.GuildID+") Could not insert participant candidate", err)
		str := "Coś poszło nie tak przy dodawaniu kandydata do podziękowania :("
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &str,
		})
		return
	}
	log.Println("(" + i.GuildID + ") " + author.Username + " has requested thx from " + selectedUser.Username)

	for err = s.MessageReactionAdd(i.ChannelID, response.ID, "✅"); err != nil; err = s.MessageReactionAdd(i.ChannelID, response.ID, "✅") {
	}
	for err = s.MessageReactionAdd(i.ChannelID, response.ID, "⛔"); err != nil; err = s.MessageReactionAdd(i.ChannelID, response.ID, "⛔") {
	}
}
