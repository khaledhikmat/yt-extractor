The `yt-extractor` backend is written in Golang. It is a combination of API Endpoints and background processing.

## Go Module

```bash
go mod init github.com/khaledhikmat/yt-extractor
go get -u github.com/jmoiron/sqlx
go get -u github.com/lib/pq
go get -u github.com/gin-gonic/gin 
go get -u github.com/gin-contrib/cors
go get -u github.com/mdobak/go-xerrors
go get -u github.com/fatih/color
go get -u go.opentelemetry.io/otel
go get -u go.opentelemetry.io/contrib/exporters/autoexport
go get -u go.opentelemetry.io/contrib/propagators/autoprop
go get -u github.com/aws/aws-sdk-go
go get -u github.com/aws/aws-sdk-go-v2
go get -u github.com/aws/aws-sdk-go-v2/config
go get -u github.com/aws/aws-sdk-go-v2/service/s3
go get -u github.com/aws/aws-sdk-go-v2/feature/s3/manager
go get -u github.com/joho/godotenv
go get -u github.com/google/uuid
```

## Env Variables

| NAME           | DEFAULT | DESCRIPTION       |
|----------------|-----|------------------|
| YOUTUBE_API_KEY       | `youtube-api-key`  | Name of the microservice to appear in OTEL. |
| NEON_DSN       | `neon-postgres-db`  | HTTP Server port. Required to expose API Endpoints. |
| APP_NAME       | `none`  | Name of the microservice to appear in OTEL. |
| API_PORT       | `8080`  | HTTP Server port. Required to expose API Endpoints. |
| RUN_TIME_ENV  | `dev`  | Runetime env name.  |
| PARSE_CODEC  | `false`  | Whether to parse codec or not.  |
| CONTINEOUS_EXTRACTION | `false` | Whether to run a contineous extraction. Please see note below.|
| PERIODIC_EXTRACTION | `true` | Whether to run a periodic extraction |
| EXTRACTION_PERIOD | 15 | Number of minutes for extraction interval |
| EXTRACTION_CHANNEL_ID | `yt_channel_id` | Youtune channel ID to use for the periodic extraction |
| LOCAL_CODEC_FOLDER  | `codecs`  | folder  |
| LOCAL_VIDEOS_FOLDER  | `videos`  | folder  |
| LOCAL_AUDIO_FOLDER | `audio` | folder |
| VIDEO_TRANSCRIPTION_CUTOFF_DATE | `2025-01-01 00:00:00` | Denotes the video transcription cutoff date |
| STORAGE_PROVIDER | `s3` | |
| STORAGE_BUCKET | `yt-extractor-videos` | |
| STORAGE_REGION | `us-east-2` | |
| AWS_ACCESS_KEY_ID | `aws-access-key-id` | |
| AWS_ACCESS_SECRET_KEY_ID | `aws-access-secret-key-id` | |
| OPEN_TELEMETRY     | `false`  | If `true`, it disables collecting OTEL telemetry signals.   |
| OTEL_EXPORTER_OTLP_ENDPOINT     | `http://localhost:4318`  | OTEL endpoint.   |
| OTEL_SERVICE_NAME     | `yt-extractor-backend`  | OTEL application name.   |
| OTEL_GO_X_EXEMPLAR     | `true`  | OTEL GO.   |

**Please note** that, needless to mention, running the application in `CONTINEOUS_EXTRACTION` mode requires dedication as it is pretty intensive.

## Build and Push to Docker Hub

```bash
make push-2-hub
```


