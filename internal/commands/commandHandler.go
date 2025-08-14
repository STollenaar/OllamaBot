package commands

import (
	"reflect"

	"github.com/bwmarrin/discordgo"
	"github.com/stollenaar/ollamabot/internal/commands/admincommand"
	"github.com/stollenaar/ollamabot/internal/commands/listcommand"
	"github.com/stollenaar/ollamabot/internal/util"
)

type CommandI interface {
	Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate)
	CreateCommandArguments() []*discordgo.ApplicationCommandOption
	ParseArguments(bot *discordgo.Session, interaction *discordgo.InteractionCreate) interface{}
}

var (
	Commands = []CommandI{
		admincommand.AdminCmd,
		listcommand.ListCmd,
	}
	ApplicationCommands []*discordgo.ApplicationCommand
	CommandHandlers     = make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))
	ModalSubmitHandlers = make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))
)

func init() {
	for _, cmd := range Commands {
		ApplicationCommands = append(ApplicationCommands, &discordgo.ApplicationCommand{
			Name:        reflect.ValueOf(cmd).FieldByName("Name").String(),
			Description: reflect.ValueOf(cmd).FieldByName("Description").String(),
			Options:     cmd.CreateCommandArguments(),
		})
		CommandHandlers[reflect.ValueOf(cmd).FieldByName("Name").String()] = cmd.Handler

		if _, ok := reflect.TypeOf(cmd).MethodByName("ModalHandler"); ok {
			ModalSubmitHandlers[reflect.ValueOf(cmd).FieldByName("Name").String()] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
				reflect.ValueOf(cmd).MethodByName("ModalHandler").Call([]reflect.Value{
					reflect.ValueOf(s),
					reflect.ValueOf(i),
				})
			}
		}
	}

	ApplicationCommands = append(ApplicationCommands,
		&discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "pong",
		},
	)

	CommandHandlers["ping"] = PingCommand
}

// PingCommand sends back the pong
func PingCommand(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong",
			Flags:   util.ConfigFile.SetEphemeral(),
		},
	})
}
