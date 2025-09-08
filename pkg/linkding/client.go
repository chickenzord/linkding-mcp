package linkding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

type Bookmark struct {
	ID                    int       `json:"id"`
	URL                   string    `json:"url"`
	Title                 string    `json:"title"`
	Description           string    `json:"description"`
	Notes                 string    `json:"notes"`
	WebArchiveSnapshotURL string    `json:"web_archive_snapshot_url"`
	FaviconURL            string    `json:"favicon_url"`
	PreviewImageURL       string    `json:"preview_image_url"`
	IsArchived            bool      `json:"is_archived"`
	Unread                bool      `json:"unread"`
	Shared                bool      `json:"shared"`
	TagNames              []string  `json:"tag_names"`
	DateAdded             time.Time `json:"date_added"`
	DateModified          time.Time `json:"date_modified"`
}

type BookmarkResponse struct {
	Count    int        `json:"count"`
	Next     *string    `json:"next"`
	Previous *string    `json:"previous"`
	Results  []Bookmark `json:"results"`
}

type CreateBookmarkRequest struct {
	URL             string   `json:"url"`
	Title           string   `json:"title,omitempty"`
	Description     string   `json:"description,omitempty"`
	Notes           string   `json:"notes,omitempty"`
	TagNames        []string `json:"tag_names,omitempty"`
	Unread          bool     `json:"unread,omitempty"`
	Shared          bool     `json:"shared,omitempty"`
	IsArchived      bool     `json:"is_archived,omitempty"`
	DisableScraping bool     `json:"disable_scraping,omitempty"`
}

type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	DateAdded time.Time `json:"date_added"`
}

type CreateTagRequest struct {
	Name string `json:"name"`
}

type TagResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []Tag   `json:"results"`
}

func NewClient(baseURL, apiToken string) *Client {
	return &Client{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	url := c.baseURL + endpoint

	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	var req *http.Request
	var err error

	if reqBody != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, reqBody)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.apiToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *Client) GetBookmarks(ctx context.Context, limit, offset int, query string) (*BookmarkResponse, error) {
	endpoint := "/api/bookmarks/"
	params := url.Values{}

	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	if query != "" {
		params.Set("q", query)
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var bookmarkResponse BookmarkResponse
	if err := json.NewDecoder(resp.Body).Decode(&bookmarkResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bookmarkResponse, nil
}

func (c *Client) CreateBookmark(ctx context.Context, req CreateBookmarkRequest) (*Bookmark, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/bookmarks/", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var bookmark Bookmark
	if err := json.NewDecoder(resp.Body).Decode(&bookmark); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bookmark, nil
}

func (c *Client) UpdateBookmark(ctx context.Context, id int, req CreateBookmarkRequest) (*Bookmark, error) {
	endpoint := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.makeRequest(ctx, "PUT", endpoint, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var bookmark Bookmark
	if err := json.NewDecoder(resp.Body).Decode(&bookmark); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bookmark, nil
}

func (c *Client) DeleteBookmark(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.makeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) ArchiveBookmark(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/api/bookmarks/%d/archive/", id)

	resp, err := c.makeRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) UnarchiveBookmark(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/api/bookmarks/%d/unarchive/", id)

	resp, err := c.makeRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetTags(ctx context.Context, limit, offset int) (*TagResponse, error) {
	endpoint := "/api/tags/"
	params := url.Values{}

	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var tagResponse TagResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tagResponse, nil
}
