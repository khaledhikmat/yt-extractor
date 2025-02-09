package youtube

import "context"

type IService interface {
	PrintExtractorVersion() error
	RetrieveVideos(channelID string, max int) ([]Video, error)
	ExtractVideos(ctx context.Context, errorStream chan error, videoURLs []string) (map[string]string, error)

	Finalize()
}
