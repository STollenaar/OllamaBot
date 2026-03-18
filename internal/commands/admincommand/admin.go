package admincommand

import (
	"log"
	"log/slog"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	ollamaApi "github.com/ollama/ollama/api"
	"github.com/stollenaar/ollamabot/internal/util"
)

var (
	AdminCmd = AdminCommand{
		Name:        "admin",
		Description: "Admin command to manage to ollamabot",
	}
	OllamaClient *ollamaApi.Client
)

type AdminCommand struct {
	Name        string
	Description string
}

func init() {
	client, err := ollamaApi.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	OllamaClient = client
}

func (a AdminCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	if event.Member().User.ID.String() != util.ConfigFile.ADMIN_USER_ID {
		event.CreateMessage(discord.MessageCreate{
			Content: "You are not the boss of me",
			Flags:   discord.MessageFlagEphemeral | discord.MessageFlagIsComponentsV2,
		})
		return
	}
	err := event.DeferCreateMessage(true)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	sub := event.SlashCommandInteractionData()

	var components []discord.LayoutComponent
	switch *sub.SubCommandGroupName {
	case "platform":
		components = platformHandler(sub, event)
	case "model":
		components = modelHandler(sub, event)
	case "ollama":
		components = ollamaHandler(sub, event)
	case "platform_model":
		components = platformModelHandler(sub, event)
	case "prompt":
		components = promptHandler(sub, event)
	}
	util.UpdateInteractionResponse(event, components)
}

func (a AdminCommand) ComponentHandler(event *events.ComponentInteractionCreate) {
	if event.Member().User.ID.String() != util.ConfigFile.ADMIN_USER_ID {
		return
	}

	err := event.DeferUpdateMessage()

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	var components []discord.LayoutComponent
	switch strings.Split(event.Data.CustomID(), "_")[1] {
	case "prompt":
		components = promptButtonHandler(event)
	default:
		components = append(components, discord.ContainerComponent{
			Components: []discord.ContainerSubComponent{
				discord.TextDisplayComponent{
					Content: "Unknown button interaction",
				},
			},
		})
	}

	util.UpdateComponentInteractionResponse(event, components)
}

func (a AdminCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "ollama",
			Description: "ollama subcommands",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "list",
					Description: "Lists Current pulled models",
				},
				{
					Name:        "pull",
					Description: "Downloads a model to use",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "model",
							Description: "Model to pull to use",
							Required:    true,
						},
					},
				},
			},
		},
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "model",
			Description: "model subcommands",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "list",
					Description: "Lists Current loaded models",
				},
				{
					Name:        "add",
					Description: "Add a model to use with the bot",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "model",
							Description: "Model to add",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Remove a llm model",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "Name of the model",
							Required:    true,
						},
					},
				},
			},
		},
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "platform",
			Description: "platform subcommands",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "add",
					Description: "Add a coin platform",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "id",
							Description: "ID of the platform",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "Name of the platform",
							Required:    true,
						},
						discord.ApplicationCommandOptionInt{
							Name:        "buying_power",
							Description: "Buying power of the platform",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Remove a coin platform",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "id",
							Description: "ID of the platform",
							Required:    true,
						},
					},
				},
				{
					Name:        "list",
					Description: "List all platforms",
				},
			},
		},
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "platform_model",
			Description: "platform model subcommands",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "set",
					Description: "Set a coin platform model settings",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "id",
							Description: "ID of the platform",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "Name of the model",
							Required:    true,
						},
						discord.ApplicationCommandOptionInt{
							Name:        "tokens",
							Description: "tokens of tokens per coin",
							Required:    true,
						},
					},
				},
			},
		},
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "prompt",
			Description: "prompt admin subcommands",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "replay",
					Description: "Replay a previous done prompt",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionInt{
							Name:        "id",
							Description: "ID of the prompt",
							Required:    true,
						},
					},
				},
				{
					Name:        "list",
					Description: "List all prompt",
				},
			},
		},
	}
}

func containsModel(model string, models []ollamaApi.ListModelResponse) bool {
	for _, m := range models {
		if model == m.Model {
			return true
		}
	}
	return false
}
