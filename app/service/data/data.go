package data

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Import the PostgreSQL driver

	"github.com/khaledhikmat/yt-extractor/service"
	"github.com/khaledhikmat/yt-extractor/service/config"
)

var mutex = &sync.Mutex{}

//go:embed sql/reset_factory.sql
var resetfactorySQL string

//go:embed sql/insertvideo.sql
var insertVideoSQL string

//go:embed sql/updatevideo_ytattributes.sql
var updateytattributesSQL string

//go:embed sql/updatevideo_ytexternalization.sql
var updateytexternalizationSQL string

//go:embed sql/updatevideo_ytextraction.sql
var updateytextractionSQL string

//go:embed sql/updatevideo_yterroredextraction.sql
var updateyterroredextractionSQL string

//go:embed sql/updatevideo_yttranscription.sql
var updateyttranscriptionSQL string

//go:embed sql/updatevideo_ytprocessing.sql
var updateytprocessingSQL string

//go:embed sql/insertjob.sql
var insertjobSQL string

//go:embed sql/updatejob.sql
var updatejobSQL string

//go:embed sql/insertapikey.sql
var insertapikeySQL string

//go:embed sql/inserterror.sql
var inserterrorSQL string

type dataService struct {
	ConfigSvc config.IService
	Db        *sqlx.DB
}

func New(cfgsvc config.IService) IService {
	return &dataService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *dataService) ResetFactory() error {
	err := svc.dbConnection()
	if err != nil {
		return err
	}

	_, err = svc.Db.Exec(resetfactorySQL)
	if err != nil {
		return err
	}

	return nil
}

func (svc *dataService) NewVideo(video Video) (int64, error) {
	err := svc.dbConnection()
	if err != nil {
		return -1, err
	}

	// Make sure that video does not already exist
	vid, err := svc.RetrieveVideoByIDs(video.ChannelID, video.VideoID)
	if err != nil {
		return -1, fmt.Errorf("Error fetching video by ID: %v", err)
	}

	// If the video already exists, switch to update the attributes
	if vid.VideoID != "" {
		err = svc.UpdateVideo(&video, JobTypeAttributes)
		return vid.ID, err
	}

	// Execute the insert query using NamedExec or NamedQuery
	rows, err := svc.Db.NamedQuery(insertVideoSQL, video)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	// Fetch the newly inserted ID if needed
	if rows.Next() {
		err = rows.Scan(&video.ID)
		if err != nil {
			return -1, err
		}
	}

	return video.ID, nil
}

func (svc *dataService) UpdateVideo(video *Video, jobType JobType) error {
	err := svc.dbConnection()
	if err != nil {
		return err
	}

	// Make sure the video does exist
	vid, err := svc.RetrieveVideoByIDs(video.ChannelID, video.VideoID)
	if err != nil {
		return fmt.Errorf("Error fetching video by ID: %v", err)
	}

	if vid.VideoID == "" {
		return fmt.Errorf("Video ID %s does not exist", vid.VideoID)
	}

	// We have several video update types:
	if jobType == JobTypeAttributes {
		if vid.Views == video.Views && vid.Comments == video.Comments && vid.Likes == video.Likes {
			// Skip the update if the attributes are the same
			return nil
		}
		_, err = svc.Db.Exec(updateytattributesSQL, video.Views, video.Comments, video.Likes, video.ID)
	} else if jobType == JobTypeExternalization {
		_, err = svc.Db.Exec(updateytexternalizationSQL, video.ID)
	} else if jobType == JobTypeExtraction {
		_, err = svc.Db.Exec(updateytextractionSQL, video.ExtractionURL, video.ID)
	} else if jobType == JobTypeErroredExtraction {
		_, err = svc.Db.Exec(updateyterroredextractionSQL, video.ExtractionURL, video.ID)
	} else if jobType == JobTypeTranscription {
		_, err = svc.Db.Exec(updateyttranscriptionSQL, video.AudioURL, video.TranscriptionURL, video.ID)
	} else if jobType == JobTypeProcessing {
		_, err = svc.Db.Exec(updateytprocessingSQL, video.ID)
	} else {
		return fmt.Errorf("Invalid job type %s", jobType)
	}

	if err != nil {
		return err
	}

	return nil
}

func (svc *dataService) RetrieveVideos(channelID string, page, pageSize int, orderBy, orderDir string) ([]Video, error) {
	videos := []Video{}
	err := svc.dbConnection()
	if err != nil {
		return videos, err
	}

	if page < 1 {
		return videos, fmt.Errorf("Invalid page number %d", page)
	}

	if pageSize <= 0 {
		return videos, fmt.Errorf("Invalid page size %d", pageSize)
	}

	if orderBy != "published_at" &&
		orderBy != "views" &&
		orderBy != "comments" &&
		orderBy != "likes" {
		return videos, fmt.Errorf("Invalid order by %s", orderBy)
	}

	if orderDir != "asc" && orderDir != "desc" {
		return videos, fmt.Errorf("Invalid order direction %s", orderDir)
	}

	// Calculate the offset
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
        SELECT * 
		FROM videos 
		WHERE channel_id = $1 
		ORDER BY %s %s 
		LIMIT $2 OFFSET $3 
    `, orderBy, orderDir)

	err = svc.Db.Select(&videos, query, channelID, pageSize, offset)
	if err != nil {
		return videos, err
	}

	return videos, nil
}

func (svc *dataService) RetrieveVideoByIDs(channelID string, videoID string) (Video, error) {
	err := svc.dbConnection()
	if err != nil {
		return Video{}, err
	}

	var videos []Video
	query := `
        SELECT * FROM videos 
		WHERE channel_id = $1 
		AND video_id = $2 
		LIMIT 1
    `

	err = svc.Db.Select(&videos, query, channelID, videoID)
	if err != nil {
		return Video{}, err
	}

	if len(videos) == 0 {
		return Video{}, nil
	}

	return videos[0], nil
}

func (svc *dataService) RetrieveVideoByID(id int64) (Video, error) {
	err := svc.dbConnection()
	if err != nil {
		return Video{}, err
	}

	var videos []Video
	query := `
        SELECT * FROM videos 
		WHERE id = $1 
		LIMIT 1
    `

	err = svc.Db.Select(&videos, query, id)
	if err != nil {
		return Video{}, err
	}

	if len(videos) == 0 {
		return Video{}, fmt.Errorf("Video ID %d does not exist", id)
	}

	return videos[0], nil
}

func (svc *dataService) RetrieveUnextractedVideos(channelID string, max int) ([]Video, error) {
	videos := []Video{}
	err := svc.dbConnection()
	if err != nil {
		return videos, err
	}

	query := `
        SELECT * FROM videos 
		WHERE channel_id = $1 
		AND extracted_at is null 
		ORDER BY published_at DESC 
		LIMIT $2 
    `

	err = svc.Db.Select(&videos, query, channelID, max)
	if err != nil {
		return videos, err
	}

	return videos, nil
}

func (svc *dataService) RetrieveErroredVideos(channelID string, max int) ([]Video, error) {
	videos := []Video{}
	err := svc.dbConnection()
	if err != nil {
		return videos, err
	}

	// The errored videos have not-available extraction-URL
	// and it has not been more than 24 hours since the extraction.
	// The last clause is to prevent the errored videos from being picked up perpetually.
	query := `
        SELECT * FROM videos 
		WHERE channel_id = $1 
		AND extracted_at is not null 
		AND extraction_url = $2
		AND extracted_at >= NOW() - INTERVAL '24 HOURS'
		ORDER BY published_at DESC 
		LIMIT $3 
    `

	err = svc.Db.Select(&videos, query, channelID, service.UnextractedVideoURL, max)
	if err != nil {
		return videos, err
	}

	return videos, nil
}

func (svc *dataService) RetrieveUnexternalizedVideos(channelID string, max int) ([]Video, error) {
	videos := []Video{}
	err := svc.dbConnection()
	if err != nil {
		return videos, err
	}

	query := `
        SELECT * FROM videos 
		WHERE channel_id = $1 
		AND extracted_at is not null 
		AND externalized_at is null 
		ORDER BY published_at DESC 
		LIMIT $2 
    `

	err = svc.Db.Select(&videos, query, channelID, max)
	if err != nil {
		return videos, err
	}

	return videos, nil
}

// Used for transcription within the backend
func (svc *dataService) RetrieveUntranscribedVideos(channelID string, max int) ([]Video, error) {
	videos := []Video{}
	err := svc.dbConnection()
	if err != nil {
		return videos, err
	}

	// Prevent unprocessed query to pick up errored extractions
	query := fmt.Sprintf(`
        SELECT * FROM videos 
		WHERE channel_id = $1 
		AND externalized_at is not null 
		AND extracted_at is not null 
		AND extraction_url != '%s' 
		AND processed_at is null 
		AND published_at >= '%s'
		ORDER BY published_at DESC 
		LIMIT $2 
    `, service.UnextractedVideoURL, svc.ConfigSvc.GetVideoTranscriptionCutoffDate())

	err = svc.Db.Select(&videos, query, channelID, max)
	if err != nil {
		return videos, err
	}

	return videos, nil
}

// Used for transcription within the external automation tool
// Abandoned
func (svc *dataService) RetrieveUnprocessedVideos(channelID string, max int) ([]Video, error) {
	videos := []Video{}
	err := svc.dbConnection()
	if err != nil {
		return videos, err
	}

	// Prevent unprocessed query to pick up errored extractions
	// Restrict transcription processing to cutoff datetime
	query := fmt.Sprintf(`
        SELECT * FROM videos 
		WHERE channel_id = $1 
		AND externalized_at is not null 
		AND extracted_at is not null 
		AND extraction_url != '%s' 
		AND processed_at is null 
		AND published_at >= '%s'
		ORDER BY published_at DESC 
		LIMIT $2 
    `, service.UnextractedVideoURL, svc.ConfigSvc.GetVideoTranscriptionCutoffDate())

	err = svc.Db.Select(&videos, query, channelID, max)
	if err != nil {
		return videos, err
	}

	return videos, nil
}

func (svc *dataService) RetrieveUpdatedVideos(channelID string, max int) ([]Video, error) {
	videos := []Video{}
	err := svc.dbConnection()
	if err != nil {
		return videos, err
	}

	query := `
        SELECT * FROM videos 
		WHERE channel_id = $1 
		AND extracted_at is not null 
		AND externalized_at is not null 
		AND processed_at is not null 
		AND updated_at >= NOW() - INTERVAL '24 HOURS'
		ORDER BY published_at DESC 
		LIMIT $2 
    `

	err = svc.Db.Select(&videos, query, channelID, max)
	if err != nil {
		return videos, err
	}

	return videos, nil
}

func (svc *dataService) NewJob(job Job) (int64, error) {
	err := svc.dbConnection()
	if err != nil {
		return -1, err
	}

	// Execute the insert query using NamedExec or NamedQuery
	rows, err := svc.Db.NamedQuery(insertjobSQL, job)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	// Fetch the newly inserted ID if needed
	if rows.Next() {
		err = rows.Scan(&job.ID)
		if err != nil {
			return -1, err
		}
	}

	return job.ID, nil
}

func (svc *dataService) UpdateJob(job *Job) error {
	err := svc.dbConnection()
	if err != nil {
		return err
	}

	_, err = svc.Db.Exec(updatejobSQL, job.State, job.Videos, job.Errors, job.CompletedAt, job.ID)
	if err != nil {
		return err
	}

	return nil
}

func (svc *dataService) RetrieveJobByID(id int64) (Job, error) {
	err := svc.dbConnection()
	if err != nil {
		return Job{}, err
	}

	var jobs []Job
	query := `
        SELECT * FROM jobs 
		WHERE id = $1 
		LIMIT 1
    `

	err = svc.Db.Select(&jobs, query, id)
	if err != nil {
		return Job{}, err
	}

	if len(jobs) == 0 {
		return Job{}, fmt.Errorf("Job ID %d does not exist", id)
	}

	return jobs[0], nil
}

func (svc *dataService) IsPendingJobsByTypeNChannel(channelID string, jobType JobType) (bool, error) {
	err := svc.dbConnection()
	if err != nil {
		return false, err
	}

	var jobs []Job
	query := `
        SELECT * FROM jobs 
		WHERE  channel_id = $1
		AND type = $2
		AND state IN ($3, $4) 
		LIMIT 1
    `

	err = svc.Db.Select(&jobs, query, channelID, jobType, JobStateQueued, JobStateRunning)
	if err != nil {
		return false, err
	}

	return len(jobs) > 0, nil
}

func (svc *dataService) NewAPIKey(key string) error {
	err := svc.dbConnection()
	if err != nil {
		return err
	}

	// Execute the insert query using NamedExec or NamedQuery
	_, err = svc.Db.Exec(insertapikeySQL, key)
	if err != nil {
		return err
	}

	return nil
}

func (svc *dataService) IsAPIKeyValid(key string) (bool, error) {
	err := svc.dbConnection()
	if err != nil {
		return false, err
	}

	var keys []string
	query := `
        SELECT key 
		FROM api_keys 
		WHERE key = $1
		AND expires_at > now() 
		LIMIT 1
    `

	err = svc.Db.Select(&keys, query, key)
	if err != nil {
		return false, err
	}

	if len(keys) == 0 {
		return false, fmt.Errorf("APP KEY %s is not valid", key)
	}

	return true, nil
}

func (svc *dataService) NewError(source, body string) error {
	err := svc.dbConnection()
	if err != nil {
		return err
	}

	// Execute the insert query using NamedExec or NamedQuery
	_, err = svc.Db.Exec(inserterrorSQL, source, body)
	if err != nil {
		return err
	}

	return nil
}

func (svc *dataService) Finalize() {
	if svc.Db != nil {
		svc.Db.Close()
	}
}

func (svc *dataService) dbConnection() error {
	var err error
	if svc.Db != nil {
		return nil
	}

	// Allow one thread to access the database at a time
	mutex.Lock()
	defer mutex.Unlock()

	svc.Db, err = sqlx.Connect("postgres", svc.ConfigSvc.GetNeonDSN())
	if err != nil {
		return err
	}

	return nil
}
