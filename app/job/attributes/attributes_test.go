package jobattributes

import (
	"testing"
)

func TestPosting2Automation(t *testing.T) {
	err := postToAutomationWebhook([]int64{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
	}, "https://hook.us2.make.com/uhyw3mqps6sbgnzbuxjiltiveiwu34rh")
	if err != nil {
		t.Error(err)
	}
}
