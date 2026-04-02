package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/box/internal/types"
)

const (
	APIBaseURL    = "https://api.box.com/2.0"
	UploadBaseURL = "https://upload.box.com/api/2.0"
	maxRetries    = 3
)

type BoxClient struct {
	HTTP *http.Client
}

func New(httpClient *http.Client) *BoxClient {
	return &BoxClient{HTTP: httpClient}
}

func (c *BoxClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		retryAfter := resp.Header.Get("Retry-After")
		resp.Body.Close()

		wait := 2 * time.Second
		if retryAfter != "" {
			if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
				wait = time.Duration(seconds) * time.Second
			}
		}
		if attempt < maxRetries {
			log.Debug().Dur("wait", wait).Msg("rate limited, retrying")
			time.Sleep(wait)
		}
	}
	return resp, nil
}

func (c *BoxClient) DoJSON(req *http.Request, target interface{}) (*http.Response, error) {
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return resp, HandleError("request", resp)
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return resp, fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return resp, nil
}

func HandleError(action string, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("api %s failed with status %s (could not read error body)", action, resp.Status)
	}
	var boxErr types.BoxError
	if json.Unmarshal(body, &boxErr) == nil && boxErr.Code != "" {
		return fmt.Errorf("api %s failed: %s - %s", action, boxErr.Code, boxErr.Message)
	}
	return fmt.Errorf("api %s failed with status %s: %s", action, resp.Status, string(body))
}
