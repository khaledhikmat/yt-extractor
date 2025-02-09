package config

import (
	"os"
	"strconv"
)

type configService struct {
}

func New() IService {
	return &configService{}
}

func (svc *configService) GetRuntimeEnvironment() string {
	if os.Getenv("RUN_TIME_ENV") == "" {
		return "dev"
	}

	return os.Getenv("RUN_TIME_ENV")
}

func (svc *configService) GetAPIPort() string {
	return os.Getenv("API_PORT")
}

func (svc *configService) GetYoutubeAPIKey() string {
	return os.Getenv("YOUTUBE_API_KEY")
}

func (svc *configService) GetNeonDSN() string {
	return os.Getenv("NEON_DSN")
}

func (svc *configService) IsOpenTelemetry() bool {
	return os.Getenv("OPEN_TELEMETRY") == "true"
}

func (svc *configService) IsParseCodecEnabled() bool {
	return os.Getenv("PARSE_CODEC") == "true"
}

func (svc *configService) IsPeriodicExtraction() bool {
	return os.Getenv("PERIODIC_EXTRACTION") == "true"
}

func (svc *configService) GetExtractionPeriod() int {
	w, err := strconv.Atoi(os.Getenv("EXTRACTION_PERIOD"))
	if err != nil {
		return 15
	}

	return w
}

func (svc *configService) GetExtractionChannelID() string {
	return os.Getenv("EXTRACTION_CHANNEL_ID")
}

func (svc *configService) GetLocalCodecsFolder() string {
	return os.Getenv("LOCAL_CODECS_FOLDER")
}

func (svc *configService) GetLocalVideosFolder() string {
	return os.Getenv("LOCAL_VIDEOS_FOLDER")
}

func (svc *configService) GetStorageProvider() string {
	return os.Getenv("STORAGE_PROVIDER")
}

func (svc *configService) GetStorageBucket() string {
	return os.Getenv("STORAGE_BUCKET")
}

func (svc *configService) GetStorageRegion() string {
	return os.Getenv("STORAGE_REGION")
}

func (svc *configService) Finalize() {
}
