package battlenet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ghaugen/d4-stats-api/internal/model"
	"github.com/ghaugen/d4-stats-api/internal/platform"
)

type Client struct {
	tokenURL string
	apiURL   string
	clientID string
	secret   string
	http     *http.Client

	mu        sync.Mutex
	token     string
	tokenExp  time.Time
}

func NewClient(tokenURL, apiURL, clientID, secret string) *Client {
	return &Client{
		tokenURL: tokenURL,
		apiURL:   apiURL,
		clientID: clientID,
		secret:   secret,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) FetchStats(ctx context.Context, username string) (*model.PlayerStats, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("battlenet: get token: %w", err)
	}

	stats, status, err := c.fetchProfile(ctx, username, token)
	if err != nil {
		return nil, err
	}

	if status == http.StatusUnauthorized {
		// Token expired — refresh once and retry.
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()

		token, err = c.getToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("battlenet: refresh token: %w", err)
		}
		stats, status, err = c.fetchProfile(ctx, username, token)
		if err != nil {
			return nil, err
		}
	}

	switch status {
	case http.StatusOK:
		return stats, nil
	case http.StatusNotFound:
		return nil, platform.ErrPlayerNotFound
	case http.StatusTooManyRequests:
		return nil, platform.ErrPlatformRateLimited
	default:
		return nil, fmt.Errorf("battlenet: unexpected status %d", status)
	}
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExp) {
		return c.token, nil
	}

	form := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL+"/oauth/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.clientID, c.secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tr struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}

	c.token = tr.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	return c.token, nil
}

type bnetProfile struct {
	BattleTag  string `json:"battletag"`
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

func (c *Client) fetchProfile(ctx context.Context, username, token string) (*model.PlayerStats, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.apiURL+"/d4/profile/"+url.PathEscape(username), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, resp.StatusCode, nil
	}

	var p bnetProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, resp.StatusCode, err
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
		Username:   p.BattleTag,
		Platform:   "battlenet",
		Characters: chars,
		Season: model.SeasonData{
			SeasonNumber:    p.Season.SeasonNumber,
			JourneyProgress: p.Season.JourneyProgress,
		},
	}, resp.StatusCode, nil
}
