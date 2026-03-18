package admincommand

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/ollamabot/internal/database"
	"github.com/stollenaar/ollamabot/internal/util"
)

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
			util.UpdateInteractionResponse(event, components)
			return
		}

		err = database.AddModel(model)
		if err != nil {
			slog.Error("Error creating model: ", slog.Any("err", err))
			util.RespondWithError(event, err)
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Successfully added the model",
				},
			}
			util.UpdateInteractionResponse(event, components)
		}
	case "list":
		models, err := database.ListModels()

		if err != nil {
			slog.Error("Error listing platforms: ", slog.Any("err", err))
			util.RespondWithError(event, err)
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
			util.RespondWithError(event, err)
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Successfully removed the model",
				},
			}
			util.UpdateInteractionResponse(event, components)
		}
	}
	return
}
