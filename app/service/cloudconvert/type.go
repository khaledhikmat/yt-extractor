package cloudconvert

type IService interface {
	ConvertVideoToAudio(channelID, videoID string) (string, error)
}
