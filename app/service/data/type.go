package data

type IService interface {
	ResetFactory() error

	NewVideo(video Video) (int64, error)
	UpdateVideo(video *Video, jobType JobType) error

	RetrieveVideos(channelID string, page, pageSize int, orderBy, orderDir string) ([]Video, error)
	RetrieveUnextractedVideos(channelID string, max int) ([]Video, error)
	RetrieveExtractErroredVideos(channelID string, max int) ([]Video, error)
	RetrieveUnexternalizedVideos(channelID string, max int) ([]Video, error)
	RetrieveUnaudioedVideos(channelID string, max int) ([]Video, error)
	RetrieveAudioErroredVideos(channelID string, max int) ([]Video, error)
	RetrieveUntranscribedVideos(channelID string, max int) ([]Video, error)
	RetrieveTranscribeErroredVideos(channelID string, max int) ([]Video, error)
	RetrieveUpdatedVideos(channelID string, max int) ([]Video, error)

	RetrieveVideoByIDs(channelID string, videoID string) (Video, error)
	RetrieveVideoByID(id int64) (Video, error)

	NewJob(job Job) (int64, error)
	UpdateJob(job *Job) error
	RetrieveJobByID(id int64) (Job, error)
	IsPendingJobsByTypeNChannel(channelID string, jobType JobType) (bool, error)

	NewAPIKey(key string) error
	IsAPIKeyValid(key string) (bool, error)
	NewError(source, body string) error

	Finalize()
}
