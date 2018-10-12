package disney

import (
	"context"
	"fmt"
)

const (
	waitTimesURLFormat string = apiBaseURL + "facility-service/theme-parks/%d;destination=%d/wait-times"
)

type Park struct {
	resortID int
	parkID   int
}

type WaitTime struct {
	AttractionID string `datastore:",noindex" json:"-"`
	PostedWait   int    `datastore:",noindex" json:"postedWait"`
	Operating    bool   `datastore:",noindex" json:"operating"`
	Timestamp    int64  `json:"timestamp"`
}

type waitTimesResponse struct {
	Attractions []waitTimesResponseEntry `json:"entries"`
}

type waitTimesResponseEntry struct {
	Type     string   `json:"type"`
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	WaitTime waitTime `json:"waitTime"`
}

type waitTime struct {
	PostedWaitMinutes int    `json:"postedWaitMinutes"`
	Status            string `json:"status"`
}

// FetchWaitTimes returns the current wait times for each attraction in the park.
func (park Park) FetchWaitTimes(ctx context.Context, timestamp int64) ([]WaitTime, error) {
	var resp waitTimesResponse
	err := fetchDisneyURL(ctx, park.waitTimeURL(), &resp)
	if err != nil {
		return nil, err
	}

	results := make([]WaitTime, 0, len(resp.Attractions))
	for _, attraction := range resp.Attractions {
		if attraction.Type != "Attraction" {
			continue
		}
		waitTime := WaitTime{
			AttractionID: attraction.Name,
			PostedWait:   attraction.WaitTime.PostedWaitMinutes,
			Operating:    attraction.WaitTime.Status == "Operating",
			Timestamp:    timestamp,
		}
		results = append(results, waitTime)
	}

	return results, nil
}

func (park Park) waitTimeURL() string {
	return fmt.Sprintf(waitTimesURLFormat, park.parkID, park.resortID)
}
