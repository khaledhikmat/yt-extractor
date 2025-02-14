package config

type IService interface {
	GetRuntimeEnvironment() string
	IsProduction() bool
	GetAPIPort() string
	IsOpenTelemetry() bool
	IsParseCodecEnabled() bool
	IsContineousExtraction() bool
	IsPeriodicExtraction() bool
	GetExtractionPeriod() int
	GetExtractionChannelID() string

	GetLocalCodecsFolder() string
	GetLocalVideosFolder() string
	GetLocalAudioFolder() string
	GetLocalTranscriptionFolder() string

	GetUpdatePeriod() string
	GetReattemptPeriod() string
	GetVideoTranscriptionCutoffDate() string

	GetTranscriptionProvider() string

	GetStorageProvider() string
	GetStorageBucket() string
	GetStorageRegion() string

	GetAWSAccessKeyID() string
	GetAWSSecretAccessKey() string
	GetAWSRegion() string

	GetDbDSN() string
	GetNeonDSN() string
	GetRailwayDSN() string
	GetYoutubeAPIKey() string
	GetOpenAIKey() string
	GetCloudConvertKey() string

	GetCloudConvertAttempts() int

	Finalize()
}
