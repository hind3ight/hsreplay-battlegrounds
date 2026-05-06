package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/hind3ight/hsreplay-battlegrounds/data"
	"github.com/hind3ight/hsreplay-battlegrounds/models"
)

func main() {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	tavernCmd := flag.NewFlagSet("tavern", flag.ExitOnError)
	detailCmd := flag.NewFlagSet("detail", flag.ExitOnError)

	listTier := listCmd.String("tier", "", "Filter by tier (S, A, B)")
	tavernTier := tavernCmd.Int("tier", 0, "Show comps that have key units at this tavern tier (1-5)")
	detailName := detailCmd.String("name", "", "Comp name to show detail")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	seasonData := data.LoadSeasonData()

	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		listComps(seasonData.Comps, *listTier)

	case "tavern":
		tavernCmd.Parse(os.Args[2:])
		showTavernGuides(seasonData.Comps, *tavernTier)

	case "detail":
		detailCmd.Parse(os.Args[2:])
		showDetail(seasonData.Comps, *detailName)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`HSReplay Battlegrounds 流派查询工具 (Season 13)
==============================================

用法:
  hsreplay-battlegrounds list              列出所有流派
  hsreplay-battlegrounds list --tier S   只显示 S 级流派
  hsreplay-battlegrounds tavern --tier 3 显示 3 本酒馆等级的关键流派
  hsreplay-battlegrounds detail --name 恶魔  显示恶魔流派的详细信息

选项:
  list:
    --tier S|A|B  按强度筛选流派

  tavern:
    --tier 1-5   显示在该酒馆等级有核心随从的流派

  detail:
    --name <名称> 显示特定流派的详细攻略
`)
}

func listComps(comps []models.Comp, tierFilter string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Tier\t流派\t难度\t描述\n")
	fmt.Fprintf(w, "----\t----\t----\t----\n")

	for _, comp := range comps {
		if tierFilter != "" && comp.Tier != tierFilter {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", comp.Tier, comp.Name, comp.Difficulty, comp.Description)
	}
	w.Flush()
}

func showTavernGuides(comps []models.Comp, tierFilter int) {
	if tierFilter < 1 || tierFilter > 5 {
		// Show all tavern guides
		showAllTavernGuides(comps)
		return
	}

	// Show which comps have key units at this tier
	fmt.Printf("\n=== 酒馆等级 %d 的关键流派 ===\n\n", tierFilter)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "流派\tTier\t难度\t该等级关键随从\n")
	fmt.Fprintf(w, "----\t----\t----\t----------\n")

	tierKey := []string{"OneStar", "TwoStar", "ThreeStar", "FourStar", "FiveStar"}
	key := tierKey[tierFilter-1]

	for _, comp := range comps {
		var units []string
		switch key {
		case "OneStar":
			units = comp.KeyTransition.OneStar
		case "TwoStar":
			units = comp.KeyTransition.TwoStar
		case "ThreeStar":
			units = comp.KeyTransition.ThreeStar
		case "FourStar":
			units = comp.KeyTransition.FourStar
		case "FiveStar":
			units = comp.KeyTransition.FiveStar
		}

		if len(units) > 0 {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", comp.Name, comp.Tier, comp.Difficulty, strings.Join(units, ", "))
		}
	}
	w.Flush()
}

func showAllTavernGuides(comps []models.Comp) {
	for _, comp := range comps {
		fmt.Printf("\n=== %s (%s级) ===\n", comp.Name, comp.Tier)
		fmt.Printf("难度: %s | %s\n\n", comp.Difficulty, comp.Description)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "酒馆等级\t关键随从\n")
		fmt.Fprintf(w, "--------\t--------\n")
		fmt.Fprintf(w, "1星\t%s\n", strings.Join(comp.KeyTransition.OneStar, ", "))
		fmt.Fprintf(w, "2星\t%s\n", strings.Join(comp.KeyTransition.TwoStar, ", "))
		fmt.Fprintf(w, "3星\t%s\n", strings.Join(comp.KeyTransition.ThreeStar, ", "))
		fmt.Fprintf(w, "4星\t%s\n", strings.Join(comp.KeyTransition.FourStar, ", "))
		fmt.Fprintf(w, "5星\t%s\n", strings.Join(comp.KeyTransition.FiveStar, ", "))
		w.Flush()
		fmt.Println()
	}
}

func showDetail(comps []models.Comp, nameFilter string) {
	if nameFilter == "" {
		fmt.Println("请使用 --name 指定流派名称")
		return
	}

	nameFilter = strings.ToLower(nameFilter)

	for _, comp := range comps {
		if strings.Contains(strings.ToLower(comp.Name), nameFilter) {
			printCompDetail(comp)
			return
		}
	}

	fmt.Printf("未找到流派: %s\n", nameFilter)
}

func printCompDetail(comp models.Comp) {
	fmt.Printf("\n=== %s ===\n", comp.Name)
	fmt.Printf("强度: %s | 难度: %s\n", comp.Tier, comp.Difficulty)
	fmt.Printf("描述: %s\n", comp.Description)
	fmt.Printf("URL: %s\n\n", comp.URL)

	// Core cards
	fmt.Printf("核心随从:\n")
	for _, card := range comp.CoreCards {
		if card.Tier > 0 {
			fmt.Printf("  [%d星] %s (%d/%d)\n", card.Tier, card.Name, card.Attack, card.Health)
		} else {
			fmt.Printf("  %s\n", card.Name)
		}
	}
	fmt.Println()

	// Addon cards
	if len(comp.AddonCards) > 0 {
		fmt.Printf("扩展随从:\n")
		for _, card := range comp.AddonCards {
			fmt.Printf("  %s\n", card.Name)
		}
		fmt.Println()
	}

	// Tavern transition guide
	fmt.Printf("酒馆等级转型指南:\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "酒馆等级\t关键随从\n")
	fmt.Fprintf(w, "--------\t--------\n")
	fmt.Fprintf(w, "1星\t%s\n", strings.Join(comp.KeyTransition.OneStar, ", "))
	fmt.Fprintf(w, "2星\t%s\n", strings.Join(comp.KeyTransition.TwoStar, ", "))
	fmt.Fprintf(w, "3星\t%s\n", strings.Join(comp.KeyTransition.ThreeStar, ", "))
	fmt.Fprintf(w, "4星\t%s\n", strings.Join(comp.KeyTransition.FourStar, ", "))
	fmt.Fprintf(w, "5星\t%s\n", strings.Join(comp.KeyTransition.FiveStar, ", "))
	w.Flush()
	fmt.Println()

	// How to play
	if comp.HowToPlay != "" {
		fmt.Printf("玩法:\n%s\n\n", comp.HowToPlay)
	}

	// When to commit
	if comp.WhenToCommit != "" {
		fmt.Printf("转型时机:\n%s\n\n", comp.WhenToCommit)
	}
}
