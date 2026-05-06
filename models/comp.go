package models

// Comp represents a Battlegrounds comp/archetype
type Comp struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Tier        string      `json:"tier"` // S, A, B
	Difficulty  string      `json:"difficulty"`
	Description string      `json:"description"`
	URL         string      `json:"url"`
	CoreCards   []Card      `json:"core_cards"`
	AddonCards  []Card      `json:"addon_cards"`
	HowToPlay   string      `json:"how_to_play"`
	WhenToCommit string     `json:"when_to_commit"`
	KeyTransition TavernGuide `json:"key_transition"`
}

// Card represents a minion card
type Card struct {
	Name   string `json:"name"`
	Tier   int    `json:"tier,omitempty"`
	Attack int    `json:"attack,omitempty"`
	Health int    `json:"health,omitempty"`
}

// TavernGuide shows key units at each tavern tier
type TavernGuide struct {
	OneStar  []string `json:"1_star"`
	TwoStar  []string `json:"2_star"`
	ThreeStar []string `json:"3_star"`
	FourStar  []string `json:"4_star"`
	FiveStar  []string `json:"5_star"`
}

// SeasonData represents the full season data
type SeasonData struct {
	Season      int    `json:"season"`
	LastUpdated string `json:"last_updated"`
	Source      string `json:"source"`
	PoweredBy   string `json:"powered_by"`
	Comps       []Comp `json:"comps"`
}

// GetTavernGuide returns a guide for each tavern tier
func (c *Comp) GetTavernGuide() map[string][]string {
	return map[string][]string{
		"Tier 1": c.KeyTransition.OneStar,
		"Tier 2": c.KeyTransition.TwoStar,
		"Tier 3": c.KeyTransition.ThreeStar,
		"Tier 4": c.KeyTransition.FourStar,
		"Tier 5": c.KeyTransition.FiveStar,
	}
}
