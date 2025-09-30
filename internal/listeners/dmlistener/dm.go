package dmlistener

import (
	"log"

	"github.com/disgoorg/disgo/events"
	ollamaApi "github.com/ollama/ollama/api"
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

func Listener(dm *events.DMMessageCreate) {
}
