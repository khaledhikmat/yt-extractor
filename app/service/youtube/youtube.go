package youtube

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/khaledhikmat/yt-extractor/service"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
	"github.com/khaledhikmat/yt-extractor/utils"
)

const (
	defaultCodecIDs = "137+140"
)

type youtubService struct {
	ConfigSvc config.IService
}

func New(cfgsvc config.IService) IService {
	return &youtubService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *youtubService) PrintExtractorVersion() error {
	lgr.Logger.Debug("Checking yt-dlp and ffmpeg versions")

	cmd := exec.Command("yt-dlp", "--version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing yt-dlp: %v", err)
	}

	cmd = exec.Command("ffmpeg", "-version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing ffmpeg: %v", err)
	}

	return nil
}

func (svc *youtubService) RetrieveVideos(channelID string, max int) ([]Video, error) {
	var results []Video

	// Every Youtube channel has a main play list ID that stores all videos!!
	playlistID, err := getUploadsPlaylistID(svc.ConfigSvc.GetYoutubeAPIKey(), channelID)
	if err != nil {
		return results, err
	}

	results, err = getVideosFromPlaylist(svc.ConfigSvc.GetYoutubeAPIKey(), playlistID, max)
	if err != nil {
		return results, err
	}

	return results, nil
}

func (svc *youtubService) ExtractVideos(ctx context.Context, errorStream chan error, videoURLs []string) (map[string]string, error) {
	results := map[string]string{}

	for _, URL := range videoURLs {
		// If the context is cancelled, exit the loop
		select {
		case <-ctx.Done():
			return results, fmt.Errorf("context cancelled")
		default:
		}

		lgr.Logger.Debug("Extracting video",
			slog.String("URL", URL),
		)

		// Run yt-dlp to extract the video and save it to an output file
		mp4File, err := runYTDLPExtractor(extractVideoID(URL), URL, svc.ConfigSvc.GetLocalVideosFolder(), svc.ConfigSvc.GetLocalCodecsFolder(), svc.ConfigSvc.IsProduction())
		if err != nil {
			errorStream <- fmt.Errorf("error extracting video %s: %v", URL, err)
			// Indicate an error by mapping the video URL to not available extraction URL
			results[URL] = service.InvalidURL
			continue
		}

		results[URL] = mp4File
	}

	return results, nil
}

func (svc *youtubService) Finalize() {
}

// PRIVATE

func getUploadsPlaylistID(apiKey, channelID string) (string, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=contentDetails&id=%s&key=%s", channelID, apiKey)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ContentDetails struct {
				RelatedPlaylists struct {
					Uploads string `json:"uploads"`
				} `json:"relatedPlaylists"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "", fmt.Errorf("no channel found with the provided ID")
	}

	return result.Items[0].ContentDetails.RelatedPlaylists.Uploads, nil
}

// getVideosFromPlaylist retrieves all videos from a playlist with pagination
func getVideosFromPlaylist(apiKey, playlistID string, maxVideos int) ([]Video, error) {
	var videos []Video
	nextPageToken := ""

	// WARNING: If maxVideos is -1, it implies that we want to fetch all videos in the playlist

	for {
		// Construct API request URL
		apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=50&playlistId=%s&key=%s&pageToken=%s",
			playlistID, apiKey, nextPageToken)

		if maxVideos > 0 && maxVideos <= 50 {
			apiURL = fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=%d&playlistId=%s&key=%s&pageToken=%s",
				maxVideos, playlistID, apiKey, nextPageToken)
		}

		// Make the HTTP request
		resp, err := http.Get(apiURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Decode the response
		var playlistResponse PlaylistItemsResponse
		if err := json.NewDecoder(resp.Body).Decode(&playlistResponse); err != nil {
			return nil, err
		}

		// Process each video in the response
		pageVideos := []Video{}
		for _, item := range playlistResponse.Items {
			video := Video{
				Title:       item.Snippet.Title,
				PublishedAt: item.Snippet.PublishedAt,
				URL:         fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Snippet.ResourceID.VideoID),
			}
			pageVideos = append(pageVideos, video)
		}

		// Get video stats
		videoIDs := []string{}
		for _, video := range pageVideos {
			videoIDs = append(videoIDs, extractVideoID(video.URL))
		}

		statistics, err := getVideoStatistics(apiKey, videoIDs)
		if err != nil {
			return videos, err
		}

		// Get video details
		for _, video := range pageVideos {
			video.ID = extractVideoID(video.URL)
			video.Views = statistics[extractVideoID(video.URL)].Views
			video.Comments = statistics[extractVideoID(video.URL)].Comments
			video.Likes = statistics[extractVideoID(video.URL)].Likes
			video.Duration = statistics[extractVideoID(video.URL)].Duration
			video.Short = statistics[extractVideoID(video.URL)].Short

			videos = append(videos, video)
		}

		// Check if there is a next page token
		if playlistResponse.NextPageToken == "" ||
			(len(videos) >= maxVideos && maxVideos > 0) {
			break
		}
		nextPageToken = playlistResponse.NextPageToken
	}

	return videos, nil
}

func extractVideoID(url string) string {
	// Example: https://www.youtube.com/watch?v=abc123
	if strings.Contains(url, "watch?v=") {
		parts := strings.Split(url, "watch?v=")
		return strings.Split(parts[1], "&")[0] // Extract ID before any extra query params
	}
	return ""
}

func getVideoStatistics(apiKey string, videoIDs []string) (map[string]Video, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=contentDetails,statistics&id=%s&key=%s", url.QueryEscape(strings.Join(videoIDs, ",")), apiKey)
	//fmt.Printf("Statistics API URL: %s\n", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var statsResponse VideoStatisticsResponse
	// if err := json.NewDecoder(resp.Body).Decode(&statsResponse); err != nil {
	// 	return nil, err
	// }
	err = json.Unmarshal(body, &statsResponse)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]Video)
	for _, item := range statsResponse.Items {
		stats[item.ID] = Video{
			Views:    item.Statistics.ViewCount,
			Comments: item.Statistics.CommentCount,
			Likes:    item.Statistics.LikeCount,
			Duration: item.ContentDetails.Duration,
			Short:    utils.ExtractDurationInSecs(item.ContentDetails.Duration) < 100,
		}
	}

	return stats, nil
}

func extractFirstID(line string) string {
	// Define a regular expression to match the first number at the beginning of the line
	re := regexp.MustCompile(`^\s*(\d+)\s+`)

	// Find the first match
	match := re.FindStringSubmatch(line)
	if len(match) > 1 {
		return match[1] // Return the captured group (the number)
	}

	return "" // Return an empty string if no match is found
}

func scrapeCodecIDs(videoURL, codecsFolder string) (string, error) {
	codecIDs := ""

	// Generate codec file
	codecFile := fmt.Sprintf("./%s/%s.txt", codecsFolder, extractVideoID(videoURL))

	// Run yt-dlp to fetch available formats
	err := runYTDLPWithOutput(videoURL, codecFile)
	if err != nil {
		return "", err
	}

	defer os.Remove(codecFile)

	lgr.Logger.Debug("Produced codec file",
		slog.String("file", codecFile),
	)

	// Process the codec file to produce codec IDs
	codecIDs, err = processCodecFile(codecFile)
	if err != nil {
		return "", err
	}

	return codecIDs, nil
}

func processCodecFile(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Define slices to hold the matching lines
	var audioLines []string
	var videoLines []string

	// Regular expressions to match the desired conditions
	//audioRegex := regexp.MustCompile(`\s+m4a\s+audio only\s+`)
	//videoRegex := regexp.MustCompile(`\s+mp4\s+\S+\s+\d+p\s+.*video only.*https`)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for audio-only match
		// if audioRegex.MatchString(line) {
		// 	audioLines = append(audioLines, line)
		// }

		if strings.Contains(line, "m4a") &&
			strings.Contains(line, "audio only") {
			audioLines = append(audioLines, line)
		}

		// Check for video-only match
		// if videoRegex.MatchString(line) {
		// 	videoLines = append(videoLines, line)
		// }

		if strings.Contains(line, "mp4") &&
			strings.Contains(line, "https") &&
			strings.Contains(line, "video only") {
			videoLines = append(videoLines, line)
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("Error reading file: %v", err)
	}

	return extractFirstID(videoLines[len(videoLines)-1]) + "+" + extractFirstID(audioLines[len(audioLines)-1]), nil
}

func runYTDLPWithOutput(videoURL, outputFile string) error {
	// Define video URL and output file name
	// videoURL := "https://www.youtube.com/watch?v=wQlek65Hp2w"
	// outputFile := "wQlek65Hp2w.txt"

	// Create or open the output file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// Construct the command
	cmd := exec.Command("yt-dlp", "-F", videoURL)

	// Redirect the command's output to the file
	cmd.Stdout = file
	cmd.Stderr = os.Stderr // Optional: log errors to stderr

	// Run the command
	fmt.Println("Running yt-dlp to fetch available formats...")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing yt-dlp: %v", err)
	}

	return nil
}

func runYTDLPExtractor(videoID, videoURL, videosFolder, codecsFolder string, isProd bool) (string, error) {
	outputFile := fmt.Sprintf("./%s/%s.mp4", videosFolder, videoID)

	// Two attempts to extract the video:
	// 1. first with the default codec IDs,
	// 2. with scraped codec IDs

	scrapedCodecIDs, err := scrapeCodecIDs(videoURL, codecsFolder)
	if err != nil {
		scrapedCodecIDs = ""
	}

	lgr.Logger.Debug("runYTDLPExtractor",
		slog.Bool("Prod", isProd),
		slog.String("videoID", videoID),
		slog.String("videoURL", videoURL),
	)

	codecIDsList := []string{defaultCodecIDs, scrapedCodecIDs}
	for _, codecIDs := range codecIDsList {
		if codecIDs == "" {
			continue
		}

		// Construct the extract command using default codec IDs
		var cmd *exec.Cmd
		if isProd {
			// In production running in Docker, we must spoof headers and user agent to prevent
			// triggering YouTube's anti-bot measures
			// Option1: Try using a realistic browser User-Agent to make requests look more like a human browsing.
			// userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
			// lgr.Logger.Debug("runYTDLPExtractor",
			// 	slog.String("userAgent", userAgent),
			// )
			// cmd = exec.Command("yt-dlp", "--user-agent", userAgent, "-f", codecIDs, "--merge-output-format", "mp4", videoURL, "-o", outputFile)
			// Option2: Documented in https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp
			lgr.Logger.Debug("runYTDLPExtractor",
				slog.String("cookies", "./cookies.txt"),
			)
			cmd = exec.Command("yt-dlp", "--cookies", "./cookies.txt", "-f", codecIDs, "--merge-output-format", "mp4", videoURL, "-o", outputFile)
		} else {
			cmd = exec.Command("yt-dlp", "-f", codecIDs, "--merge-output-format", "mp4", videoURL, "-o", outputFile)
		}

		// Set command output to the standard output (for debugging/logging)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the command
		fmt.Println("Running yt-dlp to extract video...")
		err := cmd.Run()
		if err == nil {
			return outputFile, nil
		}
	}

	return "", fmt.Errorf("error executing yt-dlp: %v", err)
}
