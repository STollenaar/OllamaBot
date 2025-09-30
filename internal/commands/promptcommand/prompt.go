package promptcommand

import (
	"context"
	"iter"
	"log"
	"log/slog"
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
		slog.Error("Error fetching models: ", slog.Any("err", err))

		err := event.CreateMessage(discord.MessageCreate{
			Flags: util.ConfigFile.SetEphemeral() | discord.MessageFlagIsComponentsV2,
			Components: []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "error fetching models",
				},
			},
		})

		if err != nil {
			slog.Error("Error deferring: ", slog.Any("err", err))
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
			discord.LabelComponent{
				Label: "Context",
				Component: discord.StringSelectMenuComponent{
					CustomID: "context",
					Options: []discord.StringSelectMenuOption{
						{
							Label: "New Context",
							Value: "new_context",
						},
						{
							Label:   "Current Context",
							Value:   "current_context",
							Default: true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		slog.Error("Error creating modal: ", slog.Any("err", err))
	}
}

func (p PromptCommand) ModalHandler(event *events.ModalSubmitInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	submittedData := extractModalSubmitData(event.Data.AllComponents())
	slog.Info("Received prompt submission",
		slog.String("model", submittedData["model"]),
		slog.String("prompt", submittedData["prompt"]),
	)

	err = database.AddHistory(database.History{
		ModelName: submittedData["model"],
		UserID:    event.User().ID.String(),
		Prompt:    submittedData["prompt"],
	})

	if err != nil {
		slog.Error("Error saving history: ", slog.Any("err", err))
	}

	var ollamaContext []int32
	if submittedData["context"] == "current_context" {
		ollamaContext = database.GetContext(event.User().ID.String(), submittedData["model"])
	}

	OllamaClient.Generate(context.TODO(), &ollamaApi.GenerateRequest{
		Model:   submittedData["model"],
		Prompt:  submittedData["prompt"],
		Stream:  new(bool),
		Context: int32ToIntSlice(ollamaContext),
	}, func(gr ollamaApi.GenerateResponse) error {
		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: &gr.Response,
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", gr.Response))
		}

		err = database.SetContext(database.UserContext{
			UserID:    event.User().ID.String(),
			ModelName: submittedData["model"],
			Context:   intToInt32Slice(gr.Context),
		})
		if err != nil {
			slog.Error("Error updating context:", slog.Any("err", err))
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

func intToInt32Slice(input []int) []int32 {
	output := make([]int32, len(input))
	for i, v := range input {
		output[i] = int32(v)
	}
	return output
}

func int32ToIntSlice(input []int32) []int {
	output := make([]int, len(input))
	for i, v := range input {
		output[i] = int(v)
	}
	return output
}
