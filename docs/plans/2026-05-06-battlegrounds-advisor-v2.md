# Battlegrounds Advisor v2 实现计划

> **Hermes:** 使用 subagent-driven-development skill 逐任务执行此计划。

**目标：** 构建一个商业化的炉石传说酒馆战旗辅助工具，在每局游戏开局根据元数据生成随从选取推荐，核心规则是多阵容共用的关键随从优先。

**架构概述：**
- 模块化分层：数据层 → 引擎层 → 协议层 → UI 层
- 游戏状态获取：屏幕捕捉 + OCR / Blizzard API / HSReplay 数据同步
- 推荐引擎：基于规则 + 可配置权重，支持插件扩展
- UI 层：游戏内 Overlay（悬浮推荐面板）
- 商业化：免费版（基础推荐）+ 付费版（高级规则、AI 辅助、实时数据）

**技术栈：**
- 核心逻辑：Go
- 游戏状态识别：Python（OCR/Screen capture）+ Go FFI
- Overlay UI：Tauri（Rust + Web）或直接 Python（overlaytk）
- 数据存储：SQLite（本地）+ 可选云端同步

---

## 阶段一：核心数据与抓取（数据层）

### Task 1: 升级 HSReplay 抓取器，支持所有赛季和完整阵容详情

**目标：** 将抓取器升级为生产级，稳定获取完整阵容数据。

**文件：**
- 修改: `cmd/fetch/main.go`
- 新建: `fetcher/parser.go`
- 新建: `fetcher/auth.go`（HSReplay 认证）

**Step 1: 分析 HSReplay 阵容页面结构**
```bash
# 抓取一个具体阵容页面
cd ~/work/Projects/GoProjects/src/github.com/hind3ight/hsreplay-battlegrounds
xvfb-run ./bin/fetch --url "https://hsreplay.net/battlegrounds/comps/41/demons-shop-buff" -o data/comps/41.json
```

**Step 2: 扩展抓取逻辑**

阵容详情页包含：核心随从、可选随从、游戏技巧、转型时机、各星级关键随从。需要解析这些字段并持久化到 SQLite。

**Step 3: 验证数据完整性**
```bash
go run cmd/fetch/main.go --season=latest --output=data/comps/
# 验证：应获取 50+ 流派
```

---

### Task 2: 构建随从元数据库

**目标：** 建立完整的随从元数据，包括中英文名称、种族、等级、关键机制描述。

**文件：**
- 新建: `data/minions_metadata.json`（从 hsreplay.net/minions 抓取）
- 新建: `cmd/update-minions/main.go`

**字段设计：**
```json
{
  "id": "BG32_873",
  "name_en": "Ashen Corruptor",
  "name_cn": "灰葬腐蚀者",
  "tribe": "demon",
  "tier": 6,
  "attack": 6,
  "health": 6,
  "mechanics": ["battlecry", "consume"],
  "key_for_comps": ["demons-shop-buff", "demons-consume"]
}
```

---

### Task 3: 构建阵容-随从交叉分析引擎

**目标：** 量化每个随从的"通用性"分数，识别跨阵容关键随从。

**文件：**
- 新建: `engine/cross_analyzer.go`
- 新建: `cmd/analyze-cross/main.go`

**核心算法：**
1. 加载所有阵容数据
2. 统计每个随从在多少个阵容中出现（核心+addon）
3. 按种族分组统计（处理5种族池限制）
4. 输出优先级列表：`minion_priority = cross_comp_count * tribe_diversity * tier_weight`

**Tier 权重：** 低星随从（T1-T3）开局价值更高，权重系数更大。
**种族跨越数：** 同时属于多个种族的随从优先（如布莱恩属于多个阵容）。

---

## 阶段二：游戏状态识别（协议层）

### Task 4: 设计游戏状态数据结构

**目标：** 定义游戏状态的数据模型，包括当前回合、可用种族、商店随从等。

**文件：**
- 新建: `models/game_state.go`

```go
type GameState struct {
    GameID       string    // 对局唯一ID
    Phase        GamePhase //开局/中期/后期
    Turn         int       // 当前回合
    AvailableTribes []Tribe // 本局5个种族
    ShopMinions  []Minion  // 当前商店
    BoardMinions []Minion  // 已有随从
    TavernTier   int       // 当前酒馆等级
    HeroID       string    // 英雄ID
    HeroPowerUsed bool     // 英雄技能是否使用
}
```

### Task 5: 实现屏幕捕捉 + OCR 状态读取

**目标：** 从炉石传说游戏窗口读取当前游戏状态。

**文件：**
- 新建: `reader/screen_capture.go`（Python）
- 新建: `reader/ocr.go`（Python）
- 新建: `reader/game_state_reader.go`（Go wrapper）

**技术方案：**
- `mss` + `PIL` 截图
- `pytesseract` OCR 识别随从名称
- 区域模板匹配确定 UI 元素位置
- Windows: Win32 API；Linux: X11/Wayland

**验证：**
```bash
cd reader/
python3 test_capture.py --window="Hearthstone"
# 输出当前商店随从列表
```

### Task 6: 标准化随从识别（名称匹配）

**目标：** 将 OCR 识别的随从名称与元数据库匹配，处理拼写变体。

**文件：**
- 新建: `reader/name_matcher.go`

**规则：**
- 精确匹配
- 模糊匹配（Levenshtein 距离 ≤ 2）
- 别名支持（如"小瞎眼" = "小瞎眼"）
- 英文名兼容（OCR 可能返回英文）

---

## 阶段三：推荐引擎（引擎层）

### Task 7: 构建推荐引擎核心

**目标：** 根据游戏状态和元数据，生成随从选取优先级列表。

**文件：**
- 新建: `engine/recommender.go`
- 新建: `engine/rules.go`
- 新建: `engine/weights.go`

**推荐算法：**

```
Score(minion) = Σ(comp_relevance) × tier_weight × tribe_bonus × synergy_bonus

其中：
- comp_relevance: 该随从在当前阵容方向中的重要性（0-1）
- tier_weight: 低星优先（T1=3, T2=2.5, T3=2, T4=1.5, T5=1, T6=0.5）
- tribe_bonus: 当前种族匹配 +0.5，不匹配 -0.3
- synergy_bonus: 与已有随从协同 +0.3
```

**开局规则（核心）：**
1. 识别本局5个种族
2. 加载跨阵容通用随从（出现≥2个阵容的随从）
3. 按 `cross_tribe_count × tier_weight` 排序
4. 输出前三推荐，并说明可转型方向

### Task 8: 设计规则引擎（可扩展性核心）

**目标：** 支持 YAML/JSON 配置文件定义推荐规则，无需修改代码。

**文件：**
- 新建: `engine/rule_engine.go`
- 新建: `rules/default_rules.yaml`

**规则格式：**
```yaml
rules:
  - name: "开局优先低星关键随从"
    condition:
      turn: [1, 3]
      phase: early
    priority_boost:
      tier_1: 3.0
      tier_2: 2.5
      tier_3: 2.0

  - name: "恶魔流优先布莱恩"
    condition:
      tribe: demon
      comps: ["demons-shop-buff", "demons-consume"]
    priority_boost:
      "Brann Bronzebeard": 2.0
```

---

## 阶段四：UI 层（Overlay）

### Task 9: 构建 Overlay UI

**目标：** 在游戏上方显示实时推荐面板。

**文件：**
- 新建: `overlay/`（Tauri 项目或 Python overlay）

**技术选型：**
- **推荐：Tauri**（Rust 后端 + Web 前端，轻量、跨平台）
- 或：**Python overlay**（overlaytk、screeninfo，轻量但兼容性差）

**UI 设计：**
```
┌─────────────────────────────────┐
│  🏆 本局推荐（开局）             │
├─────────────────────────────────┤
│  1. 布莱恩·铜须 (T2)            │
│     适用：恶魔/海盗/元素/鱼人    │
│     原因：跨5大阵容核心         │
│                                │
│  2. 死亡之翼俯冲 (T3)           │
│     适用：龙/野兽               │
│     原因：开局优质亡语         │
│                                │
│  3. 撕心狼队长 (T3)             │
│     适用：野兽/亡灵             │
│     原因：跨2种族核心           │
└─────────────────────────────────┘
```

**功能：**
- 悬浮可拖拽
- 透明背景（不影响游戏）
- 快捷键显示/隐藏（默认 `Ctrl+Shift+H`）
- 推荐理由悬停显示

---

## 阶段五：商业化与可扩展性

### Task 10: 设计插件系统

**目标：** 支持第三方或用户自定义推荐规则和阵容数据。

**文件：**
- 新建: `plugin/loader.go`
- 新建: `plugins/` 目录

**插件结构：**
```
plugins/
├── my-custom-rules.yaml
└── community/
    └── top500-rules.yaml
```

### Task 11: 数据同步服务（可选云端）

**目标：** 支持用户账号、数据同步、高级规则订阅。

**架构：**
- 轻量后端（Go + PostgreSQL 或 SQLite on cloud）
- REST API 提供阵容数据更新
- 用户自定义规则云端备份

### Task 12: 构建配置管理 UI

**目标：** 提供图形化配置界面，管理规则权重、插件、显示偏好。

**技术：**
- Tauri 内嵌 Web UI 或独立 Electron 窗口
- 实时预览推荐结果

---

## 执行顺序

```
Phase 1（数据层）     → Phase 2（协议层）  → Phase 3（引擎层） → Phase 4（UI） → Phase 5
Task 1-3             Task 4-6              Task 7-8            Task 9         Task 10-12
```

**推荐 MVP 路线：**
1. 先完成 Task 1-3（数据就绪）
2. Task 4-6（状态读取）用手动输入替代（先用 CLI 输入本局种族）
3. Task 7-8（引擎核心）
4. Task 9（Overlay）← 这个是用户体验核心
5. Task 10-12（扩展性和商业化）

---

## 技术风险与备选方案

| 风险 | 缓解方案 |
|------|---------|
| OCR 识别率低 | 使用 BLP 图像解析替代 OCR，或调用 HSReplay API |
| Blizzard 反作弊检测 | 只读屏幕像素，不注入内存 |
| Overlay 被游戏遮挡 | 使用置顶窗口 + 全透明 |
| HSReplay 数据版权 | 只做数据分析工具，不直接复制内容 |
