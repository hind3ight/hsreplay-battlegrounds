# Battlegrounds Advisor

酒馆战棋随从购买辅助工具。根据 HSReplay 元数据，自动生成随从购买优先级推荐，帮助在开局种族池确定后做出最优选择。

## 项目架构

```
hsreplay-battlegrounds/
├── cmd/                        # 命令行入口（Go）
│   ├── fetch/                  # HSReplay 数据抓取
│   ├── analyze/                # 阵容分析与推荐生成
│   └── interactive/            # 交互式推荐模式
├── engine/                     # 推荐引擎核心（Python）
│   ├── models.py               # 数据结构定义
│   ├── recommender.py          # 推荐逻辑
│   ├── rule_engine.py          # 规则引擎
│   └── comp_matcher.py         # 阵容匹配
├── reader/                     # 游戏状态识别（Python）
│   ├── screen_capture.py       # 屏幕截图
│   ├── ocr_minions.py          # OCR 识别随从
│   ├── name_matcher.py         # 中英文名称匹配
│   └── tribe_detector.py       # 种族检测
├── overlay/                    # 游戏内悬浮面板（Python）
│   ├── window.py               # 窗口管理
│   ├── panel.py                # 推荐面板
│   └── styles.py               # 样式
├── plugins/                    # 插件系统
│   └── loader.py               # 插件加载器
├── data/                      # 数据文件
│   └── season13_comps.json     # 赛季阵容数据
├── bin/                       # 编译后的二进制文件
│   ├── fetch                  # 数据抓取工具
│   ├── analyze                # 分析工具
│   └── interactive            # 交互模式
└── scripts/                   # 辅助脚本
    ├── setup.sh               # Linux/macOS 安装脚本
    ├── setup.bat              # Windows 安装脚本
    └── update-data.sh         # 数据更新脚本
```

## 技术栈

- **核心逻辑**: Go
- **游戏状态识别**: Python (OCR/Screen Capture)
- **Overlay UI**: Python (tkinter)
- **数据**: HSReplay, JSON/SQLite

## 快速开始

### Linux / macOS

```bash
./scripts/setup.sh
```

### Windows

```powershell
.\scripts\setup.bat
```

或手动安装：

```bash
git clone https://github.com/hind3ight/hsreplay-battlegrounds.git
cd hsreplay-battlegrounds
python -m venv .venv
.venv\Scripts\activate
pip install -r reader\requirements.txt
python tests\test_mock.py
```

## 使用方法

### 抓取最新阵容数据

```bash
./bin/fetch -list           # 仅列出阵容
./bin/fetch -o data/comps/  # 抓取详情并保存
```

### 生成随从购买指南

```bash
./bin/analyze               # 输出到默认目录
./bin/analyze -o ./output/  # 指定输出目录
```

### 交互式推荐

```bash
./bin/interactive -o ./output/
```

### 运行测试

```bash
.venv\Scripts\python tests\test_mock.py
```

## 核心思路

酒馆战棋每局只有 5 个种族可用。推荐引擎的核心逻辑是：**优先选择跨种族多的随从**，在种族池确定后可以灵活转型。

推荐等级：
- **T0**: 跨 5 种族（几乎必拿）
- **T1-T2**: 跨 3-4 种族
- **T3-T5**: 特定种族核心随从

## License

MIT
