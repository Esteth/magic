package disney

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"google.golang.org/appengine/urlfetch"
)

const (
	apiAuthorizationBody = "grant_type=assertion&assertion_type=public&client_id=WDPRO-MOBILE.MDX.WDW.ANDROID-PROD"
	apiBaseURL           = "https://api.wdpro.disney.go.com/"
)

type authResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in,string"`
}

func fetchAccessToken(ctx context.Context) (string, error) {
	client := urlfetch.Client(ctx)
	resp, err :=
		client.Post(
			"https://authorization.go.com/token",
			"raw",
			strings.NewReader(apiAuthorizationBody))
	if err != nil {
		return "", errors.Wrap(err, "Error fetching auth token")
	}
	defer resp.Body.Close()

	decodedResponse := authResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&decodedResponse); err != nil {
		return "", errors.Wrap(err, "Error decoding auth token response")
	}

	// TODO: cache results according to expiresIn.
	return decodedResponse.AccessToken, nil
}

func fetchDisneyURL(ctx context.Context, url string, out interface{}) error {
	accessToken, err := fetchAccessToken(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "BEARER "+accessToken)
	req.Header.Add("Accept", "application/json;apiversion=1")
	req.Header.Add("X-Conversation-Id", "WDPRO-MOBILE.MDX.CLIENT-PROD")
	req.Header.Add("X-App-Id", "WDW-MDX-ANDROID-3.4.1")
	req.Header.Add("X-Correlation-ID", strconv.FormatInt(time.Now().UTC().Unix(), 10))

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return errors.Wrap(err, "Error decoding API response")
	}

	return nil
}
