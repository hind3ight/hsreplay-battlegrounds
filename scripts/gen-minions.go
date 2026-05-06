package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hind3ight/hsreplay-battlegrounds/models"
)

func extractTribe(compName string) string {
	nameLower := strings.ToLower(compName)
	if strings.Contains(nameLower, "demon") {
		return "Demon"
	} else if strings.Contains(nameLower, "undead") {
		return "Undead"
	} else if strings.Contains(nameLower, "elemental") {
		return "Elemental"
	} else if strings.Contains(nameLower, "pirate") || strings.Contains(nameLower, "bounty") {
		return "Pirate"
	} else if strings.Contains(nameLower, "beast") {
		return "Beast"
	} else if strings.Contains(nameLower, "murloc") {
		return "Murloc"
	} else if strings.Contains(nameLower, "mech") {
		return "Mech"
	} else if strings.Contains(nameLower, "quilboar") {
		return "Quilboar"
	} else if strings.Contains(nameLower, "naga") {
		return "Naga"
	} else if strings.Contains(nameLower, "dragon") {
		return "Dragon"
	} else if strings.Contains(nameLower, "back to back") {
		return "Elemental"
	}
	return "Other"
}

type MinionMeta struct {
	Name       string            `json:"name"`
	NameCN     string            `json:"name_cn,omitempty"`
	Tier       int               `json:"tier"`
	Attack     int               `json:"attack"`
	Health     int               `json:"health"`
	Tribe      string            `json:"tribe"`
	CompCount  int               `json:"comp_count"`
	Comps      []string          `json:"comps"`
	TribesSeen []string          `json:"tribes_seen"`
	Tags       []string          `json:"tags,omitempty"`
}

type MinionsMetadata struct {
	GeneratedAt  string                 `json:"generated_at"`
	Source       string                 `json:"source"`
	TotalMinions int                    `json:"total_minions"`
	Minions      map[string]MinionMeta  `json:"minions"`
}

func main() {
	data, err := os.ReadFile("data/season13_comps.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data file: %v\n", err)
		os.Exit(1)
	}

	var season struct {
		Comps []models.Comp `json:"comps"`
	}
	if err := json.Unmarshal(data, &season); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	zhData, _ := os.ReadFile("data/minion_names.json")
	var zhNames map[string]string
	json.Unmarshal(zhData, &zhNames)

	type minionEntry struct {
		meta MinionMeta
	}
	minionMap := make(map[string]*MinionMeta)

	for _, comp := range season.Comps {
		tribe := extractTribe(comp.Name)
		for _, card := range comp.CoreCards {
			m, ok := minionMap[card.Name]
			if !ok {
				m = &MinionMeta{Name: card.Name}
				minionMap[card.Name] = m
			}
			m.Tier = card.Tier
			m.Attack = card.Attack
			m.Health = card.Health
			m.Tribe = tribe
			m.CompCount++
			if !contains(m.Comps, comp.Name) {
				m.Comps = append(m.Comps, comp.Name)
			}
			if !contains(m.TribesSeen, tribe) {
				m.TribesSeen = append(m.TribesSeen, tribe)
			}
		}
		for _, card := range comp.AddonCards {
			m, ok := minionMap[card.Name]
			if !ok {
				m = &MinionMeta{Name: card.Name}
				minionMap[card.Name] = m
			}
			m.CompCount++
			m.Tribe = tribe // Use comp's tribe for addons
			if !contains(m.Comps, comp.Name) {
				m.Comps = append(m.Comps, comp.Name)
			}
			if !contains(m.TribesSeen, tribe) {
				m.TribesSeen = append(m.TribesSeen, tribe)
			}
		}
	}

	// Add CN names and tags
	for name, meta := range minionMap {
		if cn, ok := zhNames[name]; ok {
			meta.NameCN = cn
		}
		// Add tags
		if meta.Tier >= 5 {
			meta.Tags = append(meta.Tags, "high_tier")
		}
		if len(meta.TribesSeen) >= 2 {
			meta.Tags = append(meta.Tags, "cross_tribe")
		}
		if contains(meta.Comps, "Back to Back") || contains(meta.Comps, "Elementals - Shop Buff/Spells") {
			meta.Tags = append(meta.Tags, "spell_scaling")
		}
	}

	// Sort comps for each minion
	for _, m := range minionMap {
		sort.Strings(m.Comps)
		sort.Strings(m.TribesSeen)
	}

	// Build final map sorted by name
	sortedMinions := make(map[string]MinionMeta)
	var names []string
	for name := range minionMap {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		m := minionMap[name]
		sortedMinions[name] = *m
	}

	result := MinionsMetadata{
		GeneratedAt:  "2025-05-06",
		Source:       "season13_comps.json analysis",
		TotalMinions: len(sortedMinions),
		Minions:      sortedMinions,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	os.WriteFile("data/minions_metadata.json", out, 0644)
	fmt.Printf("Generated data/minions_metadata.json with %d minions\n", len(sortedMinions))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
