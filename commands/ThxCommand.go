package commands

import (
	"csrvbot/internal/repos"
	"csrvbot/pkg"
	"csrvbot/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"log"
)

type ThxCommand struct {
	Name          string
	Description   string
	DMPermission  bool
	GiveawayHours string
	GiveawayRepo  repos.GiveawayRepo
	UserRepo      repos.UserRepo
	ServerRepo    repos.ServerRepo
}

func NewThxCommand(giveawayRepo *repos.GiveawayRepo, userRepo *repos.UserRepo, serverRepo *repos.ServerRepo, giveawayHours string) ThxCommand {
	return ThxCommand{
		Name:          "thx",
		Description:   "Podziękowanie innemu użytkownikowi",
		DMPermission:  false,
		GiveawayRepo:  *giveawayRepo,
		UserRepo:      *userRepo,
		ServerRepo:    *serverRepo,
		GiveawayHours: giveawayHours,
	}
}

func (h ThxCommand) Register(s *discordgo.Session) {
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:         h.Name,
		Description:  h.Description,
		DMPermission: &h.DMPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Użytkownik, któremu chcesz podziękować",
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

func (h ThxCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := pkg.CreateContext()
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.Println("("+i.GuildID+") handleThxCommand#session.Guild", err)
		return
	}
	var selectedUser *discordgo.User
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		selectedUser = data.Resolved.Messages[data.TargetID].Author
	} else {
		selectedUser = data.Options[0].UserValue(s)
	}
	author := i.Member.User
	if author.ID == selectedUser.ID {
		discord.RespondWithMessage(s, i, "Nie można dziękować sobie!")
		return
	}
	if selectedUser.Bot {
		discord.RespondWithMessage(s, i, "Nie można dziękować botom!")
		return
	}
	isUserBlacklisted, err := h.UserRepo.IsUserBlacklisted(ctx, i.GuildID, selectedUser.ID)
	if err != nil {
		log.Println("("+i.GuildID+") handleThxCommand#UserRepo.IsUserBlacklisted", err)
		return
	}
	if isUserBlacklisted {
		discord.RespondWithMessage(s, i, "Ten użytkownik jest na czarnej liście i nie może brać udziału :(")
		return
	}
	giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, i.GuildID)
	if err != nil || giveaway == nil {
		log.Println("("+i.GuildID+") Could not get giveaway", err)
		return
	}
	participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, giveaway.Id)
	if err != nil {
		log.Println("("+i.GuildID+") Could not get participants", err)
		return
	}

	embed := discord.ConstructThxEmbed(participants, h.GiveawayHours, selectedUser.ID, "", "wait")

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Println("("+i.GuildID+") Could not respond to interaction ("+i.ID+")", err)
		return
	}

	response, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.Println("("+i.GuildID+") Could not fetch a response of interaction ("+i.ID+")", err)
		return
	}

	err = h.GiveawayRepo.InsertParticipant(ctx, giveaway.Id, i.GuildID, guild.Name, selectedUser.ID, selectedUser.Username, i.ChannelID, response.ID)
	if err != nil {
		log.Println("("+i.GuildID+") Could not insert participant", err)
		str := "Coś poszło nie tak przy dodawaniu podziękowania :("
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &str,
		})
		return
	}
	log.Println("(" + i.GuildID + ") " + author.Username + " has thanked " + selectedUser.Username)
	discord.NotifyThxOnThxInfoChannel(ctx, s, h.ServerRepo, h.GiveawayRepo, i.GuildID, i.ChannelID, response.ID, selectedUser.ID, "", "wait")

	for err = s.MessageReactionAdd(i.ChannelID, response.ID, "✅"); err != nil; err = s.MessageReactionAdd(i.ChannelID, response.ID, "✅") {
	}
	for err = s.MessageReactionAdd(i.ChannelID, response.ID, "⛔"); err != nil; err = s.MessageReactionAdd(i.ChannelID, response.ID, "⛔") {
	}
}

func (h ThxCommand) HandleContext(s *discordgo.Session, i *discordgo.InteractionCreate) {

}
