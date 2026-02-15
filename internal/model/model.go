package model

import "time"

type Comment struct {
	Text    string    `json:"text"`
	Created time.Time `json:"created"`
	By      string    `json:"by"`
}

type Issue struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	Type        string    `json:"type"`
	Priority    int       `json:"priority"`
	Labels      []string  `json:"labels"`
	Assignee    string    `json:"assignee,omitempty"`
	ParentID    string    `json:"parent_id,omitempty"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Description string    `json:"description,omitempty"`
	Comments    []Comment `json:"comments,omitempty"`
}

type Event struct {
	Timestamp time.Time `json:"ts"`
	Op        string    `json:"op"`
	Fields    []string  `json:"fields,omitempty"`
	From      string    `json:"from,omitempty"`
	To        string    `json:"to,omitempty"`
	Text      string    `json:"text,omitempty"`
	By        string    `json:"by,omitempty"`
}

type Config struct {
	Transitions  map[string][]string `json:"transitions"`
	DefaultState string              `json:"default_state"`
	Types        []string            `json:"types"`
	DefaultType  string              `json:"default_type"`
	IDLength     int                 `json:"id_length"`
}

func DefaultConfig() Config {
	return Config{
		Transitions: map[string][]string{
			"open":      {"active", "done", "cancelled"},
			"active":    {"done", "cancelled", "open", "review"},
			"review":    {"done", "active"},
			"done":      {"open"},
			"cancelled": {"open"},
		},
		DefaultState: "open",
		Types:        []string{"feature", "bug", "chore"},
		DefaultType:  "feature",
		IDLength:     6,
	}
}
