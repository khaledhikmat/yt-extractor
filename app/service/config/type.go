package config

type IService interface {
	GetRuntimeEnvironment() string
	GetAPIPort() string
	GetYoutubeAPIKey() string
	GetNeonDSN() string
	IsOpenTelemetry() bool
	IsParseCodecEnabled() bool
	IsPeriodicExtraction() bool
	GetExtractionPeriod() int
	GetExtractionChannelID() string

	GetLocalCodecsFolder() string
	GetLocalVideosFolder() string

	GetStorageProvider() string
	GetStorageBucket() string
	GetStorageRegion() string

	Finalize()
}
