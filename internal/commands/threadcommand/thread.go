package threadcommand

import (
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
	ThreadCmd = ThreadCommand{
		Name:        "thread",
		Description: "spawn a thread for a contained conversation",
	}
	OllamaClient *ollamaApi.Client
)

type ThreadCommand struct {
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

func (t ThreadCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
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
		CustomID: "thread",
		Title:    "Create LLM Thread",
		Components: []discord.LayoutComponent{
			discord.LabelComponent{
				Label: "Select Model",
				Component: discord.StringSelectMenuComponent{
					CustomID: "model",
					Options:  modelsToOptions(slices.Collect(maps.Keys(models))),
				},
			},
			discord.LabelComponent{
				Label: "Title",
				Component: discord.TextInputComponent{
					CustomID: "title",
					Style:    discord.TextInputStyleShort,
					Required: true,
				},
			},
			discord.LabelComponent{
				Label: "System prompt",
				Component: discord.TextInputComponent{
					CustomID: "system",
					Style:    discord.TextInputStyleParagraph,
					Required: true,
				},
			},
		},
	})
	if err != nil {
		slog.Error("Error creating modal: ", slog.Any("err", err))
	}
}

func (t ThreadCommand) ModalHandler(event *events.ModalSubmitInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	submittedData := extractModalSubmitData(event.Data.AllComponents())
	slog.Info("Received model submission",
		slog.String("model", submittedData["model"]),
		slog.String("system", submittedData["system"]),
		slog.String("title", submittedData["title"]),
	)

	thread, err := event.Client().Rest.CreateThread(event.Channel().ID(), discord.GuildPublicThreadCreate{
		Name:                submittedData["title"],
		AutoArchiveDuration: discord.AutoArchiveDuration24h,
	})

	if err != nil {
		slog.Error("Error creating thread: ", slog.Any("err", err))
		return
	}

	err = database.AddThread(submittedData["model"], submittedData["system"], thread.ID().String())

	if err != nil {
		slog.Error("Error saving thread info: ", slog.Any("err", err))
		return
	}
}

func (t ThreadCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
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
