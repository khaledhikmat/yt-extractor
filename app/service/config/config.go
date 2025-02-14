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

func (svc *configService) IsProduction() bool {
	return svc.GetRuntimeEnvironment() == "prod"
}

func (svc *configService) GetAPIPort() string {
	return os.Getenv("API_PORT")
}

func (svc *configService) IsOpenTelemetry() bool {
	return os.Getenv("OPEN_TELEMETRY") == "true"
}

func (svc *configService) IsParseCodecEnabled() bool {
	return os.Getenv("PARSE_CODEC") == "true"
}

func (svc *configService) IsContineousExtraction() bool {
	return os.Getenv("CONTINEOUS_EXTRACTION") == "true"
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

func (svc *configService) GetLocalAudioFolder() string {
	return os.Getenv("LOCAL_AUDIO_FOLDER")
}

func (svc *configService) GetLocalTranscriptionFolder() string {
	return os.Getenv("LOCAL_TRANSCRIPTION_FOLDER")
}

func (svc *configService) GetUpdatePeriod() string {
	if os.Getenv("UPDATE_PERIOD") == "" {
		return "48 HOURS"
	}

	return os.Getenv("UPDATE_PERIOD")
}

func (svc *configService) GetReattemptPeriod() string {
	if os.Getenv("REATTEMPT_PERIOD") == "" {
		return "48 HOURS"
	}

	return os.Getenv("REATTEMPT_PERIOD")
}

func (svc *configService) GetVideoTranscriptionCutoffDate() string {
	return os.Getenv("VIDEO_TRANSCRIPTION_CUTOFF_DATE")
}

func (svc *configService) GetTranscriptionProvider() string {
	return os.Getenv("TRANSCRIPTION_PROVIDER")
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

func (svc *configService) GetAWSAccessKeyID() string {
	return os.Getenv("AWS_ACCESS_KEY_ID")
}

func (svc *configService) GetAWSSecretAccessKey() string {
	return os.Getenv("AWS_SECRET_ACCESS_KEY")
}

func (svc *configService) GetAWSRegion() string {
	return os.Getenv("AWS_REGION")
}

func (svc *configService) GetYoutubeAPIKey() string {
	return os.Getenv("YOUTUBE_API_KEY")
}

func (svc *configService) GetDbDSN() string {
	return os.Getenv(os.Getenv("DB_DSN"))
}

func (svc *configService) GetNeonDSN() string {
	return os.Getenv("NEON_DSN")
}

func (svc *configService) GetRailwayDSN() string {
	return os.Getenv("RAILWAY_DSN")
}

func (svc *configService) GetOpenAIKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

func (svc *configService) GetCloudConvertKey() string {
	return os.Getenv("CLOUDCONVERT_API_KEY")
}

func (svc *configService) GetCloudConvertAttempts() int {
	w, err := strconv.Atoi(os.Getenv("CLOUDCONVERT_ATTEMPTS"))
	if err != nil {
		return 180
	}

	return w
}

func (svc *configService) Finalize() {
}
