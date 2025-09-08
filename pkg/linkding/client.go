// Package linkding provides a client for interacting with the Linkding bookmark manager API.
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

// Client represents a Linkding API client.
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// Bookmark represents a bookmark from the Linkding API.
type Bookmark struct {
	ID                    int       `json:"id"`                      // Unique identifier for the bookmark
	URL                   string    `json:"url"`                     // The bookmarked URL
	Title                 string    `json:"title"`                   // Title of the bookmark
	Description           string    `json:"description"`             // User-provided description
	Notes                 string    `json:"notes"`                   // User-provided notes
	WebArchiveSnapshotURL string    `json:"web_archive_snapshot_url"` // Web archive snapshot URL
	FaviconURL            string    `json:"favicon_url"`             // URL to the site's favicon
	PreviewImageURL       string    `json:"preview_image_url"`       // URL to a preview image
	IsArchived            bool      `json:"is_archived"`             // Whether the bookmark is archived
	Unread                bool      `json:"unread"`                  // Whether the bookmark is marked as unread
	Shared                bool      `json:"shared"`                  // Whether the bookmark is shared
	TagNames              []string  `json:"tag_names"`               // List of associated tag names
	DateAdded             time.Time `json:"date_added"`              // When the bookmark was created
	DateModified          time.Time `json:"date_modified"`           // When the bookmark was last modified
}

// BookmarkResponse represents the response from the bookmarks list API endpoint.
type BookmarkResponse struct {
	Count    int        `json:"count"`    // Total number of bookmarks matching the query
	Next     *string    `json:"next"`     // URL for the next page of results, if any
	Previous *string    `json:"previous"` // URL for the previous page of results, if any
	Results  []Bookmark `json:"results"`  // Array of bookmark objects
}

// CreateBookmarkRequest represents the request payload for creating or updating a bookmark.
type CreateBookmarkRequest struct {
	URL             string   `json:"url"`                        // The URL to bookmark (required)
	Title           string   `json:"title,omitempty"`           // Optional title for the bookmark
	Description     string   `json:"description,omitempty"`     // Optional description
	Notes           string   `json:"notes,omitempty"`           // Optional notes
	TagNames        []string `json:"tag_names,omitempty"`       // Optional list of tag names
	Unread          bool     `json:"unread,omitempty"`          // Whether to mark as unread
	Shared          bool     `json:"shared,omitempty"`          // Whether to share the bookmark
	IsArchived      bool     `json:"is_archived,omitempty"`     // Whether the bookmark should be archived
	DisableScraping bool     `json:"disable_scraping,omitempty"` // Whether to disable metadata scraping
}

// Tag represents a tag from the Linkding API.
type Tag struct {
	ID        int       `json:"id"`         // Unique identifier for the tag
	Name      string    `json:"name"`       // Name of the tag
	DateAdded time.Time `json:"date_added"` // When the tag was created
}

// CreateTagRequest represents the request payload for creating a tag.
type CreateTagRequest struct {
	Name string `json:"name"` // Name of the tag to create (required)
}

// TagResponse represents the response from the tags list API endpoint.
type TagResponse struct {
	Count    int     `json:"count"`    // Total number of tags matching the query
	Next     *string `json:"next"`     // URL for the next page of results, if any
	Previous *string `json:"previous"` // URL for the previous page of results, if any
	Results  []Tag   `json:"results"`  // Array of tag objects
}

// NewClient creates a new Linkding API client with the provided base URL and API token.
// The baseURL should include the protocol (e.g., "https://linkding.example.com").
// The apiToken can be obtained from the Linkding admin panel under Settings > Integrations.
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

// GetBookmarks retrieves bookmarks from the Linkding API.
// Parameters:
//   - limit: Maximum number of bookmarks to return (0 for default)
//   - offset: Number of bookmarks to skip (for pagination)
//   - query: Search query to filter bookmarks (empty string for no filter)
//
// Returns a BookmarkResponse containing the results and pagination information.
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

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var bookmarkResponse BookmarkResponse
	if err := json.NewDecoder(resp.Body).Decode(&bookmarkResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bookmarkResponse, nil
}

// CreateBookmark creates a new bookmark in Linkding.
// The URL field in the request is required; all other fields are optional.
// Returns the created bookmark with server-generated fields populated.
func (c *Client) CreateBookmark(ctx context.Context, req CreateBookmarkRequest) (*Bookmark, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/bookmarks/", req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var bookmark Bookmark
	if err := json.NewDecoder(resp.Body).Decode(&bookmark); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bookmark, nil
}

// UpdateBookmark updates an existing bookmark in Linkding.
// The id parameter specifies which bookmark to update.
// Returns the updated bookmark with all current field values.
func (c *Client) UpdateBookmark(ctx context.Context, id int, req CreateBookmarkRequest) (*Bookmark, error) {
	endpoint := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.makeRequest(ctx, "PUT", endpoint, req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var bookmark Bookmark
	if err := json.NewDecoder(resp.Body).Decode(&bookmark); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bookmark, nil
}

// DeleteBookmark permanently deletes a bookmark from Linkding.
// The id parameter specifies which bookmark to delete.
// Returns an error if the bookmark doesn't exist or deletion fails.
func (c *Client) DeleteBookmark(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.makeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

// ArchiveBookmark archives a bookmark in Linkding.
// Archived bookmarks are hidden from the main bookmark list but not deleted.
// The id parameter specifies which bookmark to archive.
func (c *Client) ArchiveBookmark(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/api/bookmarks/%d/archive/", id)

	resp, err := c.makeRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

// UnarchiveBookmark unarchives a previously archived bookmark in Linkding.
// This restores the bookmark to the main bookmark list.
// The id parameter specifies which bookmark to unarchive.
func (c *Client) UnarchiveBookmark(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/api/bookmarks/%d/unarchive/", id)

	resp, err := c.makeRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetTags retrieves tags from the Linkding API.
// Parameters:
//   - limit: Maximum number of tags to return (0 for default)
//   - offset: Number of tags to skip (for pagination)
//
// Returns a TagResponse containing the results and pagination information.
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

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var tagResponse TagResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tagResponse, nil
}
