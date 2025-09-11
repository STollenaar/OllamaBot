package promptcommand

import (
	"context"
	"fmt"
	"log"
	"maps"
	"slices"

	"github.com/bwmarrin/discordgo"
	ollamaApi "github.com/ollama/ollama/api"
	"github.com/stollenaar/ollamabot/internal/database"
	"github.com/stollenaar/ollamabot/internal/util"
)

var (
	PromptCmd = PromptCommand{
		Name:        "prompt",
		Description: "Prompt command to query ollama",
	}
	OllamaClient *ollamaApi.Client
)

type PromptCommand struct {
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

func (p PromptCommand) Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	models, err := database.ListPlatformModels()

	if err != nil {
		fmt.Printf("Error fetching models: %e\n", err)
		err := bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: util.ConfigFile.SetEphemeral() | discordgo.MessageFlagsIsComponentsV2,
				Title: "tmp",
				Components: []discordgo.MessageComponent{
					discordgo.TextDisplay{
						Content: "error fetching models",
					},
				},
			},
		})

		if err != nil {
			fmt.Printf("Error deferring: %s\n", err)
		}
		return
	}

	err = bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "prompt",
			Title:    "Submit Prompt",
			Flags:    discordgo.MessageFlagsIsComponentsV2,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID: "model",
							MenuType: discordgo.StringSelectMenu,
							Options:  modelsToOptions(slices.Collect(maps.Keys(models))),
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "prompt",
							Label:    "Prompt",
							Style:    discordgo.TextInputParagraph,
							Required: true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (p PromptCommand) ModalHandler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	err := bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
		Data: &discordgo.InteractionResponseData{
			Flags:   util.ConfigFile.SetEphemeral(),
			Content: "Loading...",
		},
	})

	if err != nil {
		fmt.Printf("Error deferring: %s\n", err)
		return
	}

	submittedData := extractModalSubmitData(interaction.ModalSubmitData().Components)

	OllamaClient.Generate(context.TODO(), &ollamaApi.GenerateRequest{
		Model:  submittedData["model"],
		Prompt: submittedData["prompt"],
	}, func(gr ollamaApi.GenerateResponse) error {
		bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &gr.Response,
		})
		return nil
	})
}

func (p PromptCommand) CreateCommandArguments() []*discordgo.ApplicationCommandOption {
	return nil
	// return []*discordgo.ApplicationCommandOption{
	// 	{
	// 		Name:        "name",
	// 		Description: "Name of the model",
	// 		Type:        discordgo.ApplicationCommandOptionString,
	// 		Required:    true,
	// 	},
	// 	{
	// 		Name:        "prompt",
	// 		Description: "prompt you want to send",
	// 		Type:        discordgo.ApplicationCommandOptionString,
	// 		Required:    true,
	// 	},
	// }
}

func (p PromptCommand) ParseArguments(bot *discordgo.Session, interaction *discordgo.InteractionCreate) interface{} {
	return nil
}

func modelsToOptions(models []string) (options []discordgo.SelectMenuOption) {
	for _, model := range models {
		options = append(options, discordgo.SelectMenuOption{
			Label: model,
			Value: model,
		})
	}
	return
}

func extractModalSubmitData(components []discordgo.MessageComponent) map[string]string {
	formData := make(map[string]string)
	for _, component := range components {
		input := component.(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput)
		formData[input.CustomID] = input.Value
	}
	return formData
}
