package cheats

// Cheatsheet holds metadata and content for a single topic.
type Cheatsheet struct {
	Topic       string   `json:"topic"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
	Raw         string   `json:"raw,omitempty"`
}
