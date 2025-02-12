package config

type IService interface {
	GetRuntimeEnvironment() string
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

	GetVideoTranscriptionCutoffDate() string

	GetTranscriptionProvider() string

	GetStorageProvider() string
	GetStorageBucket() string
	GetStorageRegion() string

	GetAWSAccessKeyID() string
	GetAWSSecretAccessKey() string
	GetAWSRegion() string

	GetNeonDSN() string
	GetYoutubeAPIKey() string
	GetOpenAIKey() string
	GetCloudConvertKey() string

	Finalize()
}
