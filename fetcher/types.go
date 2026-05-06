package fetcher

// RawCompSummary represents a comp listed on the main page
type RawCompSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// RawCompDetail represents detailed comp data from detail page
type RawCompDetail struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Difficulty   string            `json:"difficulty"`
	CoreCards    []RawCard         `json:"core_cards"`
	AddonCards   []RawCard         `json:"addon_cards"`
	HowToPlay    string            `json:"how_to_play"`
	WhenToCommit string            `json:"when_to_commit"`
	Enablers     []RawCard         `json:"enablers"`
	TavernTier   map[string][]string `json:"tavern_tier"`
}

// RawCard represents a card extracted from page
type RawCard struct {
	Name  string `json:"name"`
	Tier  int    `json:"tier,omitempty"`
	Attack int   `json:"attack,omitempty"`
	Health int   `json:"health,omitempty"`
	URL   string `json:"url,omitempty"`
}

// RawMinionDetail represents detailed minion info fetched from minion page
type RawMinionDetail struct {
	Name      string   `json:"name"`
	Tier      int      `json:"tier"`
	Attack    int      `json:"attack"`
	Health    int      `json:"health"`
	Tribe     string   `json:"tribe"`
	Abilities []string `json:"abilities,omitempty"`
}

// ScrapedResult represents the complete scraped result
type ScrapedResult struct {
	Summaries []RawCompSummary   `json:"summaries"`
	Comps     []RawCompDetail    `json:"comps"`
	ScrapedAt string             `json:"scraped_at"`
}
