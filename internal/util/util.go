package util

import (
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const (
	DISCORD_EMOJI_URL       = "https://cdn.discordapp.com/emojis/%s.%s"
	DiscordEpoch      int64 = 1420070400000
)

// Contains check slice contains want string
func Contains(slice []string, want string) bool {
	for _, element := range slice {
		if element == want {
			return true
		}
	}
	return false
}

// DeleteEmpty deleting empty strings in string slice
func DeleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// Elapsed timing time till function completion
func Elapsed(channel string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("Loading %s took %v to complete\n", channel, time.Since(start))
	}
}

// FilterDiscordMessages filtering specific messages out of message slice
func FilterDiscordMessages(messages []*discord.Message, condition func(*discord.Message) bool) (result []*discord.Message) {
	for _, message := range messages {
		if condition(message) {
			result = append(result, message)
		}
	}
	return result
}

// SnowflakeToTimestamp converts a Discord snowflake ID to a timestamp
func SnowflakeToTimestamp(snowflakeID string) (time.Time, error) {
	id, err := strconv.ParseInt(snowflakeID, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	timestamp := (id >> 22) + DiscordEpoch
	return time.Unix(0, timestamp*int64(time.Millisecond)), nil
}

// FetchDiscordEmojiImage fetches the raw image bytes for a given emoji ID and animation status.
func FetchDiscordEmojiImage(emojiID string, isAnimated bool) (string, error) {
	ext := "png"
	if isAnimated {
		ext = "gif"
	}
	url := fmt.Sprintf(DISCORD_EMOJI_URL, emojiID, ext)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch emoji from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}
	base64Data := base64.StdEncoding.EncodeToString(data)

	return base64Data, nil
}

func GetSeparator() discord.SeparatorComponent {
	divider := true

	return discord.SeparatorComponent{
		Divider: &divider,
		Spacing: discord.SeparatorSpacingSizeLarge,
	}
}


func IntToInt32Slice(input []int) []int32 {
	output := make([]int32, len(input))
	for i, v := range input {
		output[i] = int32(v)
	}
	return output
}

func Int32ToIntSlice(input []int32) []int {
	output := make([]int, len(input))
	for i, v := range input {
		output[i] = int(v)
	}
	return output
}

// UpdateInteractionResponse updates the interaction response for application command events.
func UpdateInteractionResponse(event *events.ApplicationCommandInteractionCreate, components []discord.LayoutComponent) {
	_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Components: &components,
		Flags:      ConfigFile.SetComponentV2Flags(),
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
	}
}

// RespondWithError builds a simple error component and updates the interaction response.
func RespondWithError(event *events.ApplicationCommandInteractionCreate, err error) {
	components := []discord.LayoutComponent{
		discord.TextDisplayComponent{Content: err.Error()},
	}
	UpdateInteractionResponse(event, components)
}

// UpdateComponentInteractionResponse updates the interaction response for component events.
func UpdateComponentInteractionResponse(event *events.ComponentInteractionCreate, components []discord.LayoutComponent) {
	_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Components: &components,
		Flags:      ConfigFile.SetComponentV2Flags(),
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err), slog.Any(". With body:", components))
	}
}

// RespondWithErrorComponent builds a simple error component and updates the component interaction response.
func RespondWithErrorComponent(event *events.ComponentInteractionCreate, err error) {
	components := []discord.LayoutComponent{
		discord.TextDisplayComponent{Content: err.Error()},
	}
	UpdateComponentInteractionResponse(event, components)
}

func BreakContent(content string, maxLength int) (result []string) {
	words := strings.Split(content, " ")

	var tmp string
	for i, word := range words {
		if i == 0 {
			tmp = word
		} else if len(tmp)+len(word) < maxLength {
			tmp += " " + word
		} else {
			result = append(result, tmp)
			tmp = word
		}
	}
	result = append(result, tmp)
	return result
}
