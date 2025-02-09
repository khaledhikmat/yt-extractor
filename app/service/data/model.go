package data

import "time"

type Video struct {
	ID             int64      `json:"id" db:"id"`
	ChannelID      string     `json:"channelId" db:"channel_id"`
	VideoID        string     `json:"videoId" db:"video_id"`
	VideoURL       string     `json:"videoUrl" db:"video_url"`
	Title          string     `json:"title" db:"title"`
	PublishedAt    time.Time  `json:"publishedAt" db:"published_at"`
	Views          int64      `json:"views" db:"views"`
	Comments       int64      `json:"comments" db:"comments"`
	Likes          int64      `json:"likes" db:"likes"`
	Duration       int64      `json:"duration" db:"duration"`
	Short          bool       `json:"short" db:"short"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
	ExtractionURL  *string    `json:"extractionUrl" db:"extraction_url"`
	ExtractedAt    *time.Time `json:"extractedAt" db:"extracted_at"`
	ExternalizedAt *time.Time `json:"externalizedAt" db:"externalized_at"`
	ProcessedAt    *time.Time `json:"processedAt" db:"processed_at"`
}

type JobState string

const (
	JobStateQueued    JobState = "queued"
	JobStateRunning   JobState = "running"
	JobStateCancelled JobState = "cancelled"
	JobStateCompleted JobState = "completed"
)

type JobType string

const (
	JobTypeAttributes        JobType = "attributes"
	JobTypeExtraction        JobType = "extraction"
	JobTypeErroredExtraction JobType = "erroredextraction"
	JobTypeProcessing        JobType = "processing"
	JobTypeExternalization   JobType = "externalization"
)

type Job struct {
	ID          int64      `json:"id" db:"id"`
	ChannelID   string     `json:"channelId" db:"channel_id"`
	Type        JobType    `json:"type" db:"type"`
	State       JobState   `json:"state" db:"state"`
	Videos      int64      `json:"videos" db:"videos"`
	Errors      int64      `json:"errors" db:"errors"`
	StartedAt   time.Time  `json:"startedAt" db:"started_at"`
	CompletedAt *time.Time `json:"completedAt" db:"completed_at"`
}

type JobVideo struct {
	ID          int64      `json:"id" db:"id"`
	JobID       int64      `json:"jobId" db:"job_id"`
	ChannelID   string     `json:"channelId" db:"channel_id"`
	VideolID    string     `json:"videoId" db:"video_id"`
	StartedAt   time.Time  `json:"startedAt" db:"started_at"`
	Error       *string    `json:"error" db:"error"`
	CompletedAt *time.Time `json:"completedAt" db:"completed_at"`
}

type Error struct {
	ID         int64     `json:"id" db:"id"`
	Source     string    `json:"source" db:"source"`
	Body       string    `json:"body" db:"body"`
	OccurredAt time.Time `json:"occurredAt" db:"occurred_at"`
}
