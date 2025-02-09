package audio

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
)

type audioService struct {
	ConfigSvc config.IService
}

func New(cfgsvc config.IService) IService {
	return &audioService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *audioService) SplitAudio(URL string) ([]string, error) {
	lgr.Logger.Debug("Split audio",
		slog.String("URL", URL),
	)
	files := []string{}
	jobID := uuid.New().String()

	// Split audio into 10 minute segments
	outputPattern := fmt.Sprintf("./audio/%s_%%03d.mp3", jobID)
	cmd := exec.Command("ffmpeg", "-i", URL, "-f", "segment", "-segment_time", "600", "-c", "copy", outputPattern)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return files, fmt.Errorf("error executing ffmpeg: %v", err)
	}

	// Collect the generated file names
	pattern := fmt.Sprintf("./audio/%s_*.mp3", jobID)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return files, fmt.Errorf("error collecting generated files: %v", err)
	}

	files = append(files, matches...)

	return files, nil
}

func (svc *audioService) Finalize() {
}
