package admincommand

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
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

// CommandParsed parsed struct for count command
type CommandParsed struct {
	SubCommand string
	Arguments  map[string]string
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
		err := OllamaClient.Pull(context.TODO(), &ollamaApi.PullRequest{
			Model: parsedArguments.Arguments["model"],
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
			return
		} else {
			components = []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: "Pulled model",
				},
			}
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
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
			parsedArguments.SubCommand = opt.Name
			for _, option := range opt.Options {
				parsedArguments.Arguments[option.Name] = option.StringValue()
			}
		} else {
			optionMap[opt.Name] = opt
		}
	}

	return parsedArguments
}
