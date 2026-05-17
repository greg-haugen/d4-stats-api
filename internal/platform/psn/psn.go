package psn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ghaugen/d4-stats-api/internal/model"
	"github.com/ghaugen/d4-stats-api/internal/platform"
)

type Client struct {
	apiURL    string
	authToken string
	http      *http.Client
}

func NewClient(apiURL, authToken string) *Client {
	return &Client{
		apiURL:    apiURL,
		authToken: authToken,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

type psnProfile struct {
	OnlineID   string `json:"onlineId"`
	Characters []struct {
		Name         string `json:"name"`
		Class        string `json:"class"`
		Level        int    `json:"level"`
		ParagonLevel int    `json:"paragonLevel"`
	} `json:"characters"`
	Season struct {
		SeasonNumber    int `json:"seasonNumber"`
		JourneyProgress int `json:"journeyProgress"`
	} `json:"season"`
}

func (c *Client) FetchStats(ctx context.Context, username string) (*model.PlayerStats, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.apiURL+"/d4/profile/"+url.PathEscape(username), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		io.Copy(io.Discard, resp.Body)
		return nil, platform.ErrPSNTokenInvalid
	case http.StatusNotFound:
		io.Copy(io.Discard, resp.Body)
		return nil, platform.ErrPlayerNotFound
	case http.StatusTooManyRequests:
		io.Copy(io.Discard, resp.Body)
		return nil, platform.ErrPlatformRateLimited
	case http.StatusOK:
		// handled below
	default:
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("psn: unexpected status %d", resp.StatusCode)
	}

	var p psnProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}

	chars := make([]model.Character, len(p.Characters))
	for i, ch := range p.Characters {
		chars[i] = model.Character{
			Name:         ch.Name,
			Class:        ch.Class,
			Level:        ch.Level,
			ParagonLevel: ch.ParagonLevel,
		}
	}

	return &model.PlayerStats{
		Username:   p.OnlineID,
		Platform:   "psn",
		Characters: chars,
		Season: model.SeasonData{
			SeasonNumber:    p.Season.SeasonNumber,
			JourneyProgress: p.Season.JourneyProgress,
		},
	}, nil
}
