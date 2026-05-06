package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// PlaywrightFetcher uses browser automation to fetch HSReplay data
type PlaywrightFetcher struct {
	browser playwright.Browser
	page    playwright.Page
}

// NewPlaywrightFetcher creates a new Playwright-based fetcher
func NewPlaywrightFetcher() (*PlaywrightFetcher, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to launch playwright: %w\n"+
			"Please install browsers: ~/go/bin/playwright install chromium", err)
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	context, err := browser.NewContext()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	page, err := context.NewPage()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	context.SetDefaultTimeout(60000)

	return &PlaywrightFetcher{
		browser: browser,
		page:    page,
	}, nil
}

// Close releases browser resources
func (f *PlaywrightFetcher) Close() error {
	if f.page != nil {
		f.page.Close()
	}
	if f.browser != nil {
		return f.browser.Close()
	}
	return nil
}

// FetchCompList fetches all comp summaries from the main page
func (f *PlaywrightFetcher) FetchCompList(ctx context.Context) ([]RawCompSummary, error) {
	timeout := 60000.0
	_, err := f.page.Goto("https://hsreplay.net/battlegrounds/comps/", playwright.PageGotoOptions{Timeout: &timeout})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	f.page.WaitForSelector("main")
	f.page.WaitForTimeout(2000)

	result, err := f.page.Evaluate(`() => {
		const results = [];
		const seen = new Set();
		const baseUrl = 'https://hsreplay.net';

		// Try multiple selectors to find comp links
		const selectors = [
			'main a[href*="/battlegrounds/comps/"]',
			'a[href*="/battlegrounds/comps/"]',
			'.comp-card a[href*="/battlegrounds/comps/"]',
			'.comps-list a[href*="/battlegrounds/comps/"]',
		];

		for (const sel of selectors) {
			document.querySelectorAll(sel).forEach(a => {
				const href = a.getAttribute('href');
				if (!href) return;
				const parts = href.split('/').filter(Boolean);

				if (parts.length >= 4 && /^\d+$/.test(parts[2])) {
					const id = parseInt(parts[2]);
					if (seen.has(id)) return;
					seen.add(id);

					const slug = parts[3];
					const name = slug.replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
					const fullText = (a.textContent || '').replace(/\s+/g, ' ').trim();

					const diffMatch = fullText.match(/(Easy|Medium|Hard)\s*$/);
					const difficulty = diffMatch ? diffMatch[1] : '';

					let description = fullText.replace(/BG[\w_]+\s+\d+\s+\d+\s+Tier/gi, '');
					description = description.replace(/(Easy|Medium|Hard)$/g, '').trim();

					const compNameMatch = name.replace(/-/g, ' ');
					description = description.replace(new RegExp('^' + compNameMatch + '\\s*', 'i'), '');
					description = description.replace(/\s+/g, ' ').trim();

					const fullUrl = baseUrl + (href.startsWith('/') ? href : '/' + href);

					if (id && name) {
						results.push({ id, name, description, difficulty, url: fullUrl });
					}
				}
			});
			if (results.length > 0) break;
		}

		return JSON.stringify(results);
	}`)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate JS: %w", err)
	}

	return parseCompSummaries(result)
}

// FetchCompDetail fetches detailed info for a specific comp
func (f *PlaywrightFetcher) FetchCompDetail(ctx context.Context, compURL string) (*RawCompDetail, error) {
	f.page.WaitForTimeout(500)

	_, err := f.page.Goto(compURL)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	f.page.WaitForTimeout(1000)

	result, err := f.page.Evaluate(`() => {
		const getSectionContent = (headingText) => {
			const headings = [...document.querySelectorAll('h2, h3, h4')];
			const heading = headings.find(h => h.textContent.includes(headingText));
			if (heading) {
				const next = heading.nextElementSibling;
				return next ? next.textContent.replace(/\s+/g, ' ').trim() : '';
			}
			return '';
		};

		const getCardsFromSection = (sectionHeading) => {
			const headings = [...document.querySelectorAll('h2, h3, h4')];
			const heading = headings.find(h => h.textContent.includes(sectionHeading));
			if (!heading) return [];
			
			const cards = [];
			let sibling = heading.nextElementSibling;
			let count = 0;
			while (sibling && !['H2', 'H3', 'H4'].includes(sibling.tagName) && count < 10) {
				sibling.querySelectorAll('a[href*="/battlegrounds/minions/"]').forEach(a => {
					const href = a.getAttribute('href');
					const name = a.textContent.replace(/\s+/g, ' ').trim();
					if (name && !cards.find(c => c.name === name)) {
						const tierMatch = href.match(/\/tier-(\d+)\//);
						const tier = tierMatch ? parseInt(tierMatch[1]) : 0;
						cards.push({ name, tier, url: href });
					}
				});
				sibling = sibling.nextElementSibling;
				count++;
			}
			return cards;
		};

		// Extract tavern tier guidance from key_transition if present
		const getTavernTier = () => {
			const result = {};
			const headings = [...document.querySelectorAll('h2, h3, h4')];
			const ttHeading = headings.find(h => 
				h.textContent.includes('Key Transition') || 
				h.textContent.includes('Tavern Tier') ||
				h.textContent.includes('Turn × Star')
			);
			if (!ttHeading) return result;
			
			let sibling = ttHeading.nextElementSibling;
			let count = 0;
			while (sibling && !['H2', 'H3', 'H4'].includes(sibling.tagName) && count < 5) {
				const tierMatch = sibling.textContent.match(/(?:Star|Tier)\s*[-–]?\s*(\d)/i);
				if (tierMatch) {
					const tier = 'tier_' + tierMatch[1];
					const names = [];
					sibling.querySelectorAll('a[href*="/battlegrounds/minions/"]').forEach(a => {
						names.push(a.textContent.replace(/\s+/g, ' ').trim());
					});
					if (names.length > 0) {
						result[tier] = names;
					}
				}
				sibling = sibling.nextElementSibling;
				count++;
			}
			return result;
		};

		const urlMatch = window.location.pathname.match(/\/battlegrounds\/comps\/(\d+)\//);
		const id = urlMatch ? urlMatch[1] : '';

		const data = {
			id,
			name: document.querySelector('h1')?.textContent.replace(/\s+/g, ' ').trim().replace(' Comp', '') || '',
			difficulty: document.body.textContent.includes('Medium') ? 'Medium' : 
			           document.body.textContent.includes('Easy') ? 'Easy' : 
			           document.body.textContent.includes('Hard') ? 'Hard' : 'Unknown',
			core_cards: getCardsFromSection('Core Cards'),
			addon_cards: getCardsFromSection('Addon Cards'),
			enablers: getCardsFromSection('Common Enablers'),
			how_to_play: getSectionContent('How to Play') || getSectionContent('How to Play this Comp'),
			when_to_commit: getSectionContent('When to Commit') || getSectionContent('When to Commit this Comp'),
			tavern_tier: getTavernTier(),
		};

		return JSON.stringify(data);
	}`)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate JS: %w", err)
	}

	return parseCompDetail(result)
}

// FetchMinionDetail fetches detailed stats for a specific minion from its page
func (f *PlaywrightFetcher) FetchMinionDetail(ctx context.Context, minionURL string) (*RawMinionDetail, error) {
	f.page.WaitForTimeout(300)

	_, err := f.page.Goto(minionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	f.page.WaitForTimeout(800)

	result, err := f.page.Evaluate(`() => {
		const text = document.body.textContent;
		
		// Extract tribe info using JS regex
		const tribeMatch = text.match(/(Beast|Demon|Undead|Elemental|Dragon|Murloc|Mech|Pirate|Quilboar|Naga)/i);
		let tribe = tribeMatch ? tribeMatch[0] : 'None';
		if (tribe) {
			tribe = tribe.charAt(0).toUpperCase() + tribe.slice(1).toLowerCase();
		}
		
		// Extract attack/health from stats boxes
		const statsMatch = text.match(/Attack\s*[:\-]?\s*(\d+).*?Health\s*[:\-]?\s*(\d+)/s);
		let attack = 0, health = 0;
		if (statsMatch) {
			attack = parseInt(statsMatch[1]) || 0;
			health = parseInt(statsMatch[2]) || 0;
		}
		
		// Try to get exact stats from prominent display
		const atkEl = document.querySelector('[data-testid="attack"]');
		const hpEl = document.querySelector('[data-testid="health"]');
		if (atkEl) attack = parseInt(atkEl.textContent) || attack;
		if (hpEl) health = parseInt(hpEl.textContent) || health;

		// Extract tier from URL
		const tierMatch = window.location.pathname.match(/tier-(\d+)/);
		const tier = tierMatch ? parseInt(tierMatch[1]) : 0;

		// Extract abilities/text
		const abilities = [];
		document.querySelectorAll('[class*="ability"], [class*="text"], [class*="description"]').forEach(el => {
			const t = el.textContent.replace(/\s+/g, ' ').trim();
			if (t.length > 5 && t.length < 300 && !abilities.includes(t)) {
				abilities.push(t);
			}
		});

		return JSON.stringify({
			tribe,
			attack,
			health,
			tier,
			abilities: abilities.slice(0, 3),
		});
	}`)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate JS for minion: %w", err)
	}

	return parseMinionDetail(result)
}

// FetchAllComps fetches all comps with details
func (f *PlaywrightFetcher) FetchAllComps(ctx context.Context) (*ScrapedResult, error) {
	summaries, err := f.FetchCompList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comp list: %w", err)
	}

	fmt.Printf("Found %d comps\n", len(summaries))

	result := &ScrapedResult{
		Summaries: summaries,
		Comps:     make([]RawCompDetail, 0, len(summaries)),
		ScrapedAt: time.Now().Format(time.RFC3339),
	}

	for i, summary := range summaries {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		fmt.Printf("[%d/%d] Fetching: %s\n", i+1, len(summaries), summary.Name)

		detail, err := f.FetchCompDetail(ctx, summary.URL)
		if err != nil {
			fmt.Printf("  Warning: failed to fetch detail for %s: %v\n", summary.Name, err)
			continue
		}
		result.Comps = append(result.Comps, *detail)
	}

	return result, nil
}

// FetchAllMinionStats fetches stats for all minions mentioned in scraped comps
func (f *PlaywrightFetcher) FetchAllMinionStats(ctx context.Context, comps []RawCompDetail) (map[string]*RawMinionDetail, error) {
	// Collect unique minion URLs
	minionURLs := make(map[string]string)
	for _, comp := range comps {
		for _, card := range comp.CoreCards {
			if card.URL != "" {
				minionURLs[card.Name] = card.URL
			}
		}
		for _, card := range comp.AddonCards {
			if card.URL != "" {
				minionURLs[card.Name] = card.URL
			}
		}
		for _, card := range comp.Enablers {
			if card.URL != "" {
				minionURLs[card.Name] = card.URL
			}
		}
	}

	results := make(map[string]*RawMinionDetail)
	var names []string
	for name := range minionURLs {
		names = append(names, name)
	}

	fmt.Printf("Fetching stats for %d unique minions...\n", len(names))

	for i, name := range names {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		url := minionURLs[name]
		detail, err := f.FetchMinionDetail(ctx, url)
		if err != nil {
			fmt.Printf("  [%d/%d] Warning: failed to fetch %s: %v\n", i+1, len(names), name, err)
			// Fill with basic data from comp
			results[name] = &RawMinionDetail{Name: name}
			continue
		}
		detail.Name = name
		results[name] = detail
		fmt.Printf("  [%d/%d] %s (Tier %d, %s)\n", i+1, len(names), name, detail.Tier, detail.Tribe)
	}

	return results, nil
}

func parseCompSummaries(result interface{}) ([]RawCompSummary, error) {
	jsonStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	var comps []RawCompSummary
	if err := json.Unmarshal([]byte(jsonStr), &comps); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return comps, nil
}

func parseCompDetail(result interface{}) (*RawCompDetail, error) {
	jsonStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	var comp RawCompDetail
	if err := json.Unmarshal([]byte(jsonStr), &comp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &comp, nil
}

func parseMinionDetail(result interface{}) (*RawMinionDetail, error) {
	jsonStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	var minion RawMinionDetail
	if err := json.Unmarshal([]byte(jsonStr), &minion); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &minion, nil
}

// NormalizeMinionName cleans up inconsistent minion name spellings
func NormalizeMinionName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "  ", " ")
	return name
}