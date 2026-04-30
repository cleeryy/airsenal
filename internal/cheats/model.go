package cheats

// Cheatsheet holds metadata and content for a single topic.
type Cheatsheet struct {
	Topic       string   `json:"topic"`
	Category    string   `json:"category,omitempty"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Content     string   `json:"content"`
	Raw         string   `json:"raw,omitempty"`
}

type CategorySummary struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type TagSummary struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}
