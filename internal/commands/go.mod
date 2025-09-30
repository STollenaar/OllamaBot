module github.com/stollenaar/ollamabot/internal/commands

go 1.25.0

require (
	github.com/disgoorg/disgo v0.19.0-rc.6
	github.com/stollenaar/ollamabot/internal/commands/admincommand v0.0.0-20250929215636-9407d818cd52
	github.com/stollenaar/ollamabot/internal/commands/listcommand v0.0.0-20250929215636-9407d818cd52
	github.com/stollenaar/ollamabot/internal/commands/promptcommand v0.0.0-20250929215636-9407d818cd52
	github.com/stollenaar/ollamabot/internal/util v0.0.0-20250929215636-9407d818cd52
)

require (
	github.com/apache/arrow-go/v18 v18.4.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.39.2 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.12 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.16 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.65.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.6 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/disgoorg/json/v2 v2.0.0 // indirect
	github.com/disgoorg/omit v1.0.0 // indirect
	github.com/disgoorg/snowflake/v2 v2.0.3 // indirect
	github.com/duckdb/duckdb-go-bindings v0.1.20 // indirect
	github.com/duckdb/duckdb-go-bindings/darwin-amd64 v0.1.20 // indirect
	github.com/duckdb/duckdb-go-bindings/darwin-arm64 v0.1.20 // indirect
	github.com/duckdb/duckdb-go-bindings/linux-amd64 v0.1.20 // indirect
	github.com/duckdb/duckdb-go-bindings/linux-arm64 v0.1.20 // indirect
	github.com/duckdb/duckdb-go-bindings/windows-amd64 v0.1.20 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/encoding/ini v0.1.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/flatbuffers v25.9.23+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/marcboeker/go-duckdb/arrowmapping v0.0.20 // indirect
	github.com/marcboeker/go-duckdb/mapping v0.0.20 // indirect
	github.com/marcboeker/go-duckdb/v2 v2.4.1 // indirect
	github.com/ollama/ollama v0.12.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/sasha-s/go-csync v0.0.0-20240107134140-fcbab37b09ad // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/stollenaar/aws-rotating-credentials-provider/credentials v0.0.0-20250330204128-299effe6093c // indirect
	github.com/stollenaar/ollamabot/internal/database v0.0.0-20250929215636-9407d818cd52 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/exp v0.0.0-20250911091902-df9299821621 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/telemetry v0.0.0-20250908211612-aef8a434d053 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

replace (
	github.com/stollenaar/ollamabot/internal/commands/admincommand => ./admincommand
	github.com/stollenaar/ollamabot/internal/commands/listcommand => ./listcommand
	github.com/stollenaar/ollamabot/internal/commands/promptcommand => ./promptcommand
	github.com/stollenaar/ollamabot/internal/database => ../database
	github.com/stollenaar/ollamabot/internal/util => ../util
)
