package fetcher

import (
	"fmt"

	"github.com/hind3ight/hsreplay-battlegrounds/models"
)

// ToModels converts scraped raw comps to models.Comp
func ToModels(result *ScrapedResult) []models.Comp {
	comps := make([]models.Comp, 0, len(result.Comps))

	summaryMap := make(map[string]RawCompSummary)
	for _, s := range result.Summaries {
		summaryMap[s.URL] = s
	}

	for _, raw := range result.Comps {
		comp := models.Comp{
			Name:         raw.Name,
			Difficulty:   raw.Difficulty,
			Description:  getDescription(raw, summaryMap),
			HowToPlay:    raw.HowToPlay,
			WhenToCommit: raw.WhenToCommit,
			CoreCards:    toModelCards(raw.CoreCards),
			AddonCards:   toModelCards(raw.AddonCards),
			URL:          getURL(raw, summaryMap),
		}

		// Parse ID from URL if available
		if raw.ID != "" {
			fmt.Sscanf(raw.ID, "%d", &comp.ID)
		}

		// Determine tier based on content analysis
		comp.Tier = determineTier(comp)

		// Build tavern guide from card tiers
		comp.KeyTransition = buildTavernGuide(comp)

		comps = append(comps, comp)
	}

	return comps
}

func toModelCards(rawCards []RawCard) []models.Card {
	cards := make([]models.Card, len(rawCards))
	for i, c := range rawCards {
		cards[i] = models.Card{
			Name:   c.Name,
			Tier:   c.Tier,
			Attack: c.Attack,
			Health: c.Health,
		}
	}
	return cards
}

func getDescription(raw RawCompDetail, summaryMap map[string]RawCompSummary) string {
	// Try to find summary info
	for _, s := range summaryMap {
		if raw.Name == s.Name || raw.Name == s.Name+" Comp" {
			return s.Description
		}
	}
	return raw.Name
}

func getURL(raw RawCompDetail, summaryMap map[string]RawCompSummary) string {
	for _, s := range summaryMap {
		if raw.Name == s.Name || raw.Name == s.Name+" Comp" {
			return s.URL
		}
	}
	return ""
}

func determineTier(comp models.Comp) string {
	// Simple heuristic: check if comp has high-tier cards
	hasHighTier := false
	for _, card := range comp.CoreCards {
		if card.Tier >= 6 {
			hasHighTier = true
			break
		}
	}
	if hasHighTier {
		return "S"
	}
	return "A"
}

func buildTavernGuide(comp models.Comp) models.TavernGuide {
	guide := models.TavernGuide{
		OneStar:   []string{},
		TwoStar:   []string{},
		ThreeStar: []string{},
		FourStar:  []string{},
		FiveStar:  []string{},
	}

	seen := make(map[string]bool)

	// Group core cards by tier
	for _, card := range comp.CoreCards {
		if card.Tier == 0 || seen[card.Name] {
			continue
		}
		seen[card.Name] = true

		switch card.Tier {
		case 1:
			guide.OneStar = append(guide.OneStar, card.Name)
		case 2:
			guide.TwoStar = append(guide.TwoStar, card.Name)
		case 3:
			guide.ThreeStar = append(guide.ThreeStar, card.Name)
		case 4:
			guide.FourStar = append(guide.FourStar, card.Name)
		case 5, 6:
			guide.FiveStar = append(guide.FiveStar, card.Name)
		}
	}

	// Ensure no empty slices
	if len(guide.OneStar) == 0 {
		guide.OneStar = []string{}
	}
	if len(guide.TwoStar) == 0 {
		guide.TwoStar = []string{}
	}
	if len(guide.ThreeStar) == 0 {
		guide.ThreeStar = []string{}
	}
	if len(guide.FourStar) == 0 {
		guide.FourStar = []string{}
	}
	if len(guide.FiveStar) == 0 {
		guide.FiveStar = []string{}
	}

	return guide
}
