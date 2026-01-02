package threadlistener

import (
	"context"
	"database/sql"
	"log"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	ollamaApi "github.com/ollama/ollama/api"
	"github.com/stollenaar/ollamabot/internal/database"
	"github.com/stollenaar/ollamabot/internal/util"
)

var (
	OllamaClient *ollamaApi.Client
)

func init() {
	client, err := ollamaApi.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	OllamaClient = client
}

func Listener(event *events.GuildMessageCreate) {
	if event.Message.Author.ID == event.Client().ID() {
		return
	}

	thread, err := database.GetThread(event.ChannelID.String())

	if err != nil {
		if err == sql.ErrNoRows {
			return
		} else {
			slog.Error("Error fetching contexts:", slog.Any("err", err))
		}
	}

	OllamaClient.Generate(context.TODO(), &ollamaApi.GenerateRequest{
		Model:   thread.ModelName,
		System:  thread.Prompt,
		Prompt:  event.Message.Content,
		Stream:  new(bool),
		Context: util.Int32ToIntSlice(thread.Context),
	}, func(gr ollamaApi.GenerateResponse) error {
		_, err := event.Client().Rest.CreateMessage(event.ChannelID, discord.MessageCreate{
			MessageReference: &discord.MessageReference{
				MessageID: &event.MessageID,
				ChannelID: &event.ChannelID,
				GuildID:   &event.GuildID,
			},
			Content: gr.Response,
		})

		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", gr.Response))
		}

		err = database.UpdateThreadContext(thread.ThreadID, util.IntToInt32Slice(gr.Context))
		if err != nil {
			slog.Error("Error updating context:", slog.Any("err", err))
		}
		return err
	})
}
