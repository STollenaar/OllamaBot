package listcommand

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	ollamaApi "github.com/ollama/ollama/api"
	"github.com/stollenaar/ollamabot/internal/database"
)

var (
	ListCmd = ListCommand{
		Name:        "list",
		Description: "List command to see what models are available",
	}
	OllamaClient *ollamaApi.Client
)

type ListCommand struct {
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

func (l ListCommand) Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
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

	models, err := database.ListPlatformModels()
	var components []discordgo.MessageComponent

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
	}

	for model, platforms := range models {
		var costs []string
		for _, platform := range platforms {
			costs = append(costs, fmt.Sprintf("### Platform: %s\n### Cost: %d/token", platform.PlatformName, platform.Cost))
		}
		container := discordgo.Container{
			Components: []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: fmt.Sprintf("### Name: %s\n%s", model, strings.Join(costs, "\n")),
				},
			},
		}
		components = append(components, container)
	}

	if len(components) == 0 {
		container := discordgo.Container{
			Components: []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: fmt.Sprintf("No models are available at the moment"),
				},
			},
		}
		components = append(components, container)
	}
	_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Components: &components,
	})

	if err != nil {
		fmt.Printf("Error editing the response: %s\n", err)
	}
}

func (l ListCommand) CreateCommandArguments() []*discordgo.ApplicationCommandOption {
	return nil
}

func (l ListCommand) ParseArguments(bot *discordgo.Session, interaction *discordgo.InteractionCreate) interface{} {
	return nil
}
