package admincommand

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	ollamaApi "github.com/ollama/ollama/api"
	"github.com/stollenaar/ollamabot/internal/util"
)

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
			util.RespondWithError(event, err)
			return
		}
		components = []discord.LayoutComponent{
			discord.TextDisplayComponent{
				Content: "Pulled model",
			},
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
