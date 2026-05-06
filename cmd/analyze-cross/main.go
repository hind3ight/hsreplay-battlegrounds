package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ============================================================================
// Data structures
// ============================================================================

type SeasonData struct {
	Season      int    `json:"season"`
	LastUpdated string `json:"last_updated"`
	Source      string `json:"source"`
	PoweredBy   string `json:"powered_by"`
	Comps       []Comp `json:"comps"`
}

type Comp struct {
	ID            int               `json:"id"`
	Name          string            `json:"name"`
	Tier          string            `json:"tier"`
	Difficulty    string            `json:"difficulty"`
	Description   string            `json:"description"`
	URL           string            `json:"url"`
	CoreCards     []Card            `json:"core_cards"`
	AddonCards    []Card            `json:"addon_cards"`
	HowToPlay     string            `json:"how_to_play"`
	WhenToCommit  string            `json:"when_to_commit"`
	KeyTransition map[string][]string `json:"key_transition"`
}

type Card struct {
	Name   string `json:"name"`
	Tier   int    `json:"tier,omitempty"`
	Attack int    `json:"attack,omitempty"`
	Health int    `json:"health,omitempty"`
}

type MinionMeta struct {
	Name       string   `json:"name"`
	NameCN     string   `json:"name_cn,omitempty"`
	Tier       int      `json:"tier"`
	Attack     int      `json:"attack"`
	Health     int      `json:"health"`
	Tribe      string   `json:"tribe"`
	CompCount  int      `json:"comp_count"`
	Comps      []string `json:"comps"`
	TribesSeen []string `json:"tribes_seen"`
	Tags       []string `json:"tags,omitempty"`
}

type MinionsMetadata struct {
	GeneratedAt  string                 `json:"generated_at"`
	Source       string                 `json:"source"`
	TotalMinions int                    `json:"total_minions"`
	Minions      map[string]MinionMeta `json:"minions"`
}

// ============================================================================
// Analysis result types
// ============================================================================

// CrossCompEntry: a minion's role in a cross-comp analysis
type CrossCompEntry struct {
	Minion        string   `json:"minion"`
	NameCN        string   `json:"name_cn,omitempty"`
	Tier          int      `json:"tier"`
	AsCoreIn      []string `json:"as_core_in"`
	AsAddonIn     []string `json:"as_addon_in"`
	CompCount     int      `json:"comp_count"`
	CrossTribe    bool     `json:"cross_tribe"`
	TribesSeen    []string `json:"tribes_seen"`
	TotalAppearances int   `json:"total_appearances"`
}

// PivotOpportunity: a card that bridges two comps
type PivotOpportunity struct {
	Minion       string   `json:"minion"`
	NameCN       string   `json:"name_cn,omitempty"`
	Tier         int      `json:"tier"`
	FromComps    []string `json:"from_comps"`
	ToComps      []string `json:"to_comps"`
	PivotType    string   `json:"pivot_type"` // "transition", "fodder", "enabler"
	Score        float64  `json:"pivot_score"`
}

// CompRelation: similarity between two comps
type CompRelation struct {
	CompA        string   `json:"comp_a"`
	CompB        string   `json:"comp_b"`
	SharedCards  []string `json:"shared_cards"`
	SharedCount  int      `json:"shared_count"`
	Similarity   float64  `json:"similarity"` // 0-1
	CanTransition bool    `json:"can_transition"`
}

// CompAnalysis: complete analysis for one comp
type CompAnalysis struct {
	Comp        Comp             `json:"comp"`
	AllCards    []Card           `json:"all_cards"`
	Pivots      []PivotOpportunity `json:"pivots"`
	EconomyUnits []string        `json:"economy_units"`
	CoreUnits   []Card           `json:"core_units"`
	AddonUnits  []Card           `json:"addon_units"`
}

// ============================================================================
// Main
// ============================================================================

var (
	compsData    SeasonData
	zhNames      map[string]string
	minionsMeta  MinionsMetadata
)

func main() {
	compsFile := flag.String("comps", "data/season13_comps.json", "Comps JSON file")
	namesFile := flag.String("names", "data/minion_names.json", "Minion names mapping file")
	minionsFile := flag.String("minions", "data/minions_metadata.json", "Minions metadata file")
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	minionFilter := flag.String("minion", "", "Show only this minion's cross-comp analysis")
	compFilter := flag.String("comp", "", "Show only this comp's analysis")
	pivotOnly := flag.Bool("pivots", false, "Show only pivot opportunities")
	sharedOnly := flag.Bool("shared", false, "Show only shared cards between comps")
	flag.Parse()

	// Load data
	loadComps(*compsFile)
	loadZhNames(*namesFile)
	if _, err := os.Stat(*minionsFile); err == nil {
		loadMinionsMeta(*minionsFile)
	}

	// Build all-card set per comp
	compCards := buildCompCards()

	if *minionFilter != "" {
		analyzeMinionCross(*minionFilter, compCards)
		return
	}

	if *pivotOnly {
		analyzePivots(compCards)
		return
	}

	if *sharedOnly {
		analyzeSharedCards(compCards)
		return
	}

	if *compFilter != "" {
		analyzeComp(*compFilter, compCards)
		return
	}

	// Default: full analysis
	runFullAnalysis(compCards, *outputFile)
}

// ============================================================================
// Data loading
// ============================================================================

func loadComps(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading comps file: %v\n", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(data, &compsData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing comps JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d comps (Season %d)\n", len(compsData.Comps), compsData.Season)
}

func loadZhNames(path string) {
	zhNames = make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Warning: could not load %s: %v\n", path, err)
		return
	}
	json.Unmarshal(data, &zhNames)
}

func loadMinionsMeta(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Warning: could not load %s: %v\n", path, err)
		return
	}
	json.Unmarshal(data, &minionsMeta)
	fmt.Printf("Loaded %d minions metadata\n", minionsMeta.TotalMinions)
}

// ============================================================================
// Core analysis
// ============================================================================

func buildCompCards() map[string][]string {
	result := make(map[string][]string)
	for _, comp := range compsData.Comps {
		cardSet := make(map[string]bool)
		for _, c := range comp.CoreCards {
			cardSet[c.Name] = true
		}
		for _, c := range comp.AddonCards {
			cardSet[c.Name] = true
		}
		var cards []string
		for name := range cardSet {
			cards = append(cards, name)
		}
		result[comp.Name] = cards
	}
	return result
}

// getCompRole returns whether a card is core or addon in a comp
func getCompRole(compName, cardName string) string {
	for _, c := range compsData.Comps {
		if c.Name != compName {
			continue
		}
		for _, card := range c.CoreCards {
			if card.Name == cardName {
				return "core"
			}
		}
		for _, card := range c.AddonCards {
			if card.Name == cardName {
				return "addon"
			}
		}
	}
	return "unknown"
}

// getCardInfo returns Card for a given name from a specific comp
func getCardInfo(compName, cardName string) Card {
	for _, comp := range compsData.Comps {
		if comp.Name != compName {
			continue
		}
		for _, c := range comp.CoreCards {
			if c.Name == cardName {
				return c
			}
		}
		for _, c := range comp.AddonCards {
			if c.Name == cardName {
				return c
			}
		}
	}
	return Card{Name: cardName}
}

// ============================================================================
// Minion cross-comp analysis
// ============================================================================

func analyzeMinionCross(minionName string, compCards map[string][]string) {
	fmt.Printf("\n============================================================\n")
	fmt.Printf("随从跨流派分析: %s\n", minionName)
	fmt.Printf("============================================================\n")

	cn := getCN(minionName)
	if cn != "" {
		fmt.Printf("中文名: %s\n", cn)
	}

	var asCore, asAddon []string
	for compName, cards := range compCards {
		role := getCompRole(compName, minionName)
		for _, c := range cards {
			if c == minionName {
				if role == "core" {
					asCore = append(asCore, compName)
				} else {
					asAddon = append(asAddon, compName)
				}
			}
		}
	}

	if len(asCore) > 0 {
		fmt.Printf("\n核心卡所在流派 (%d个):\n", len(asCore))
		for _, c := range asCore {
			tier := ""
			for _, comp := range compsData.Comps {
				if comp.Name == c {
					tier = comp.Tier
					break
				}
			}
			fmt.Printf("  [%s] %s\n", tier, c)
		}
	}

	if len(asAddon) > 0 {
		fmt.Printf("\n辅助卡所在流派 (%d个):\n", len(asAddon))
		for _, c := range asAddon {
			tier := ""
			for _, comp := range compsData.Comps {
				if comp.Name == c {
					tier = comp.Tier
					break
				}
			}
			fmt.Printf("  [%s] %s\n", tier, c)
		}
	}

	// Find shared cards between all comps this minion appears in
	allComps := append(asCore, asAddon...)
	if len(allComps) > 1 {
		shared := findSharedCardsAcrossComps(allComps, compCards)
		if len(shared) > 0 {
			fmt.Printf("\n这些流派之间共享的卡牌 (%d张):\n", len(shared))
			for _, card := range shared {
				cn2 := getCN(card)
				if cn2 != "" {
					fmt.Printf("  %s (%s)\n", card, cn2)
				} else {
					fmt.Printf("  %s\n", card)
				}
			}
		}
	}
}

// ============================================================================
// Pivot analysis
// ============================================================================

func analyzePivots(compCards map[string][]string) {
	fmt.Printf("\n============================================================\n")
	fmt.Printf("转型(枢轴)机会分析\n")
	fmt.Printf("============================================================\n")

	// Find cards that appear in multiple comps
	cardComps := make(map[string][]string)
	for compName, cards := range compCards {
		for _, card := range cards {
			cardComps[card] = append(cardComps[card], compName)
		}
	}

	type pivot struct {
		card      string
		comps     []string
		tier      int
		score     float64
		pivotType string
	}
	var pivots []pivot

	for card, comps := range cardComps {
		if len(comps) < 2 {
			continue
		}

		tier := getMinionTier(card)
		score := float64(len(comps)) * 10
		if tier <= 2 {
			score += 5 // Early tier = good pivot
		}
		if tier >= 5 {
			score -= 3 // Late tier = less flexible pivot
		}

		// Determine pivot type
		pivotType := "fodder"
		coreCount := 0
		for _, compName := range comps {
			if getCompRole(compName, card) == "core" {
				coreCount++
			}
		}
		if coreCount >= 2 {
			pivotType = "core_pivot"
		} else if coreCount == 1 {
			pivotType = "transition"
		}

		pivots = append(pivots, pivot{card, comps, tier, score, pivotType})
	}

	// Sort by score descending
	sort.Slice(pivots, func(i, j int) bool {
		return pivots[i].score > pivots[j].score
	})

	fmt.Printf("\n推荐转型卡牌 (按转型价值排序):\n")
	for _, p := range pivots[:min(30, len(pivots))] {
		cn := getCN(p.card)
		cnStr := ""
		if cn != "" {
			cnStr = " / " + cn
		}
		fmt.Printf("\n[%s] %s%s (Tier %d, 出现在%d个流派)\n", p.pivotType, p.card, cnStr, p.tier, len(p.comps))
		for _, c := range p.comps {
			tier := ""
			for _, comp := range compsData.Comps {
				if comp.Name == c {
					tier = comp.Tier
					break
				}
			}
			fmt.Printf("  -> %s [%s]\n", c, tier)
		}
	}
}

// ============================================================================
// Shared cards analysis
// ============================================================================

func analyzeSharedCards(compCards map[string][]string) {
	fmt.Printf("\n============================================================\n")
	fmt.Printf("流派间共享卡牌分析\n")
	fmt.Printf("============================================================\n")

	cardComps := make(map[string][]string)
	for compName, cards := range compCards {
		for _, card := range cards {
			cardComps[card] = append(cardComps[card], compName)
		}
	}

	type shared struct {
		card   string
		comps  []string
		tier   int
		coreIn int
	}
	var sharedCards []shared

	for card, comps := range cardComps {
		if len(comps) < 2 {
			continue
		}
		coreIn := 0
		for _, compName := range comps {
			if getCompRole(compName, card) == "core" {
				coreIn++
			}
		}
		sharedCards = append(sharedCards, shared{card, comps, getMinionTier(card), coreIn})
	}

	// Sort: most shared first, then by core count
	sort.Slice(sharedCards, func(i, j int) bool {
		if len(sharedCards[i].comps) != len(sharedCards[j].comps) {
			return len(sharedCards[i].comps) > len(sharedCards[j].comps)
		}
		return sharedCards[i].coreIn > sharedCards[j].coreIn
	})

	fmt.Printf("\n共享卡牌排名:\n")
	for i, s := range sharedCards {
		if i >= 30 {
			break
		}
		cn := getCN(s.card)
		cnStr := ""
		if cn != "" {
			cnStr = " (" + cn + ")"
		}
		coreStr := ""
		if s.coreIn > 0 {
			coreStr = fmt.Sprintf(", %d个作为核心", s.coreIn)
		}
		fmt.Printf("\n%d. %s%s [Tier %d] - 共享于 %d 个流派%s\n", i+1, s.card, cnStr, s.tier, len(s.comps), coreStr)
		for _, c := range s.comps {
			role := getCompRole(c, s.card)
			fmt.Printf("     %s (%s)\n", c, role)
		}
	}
}

// ============================================================================
// Single comp analysis
// ============================================================================

func analyzeComp(compName string, compCards map[string][]string) {
	var target Comp
	found := false
	for _, c := range compsData.Comps {
		if c.Name == compName {
			target = c
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Comp not found: %s\n", compName)
		return
	}

	fmt.Printf("\n============================================================\n")
	fmt.Printf("流派深度分析: %s\n", target.Name)
	fmt.Printf("============================================================\n")
	fmt.Printf("Tier: %s | Difficulty: %s\n", target.Tier, target.Difficulty)
	fmt.Printf("Description: %s\n\n", target.Description)
	fmt.Printf("When to Commit: %s\n\n", target.WhenToCommit)
	fmt.Printf("How to Play: %s\n", target.HowToPlay)

	// Core units
	fmt.Printf("\n--- 核心卡 (%d张) ---\n", len(target.CoreCards))
	for _, c := range target.CoreCards {
		cn := getCN(c.Name)
		fmt.Printf("  [%d星] %s", c.Tier, c.Name)
		if cn != "" {
			fmt.Printf(" (%s)", cn)
		}
		if c.Attack > 0 {
			fmt.Printf(" 攻:%d 血:%d", c.Attack, c.Health)
		}
		fmt.Printf("\n")
	}

	// Addon units
	if len(target.AddonCards) > 0 {
		fmt.Printf("\n--- 辅助卡 (%d张) ---\n", len(target.AddonCards))
		for _, c := range target.AddonCards {
			cn := getCN(c.Name)
			fmt.Printf("  %s", c.Name)
			if cn != "" {
				fmt.Printf(" (%s)", cn)
			}
			fmt.Printf("\n")
		}
	}

	// Key transitions
	if len(target.KeyTransition) > 0 {
		fmt.Printf("\n--- 酒馆过渡路线 ---\n")
		tiers := []string{"1_star", "2_star", "3_star", "4_star", "5_star"}
		tierNames := []string{"1星", "2星", "3星", "4星", "5星"}
		for i, t := range tiers {
			if cards, ok := target.KeyTransition[t]; ok && len(cards) > 0 {
				fmt.Printf("  %s: %s\n", tierNames[i], strings.Join(cards, ", "))
			}
		}
	}

	// Find similar comps
	fmt.Printf("\n--- 相似流派 (共享卡牌>=2) ---\n")
	cards := compCards[target.Name]
	cardSet := make(map[string]bool)
	for _, c := range cards {
		cardSet[c] = true
	}

	type sim struct {
		name    string
		shared  int
		sharedC []string
	}
	var similarities []sim

	for compName, compCardsList := range compCards {
		if compName == target.Name {
			continue
		}
		var shared []string
		for _, c := range compCardsList {
			if cardSet[c] {
				shared = append(shared, c)
			}
		}
		if len(shared) >= 2 {
			similarities = append(similarities, sim{compName, len(shared), shared})
		}
	}

	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].shared > similarities[j].shared
	})

	for _, s := range similarities {
		fmt.Printf("  %s: 共享%d张 - %s\n", s.name, s.shared, strings.Join(s.sharedC, ", "))
	}
}

// ============================================================================
// Full analysis report
// ============================================================================

func runFullAnalysis(compCards map[string][]string, outputFile string) {
	fmt.Printf("\n============================================================\n")
	fmt.Printf("酒馆战旗 - 跨阵容综合分析报告\n")
	fmt.Printf("Season %d | %s | Generated by Cross-Analyzer\n", compsData.Season, compsData.LastUpdated)
	fmt.Printf("============================================================\n")

	// 1. Most versatile minions (appear in most comps)
	fmt.Printf("\n【1】通用性最强的随从 (跨流派最多)\n")
	fmt.Printf("------------------------------------------------------------\n")
	cardCompCount := make(map[string]int)
	cardCompNames := make(map[string][]string)
	for compName, cards := range compCards {
		for _, card := range cards {
			cardCompCount[card]++
			cardCompNames[card] = append(cardCompNames[card], compName)
		}
	}

	type versatile struct {
		card  string
		count int
		comps []string
		tier  int
	}
	var versatileList []versatile
	for card, count := range cardCompCount {
		versatileList = append(versatileList, versatile{
			card, count, cardCompNames[card], getMinionTier(card),
		})
	}
	sort.Slice(versatileList, func(i, j int) bool {
		return versatileList[i].count > versatileList[j].count
	})

	for i, v := range versatileList[:15] {
		cn := getCN(v.card)
		tierStr := ""
		if v.tier > 0 {
			tierStr = fmt.Sprintf(" [T%d]", v.tier)
		}
		fmt.Printf("  %2d. %s%s - 出现在%d个流派\n", i+1, v.card, tierStr, v.count)
		if cn != "" {
			fmt.Printf("      中文: %s\n", cn)
		}
	}

	// 2. Economy units
	fmt.Printf("\n【2】经济型随从 (跨流派经济支援)\n")
	fmt.Printf("------------------------------------------------------------\n")
	economyKeywords := []string{
		"Brann", "Drakkari", "Balinda", "Felfire", "Cataclysmic",
		"Brann Bronzebeard", "Drakkari Enchanter", "Balinda Stonehearth",
	}
	for _, kw := range economyKeywords {
		count := cardCompCount[kw]
		if count > 0 {
			cn := getCN(kw)
			fmt.Printf("  %s", kw)
			if cn != "" {
				fmt.Printf(" (%s)", cn)
			}
			fmt.Printf(" - %d个流派\n", count)
		}
	}

	// 3. Comp-tier overview
	fmt.Printf("\n【3】流派Tier概览\n")
	fmt.Printf("------------------------------------------------------------\n")
	tierGroups := make(map[string][]string)
	for _, comp := range compsData.Comps {
		tierGroups[comp.Tier] = append(tierGroups[comp.Tier], comp.Name)
	}
	tierOrder := []string{"S", "A", "B", "C"}
	for _, t := range tierOrder {
		names, ok := tierGroups[t]
		if !ok || len(names) == 0 {
			continue
		}
		fmt.Printf("  Tier %s (%d个):\n", t, len(names))
		for _, n := range names {
			fmt.Printf("    - %s\n", n)
		}
	}

	// 4. Transition opportunities (comps with most shared cards)
	fmt.Printf("\n【4】最容易转型的流派对 (共享>=3张卡)\n")
	fmt.Printf("------------------------------------------------------------\n")
	type transition struct {
		compA   string
		compB   string
		shared  int
		sharedC []string
	}
	var transitions []transition

	compNames := make([]string, 0, len(compCards))
	for name := range compCards {
		compNames = append(compNames, name)
	}

	for i := 0; i < len(compNames); i++ {
		for j := i + 1; j < len(compNames); j++ {
			cA, cB := compNames[i], compNames[j]
			cardsA := make(map[string]bool)
			for _, c := range compCards[cA] {
				cardsA[c] = true
			}
			var shared []string
			for _, c := range compCards[cB] {
				if cardsA[c] {
					shared = append(shared, c)
				}
			}
			if len(shared) >= 3 {
				transitions = append(transitions, transition{cA, cB, len(shared), shared})
			}
		}
	}

	sort.Slice(transitions, func(i, j int) bool {
		return transitions[i].shared > transitions[j].shared
	})

	count := 0
	for _, tr := range transitions {
		count++
		if count > 10 {
			break
		}
		fmt.Printf("\n  %s <-> %s (共享%d张)\n", tr.compA, tr.compB, tr.shared)
		for _, c := range tr.sharedC {
			cn := getCN(c)
			if cn != "" {
				fmt.Printf("    - %s (%s)\n", c, cn)
			} else {
				fmt.Printf("    - %s\n", c)
			}
		}
	}

	// 5. Tribe analysis
	fmt.Printf("\n【5】种族构成分析\n")
	fmt.Printf("------------------------------------------------------------\n")
	tribeComps := make(map[string]int)
	for _, comp := range compsData.Comps {
		tribe := extractTribe(comp.Name)
		tribeComps[tribe]++
	}
	var tribeList []struct{ tribe string; count int }
	for t, c := range tribeComps {
		tribeList = append(tribeList, struct{ tribe string; count int }{t, c})
	}
	sort.Slice(tribeList, func(i, j int) bool {
		return tribeList[i].count > tribeList[j].count
	})
	for _, t := range tribeList {
		tribeCN := tribeCN(t.tribe)
		fmt.Printf("  %s (%s): %d个流派\n", t.tribe, tribeCN, t.count)
	}

	// 6. Minion tier distribution
	fmt.Printf("\n【6】随从星级分布\n")
	fmt.Printf("------------------------------------------------------------\n")
	tierDist := make(map[int]int)
	for card := range cardCompCount {
		t := getMinionTier(card)
		tierDist[t]++
	}
	tierLabels := []string{"1星", "2星", "3星", "4星", "5星", "6星+"}
	for i, label := range tierLabels {
		cnt := tierDist[i+1]
		if i >= 5 {
			cnt = 0
			for t := 6; t <= 9; t++ {
				cnt += tierDist[t]
			}
		}
		if cnt > 0 {
			fmt.Printf("  %s: %d个随从\n", label, cnt)
		}
	}

	fmt.Printf("\n============================================================\n")
	fmt.Printf("分析完成 | 共%d个流派 | %d个独立随从\n", len(compsData.Comps), len(cardCompCount))
	fmt.Printf("============================================================\n")
}

// ============================================================================
// Helpers
// ============================================================================

func getCN(name string) string {
	if cn, ok := zhNames[name]; ok {
		return cn
	}
	return ""
}

func getMinionTier(name string) int {
	// Check all comps
	for _, comp := range compsData.Comps {
		for _, c := range comp.CoreCards {
			if c.Name == name {
				return c.Tier
			}
		}
	}
	// Check metadata
	if meta, ok := minionsMeta.Minions[name]; ok {
		return meta.Tier
	}
	return 0
}

func findSharedCardsAcrossComps(compNames []string, compCards map[string][]string) []string {
	if len(compNames) < 2 {
		return nil
	}
	// Start with cards from first comp
	shared := make(map[string]bool)
	first := true
	for _, compName := range compNames {
		cards, ok := compCards[compName]
		if !ok {
			continue
		}
		cardSet := make(map[string]bool)
		for _, c := range cards {
			cardSet[c] = true
		}
		if first {
			for c := range cardSet {
				shared[c] = true
			}
			first = false
		} else {
			for c := range shared {
				if !cardSet[c] {
					delete(shared, c)
				}
			}
		}
	}
	var result []string
	for c := range shared {
		result = append(result, c)
	}
	sort.Strings(result)
	return result
}

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
		return "Spell"
	}
	return "Other"
}

func tribeCN(tribe string) string {
	m := map[string]string{
		"Demon":   "恶魔",
		"Undead":  "亡灵",
		"Elemental": "元素",
		"Pirate":  "海盗",
		"Beast":   "野兽",
		"Murloc":  "鱼人",
		"Mech":    "机械",
		"Quilboar": "野猪人",
		"Naga":    "娜迦",
		"Dragon":  "龙",
		"Spell":   "法术",
		"Other":   "其他",
	}
	if cn, ok := m[tribe]; ok {
		return cn
	}
	return tribe
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
