package transcription

type IService interface {
	TranscribeAudio(audioFilePath string) (string, error)
}
