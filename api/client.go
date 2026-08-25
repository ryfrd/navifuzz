package api

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

const apiVersion = "1.16.1"
const clientName = "navifuzz"

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
	Genre     string `json:"genre"`
}

type Artist struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	AlbumCount     int    `json:"albumCount"`
	AlbumID        string `json:"albumId,omitempty"`
	Genre          string `json:"genre,omitempty"`
	UserRating     int    `json:"userRating,omitempty"`
	AverageRating  float64 `json:"averageRating,omitempty"`
}

type Song struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Artist       string `json:"artist"`
	Album        string `json:"album"`
	Duration     int    `json:"duration"`
	Track        int    `json:"track"`
	DiscNumber   int    `json:"discNumber"`
	Genre        string `json:"genre"`
	Year         int    `json:"year"`
	Suffix       string `json:"suffix"`
	ContentType  string `json:"contentType"`
	Size         int64  `json:"size"`
	BitRate      int    `json:"bitRate"`
	SampleRate   int    `json:"sampleRate"`
	ChannelCount int    `json:"channelCount"`
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

type artistsResponse struct {
	Artists struct {
		Index []struct {
			Name string   `json:"name"`
			Artists []Artist `json:"artist"`
		} `json:"index"`
	} `json:"artists"`
}

type artistResponse struct {
	Artist struct {
		Album []Album `json:"album"`
	} `json:"artist"`
}

type searchResponse struct {
	SearchResult3 struct {
		Artist []Artist `json:"artist"`
		Album  []Album  `json:"album"`
		Song    []Song   `json:"song"`
	} `json:"searchResult3"`
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
	} else {
		params.Set("size", "99999")
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

func (c *Client) GetArtists() ([]Artist, error) {
	data, err := c.doRequest("getArtists", nil)
	if err != nil {
		return nil, err
	}

	var result artistsResponse
	if err := c.parseResponse(data, &result); err != nil {
		return nil, err
	}

	var artists []Artist
	for _, idx := range result.Artists.Index {
		artists = append(artists, idx.Artists...)
	}
	return artists, nil
}

func (c *Client) GetArtist(id string) ([]Album, error) {
	params := url.Values{}
	params.Set("id", id)

	data, err := c.doRequest("getArtist", params)
	if err != nil {
		return nil, err
	}

	var result artistResponse
	if err := c.parseResponse(data, &result); err != nil {
		return nil, err
	}

	return result.Artist.Album, nil
}

func (c *Client) Search3(query string) ([]Artist, []Album, []Song, error) {
	params := url.Values{}
	params.Set("query", query)

	data, err := c.doRequest("search3", params)
	if err != nil {
		return nil, nil, nil, err
	}

	var result searchResponse
	if err := c.parseResponse(data, &result); err != nil {
		return nil, nil, nil, err
	}

	return result.SearchResult3.Artist, result.SearchResult3.Album, result.SearchResult3.Song, nil
}

type randomSongsResponse struct {
	RandomSongs struct {
		Song []Song `json:"song"`
	} `json:"randomSongs"`
}

func (c *Client) GetRandomSongs(size int) ([]Song, error) {
	params := url.Values{}
	if size > 0 {
		params.Set("size", strconv.Itoa(size))
	} else {
		params.Set("size", "99999")
	}

	data, err := c.doRequest("getRandomSongs", params)
	if err != nil {
		return nil, err
	}

	var result randomSongsResponse
	if err := c.parseResponse(data, &result); err != nil {
		return nil, err
	}

	return result.RandomSongs.Song, nil
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

func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

type albumData struct {
	Album
	DurationStr string
}

type songData struct {
	Song
	DurationStr string
	TrackStr    string
	SizeStr     string
	BitRateStr  string
	FileType    string
}

type artistData struct {
	Artist
}

func RenderArtist(a Artist, format string) (string, error) {
	tmpl, err := template.New("artist").Parse(format)
	if err != nil {
		return "", fmt.Errorf("invalid artist_format: %w", err)
	}
	data := artistData{Artist: a}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("artist_format error: %w", err)
	}
	return buf.String(), nil
}

func RenderAlbum(a Album, format string) (string, error) {
	tmpl, err := template.New("album").Parse(format)
	if err != nil {
		return "", fmt.Errorf("invalid album_format: %w", err)
	}
	data := albumData{Album: a, DurationStr: FormatDuration(a.Duration)}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("album_format error: %w", err)
	}
	return buf.String(), nil
}

func RenderSong(s Song, format string) (string, error) {
	tmpl, err := template.New("song").Parse(format)
	if err != nil {
		return "", fmt.Errorf("invalid song_format: %w", err)
	}
	data := songData{
		Song:        s,
		DurationStr: FormatDuration(s.Duration),
		TrackStr:    fmt.Sprintf("%02d", s.Track),
		SizeStr:     FormatSize(s.Size),
		BitRateStr:  fmt.Sprintf("%d kbps", s.BitRate),
		FileType:    s.Suffix,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("song_format error: %w", err)
	}
	return buf.String(), nil
}
