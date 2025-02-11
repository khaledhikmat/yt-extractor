package cloudconvert

import (
	"testing"
)

// WARNIMNG: Requires that env vars be set:
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// STORAGE_REGION
// STORAGE_BUCKET
// CLOUDCONVERT_API_KEY

func TestConversion(t *testing.T) {
	svc := New(nil)
	s3URL, err := svc.ConvertVideoToAudio("UCP-PfkMcOKriSxFMH7pTxfA", "0hiQBvg5OmY")
	if err != nil {
		t.Error(err)
	}
	t.Log(s3URL)
}
