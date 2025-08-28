package admincommand

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
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

// CommandParsed parsed struct for count command
type CommandParsed struct {
	SubCommandGroup string
	SubCommand      string
	Arguments       map[string]string
}

func init() {
	client, err := ollamaApi.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	OllamaClient = client
}

func (a AdminCommand) Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Member.User.ID != util.ConfigFile.ADMIN_USER_ID {
		bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You are not the boss of me",
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsIsComponentsV2,
			},
		})
		return
	}
	err := bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsIsComponentsV2,
			Title: "tmp",
			Components: []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: "loading",
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Error deferring: %s\n", err)
		return
	}

	parsedArguments := a.ParseArguments(bot, interaction).(*CommandParsed)

	var components []discordgo.MessageComponent
	if parsedArguments.SubCommandGroup == "platform" {
		components = platformHandler(parsedArguments, bot, interaction)
	} else {
		switch parsedArguments.SubCommand {
		case "list":
			resp, err := OllamaClient.List(context.TODO())
			if err != nil {
				fmt.Printf("Error listing models: %s\n", err)
				return
			}

			for _, model := range resp.Models {
				container := discordgo.Container{
					Components: []discordgo.MessageComponent{
						discordgo.TextDisplay{
							Content: fmt.Sprintf("### Name: %s", model.Name),
						},
					},
				}
				components = append(components, container)
			}
		case "pull":
			components = pullHandler(parsedArguments, bot, interaction)
		}
	}
	_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Components: &components,
	})
	if err != nil {
		fmt.Printf("Error editing the response: %s\n", err)
	}
}

func (a AdminCommand) CreateCommandArguments() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "list",
			Description: "Lists Current loaded models",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
		{
			Name:        "pull",
			Description: "Downloads a model to use",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "model",
					Description: "Model to pull to use",
					Required:    true,
				},
			},
		},
		{
			Name:        "platform",
			Description: "platform subcommands",
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "Add a coin platform",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the platform",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "buying_power",
							Description: "Buying power of the platform",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Remove a coin platform",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "id",
							Description: "ID of the platform",
							Required:    true,
						},
					},
				},
				{
					Name:        "list",
					Description: "List all platforms",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
}

func (a AdminCommand) ParseArguments(bot *discordgo.Session, interaction *discordgo.InteractionCreate) interface{} {
	parsedArguments := new(CommandParsed)
	parsedArguments.Arguments = map[string]string{}

	// Access options in the order provided by the user.
	options := interaction.ApplicationCommandData().Options
	// parsedArguments.GuildID = interaction.GuildID
	// Or convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		switch opt.Type {
		case discordgo.ApplicationCommandOptionSubCommand:
			parsedArguments.SubCommand = opt.Name
			for _, option := range opt.Options {
				parsedArguments.Arguments[option.Name] = option.StringValue()
			}
		case discordgo.ApplicationCommandOptionSubCommandGroup:
			parsedArguments.SubCommandGroup = opt.Name
			parsedArguments.SubCommand = opt.Options[0].Name
			for _, sub := range opt.Options[0].Options {
				switch sub.Type {
				case discordgo.ApplicationCommandOptionInteger:
					parsedArguments.Arguments[sub.Name] = strconv.FormatInt(sub.IntValue(), 10)
				default:
					parsedArguments.Arguments[sub.Name] = sub.StringValue()
				}
			}
		default:
			optionMap[opt.Name] = opt
		}
	}

	return parsedArguments
}

func pullHandler(args *CommandParsed, bot *discordgo.Session, interaction *discordgo.InteractionCreate) (components []discordgo.MessageComponent) {
	err := OllamaClient.Pull(context.TODO(), &ollamaApi.PullRequest{
		Model: args.Arguments["model"],
	}, func(pr ollamaApi.ProgressResponse) error {
		return nil
	})
	if err != nil {
		fmt.Printf("Error pulling model: %s\n", err)
		components = []discordgo.MessageComponent{
			discordgo.TextDisplay{
				Content: err.Error(),
			},
		}
		_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Components: &components,
		})
		if err != nil {
			fmt.Printf("Error editing the response: %s\n", err)
		}
		return
	} else {
		components = []discordgo.MessageComponent{
			discordgo.TextDisplay{
				Content: "Pulled model",
			},
		}
	}
	return
}

func platformHandler(args *CommandParsed, bot *discordgo.Session, interaction *discordgo.InteractionCreate) (components []discordgo.MessageComponent) {
	switch args.SubCommand {
	case "add":

		bpwr, _ := strconv.Atoi(args.Arguments["buying_power"])
		platform := database.Platform{
			Name:        args.Arguments["name"],
			BuyingPower: bpwr,
		}
		err := database.AddPlatform(platform)
		if err != nil {
			fmt.Printf("Error creating platform: %s\n", err)
			components = []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: err.Error(),
				},
			}
			_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Components: &components,
			})
			if err != nil {
				fmt.Printf("Error editing the response: %s\n", err)
			}
		} else {
			components = []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: "Successfully added the platform",
				},
			}
			_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Components: &components,
			})
			if err != nil {
				fmt.Printf("Error editing the response: %s\n", err)
			}
		}
	case "list":
		platforms, err := database.ListPlatforms()

		if err != nil {
			fmt.Printf("Error listing platforms: %s\n", err)
			components = []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: err.Error(),
				},
			}
			_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Components: &components,
			})
			if err != nil {
				fmt.Printf("Error editing the response: %s\n", err)
			}

			return
		}

		for _, platform := range platforms {
			container := discordgo.Container{
				Components: []discordgo.MessageComponent{
					discordgo.TextDisplay{
						Content: fmt.Sprintf("### ID: %s\n### BuyingPower: %d", platform.Name, platform.BuyingPower),
					},
				},
			}
			components = append(components, container)
		}

	case "remove":
		err := database.RemovePlatform(args.Arguments["id"])
		if err != nil {
			fmt.Printf("Error removing platform: %s\n", err)
			components = []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: err.Error(),
				},
			}
			_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Components: &components,
			})
			if err != nil {
				fmt.Printf("Error editing the response: %s\n", err)
			}
		} else {
			components = []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: "Successfully removed the platform",
				},
			}
			_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Components: &components,
			})
			if err != nil {
				fmt.Printf("Error editing the response: %s\n", err)
			}
		}
	}
	return
}
