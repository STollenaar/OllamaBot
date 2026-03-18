package admincommand

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	ollamaApi "github.com/ollama/ollama/api"
	"github.com/stollenaar/ollamabot/internal/database"
	"github.com/stollenaar/ollamabot/internal/util"
)

func promptHandler(args discord.SlashCommandInteractionData, event *events.ApplicationCommandInteractionCreate) (components []discord.LayoutComponent) {
	switch *args.SubCommandName {
	case "list":
		history, err := database.ListHistory(0)

		if err != nil {
			slog.Error("Error listing history: ", slog.Any("err", err))
			util.RespondWithError(event, err)
			return
		}

		var container discord.ContainerComponent
		for _, hist := range history {
			user, err := event.Client().Rest.GetUser(snowflake.MustParse(hist.UserID))
			if err != nil {
				slog.Error("Error fetching user", slog.Any("err", err))
			}
			container.Components = append(container.Components, discord.SectionComponent{
				Components: []discord.SectionSubComponent{
					discord.TextDisplayComponent{
						Content: fmt.Sprintf("**ID:** %d\n**Model Name:** %s\n**User:** %s", hist.ID, hist.ModelName, user.Username),
					},
					discord.TextDisplayComponent{
						Content: fmt.Sprintf("**Prompt:**\r%s", hist.Prompt),
					},
				},
				Accessory: discord.ButtonComponent{
					Style:    discord.ButtonStylePrimary,
					Label:    "Replay",
					CustomID: fmt.Sprintf("admin_prompt_page_replay_%d", hist.ID),
				},
			},
				discord.SeparatorComponent{},
			)
		}
		maxSeq := database.CountHistory()

		container.Components = append(container.Components,
			discord.ActionRowComponent{
				Components: []discord.InteractiveComponent{
					discord.ButtonComponent{
						CustomID: fmt.Sprintf("admin_prompt_page_first_%d", 0),
						Label:    "First",
						Style:    discord.ButtonStyleSecondary,
					},
					discord.ButtonComponent{
						CustomID: fmt.Sprintf("admin_prompt_page_previous_%d", 0),
						Label:    "Previous",
						Style:    discord.ButtonStylePrimary,
					},
					discord.ButtonComponent{
						CustomID: fmt.Sprintf("admin_prompt_page_next_%d", 6),
						Label:    "Next",
						Style:    discord.ButtonStylePrimary,
					},
					discord.ButtonComponent{
						CustomID: fmt.Sprintf("admin_prompt_page_last_%d", maxSeq-6),
						Label:    "Last",
						Style:    discord.ButtonStyleSecondary,
					},
				},
			},
		)

		if len(history) == 0 {
			container.Components = []discord.ContainerSubComponent{
				discord.TextDisplayComponent{
					Content: "No History Yet",
				},
			}
		}
		components = append(components, container)
	case "replay":
		history, err := database.GetHistory(args.Options["id"].Int())

		if err != nil {
			slog.Error("Error fetching history: ", slog.Any("err", err))
			util.RespondWithError(event, err)
			return
		}

		OllamaClient.Generate(context.TODO(), &ollamaApi.GenerateRequest{
			Model:  history.ModelName,
			Prompt: history.Prompt,
			Stream: new(bool),
		}, func(gr ollamaApi.GenerateResponse) error {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: gr.Response,
				},
				discord.ActionRowComponent{
					Components: []discord.InteractiveComponent{
						discord.ButtonComponent{
							Style:    discord.ButtonStylePrimary,
							Label:    "Post Prompt",
							CustomID: "admin_prompt_page_post",
						},
						discord.ButtonComponent{
							Style:    discord.ButtonStyleDanger,
							Label:    "Retry Prompt",
							CustomID: fmt.Sprintf("admin_prompt_page_retry_%d", history.ID),
						},
					},
				},
			}
			return nil
		})
	}
	return
}

func promptButtonHandler(event *events.ComponentInteractionCreate) (components []discord.LayoutComponent) {
	if event.Member().User.ID.String() != util.ConfigFile.ADMIN_USER_ID {
		return []discord.LayoutComponent{}
	}

	maxSeq := database.CountHistory()
	switch strings.Split(event.Data.CustomID(), "_")[3] {
	case "post":
		return []discord.LayoutComponent{
			event.Message.Components[0],
		}
	case "previous":
		index, _ := strconv.Atoi(strings.Split(event.Data.CustomID(), "_")[4])
		index -= 6
		if index < 0 {
			index = 0
		}
		return promptListHandler(index, maxSeq-6, event)
	case "next":
		index, _ := strconv.Atoi(strings.Split(event.Data.CustomID(), "_")[4])
		index += 6
		if index > maxSeq {
			index -= 6
		}
		return promptListHandler(index, maxSeq-6, event)
	case "last":
		return promptListHandler(maxSeq-6, maxSeq-6, event)
	case "first":
		return promptListHandler(0, maxSeq-6, event)
	case "retry":
		fallthrough
	case "replay":
		id, _ := strconv.Atoi(strings.Split(event.Data.CustomID(), "_")[4])
		history, err := database.GetHistory(id)

		if err != nil {
			slog.Error("Error fetching history: ", slog.Any("err", err))
			util.RespondWithErrorComponent(event, err)
			return
		}

		OllamaClient.Generate(context.TODO(), &ollamaApi.GenerateRequest{
			Model:  history.ModelName,
			Prompt: history.Prompt,
			Stream: new(bool),
		}, func(gr ollamaApi.GenerateResponse) error {
			components = []discord.LayoutComponent{
				discord.TextDisplayComponent{
					Content: gr.Response,
				},
				discord.ActionRowComponent{
					Components: []discord.InteractiveComponent{
						discord.ButtonComponent{
							Style:    discord.ButtonStylePrimary,
							Label:    "Post Prompt",
							CustomID: "admin_prompt_page_post",
						},
						discord.ButtonComponent{
							Style:    discord.ButtonStyleDanger,
							Label:    "Retry Prompt",
							CustomID: fmt.Sprintf("admin_prompt_page_retry_%d", id),
						},
					},
				},
			}
			return nil
		})
		return components
	default:
		return []discord.LayoutComponent{}
	}
}

func promptListHandler(index, max int, event *events.ComponentInteractionCreate) (components []discord.LayoutComponent) {
	history, err := database.ListHistory(index)

	if err != nil {
		slog.Error("Error listing history: ", slog.Any("err", err))
		util.RespondWithErrorComponent(event, err)
		return
	}

	var container discord.ContainerComponent
	for _, hist := range history {
		user, err := event.Client().Rest.GetUser(snowflake.MustParse(hist.UserID))
		if err != nil {
			slog.Error("Error fetching user", slog.Any("err", err))
		}

		container.Components = append(container.Components, discord.SectionComponent{
			Components: []discord.SectionSubComponent{
				discord.TextDisplayComponent{
					Content: fmt.Sprintf("**ID:** %d\n**Model Name:** %s\n**User:** %s", hist.ID, hist.ModelName, user.Username),
				},
				discord.TextDisplayComponent{
					Content: fmt.Sprintf("**Prompt:**\r%s", hist.Prompt),
				},
			},
			Accessory: discord.ButtonComponent{
				Style:    discord.ButtonStylePrimary,
				Label:    "Replay",
				CustomID: fmt.Sprintf("admin_prompt_page_replay_%d", hist.ID),
			},
		},
			discord.SeparatorComponent{},
		)
	}
	container.Components = append(container.Components,
		discord.ActionRowComponent{
			Components: []discord.InteractiveComponent{
				discord.ButtonComponent{
					CustomID: fmt.Sprintf("admin_prompt_page_first_%d", index),
					Label:    "First",
					Style:    discord.ButtonStyleSecondary,
				},
				discord.ButtonComponent{
					CustomID: fmt.Sprintf("admin_prompt_page_previous_%d", index),
					Label:    "Previous",
					Style:    discord.ButtonStylePrimary,
				},
				discord.ButtonComponent{
					CustomID: fmt.Sprintf("admin_prompt_page_next_%d", index),
					Label:    "Next",
					Style:    discord.ButtonStylePrimary,
				},
				discord.ButtonComponent{
					CustomID: fmt.Sprintf("admin_prompt_page_last_%d", index),
					Label:    "Last",
					Style:    discord.ButtonStyleSecondary,
				},
			},
		},
		discord.TextDisplayComponent{
			Content: fmt.Sprintf("Page: %d/%d", index, max),
		},
	)

	if len(history) == 0 {
		container.Components = []discord.ContainerSubComponent{
			discord.TextDisplayComponent{
				Content: "No History Yet",
			},
		}
	}
	components = append(components, container)
	return
}
