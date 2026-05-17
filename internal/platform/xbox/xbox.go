package xbox

import (
	"bytes"
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
	oauthURL string
	xstsURL  string
	apiURL   string
	clientID string
	secret   string
	http     *http.Client

	mu       sync.Mutex
	xstsToken string
	xstsExp   time.Time
}

// NewClient creates an Xbox client. oauthURL, xstsURL, and apiURL are separated
// to allow test servers to intercept each endpoint independently.
func NewClient(oauthURL, xstsURL, apiURL, clientID, secret string) *Client {
	return &Client{
		oauthURL: oauthURL,
		xstsURL:  xstsURL,
		apiURL:   apiURL,
		clientID: clientID,
		secret:   secret,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) FetchStats(ctx context.Context, username string) (*model.PlayerStats, error) {
	xsts, err := c.getXSTSToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("xbox: get XSTS token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.apiURL+"/d4/profile/"+url.PathEscape(username), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "XBL3.0 x=;"+xsts)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
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
		return nil, fmt.Errorf("xbox: unexpected status %d", resp.StatusCode)
	}

	var p xboxProfile
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
		Username:   p.Gamertag,
		Platform:   "xbox",
		Characters: chars,
		Season: model.SeasonData{
			SeasonNumber:    p.Season.SeasonNumber,
			JourneyProgress: p.Season.JourneyProgress,
		},
	}, nil
}

func (c *Client) getXSTSToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.xstsToken != "" && time.Now().Before(c.xstsExp) {
		return c.xstsToken, nil
	}

	msToken, err := c.getMSToken(ctx)
	if err != nil {
		return "", err
	}

	body, _ := json.Marshal(map[string]any{
		"Properties": map[string]any{
			"AuthMethod": "RPS",
			"SiteName":   "user.auth.xboxlive.com",
			"RpsTicket":  "d=" + msToken,
		},
		"RelyingParty": "http://xboxlive.com",
		"TokenType":    "JWT",
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.xstsURL+"/xsts/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tr struct {
		Token    string `json:"Token"`
		NotAfter string `json:"NotAfter"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}

	exp, err := time.Parse(time.RFC3339, tr.NotAfter)
	if err != nil {
		exp = time.Now().Add(1 * time.Hour)
	}

	c.xstsToken = tr.Token
	c.xstsExp = exp
	return c.xstsToken, nil
}

func (c *Client) getMSToken(ctx context.Context) (string, error) {
	form := url.Values{"grant_type": {"client_credentials"}, "scope": {"https://graph.microsoft.com/.default"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.oauthURL+"/oauth/token", strings.NewReader(form.Encode()))
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
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	return tr.AccessToken, nil
}

type xboxProfile struct {
	Gamertag   string `json:"gamertag"`
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
