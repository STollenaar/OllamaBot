module github.com/stollenaar/ollamabot/internal/listeners/dmlistener

go 1.25

replace (
	github.com/stollenaar/ollamabot/internal/database => ../../database
	github.com/stollenaar/ollamabot/internal/util => ../../util
)

require (
	github.com/disgoorg/disgo v0.18.16
	github.com/ollama/ollama v0.12.3
)

require (
	github.com/disgoorg/json v1.2.0 // indirect
	github.com/disgoorg/snowflake/v2 v2.0.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/sasha-s/go-csync v0.0.0-20240107134140-fcbab37b09ad // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
)
