package promptcommand

import (
	"context"
	"fmt"
	"iter"
	"log"
	"maps"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
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

func init() {
	client, err := ollamaApi.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	OllamaClient = client
}

func (p PromptCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	models, err := database.ListPlatformModels()

	if err != nil {
		fmt.Printf("Error fetching models: %e\n", err)

		err := event.CreateMessage(discord.MessageCreate{
			Flags: util.ConfigFile.SetEphemeral() | discord.MessageFlagIsComponentsV2,
			Components: []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "error fetching models",
				},
			},
		})

		if err != nil {
			fmt.Printf("Error deferring: %s\n", err)
		}
		return
	}

	err = event.Modal(discord.ModalCreate{
		CustomID: "prompt",
		Title:    "Submit Prompt",
		Components: []discord.LayoutComponent{
			discord.LabelComponent{
				Label: "Select Model",
				Component: discord.StringSelectMenuComponent{
					CustomID: "model",
					Options:  modelsToOptions(slices.Collect(maps.Keys(models))),
				},
			},
			discord.LabelComponent{
				Label: "Prompt",
				Component: discord.TextInputComponent{
					CustomID: "prompt",
					Style:    discord.TextInputStyleParagraph,
					Required: true,
				},
			},
		},
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (p PromptCommand) ModalHandler(event *events.ModalSubmitInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		fmt.Printf("Error deferring: %s\n", err)
		return
	}

	submittedData := extractModalSubmitData(event.Data.AllComponents())

	OllamaClient.Generate(context.TODO(), &ollamaApi.GenerateRequest{
		Model:  submittedData["model"],
		Prompt: submittedData["prompt"],
		Stream: new(bool),
	}, func(gr ollamaApi.GenerateResponse) error {
		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: &gr.Response,
		})
		if err != nil {
			fmt.Printf("Error updating prompt: %s\n", err)
		}
		return err
	})
}

func (p PromptCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return nil
}

func (p PromptCommand) ParseArguments(event *events.ApplicationCommandInteractionCreate) interface{} {
	return nil
}

func modelsToOptions(models []string) (options []discord.StringSelectMenuOption) {
	for _, model := range models {
		options = append(options, discord.StringSelectMenuOption{
			Label: model,
			Value: model,
		})
	}
	return
}

func extractModalSubmitData(components iter.Seq[discord.Component]) map[string]string {
	formData := make(map[string]string)
	for component := range components {
		switch c := component.(type) {
		case discord.TextInputComponent:
			formData[c.CustomID] = c.Value
		case discord.StringSelectMenuComponent:
			formData[c.CustomID] = c.Values[0]
		}
	}
	return formData
}
