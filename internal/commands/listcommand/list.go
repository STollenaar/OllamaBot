package listcommand

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/ollamabot/internal/database"
	"github.com/stollenaar/ollamabot/internal/util"
)

var (
	ListCmd = ListCommand{
		Name:        "list",
		Description: "List command to see what models are available",
	}
)

type ListCommand struct {
	Name        string
	Description string
}

func (l ListCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)
	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	models, err := database.ListPlatformModels()
	var components []discord.LayoutComponent

	if err != nil {
		slog.Error("Error pulling model: ", slog.Any("err", err))
		components = []discord.LayoutComponent{
			discord.TextDisplayComponent{
				Content: err.Error(),
			},
		}
		util.UpdateInteractionResponse(event, components)
		return
	}

	for model, platforms := range models {
		var costs []string
		for _, platform := range platforms {
			costs = append(costs, fmt.Sprintf("### Platform: %s\n### Cost: %d/token", platform.PlatformName, platform.Tokens))
		}
		container := discord.ContainerComponent{
			Components: []discord.ContainerSubComponent{
				discord.TextDisplayComponent{
					Content: fmt.Sprintf("### Name: %s\n%s", model, strings.Join(costs, "\n")),
				},
			},
		}
		components = append(components, container)
	}

	if len(components) == 0 {
		container := discord.ContainerComponent{
			Components: []discord.ContainerSubComponent{
				discord.TextDisplayComponent{
					Content: "No models are available at the moment",
				},
			},
		}
		components = append(components, container)
	}
	util.UpdateInteractionResponse(event, components)
}

func (l ListCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return nil
}
