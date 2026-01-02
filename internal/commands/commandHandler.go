package commands

import (
	"reflect"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/ollamabot/internal/commands/admincommand"
	"github.com/stollenaar/ollamabot/internal/commands/listcommand"
	"github.com/stollenaar/ollamabot/internal/commands/promptcommand"
	"github.com/stollenaar/ollamabot/internal/commands/threadcommand"
	"github.com/stollenaar/ollamabot/internal/util"
)

type CommandI interface {
	Handler(e *events.ApplicationCommandInteractionCreate)
	CreateCommandArguments() []discord.ApplicationCommandOption
}

var (
	Commands = []CommandI{
		admincommand.AdminCmd,
		listcommand.ListCmd,
		promptcommand.PromptCmd,
		threadcommand.ThreadCmd,
	}
	ApplicationCommands []discord.ApplicationCommandCreate
	CommandHandlers     = make(map[string]func(e *events.ApplicationCommandInteractionCreate))
	ModalSubmitHandlers = make(map[string]func(e *events.ModalSubmitInteractionCreate))
	ComponentHandlers   = make(map[string]func(e *events.ComponentInteractionCreate))
)

func init() {
	for _, cmd := range Commands {
		ApplicationCommands = append(ApplicationCommands, discord.SlashCommandCreate{
			Name:        reflect.ValueOf(cmd).FieldByName("Name").String(),
			Description: reflect.ValueOf(cmd).FieldByName("Description").String(),
			Options:     cmd.CreateCommandArguments(),
		})
		CommandHandlers[reflect.ValueOf(cmd).FieldByName("Name").String()] = cmd.Handler

		if _, ok := reflect.TypeOf(cmd).MethodByName("ModalHandler"); ok {
			ModalSubmitHandlers[reflect.ValueOf(cmd).FieldByName("Name").String()] = func(e *events.ModalSubmitInteractionCreate) {
				reflect.ValueOf(cmd).MethodByName("ModalHandler").Call([]reflect.Value{
					reflect.ValueOf(e),
				})
			}
		}
		if _, ok := reflect.TypeOf(cmd).MethodByName("ComponentHandler"); ok {
			ComponentHandlers[reflect.ValueOf(cmd).FieldByName("Name").String()] = func(e *events.ComponentInteractionCreate) {
				reflect.ValueOf(cmd).MethodByName("ComponentHandler").Call([]reflect.Value{
					reflect.ValueOf(e),
				})
			}
		}
	}

	ApplicationCommands = append(ApplicationCommands,
		discord.SlashCommandCreate{
			Name:        "ping",
			Description: "pong",
		},
	)

	CommandHandlers["ping"] = PingCommand
}

// PingCommand sends back the pong
func PingCommand(event *events.ApplicationCommandInteractionCreate) {
	event.CreateMessage(discord.MessageCreate{
		Content: "Pong",
		Flags:   util.ConfigFile.SetEphemeral(),
	})
}
