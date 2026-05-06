package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// resolveDataFile 尝试在多个位置查找 data/season13_comps.json：
// 1. 当前工作目录 (data/season13_comps.json)
// 2. 可执行文件所在目录的 data/ (调试时从 bin/ 运行)
// 3. 源码目录的 data/ (IDE 调试)
func resolveDataFile(name string) (string, error) {
	paths := []string{
		name,                              // 相对 CWD
		filepath.Join("..", name),         // 相对 bin/
		filepath.Join("..", "..", name),   // 相对 cmd/analyze/
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("file not found in any of: %v", paths)
}

type Card struct {
	Name   string `json:"name"`
	Tier   *int   `json:"tier"`
	Attack *int   `json:"attack"`
	Health *int   `json:"health"`
}

type Comp struct {
	ID            any               `json:"id"`
	Name          string            `json:"name"`
	Tier          string            `json:"tier"`
	Difficulty    string            `json:"difficulty"`
	Description   string            `json:"description"`
	URL           string            `json:"url"`
	CoreCards     []Card           `json:"core_cards"`
	AddonCards    []Card           `json:"addon_cards"`
	HowToPlay     string           `json:"how_to_play"`
	WhenToCommit  string           `json:"when_to_commit"`
	Enablers      []any            `json:"enablers"`
	TavernTier    any              `json:"tavern_tier"`
	KeyTransition map[string][]string `json:"key_transition"`
}

type Data struct {
	Summary   []any  `json:"summaries"`
	Comps     []Comp `json:"comps"`
	ScrapedAt string `json:"scraped_at"`
}

type MinionInfo struct {
	Tier      *int
	Count     int
	Comps     []string
	Tribes    map[string]bool
}

func extractTribe(compName string) string {
	nameLower := strings.ToLower(compName)
	if strings.Contains(nameLower, "demon") {
		return "恶魔"
	} else if strings.Contains(nameLower, "undead") {
		return "亡灵"
	} else if strings.Contains(nameLower, "elemental") {
		return "元素"
	} else if strings.Contains(nameLower, "pirate") || strings.Contains(nameLower, "bounty") {
		return "海盗"
	} else if strings.Contains(nameLower, "beast") {
		return "野兽"
	} else if strings.Contains(nameLower, "murloc") {
		return "鱼人"
	} else if strings.Contains(nameLower, "mech") {
		return "机械"
	} else if strings.Contains(nameLower, "quilboar") {
		return "野猪人"
	} else if strings.Contains(nameLower, "naga") {
		return "娜迦"
	} else if strings.Contains(nameLower, "dragon") {
		return "龙"
	}
	return "其他"
}

func main() {
	// 加载数据
	compsFile, err := resolveDataFile("data/season13_comps.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding data file: %v\n", err)
		os.Exit(1)
	}
	compsData, err := os.ReadFile(compsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data file: %v\n", err)
		os.Exit(1)
	}

	var data Data
	if err := json.Unmarshal(compsData, &data); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// 加载中英文对照
	namesData, err := os.ReadFile("data/minion_names.json")
	var zhNames map[string]string
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load minion_names.json: %v\n", err)
		zhNames = make(map[string]string)
	} else {
		json.Unmarshal(namesData, &zhNames)
	}

	// 统计随从跨种族情况
	minionMap := make(map[string]*MinionInfo)
	for _, comp := range data.Comps {
		tribe := extractTribe(comp.Name)
		for _, card := range append(comp.CoreCards, comp.AddonCards...) {
			if _, ok := minionMap[card.Name]; !ok {
				minionMap[card.Name] = &MinionInfo{
					Tier:   card.Tier,
					Tribes: make(map[string]bool),
				}
			}
			minionMap[card.Name].Count++
			minionMap[card.Name].Tribes[tribe] = true
			minionMap[card.Name].Comps = append(minionMap[card.Name].Comps, comp.Name)
		}
	}

	// 收集并按种族跨度排序
	type minionEntry struct {
		Name      string
		Info      *MinionInfo
		TribeSpan int
	}
	var entries []minionEntry
	for name, info := range minionMap {
		entries = append(entries, minionEntry{
			Name:      name,
			Info:      info,
			TribeSpan: len(info.Tribes),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].TribeSpan != entries[j].TribeSpan {
			return entries[i].TribeSpan > entries[j].TribeSpan
		}
		return entries[i].Info.Count > entries[j].Info.Count
	})

	// 输出
	fmt.Println("============================================================")
	fmt.Println("酒馆战旗随从购买指南 (考虑5种族池限制)")
	fmt.Println("============================================================")
	fmt.Println()
	fmt.Println("【核心思路】")
	fmt.Println("每局只有5个种族可用。选择跨种族多的随从，")
	fmt.Println("可以在种族池确定后灵活转型。")
	fmt.Println()
	fmt.Println("【术语】")
	fmt.Println("- 跨种族数: 该随从能在几个种族中使用")
	fmt.Println("- 流派数: 使用该随从的流派数量")
	fmt.Println()

	// 按跨种族数分组输出
	printSection := func(title string, minTribeSpan, maxTribeSpan int) {
		fmt.Println("============================================================")
		fmt.Printf("【%s】\n", title)
		fmt.Println("============================================================")
		for _, e := range entries {
			if e.TribeSpan < minTribeSpan || e.TribeSpan > maxTribeSpan {
				continue
			}
			tier := 0
			if e.Info.Tier != nil {
				tier = *e.Info.Tier
			}
			zh := e.Name
			if v, ok := zhNames[e.Name]; ok {
				zh = v
			} else {
				zh = "[待确认] " + e.Name
			}

			// 收集种族名
			var tribes []string
			for t := range e.Info.Tribes {
				tribes = append(tribes, t)
			}
			sort.Strings(tribes)

			fmt.Printf("\n%s (Tier %d)\n", zh, tier)
			fmt.Printf("  跨种族(%d): %s\n", e.TribeSpan, strings.Join(tribes, "/"))
			fmt.Printf("  流派数: %d\n", e.Info.Count)
			fmt.Printf("  可转: %s\n", strings.Join(e.Info.Comps, ", "))
		}
		fmt.Println()
	}

	printSection("T0 必拿 - 跨5种族", 5, 999)
	printSection("T1 推荐 - 跨4种族", 4, 4)
	printSection("T2 备选 - 跨3种族", 3, 3)
	printSection("T3 特定组合 - 跨2种族", 2, 2)
}
