package data

import (
	"encoding/json"
	"strings"

	"github.com/hind3ight/hsreplay-battlegrounds/models"
)

var seasonData *models.SeasonData

// LoadSeasonData loads the season data
func LoadSeasonData() *models.SeasonData {
	if seasonData != nil {
		return seasonData
	}

	seasonData = &models.SeasonData{
		Season:      13,
		LastUpdated: "2025-05-04",
		Source:      "https://hsreplay.net/battlegrounds/comps/",
		PoweredBy:   "JeefHS",
		Comps:       loadComps(),
	}

	return seasonData
}

func loadComps() []models.Comp {
	jsonData := `{
  "season": 13,
  "last_updated": "2025-05-04",
  "source": "https://hsreplay.net/battlegrounds/comps/",
  "powered_by": "JeefHS",
  "comps": [
    {
      "id": 41,
      "name": "Demons - Shop Buff",
      "tier": "S",
      "difficulty": "Medium",
      "description": "Scale and consume shop",
      "url": "https://hsreplay.net/battlegrounds/comps/41/demons-shop-buff",
      "core_cards": [
        {"name": "Brann Bronzebeard", "tier": 2, "attack": 4, "health": 4},
        {"name": "Twisted Wrathguard", "tier": 4, "attack": 4, "health": 4},
        {"name": "Ashen Corruptor", "tier": 6, "attack": 6, "health": 6},
        {"name": "Malchezaar, Prince of Dance", "tier": 5, "attack": 4, "health": 4},
        {"name": "Consummate Conqueror", "tier": 9, "attack": 7, "health": 7}
      ],
      "addon_cards": [
        {"name": "Batty Terrorguard"},
        {"name": "Soul Rewinder"},
        {"name": "Laboratory Assistant"},
        {"name": "Rylak Metalhead"},
        {"name": "Famished Felbat"}
      ],
      "how_to_play": "Scale your shop somehow, usually through Shadowdancer, Felemental, or Consummate Conqueror. Then add fodders to your shop to capture the buffed minions. Typically you have an economy setup with Brann Bronzebeard + damage rewinder (Ashen Corruptor) which allows you to cycle a lot of cards. Every card you cycle will add a lot of fodders to eat because of Twisted Wrathguard. There's another line where you can add fodders to your shop using Rylak Metalhead + Laboratory Assistant.",
      "when_to_commit": "Buffed Shop (Shadowdancer) + Ashen Corruptor + Brann Bronzebeard",
      "key_transition": {
        "1_star": ["Brann Bronzebeard"],
        "2_star": ["Brann Bronzebeard", "Twisted Wrathguard"],
        "3_star": ["Ashen Corruptor", "Rylak Metalhead"],
        "4_star": ["Malchezaar, Prince of Dance"],
        "5_star": ["Consummate Conqueror"]
      }
    },
    {
      "id": 14,
      "name": "Undead - Attack Scaling",
      "tier": "S",
      "difficulty": "Medium",
      "description": "Summon high attack Undeads",
      "url": "https://hsreplay.net/battlegrounds/comps/14/undead-attack-scaling",
      "core_cards": [
        {"name": "Handless Forsaken", "tier": 2, "attack": 1, "health": 1},
        {"name": "Friendly Geist", "tier": 6, "attack": 3, "health": 3},
        {"name": "Balinda Stonehearth", "tier": 6, "attack": 6, "health": 6},
        {"name": "Drustfallen Butcher", "tier": 2, "attack": 9, "health": 9}
      ],
      "addon_cards": [
        {"name": "Tranquil Meditative"},
        {"name": "Eternal Summoner"},
        {"name": "Mummifier"},
        {"name": "Titus Rivendare"},
        {"name": "Forsaken Weaver"},
        {"name": "Wintergrasp Ghoul"},
        {"name": "Nightbane, Ignited"},
        {"name": "Deathly Striker"}
      ],
      "how_to_play": "Scale the attack of your Undeads and summons with the Butchering spell from Drustfallen Butcher. Butchering doubles its effect when paired with Balinda Stonehearth, and also improves with spell scaling from Friendly Geist or Tranquil Meditative. Your quality of summons come from Handless Forsaken, but also could use Deathly Striker or Eternal Summoner. Be sure to reborn as many key undeads as possible, starting with Handless Forsaken.",
      "when_to_commit": "Attack Scaling (Drustfallen Butcher) + summons (Handless Forsaken)",
      "key_transition": {
        "1_star": ["Handless Forsaken"],
        "2_star": ["Drustfallen Butcher", "Handless Forsaken"],
        "3_star": ["Friendly Geist"],
        "4_star": ["Balinda Stonehearth"],
        "5_star": ["Eternal Summoner", "Deathly Striker"]
      }
    },
    {
      "id": 82,
      "name": "Back to Back",
      "tier": "S",
      "difficulty": "Medium",
      "description": "Scale and repeatedly play specific tavern spell",
      "url": "https://hsreplay.net/battlegrounds/comps/82/back-to-back",
      "core_cards": [
        {"name": "Felfire Conjurer", "tier": 6, "attack": 5, "health": 5},
        {"name": "Drakkari Enchanter", "tier": 1, "attack": 5, "health": 5},
        {"name": "Cataclysmic Harbinger", "tier": 6, "attack": 10, "health": 10},
        {"name": "Balinda Stonehearth", "tier": 6, "attack": 6, "health": 6}
      ],
      "addon_cards": [
        {"name": "Sinrunner Blanchy"},
        {"name": "Blade Collector"}
      ],
      "how_to_play": "Scale spell power with Felfire Conjurer + Drakkari Enchanter to make your Back to Back (Tavern Spell) bigger. Then use Balinda Stonehearth to double the scaling and effectiveness of your back to backs. Cataclysmic Harbinger allows you to print multiple of them every turn. Scale premium units such as Sinrunner Blanchy.",
      "when_to_commit": "Back to Back generation (Cataclysmic Harbinger) + spell power (Felfire Conjurer) + Drakkari Enchanter",
      "key_transition": {
        "1_star": ["Drakkari Enchanter"],
        "2_star": ["Drakkari Enchanter", "Felfire Conjurer"],
        "3_star": ["Felfire Conjurer"],
        "4_star": ["Balinda Stonehearth"],
        "5_star": ["Cataclysmic Harbinger", "Sinrunner Blanchy"]
      }
    },
    {
      "id": 3,
      "name": "Pirates - Bounty APM",
      "tier": "S",
      "difficulty": "Hard",
      "description": "Bounties generate gold and stats",
      "url": "https://hsreplay.net/battlegrounds/comps/3/pirates-bounty-apm",
      "core_cards": [
        {"name": "Proud Privateer", "tier": 8, "attack": 8, "health": 8},
        {"name": "Sky Admiral Rogers", "tier": 4, "attack": 6, "health": 6},
        {"name": "Brann Bronzebeard", "tier": 2, "attack": 4, "health": 4},
        {"name": "Brazen Buccaneer", "tier": 4, "attack": 4, "health": 4}
      ],
      "addon_cards": [
        {"name": "Fire-forged Evoker"},
        {"name": "Drakkari Enchanter"},
        {"name": "Blade Collector"},
        {"name": "Batty Terrorguard"},
        {"name": "Living Azerite"},
        {"name": "Forsaken Weaver"}
      ],
      "how_to_play": "Sky Admiral Rogers combined with Proud Privateer gives you lots of gold to hit more of your setup. Cycle Tranquil Meditative, or spell buff Chromadrakes with Brann for spell scaling to make your bounties buff more. You also want to cycle every single free card (Refreshing Anomaly) and Roving Sailor so you stay infinite.",
      "when_to_commit": "Bounty Generation (Sky Admiral Rogers) + Proud Privateer",
      "key_transition": {
        "1_star": ["Brann Bronzebeard"],
        "2_star": ["Brann Bronzebeard", "Brazen Buccaneer"],
        "3_star": ["Sky Admiral Rogers"],
        "4_star": ["Sky Admiral Rogers", "Proud Privateer"],
        "5_star": ["Proud Privateer"]
      }
    },
    {
      "id": 81,
      "name": "Elementals - Shop Buff/Spells",
      "tier": "S",
      "difficulty": "Medium",
      "description": "Make big units in shop to scale",
      "url": "https://hsreplay.net/battlegrounds/comps/81/elementals-shop-buff-spells",
      "core_cards": [
        {"name": "Leyline Surfacer", "tier": 4, "attack": 6, "health": 6},
        {"name": "Living Azerite", "tier": 6, "attack": 5, "health": 5},
        {"name": "Brann Bronzebeard", "tier": 2, "attack": 4, "health": 4},
        {"name": "Balinda Stonehearth", "tier": 6, "attack": 6, "health": 6}
      ],
      "addon_cards": [
        {"name": "Wildfire Elemental"},
        {"name": "Air Revenant"},
        {"name": "Flaming Enforcer"},
        {"name": "Rylak Metalhead"},
        {"name": "Titus Rivendare"}
      ],
      "how_to_play": "First, buff the shop. This is usually done by having Living Azerite, but Air Revenant works as well. Keep Leyline Surfacer for stats. Double the effect of the Arcane Absorption you get from Leyline using Balinda Stonehearth. In the endgame, scale up Wildfire Elemental.",
      "when_to_commit": "Brann Bronzebeard + shop buff (Living Azerite)",
      "key_transition": {
        "1_star": ["Brann Bronzebeard"],
        "2_star": ["Brann Bronzebeard"],
        "3_star": ["Leyline Surfacer"],
        "4_star": ["Living Azerite"],
        "5_star": ["Balinda Stonehearth", "Wildfire Elemental"]
      }
    },
    {
      "id": 32,
      "name": "Mechs - Automaton",
      "tier": "A",
      "difficulty": "Easy",
      "description": "Summon large Automatons (SPECIFIC HEROES)",
      "url": "https://hsreplay.net/battlegrounds/comps/32/mechs-automaton-specific-heroes",
      "core_cards": [
        {"name": "Ancestral Automaton", "tier": 3, "attack": 4, "health": 4},
        {"name": "Kangor's Apprentice", "tier": 3, "attack": 6, "health": 6},
        {"name": "Auto Assembler", "tier": 2, "attack": 2, "health": 2}
      ],
      "addon_cards": [
        {"name": "Titus Rivendare"},
        {"name": "Leeroy the Reckless"}
      ],
      "how_to_play": "Reborn your Automaton as soon as possible to start scaling it. Put naked Auto Assembler on the board to also scale it. Roll on 5 for Kangors, and then you can fill out the rest of your comp with Leeroy and scam.",
      "when_to_commit": "Automaton + Enabled Hero/Trinket",
      "key_transition": {
        "1_star": ["Auto Assembler"],
        "2_star": ["Ancestral Automaton", "Auto Assembler"],
        "3_star": ["Kangor's Apprentice"],
        "4_star": ["Kangor's Apprentice"],
        "5_star": ["Titus Rivendare"]
      }
    },
    {
      "id": 83,
      "name": "Beasts - Leviathan",
      "tier": "A",
      "difficulty": "Easy",
      "description": "Make huge attack summons",
      "url": "https://hsreplay.net/battlegrounds/comps/83/beasts-leviathan",
      "core_cards": [
        {"name": "Sewer Lord", "tier": 4, "attack": 6, "health": 6},
        {"name": "Lurking Leviathan", "tier": 3, "attack": 8, "health": 8},
        {"name": "Banana Slamma", "tier": 3, "attack": 6, "health": 6}
      ],
      "addon_cards": [
        {"name": "Manasaber"},
        {"name": "Rylak Metalhead"},
        {"name": "Hunting Tiger Shark"}
      ],
      "how_to_play": "Summon as many beasts as possible after you get early Lurking Leviathan. You either use Sewer Lord and/or Rylak Metalhead + Hunting Tiger Shark. Then, paired with Banana Slamma, every summon will have a ton of attack. Reborn on Sewer Lord provides so many summons.",
      "when_to_commit": "Early Lurking Leviathan + Beast summons (Sewer Lord)",
      "key_transition": {
        "1_star": [],
        "2_star": [],
        "3_star": ["Lurking Leviathan"],
        "4_star": ["Sewer Lord", "Banana Slamma"],
        "5_star": ["Rylak Metalhead", "Manasaber"]
      }
    },
    {
      "id": 12,
      "name": "Beasts - Summons",
      "tier": "A",
      "difficulty": "Easy",
      "description": "Scale summoned beasts in combat",
      "url": "https://hsreplay.net/battlegrounds/comps/12/beasts-summons",
      "core_cards": [
        {"name": "Goldrinn, the Great Wolf", "tier": 8, "attack": 8, "health": 8},
        {"name": "Titus Rivendare", "tier": 5, "attack": 7, "health": 4},
        {"name": "Monstrous Macaw", "tier": 5, "attack": 4, "health": 4},
        {"name": "Sewer Lord", "tier": 4, "attack": 6, "health": 6}
      ],
      "addon_cards": [
        {"name": "Manasaber"},
        {"name": "Banana Slamma"},
        {"name": "Sewer Rat"}
      ],
      "how_to_play": "Goldrinn, the Great Wolf + Monstrous Macaw and Titus Rivendare gives a huge aura affect to all beasts you summon. Usually, you want to taunt and reborn the Goldrinn so the Macaw always procs it and it always dies. Sewer Lord is the best summon beast to run in this comp.",
      "when_to_commit": "Look for Goldrinn, the Great Wolf + beast summons (specifically Sewer Lord)",
      "key_transition": {
        "1_star": [],
        "2_star": ["Monstrous Macaw"],
        "3_star": ["Sewer Lord"],
        "4_star": ["Monstrous Macaw", "Titus Rivendare"],
        "5_star": ["Goldrinn, the Great Wolf"]
      }
    },
    {
      "id": 8,
      "name": "Murlocs - Venom Scam",
      "tier": "A",
      "difficulty": "Easy",
      "description": "Buy scam cards and try to preserve placements",
      "url": "https://hsreplay.net/battlegrounds/comps/8/murlocs-venom-scam",
      "core_cards": [
        {"name": "Leeroy the Reckless", "tier": 6, "attack": 2, "health": 2},
        {"name": "Bile Spitter", "tier": 2, "attack": 1, "health": 10},
        {"name": "Diremuck Forager", "tier": 1, "attack": 7, "health": 5}
      ],
      "addon_cards": [
        {"name": "Deadly Spore"},
        {"name": "Pufferquil"},
        {"name": "Brann Bronzebeard"},
        {"name": "Heroic Underdog"}
      ],
      "how_to_play": "Buy all the scam cards and try to preserve placements by killing as many units as possible. Keep your strongest units on the board (that's why you want to concentrate stats!) and fill the rest of your board with units that may be able to trade 1 for 1. Aim to golden Diremuck Forager and give it venom using Bile Spitter. Then, the Diremuck will pull many scam units from your hand.",
      "when_to_commit": "When you have no future or are clearly getting outstatted",
      "key_transition": {
        "1_star": ["Diremuck Forager"],
        "2_star": ["Bile Spitter", "Diremuck Forager"],
        "3_star": ["Leeroy the Reckless"],
        "4_star": ["Leeroy the Reckless", "Diremuck Forager (golden)"],
        "5_star": []
      }
    },
    {
      "id": 67,
      "name": "Murlocs - APM",
      "tier": "A",
      "difficulty": "Hard",
      "description": "Cycle cards to scale murlocs",
      "url": "https://hsreplay.net/battlegrounds/comps/67/murlocs-apm",
      "core_cards": [
        {"name": "Brann Bronzebeard", "tier": 2, "attack": 4, "health": 4},
        {"name": "Primitive Painter", "tier": 4, "attack": 3, "health": 8},
        {"name": "Expert Aviator", "tier": 3, "attack": 4, "health": 4},
        {"name": "Magicfin Mycologist", "tier": 4, "attack": 8, "health": 3}
      ],
      "addon_cards": [
        {"name": "Bream Counter"},
        {"name": "Mrglin' Burglar"},
        {"name": "Cousin Errgl"},
        {"name": "Diremuck Forager"},
        {"name": "Bile Spitter"},
        {"name": "Choral Mrrrglr"}
      ],
      "how_to_play": "First, have a ton of economy, typically starting from golden Brann Bronzebeard or Brann + Magicfin Mycologist, remembering premium economy spells such as Cloning Conch or Battlecry Discovery. Then, your payoff card for all that economy is typically Primitive Painter. Your payoff can also include Bream Counter + Mrglin' Burglar handbuff package, or Cousin Errgl family package. In any case, you're scaling Expert Aviator, which is the most premium murloc because it can summon any scam card (Leeroy the Reckless) from your hand.",
      "when_to_commit": "Brann Bronzebeard + lots of economy + Murloc scaling (Primitive Painter)",
      "key_transition": {
        "1_star": ["Brann Bronzebeard"],
        "2_star": ["Magicfin Mycologist", "Brann Bronzebeard"],
        "3_star": ["Expert Aviator"],
        "4_star": ["Primitive Painter"],
        "5_star": ["Expert Aviator", "Primitive Painter"]
      }
    },
    {
      "id": 59,
      "name": "Quilboar - Darkgaze",
      "tier": "B",
      "difficulty": "Medium",
      "description": "Spend gold to cast blood gems",
      "url": "https://hsreplay.net/battlegrounds/comps/59/quilboar-darkgaze",
      "core_cards": [
        {"name": "Darkgaze Elder", "tier": 6, "attack": 8, "health": 5},
        {"name": "Prickly Piper", "tier": 8, "attack": 1, "health": 2},
        {"name": "Brann Bronzebeard", "tier": 2, "attack": 4, "health": 4}
      ],
      "addon_cards": [
        {"name": "Earthsong Shaman"},
        {"name": "Bristlebach"},
        {"name": "Hired Ritualist"},
        {"name": "Pufferquil"},
        {"name": "Geomagus Roogug"}
      ],
      "how_to_play": "After getting Darkgaze Elder and buffed gems (via Prickly Piper), the objective is to spend as much gold as possible. This is usually done with Brann Bronzebeard, cycling every free battlecry such as Shell Collector and Tavern Tempest. Also cycle every Moon-Bacon Jazzer and Fearless Foodie to help scale. Any quilboar in the addon section fills out the rest of the board.",
      "when_to_commit": "Darkgaze Elder + Scaled gems (Prickly Piper)",
      "key_transition": {
        "1_star": ["Brann Bronzebeard"],
        "2_star": ["Prickly Piper"],
        "3_star": ["Darkgaze Elder"],
        "4_star": ["Darkgaze Elder"],
        "5_star": ["Bristlebach", "Hired Ritualist"]
      }
    },
    {
      "id": 60,
      "name": "Quilboar - Smuggler",
      "tier": "B",
      "difficulty": "Medium",
      "description": "Play gems during combat",
      "url": "https://hsreplay.net/battlegrounds/comps/60/quilboar-smuggler",
      "core_cards": [
        {"name": "Gem Smuggler", "tier": 4, "attack": 5, "health": 5},
        {"name": "Prickly Piper", "tier": 8, "attack": 1, "health": 2},
        {"name": "Titus Rivendare", "tier": 5, "attack": 7, "health": 4},
        {"name": "Brann Bronzebeard", "tier": 2, "attack": 4, "health": 4},
        {"name": "Rylak Metalhead", "tier": 3, "attack": 5, "health": 5},
        {"name": "Moon-Bacon Jazzer", "tier": 1, "attack": 7, "health": 7},
        {"name": "Monstrous Macaw", "tier": 5, "attack": 4, "health": 4}
      ],
      "addon_cards": [
        {"name": "Bristlebach"},
        {"name": "Vinespeaker"},
        {"name": "Tarecgosa"}
      ],
      "how_to_play": "MUST HAVE BEASTS - Before proccing Gem Smuggler, scale gems as much as possible with Prickly Piper and proccing Moon-Bacon Jazzer. In midgame, use Three Lil' Quilboar or Bristlebach to stabilize. The final variation uses Rylak Metalhead to proc Gem Smuggler. With reborn on Rylak, Monstrous Macaw, Brann Bronzebeard, and Titus Rivendare all act as multipliers to the Rylak effect. Golden Rylak can proc both Jazzer and Smuggler. Tarecgosa can be added to keep gems.",
      "when_to_commit": "Look for: Brann Bronzebeard + Gem Scaling (Prickly Piper) + Gem Smuggler + Rylak Metalhead",
      "key_transition": {
        "1_star": ["Moon-Bacon Jazzer", "Brann Bronzebeard"],
        "2_star": ["Prickly Piper", "Moon-Bacon Jazzer"],
        "3_star": ["Gem Smuggler", "Rylak Metalhead"],
        "4_star": ["Gem Smuggler", "Titus Rivendare"],
        "5_star": ["Monstrous Macaw", "Tarecgosa"]
      }
    },
    {
      "id": 6,
      "name": "Undead - Overflow",
      "tier": "B",
      "difficulty": "Easy",
      "description": "Overflow your board to scale",
      "url": "https://hsreplay.net/battlegrounds/comps/6/undead-overflow",
      "core_cards": [
        {"name": "亡灵溢出流", "tier": 4, "attack": 10, "health": 3},
        {"name": "溢出机制", "tier": 1, "attack": 7, "health": 5}
      ],
      "addon_cards": [],
      "how_to_play": "Overflow your board to scale through undead mechanics.",
      "when_to_commit": "Undead units + overflow mechanic",
      "key_transition": {
        "1_star": [],
        "2_star": [],
        "3_star": ["溢出机制"],
        "4_star": ["亡灵溢出流"],
        "5_star": []
      }
    }
  ]
}`

	var s models.SeasonData
	if err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&s); err != nil {
		// Fallback to minimal data
		return getMinimalComps()
	}

	return s.Comps
}

func getMinimalComps() []models.Comp {
	return []models.Comp{
		{ID: 41, Name: "Demons - Shop Buff", Tier: "S", Difficulty: "Medium"},
		{ID: 14, Name: "Undead - Attack Scaling", Tier: "S", Difficulty: "Medium"},
		{ID: 82, Name: "Back to Back", Tier: "S", Difficulty: "Medium"},
		{ID: 3, Name: "Pirates - Bounty APM", Tier: "S", Difficulty: "Hard"},
		{ID: 81, Name: "Elementals - Shop Buff/Spells", Tier: "S", Difficulty: "Medium"},
	}
}
