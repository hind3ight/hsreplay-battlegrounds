package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Card struct {
	Name   string `json:"name"`
	Tier   *int   `json:"tier"`
	Attack *int   `json:"attack"`
	Health *int   `json:"health"`
}

type Comp struct {
	ID             any               `json:"id"`
	Name           string            `json:"name"`
	Tier           string            `json:"tier"`
	Difficulty     string            `json:"difficulty"`
	Description    string            `json:"description"`
	URL            string            `json:"url"`
	CoreCards      []Card           `json:"core_cards"`
	AddonCards     []Card           `json:"addon_cards"`
	HowToPlay      string           `json:"how_to_play"`
	WhenToCommit   string           `json:"when_to_commit"`
	Enablers       []any            `json:"enablers"`
	TavernTier     any              `json:"tavern_tier"`
	KeyTransition  map[string][]string `json:"key_transition"`
}

type Data struct {
	Summary   []any  `json:"summaries"`
	Comps     []Comp `json:"comps"`
	ScrapedAt string `json:"scraped_at"`
}

var ALL_TRIBES = []string{
	"恶魔", "亡灵", "元素", "海盗", "野兽",
	"鱼人", "机械", "野猪人", "娜迦", "龙",
}

// isGenericTerm 检查是否是通用描述词（非真实随从）
func isGenericTerm(name string) bool {
	genericTerms := []string{
		"溢出机制", "亡灵溢出流",
		"各种龙", "各种鱼人", "各种野兽",
		"娜迦+法术组合", "龙+战吼配合",
	}
	for _, term := range genericTerms {
		if name == term {
			return true
		}
	}
	return false
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

func loadData() (*Data, map[string]string, error) {
	compsData, err := os.ReadFile("data/season13_comps.json")
	if err != nil {
		return nil, nil, err
	}

	var data Data
	if err := json.Unmarshal(compsData, &data); err != nil {
		return nil, nil, err
	}

	namesData, err := os.ReadFile("data/minion_names.json")
	var zhNames map[string]string
	if err != nil {
		zhNames = make(map[string]string)
	} else {
		json.Unmarshal(namesData, &zhNames)
	}

	return &data, zhNames, nil
}

func selectTribes() []string {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║         酒馆战旗随从购买指南 - 种族选择                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("请选择本局可用的5个种族 (输入5个数字，空格分隔):")
	fmt.Println()

	for i, tribe := range ALL_TRIBES {
		fmt.Printf("  %2d. %s\n", i+1, tribe)
	}

	fmt.Println()
	fmt.Printf("当前选择: ")

	// 先检查环境变量
	var indices []int
	if envSelected := os.Getenv("BG_TRIBES"); envSelected != "" {
		for _, s := range strings.Split(envSelected, ",") {
			var n int
			fmt.Sscanf(s, "%d", &n)
			if n >= 1 && n <= len(ALL_TRIBES) {
				indices = append(indices, n-1)
			}
		}
	}

	// 如果环境变量无效，交互式读取
	if len(indices) != 5 {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		for _, s := range strings.Fields(input) {
			var n int
			fmt.Sscanf(s, "%d", &n)
			if n >= 1 && n <= len(ALL_TRIBES) {
				indices = append(indices, n-1)
			}
		}
	}

	if len(indices) != 5 {
		fmt.Printf("\n错误: 必须选择5个种族，你选择了 %d 个\n\n", len(indices))
		return selectTribes()
	}

	sort.Ints(indices)
	var result []string
	for _, i := range indices {
		result = append(result, ALL_TRIBES[i])
		fmt.Printf("%s ", ALL_TRIBES[i])
	}
	fmt.Println()

	return result
}

type MinionInfo struct {
	Name        string
	ZHName      string
	Tier        int
	CompCount   int
	ValidTribes []string
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateDoc(data *Data, zhNames map[string]string, selectedTribes []string) (string, error) {
	selectedSet := make(map[string]bool)
	for _, t := range selectedTribes {
		selectedSet[t] = true
	}

	// 收集可玩流派
	var availableComps []Comp
	for _, comp := range data.Comps {
		if selectedSet[extractTribe(comp.Name)] {
			availableComps = append(availableComps, comp)
		}
	}

	// 统计随从
	minionMap := make(map[string]*MinionInfo)
	for _, comp := range availableComps {
		compTribe := extractTribe(comp.Name)
		for _, card := range append(comp.CoreCards, comp.AddonCards...) {
			// 跳过通用描述词
			if isGenericTerm(card.Name) {
				continue
			}
			if _, ok := minionMap[card.Name]; !ok {
				tier := 0
				if card.Tier != nil {
					tier = *card.Tier
				}
				zh := card.Name
				if v, ok := zhNames[card.Name]; ok {
					zh = v
				} else {
					zh = "[待确认] " + card.Name
				}
				minionMap[card.Name] = &MinionInfo{
					Name:        card.Name,
					ZHName:      zh,
					Tier:        tier,
					ValidTribes: []string{},
				}
			}
			minionMap[card.Name].CompCount++
			if !contains(minionMap[card.Name].ValidTribes, compTribe) {
				minionMap[card.Name].ValidTribes = append(minionMap[card.Name].ValidTribes, compTribe)
			}
		}
	}

	// 按评分排序
	type minionEntry struct {
		Name        string
		ZHName      string
		Tier        int
		CompCount   int
		ValidTribes []string
	}
	var minions []minionEntry
	for _, m := range minionMap {
		minions = append(minions, minionEntry{
			Name:        m.Name,
			ZHName:      m.ZHName,
			Tier:        m.Tier,
			CompCount:   m.CompCount,
			ValidTribes: m.ValidTribes,
		})
	}
	sort.Slice(minions, func(i, j int) bool {
		if minions[i].CompCount != minions[j].CompCount {
			return minions[i].CompCount > minions[j].CompCount
		}
		return minions[i].Tier < minions[j].Tier
	})

	// 按酒馆等级分组
	type tierGroup struct {
		Min  int
		Max  int
		List []minionEntry
	}
	var tierGroups []tierGroup
	for t := 1; t <= 6; t += 2 {
		tg := tierGroup{Min: t, Max: min(t+1, 6)}
		for _, m := range minions {
			if m.Tier >= tg.Min && m.Tier <= tg.Max {
				tg.List = append(tg.List, m)
			}
		}
		if len(tg.List) > 0 {
			tierGroups = append(tierGroups, tg)
		}
	}

	// 生成 Markdown
	var sb strings.Builder
	sb.WriteString("# 酒馆战旗随从购买指南\n\n")
	sb.WriteString(fmt.Sprintf("> 生成时间: %s<br>\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("> 本局种族: %s\n\n", strings.Join(selectedTribes, " / ")))

	// 流派列表
	sb.WriteString("## 可玩流派\n\n")
	sb.WriteString("| 流派 | 核心随从 | 特点 |\n")
	sb.WriteString("|------|----------|------|\n")
	for _, comp := range availableComps {
		var coreCards []string
		for _, c := range comp.CoreCards {
			zh := c.Name
			if v, ok := zhNames[c.Name]; ok {
				zh = v
			}
			coreCards = append(coreCards, zh)
		}
		coreStr := strings.Join(coreCards, ", ")
		if len(coreStr) > 50 {
			coreStr = coreStr[:47] + "..."
		}
		diff := comp.Difficulty
		if diff == "" {
			diff = "-"
		}
		sb.WriteString(fmt.Sprintf("| **%s** | %s | %s |\n", comp.Name, coreStr, diff))
	}

	// 随从购买优先级
	sb.WriteString("\n## 随从购买优先级\n\n")
	sb.WriteString("> 按流派覆盖率排序，流派覆盖率相同按酒馆等级排序\n\n")

	for _, tg := range tierGroups {
		sb.WriteString(fmt.Sprintf("### Tier %d-%d\n\n", tg.Min, tg.Max))
		sb.WriteString("| 随从 | Tier | 覆盖率 | 适用种族 |\n")
		sb.WriteString("|------|------|--------|----------|\n")
		for _, m := range tg.List {
			tribes := strings.Join(m.ValidTribes, "/")
			sb.WriteString(fmt.Sprintf("| **%s** | %d | %d | %s |\n",
				m.ZHName, m.Tier, m.CompCount, tribes))
		}
		sb.WriteString("\n")
	}

	// 核心随从详情
	sb.WriteString("## 核心随从详解\n\n")
	for _, m := range minions {
		if m.CompCount >= 2 {
			sb.WriteString(fmt.Sprintf("### %s\n\n", m.ZHName))
			sb.WriteString(fmt.Sprintf("- **酒馆等级**: Tier %d\n", m.Tier))
			sb.WriteString(fmt.Sprintf("- **流派覆盖率**: %d/%d 个可选流派\n", m.CompCount, len(availableComps)))
			sb.WriteString(fmt.Sprintf("- **适用种族**: %s\n", strings.Join(m.ValidTribes, ", ")))

			var compDetails []string
			for _, comp := range availableComps {
				for _, card := range append(comp.CoreCards, comp.AddonCards...) {
					if isGenericTerm(card.Name) {
						continue
					}
					if card.Name == m.Name {
						compDetails = append(compDetails, comp.Name)
						break
					}
				}
			}
			sb.WriteString(fmt.Sprintf("- **使用流派**: %s\n\n", strings.Join(compDetails, ", ")))
		}
	}

	// 转型路线
	sb.WriteString("## 流派转型建议\n\n")
	for _, comp := range availableComps {
		if len(comp.KeyTransition) > 0 {
			sb.WriteString(fmt.Sprintf("### %s\n\n", comp.Name))
			for from, to := range comp.KeyTransition {
				sb.WriteString(fmt.Sprintf("- **%s** → %s\n", from, strings.Join(to, " → ")))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func main() {
	// 默认输出到 Obsidian 笔记目录
	defaultOutputDir := filepath.Join(
		os.Getenv("HOME"),
		"work", "Projects", "GoProjects", "src", "github.com", "hind3ight",
		"Obsidian-notes", "raw", "battlegrounds",
	)

	outputDir := flag.String("o", defaultOutputDir, "输出目录")
	flag.Parse()

	data, zhNames, err := loadData()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading data: %v\n", err)
		os.Exit(1)
	}

	selectedTribes := selectTribes()

	// 生成文档
	doc, err := generateDoc(data, zhNames, selectedTribes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating doc: %v\n", err)
		os.Exit(1)
	}

	// 确定输出路径
	tribeFileName := strings.Join(selectedTribes, "-") + ".md"
	outputPath := filepath.Join(*outputDir, tribeFileName)

	// 检查文件是否存在
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("\n📄 文档已存在，跳过更新: %s\n", outputPath)
		fmt.Printf("   如需重新生成，请先删除该文件\n")
		os.Exit(0)
	}

	// 确保输出目录存在
	os.MkdirAll(*outputDir, 0755)

	// 写入文件
	if err := os.WriteFile(outputPath, []byte(doc), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ 文档已生成: %s\n", outputPath)
	fmt.Printf("   大小: %d bytes\n", len(doc))
}
