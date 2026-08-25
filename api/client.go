package api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const apiVersion = "1.16.1"
const clientName = "navifzf"

type Client struct {
	Server   string
	Username string
	Password string
}

type Album struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	ArtistID  string `json:"artistId"`
	Year      int    `json:"year"`
	SongCount int    `json:"songCount"`
	Duration  int    `json:"duration"`
}

type Song struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"`
	Track    int    `json:"track"`
}

type albumListResponse struct {
	AlbumList2 struct {
		Album []Album `json:"album"`
	} `json:"albumList2"`
}

type albumResponse struct {
	Album struct {
		Song []Song `json:"song"`
	} `json:"album"`
}

type subsonicError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type subsonicResponse struct {
	Status string        `json:"status"`
	Error  *subsonicError `json:"error,omitempty"`
}

func NewClient(server, username, password string) *Client {
	server = strings.TrimRight(server, "/")
	return &Client{Server: server, Username: username, Password: password}
}

func (c *Client) authParams() url.Values {
	salt := generateSalt()
	token := computeToken(c.Password, salt)

	params := url.Values{}
	params.Set("u", c.Username)
	params.Set("t", token)
	params.Set("s", salt)
	params.Set("v", apiVersion)
	params.Set("c", clientName)
	params.Set("f", "json")
	return params
}

func (c *Client) doRequest(endpoint string, params url.Values) ([]byte, error) {
	auth := c.authParams()
	for k, vs := range params {
		for _, v := range vs {
			auth.Set(k, v)
		}
	}

	reqURL := fmt.Sprintf("%s/rest/%s?%s", c.Server, endpoint, auth.Encode())
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return body, nil
}

func (c *Client) parseResponse(data []byte, target interface{}) error {
	// The Subsonic response wraps everything in "subsonic-response"
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	resp, ok := envelope["subsonic-response"]
	if !ok {
		return fmt.Errorf("missing subsonic-response key")
	}

	var sr subsonicResponse
	if err := json.Unmarshal(resp, &sr); err != nil {
		return fmt.Errorf("cannot parse response: %w", err)
	}

	if sr.Status != "ok" {
		msg := "unknown error"
		if sr.Error != nil {
			msg = sr.Error.Message
		}
		return fmt.Errorf("API error: %s", msg)
	}

	if target != nil {
		if err := json.Unmarshal(resp, target); err != nil {
			return fmt.Errorf("cannot parse response data: %w", err)
		}
	}

	return nil
}

func (c *Client) Ping() error {
	data, err := c.doRequest("ping", nil)
	if err != nil {
		return err
	}
	return c.parseResponse(data, nil)
}

func (c *Client) GetAlbumList2(listType string, size int) ([]Album, error) {
	params := url.Values{}
	params.Set("type", listType)
	if size > 0 {
		params.Set("size", strconv.Itoa(size))
	}

	data, err := c.doRequest("getAlbumList2", params)
	if err != nil {
		return nil, err
	}

	var result albumListResponse
	if err := c.parseResponse(data, &result); err != nil {
		return nil, err
	}

	return result.AlbumList2.Album, nil
}

func (c *Client) GetAlbum(id string) ([]Song, error) {
	params := url.Values{}
	params.Set("id", id)

	data, err := c.doRequest("getAlbum", params)
	if err != nil {
		return nil, err
	}

	var result albumResponse
	if err := c.parseResponse(data, &result); err != nil {
		return nil, err
	}

	return result.Album.Song, nil
}

func (c *Client) StreamURL(id string) string {
	params := c.authParams()
	params.Set("id", id)
	return fmt.Sprintf("%s/rest/stream?%s", c.Server, params.Encode())
}

func (c *Client) Scrobble(id string) error {
	params := url.Values{}
	params.Set("id", id)
	params.Set("submission", "true")

	data, err := c.doRequest("scrobble", params)
	if err != nil {
		return err
	}
	return c.parseResponse(data, nil)
}

func generateSalt() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = "0123456789abcdef"[rand.Intn(16)]
	}
	return string(b)
}

func computeToken(password, salt string) string {
	h := md5.Sum([]byte(password + salt))
	return fmt.Sprintf("%x", h)
}

func FormatDuration(secs int) string {
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func DisplayAlbum(a Album) string {
	return fmt.Sprintf("%s - %s (%s, %d songs)",
		a.Artist, a.Name, strconv.Itoa(a.Year), a.SongCount)
}

func DisplaySong(s Song) string {
	return fmt.Sprintf("%02d. %s (%s)",
		s.Track, s.Title, FormatDuration(s.Duration))
}
