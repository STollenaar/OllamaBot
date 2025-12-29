package admincommand

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	ollamaApi "github.com/ollama/ollama/api"
	"github.com/stollenaar/ollamabot/internal/database"
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
	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Components: &components,
		Flags:      util.ConfigFile.SetComponentV2Flags(),
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
	}
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

	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Components: &components,
		Flags:      util.ConfigFile.SetComponentV2Flags(),
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
	}
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

func ollamaHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) (components []discord.LayoutComponent) {
	switch *args.SubCommandName {
	case "pull":
		err := OllamaClient.Pull(context.TODO(), &ollamaApi.PullRequest{
			Model: args.Options["model"].String(),
		}, func(pr ollamaApi.ProgressResponse) error {
			return nil
		})
		if err != nil {
			slog.Error("Error pulling model: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
			return
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Pulled model",
				},
			}
		}
	case "list":
		resp, err := OllamaClient.List(context.TODO())
		if err != nil {
			slog.Error("Error listing models: ", slog.Any("err", err))
			return
		}

		for _, model := range resp.Models {
			container := discord.ContainerComponent{
				Components: []discord.ContainerSubComponent{
					discord.TextDisplayComponent{
						Content: fmt.Sprintf("### Name: %s", model.Name),
					},
				},
			}
			components = append(components, container)
		}

		if len(resp.Models) == 0 {
			components = append(components, discord.ContainerComponent{
				Components: []discord.ContainerSubComponent{
					discord.TextDisplayComponent{
						Content: "No Ollama Models Available",
					},
				},
			})
		}
	}
	return
}

func modelHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) (components []discord.LayoutComponent) {
	switch *args.SubCommandName {
	case "add":
		model := args.Options["model"].String()

		resp, err := OllamaClient.List(context.TODO())
		if err != nil {
			slog.Error("Error listing models: ", slog.Any("err", err))
			return
		}
		if !containsModel(model, resp.Models) {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: fmt.Sprintf("Model %s is not already in Ollama", model),
				},
			}
			return
		}

		err = database.AddModel(model)
		if err != nil {
			slog.Error("Error creating model: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Successfully added the model",
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
		}
	case "list":
		models, err := database.ListModels()

		if err != nil {
			slog.Error("Error listing platforms: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}

			return
		}

		for _, model := range models {
			container := discord.ContainerComponent{
				Components: []discord.ContainerSubComponent{
					discord.TextDisplayComponent{
						Content: fmt.Sprintf("### Name: %s", model),
					},
				},
			}
			components = append(components, container)
		}

		if len(models) == 0 {
			components = append(components, discord.ContainerComponent{
				Components: []discord.ContainerSubComponent{
					discord.TextDisplayComponent{
						Content: "No Models Configured",
					},
				},
			})
		}

	case "remove":
		err := database.RemoveModel(args.Options["name"].String())
		if err != nil {
			slog.Error("Error removing model: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Successfully removed the model",
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
		}
	}
	return
}

func platformHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) (components []discord.LayoutComponent) {
	switch *args.SubCommandName {
	case "add":

		platform := database.Platform{
			ID:          args.Options["id"].String(),
			Name:        args.Options["name"].String(),
			BuyingPower: args.Options["buying_power"].Int(),
		}
		err := database.AddPlatform(platform)
		if err != nil {
			slog.Error("Error creating platform: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Successfully added the platform",
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
		}
	case "list":
		platforms, err := database.ListPlatforms()

		if err != nil {
			slog.Error("Error listing platforms: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}

			return
		}

		for _, platform := range platforms {
			container := discord.ContainerComponent{
				Components: []discord.ContainerSubComponent{
					discord.TextDisplayComponent{
						Content: fmt.Sprintf("### ID: %s\n### Name: %s\n### BuyingPower: %d", platform.ID, platform.Name, platform.BuyingPower),
					},
				},
			}
			components = append(components, container)
		}

		if len(platforms) == 0 {
			components = append(components, discord.ContainerComponent{
				Components: []discord.ContainerSubComponent{
					discord.TextDisplayComponent{
						Content: "No Platforms Configured",
					},
				},
			})
		}

	case "remove":
		err := database.RemovePlatform(args.Options["id"].String())
		if err != nil {
			slog.Error("Error removing platform: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Successfully removed the platform",
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}
		}
	}
	return
}

func platformModelHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) (components []discord.LayoutComponent) {
	platform, err := database.GetPlatform(args.Options["id"].String())

	if err != nil {
		slog.Error("Error fetching platform: ", slog.Any("err", err))
		components = []discord.LayoutComponent{
			discord.TextDisplayComponent{
				Content: err.Error(),
			},
		}
		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Components: &components,
			Flags:      util.ConfigFile.SetComponentV2Flags(),
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
		}
		return
	}

	model, err := database.GetModel(args.Options["name"].String())

	if err != nil {
		slog.Error("Error fetching model: ", slog.Any("err", err))
		components = []discord.LayoutComponent{
			discord.TextDisplayComponent{
				Content: err.Error(),
			},
		}
		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Components: &components,
			Flags:      util.ConfigFile.SetComponentV2Flags(),
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
		}
		return
	}

	err = database.SetPlatformModels(platform.ID, model, args.Options["tokens"].Int())
	if err != nil {
		slog.Error("Error setting platform_model tokens: ", slog.Any("err", err))
		components = []discord.LayoutComponent{
			discord.TextDisplayComponent{
				Content: err.Error(),
			},
		}
		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Components: &components,
			Flags:      util.ConfigFile.SetComponentV2Flags(),
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
		}
		return
	}

	components = []discord.LayoutComponent{
		discord.TextDisplayComponent{
			Content: "Successfully added the platform",
		},
	}
	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Components: &components,
		Flags:      util.ConfigFile.SetComponentV2Flags(),
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
	}
	return
}

func promptHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) (components []discord.LayoutComponent) {
	switch *args.SubCommandName {
	case "list":
		history, err := database.ListHistory(0)

		if err != nil {
			slog.Error("Error listing history: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}

			return
		}

		var container discord.ContainerComponent
		for _, hist := range history {
			user, err := event.Client().Rest.GetUser(snowflake.MustParse(hist.UserID))
			if err != nil {
				slog.Error("Error fetching user", slog.Any("err", err))
			}
			container.Components = append(container.Components, discord.SectionComponent{
				Components: []discord.SectionSubComponent{
					discord.TextDisplayComponent{
						Content: fmt.Sprintf("**ID:** %d\n**Model Name:** %s\n**User:** %s", hist.ID, hist.ModelName, user.Username),
					},
					discord.TextDisplayComponent{
						Content: fmt.Sprintf("**Prompt:**\r%s", hist.Prompt),
					},
				},
				Accessory: discord.ButtonComponent{
					Style:    discord.ButtonStylePrimary,
					Label:    "Replay",
					CustomID: fmt.Sprintf("admin_prompt_page_replay_%d", hist.ID),
				},
			},
				discord.SeparatorComponent{},
			)
		}
		container.Components = append(container.Components,
			discord.ActionRowComponent{
				Components: []discord.InteractiveComponent{
					discord.ButtonComponent{
						CustomID: fmt.Sprintf("admin_prompt_page_previous_%d", 0),
						Label:    "Previous",
						Style:    discord.ButtonStylePrimary,
					},
					discord.ButtonComponent{
						CustomID: fmt.Sprintf("admin_prompt_page_next_%d", 0),
						Label:    "Next",
						Style:    discord.ButtonStylePrimary,
					},
				},
			},
		)

		if len(history) == 0 {
			container.Components = []discord.ContainerSubComponent{
				discord.TextDisplayComponent{
					Content: "No History Yet",
				},
			}
		}
		components = append(components, container)
	case "replay":
		history, err := database.GetHistory(args.Options["id"].Int())

		if err != nil {
			slog.Error("Error fetching history: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}

			return
		}

		OllamaClient.Generate(context.TODO(), &ollamaApi.GenerateRequest{
			Model:  history.ModelName,
			Prompt: history.Prompt,
			Stream: new(bool),
		}, func(gr ollamaApi.GenerateResponse) error {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: gr.Response,
				},
				discord.ActionRowComponent{
					Components: []discord.InteractiveComponent{
						discord.ButtonComponent{
							Style:    discord.ButtonStylePrimary,
							Label:    "Post Prompt",
							CustomID: "admin_prompt_page_post",
						},
						discord.ButtonComponent{
							Style:    discord.ButtonStyleDanger,
							Label:    "Retry Prompt",
							CustomID: fmt.Sprintf("admin_prompt_page_retry_%d", history.ID),
						},
					},
				},
			}
			return nil
		})
	}
	return
}

func promptButtonHandler(event *events.ComponentInteractionCreate) (components []discord.LayoutComponent) {
	if event.Member().User.ID.String() != util.ConfigFile.ADMIN_USER_ID {
		return []discord.LayoutComponent{}
	}

	switch strings.Split(event.Data.CustomID(), "_")[3] {
	case "post":
		return []discord.LayoutComponent{
			event.Message.Components[0],
		}
	case "previous":
		index, _ := strconv.Atoi(strings.Split(event.Data.CustomID(), "_")[4])
		index -= 6
		if index < 0 {
			index = 0
		}
		return promptListHandler(index, event)
	case "next":
		index, _ := strconv.Atoi(strings.Split(event.Data.CustomID(), "_")[4])
		index += 6
		maxSeq := database.CountHistory()
		if index > maxSeq {
			index -= 6
		}
		return promptListHandler(index, event)
	case "retry":
		fallthrough
	case "replay":
		id, _ := strconv.Atoi(strings.Split(event.Data.CustomID(), "_")[4])
		history, err := database.GetHistory(id)

		if err != nil {
			slog.Error("Error fetching history: ", slog.Any("err", err))
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: err.Error(),
				},
			}
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Components: &components,
				Flags:      util.ConfigFile.SetComponentV2Flags(),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
			}

			return
		}

		OllamaClient.Generate(context.TODO(), &ollamaApi.GenerateRequest{
			Model:  history.ModelName,
			Prompt: history.Prompt,
			Stream: new(bool),
		}, func(gr ollamaApi.GenerateResponse) error {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: gr.Response,
				},
				discord.ActionRowComponent{
					Components: []discord.InteractiveComponent{
						discord.ButtonComponent{
							Style:    discord.ButtonStylePrimary,
							Label:    "Post Prompt",
							CustomID: "admin_prompt_page_post",
						},
						discord.ButtonComponent{
							Style:    discord.ButtonStyleDanger,
							Label:    "Retry Prompt",
							CustomID: fmt.Sprintf("admin_prompt_page_retry_%d", id),
						},
					},
				},
			}
			return nil
		})
		return components
	default:
		return []discord.LayoutComponent{}
	}
}

func promptListHandler(index int, event *events.ComponentInteractionCreate) (components []discord.LayoutComponent) {
	history, err := database.ListHistory(index)

	if err != nil {
		slog.Error("Error listing history: ", slog.Any("err", err))
		components = []discord.LayoutComponent{
			discord.TextDisplayComponent{
				Content: err.Error(),
			},
		}
		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Components: &components,
			Flags:      util.ConfigFile.SetComponentV2Flags(),
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
		}

		return
	}

	var container discord.ContainerComponent
	for _, hist := range history {
		user, err := event.Client().Rest.GetUser(snowflake.MustParse(hist.UserID))
		if err != nil {
			slog.Error("Error fetching user", slog.Any("err", err))
		}

		container.Components = append(container.Components, discord.SectionComponent{
			Components: []discord.SectionSubComponent{
				discord.TextDisplayComponent{
					Content: fmt.Sprintf("**ID:** %d\n**Model Name:** %s\n**User:** %s", hist.ID, hist.ModelName, user.Username),
				},
				discord.TextDisplayComponent{
					Content: fmt.Sprintf("**Prompt:**\r%s", hist.Prompt),
				},
			},
			Accessory: discord.ButtonComponent{
				Style:    discord.ButtonStylePrimary,
				Label:    "Replay",
				CustomID: fmt.Sprintf("admin_prompt_page_replay_%d", hist.ID),
			},
		},
			discord.SeparatorComponent{},
		)
	}
	container.Components = append(container.Components,
		discord.ActionRowComponent{
			Components: []discord.InteractiveComponent{
				discord.ButtonComponent{
					CustomID: fmt.Sprintf("admin_prompt_page_previous_%d", 0),
					Label:    "Previous",
					Style:    discord.ButtonStylePrimary,
				},
				discord.ButtonComponent{
					CustomID: fmt.Sprintf("admin_prompt_page_next_%d", 0),
					Label:    "Next",
					Style:    discord.ButtonStylePrimary,
				},
			},
		},
	)

	if len(history) == 0 {
		container.Components = []discord.ContainerSubComponent{
			discord.TextDisplayComponent{
				Content: "No History Yet",
			},
		}
	}
	components = append(components, container)
	return
}

func containsModel(model string, models []ollamaApi.ListModelResponse) bool {
	for _, m := range models {
		if model == m.Model {
			return true
		}
	}
	return false
}
