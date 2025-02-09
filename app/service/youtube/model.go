package youtube

type Video struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	PublishedAt    string `json:"publishedAt"`
	URL            string `json:"url"`
	Views          string `json:"views"`
	Comments       string `json:"comments"`
	Likes          string `json:"likes"`
	Duration       string `json:"duration"`
	Short          bool   `json:"short"`
	LocalReference string `json:"localReference"`
}

type PlaylistItemsResponse struct {
	NextPageToken string `json:"nextPageToken"`
	Items         []struct {
		Snippet struct {
			Title       string `json:"title"`
			PublishedAt string `json:"publishedAt"`
			ResourceID  struct {
				VideoID string `json:"videoId"`
			} `json:"resourceId"`
		} `json:"snippet"`
	} `json:"items"`
}

type VideoStatisticsResponse struct {
	Items []struct {
		ID             string `json:"id"`
		URL            string `json:"url"`
		ContentDetails struct {
			Duration string `json:"duration"`
		} `json:"contentDetails"`
		Statistics struct {
			ViewCount    string `json:"viewCount"`
			CommentCount string `json:"commentCount"`
			LikeCount    string `json:"likeCount"`
		} `json:"statistics"`
	} `json:"items"`
}
