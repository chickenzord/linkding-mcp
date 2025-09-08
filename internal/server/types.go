package server

// GetTagsArgs defines the input structure for get_tags tool
type GetTagsArgs struct {
	Limit int `json:"limit,omitempty" jsonschema:"description:Maximum number of tags to return,default:50"`
}

// CreateBookmarkArgs defines the input structure for create_bookmark tool
type CreateBookmarkArgs struct {
	URL         string   `json:"url" jsonschema:"description:URL to bookmark"`
	Title       string   `json:"title,omitempty" jsonschema:"description:Bookmark title"`
	Description string   `json:"description,omitempty" jsonschema:"description:Bookmark description"`
	Tags        []string `json:"tags,omitempty" jsonschema:"description:List of tags"`
}

// SearchBookmarksArgs defines the input structure for search_bookmarks tool
type SearchBookmarksArgs struct {
	Query string `json:"query,omitempty" jsonschema:"description:Search query"`
	Limit int    `json:"limit,omitempty" jsonschema:"description:Maximum number of results,default:20"`
}

// BookmarkResult defines the output structure for bookmark operations
type BookmarkResult struct {
	ID          int      `json:"id"`
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Success     bool     `json:"success"`
	Message     string   `json:"message,omitempty"`
}

// TagResult defines the output structure for tag operations
type TagResult struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
