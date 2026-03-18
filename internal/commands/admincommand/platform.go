package admincommand

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/stollenaar/ollamabot/internal/database"
	"github.com/stollenaar/ollamabot/internal/util"
)

func platformHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) (components []discord.LayoutComponent) {
	switch *args.SubCommandName {
	case "add":

		platform := database.Platform{
			ID:          args.Options["id"].String(),
			Name:        args.Options["name"].String(),
			BuyingPower: args.Options["buying_power"].Int(),
		}
		err := database.AddPlatform(platform)
		if err != nil {
			slog.Error("Error creating platform: ", slog.Any("err", err))
			util.RespondWithError(event, err)
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Successfully added the platform",
				},
			}
			util.UpdateInteractionResponse(event, components)
		}
	case "list":
		platforms, err := database.ListPlatforms()

		if err != nil {
			slog.Error("Error listing platforms: ", slog.Any("err", err))
			util.RespondWithError(event, err)
			return
		}

		for _, platform := range platforms {
			container := discord.ContainerComponent{
				Components: []discord.ContainerSubComponent{
					discord.TextDisplayComponent{
						Content: fmt.Sprintf("### ID: %s\n### Name: %s\n### BuyingPower: %d", platform.ID, platform.Name, platform.BuyingPower),
					},
				},
			}
			components = append(components, container)
		}

		if len(platforms) == 0 {
			components = append(components, discord.ContainerComponent{
				Components: []discord.ContainerSubComponent{
					discord.TextDisplayComponent{
						Content: "No Platforms Configured",
					},
				},
			})
		}

	case "remove":
		err := database.RemovePlatform(args.Options["id"].String())
		if err != nil {
			slog.Error("Error removing platform: ", slog.Any("err", err))
			util.RespondWithError(event, err)
		} else {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: "Successfully removed the platform",
				},
			}
			util.UpdateInteractionResponse(event, components)
		}
	}
	return
}

func platformModelHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) (components []discord.LayoutComponent) {
	platform, err := database.GetPlatform(args.Options["id"].String())

	if err != nil {
		slog.Error("Error fetching platform: ", slog.Any("err", err))
		util.RespondWithError(event, err)
		return
	}

	model, err := database.GetModel(args.Options["name"].String())

	if err != nil {
		slog.Error("Error fetching model: ", slog.Any("err", err))
		util.RespondWithError(event, err)
		return
	}

	err = database.SetPlatformModels(platform.ID, model, args.Options["tokens"].Int())
	if err != nil {
		slog.Error("Error setting platform_model tokens: ", slog.Any("err", err))
		util.RespondWithError(event, err)
		return
	}

	components = []discord.LayoutComponent{
		discord.TextDisplayComponent{
			Content: "Successfully added the platform",
		},
	}
	util.UpdateInteractionResponse(event, components)
	return
}
