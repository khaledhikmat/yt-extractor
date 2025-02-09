package audio

type IService interface {
	SplitAudio(audioURL string) ([]string, error)

	Finalize()
}
