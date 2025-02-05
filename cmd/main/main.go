package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/yourusername/fmcl/pkg/htmlfetcher"
	"github.com/yourusername/fmcl/pkg/logger"
	"github.com/yourusername/fmcl/pkg/parser"
)

// 格式化字符串到指定宽度
func formatWidth(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// 显示模式
type DisplayMode int

const (
	ModeHighImportance DisplayMode = iota
	ModeAll
	ModeWithRates
	ModeWithImportant
)

// 应用状态
type AppState struct {
	displayMode     DisplayMode
	isPaused        bool
	shouldExit      bool
	mu              sync.Mutex
	nextRefreshTime time.Time
	startTime       time.Time
}

func (s *AppState) togglePause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isPaused = !s.isPaused
}

func (s *AppState) nextMode() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.displayMode = (s.displayMode + 1) % 4
}

// 显示帮助信息
func showHelpMenu() *widgets.Paragraph {
	help := widgets.NewParagraph()
	help.Title = "快捷键说明"
	help.Text = `
q: 退出程序
r: 强制刷新
p: 暂停/继续刷新
m: 切换显示模式
h: 显示/隐藏帮助
ESC: 关闭此帮助
`
	help.BorderStyle.Fg = termui.ColorCyan
	help.TitleStyle.Fg = termui.ColorGreen
	help.TextStyle.Fg = termui.ColorWhite

	// 居中显示帮助菜单
	termWidth, termHeight := termui.TerminalDimensions()
	helpWidth := 40
	helpHeight := 10
	x := (termWidth - helpWidth) / 2
	y := (termHeight - helpHeight) / 2
	help.SetRect(x, y, x+helpWidth, y+helpHeight)

	return help
}

// 获取终端宽度
func getTerminalWidth() int {
	cmd := exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 120
	}

	width, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 120
	}

	return width
}

func formatCountdown(nextRefresh time.Time) string {
	duration := time.Until(nextRefresh)
	if duration < 0 {
		return "即将刷新"
	}
	return fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
}

// Config 配置结构
type Config struct {
	RefreshInterval    int `yaml:"refresh_interval"`
	DefaultDisplayMode int `yaml:"default_display_mode"`
	UI                 struct {
		TimeWidth       int `yaml:"time_width"`
		ImportanceWidth int `yaml:"importance_width"`
		ValueWidth      int `yaml:"value_width"`
	} `yaml:"ui"`
}

// 加载配置
func loadConfig() (*Config, error) {
	// 默认配置
	config := &Config{
		RefreshInterval:    15,
		DefaultDisplayMode: 0,
		UI: struct {
			TimeWidth       int `yaml:"time_width"`
			ImportanceWidth int `yaml:"importance_width"`
			ValueWidth      int `yaml:"value_width"`
		}{
			TimeWidth:       8,
			ImportanceWidth: 6,
			ValueWidth:      12,
		},
	}

	// 尝试读取配置文件
	configPath := filepath.Join(".", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
		// 如果配置文件不存在，使用默认配置
		return config, nil
	}

	// 解析配置文件
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

func displayData(fetcher *htmlfetcher.DefaultFetcher, state *AppState, config *Config) {
	if err := termui.Init(); err != nil {
		logger.Error("初始化TUI失败", zap.Error(err))
		return
	}
	defer termui.Close()

	// 创建TUI组件
	header := widgets.NewParagraph()
	header.Text = fmt.Sprintf("FMCL @ %s", state.startTime.Format("2006-01-02 15:04:05"))
	header.TextStyle.Fg = termui.ColorGreen
	header.Border = false

	dataList := widgets.NewList()
	dataList.TextStyle.Fg = termui.ColorWhite
	dataList.BorderStyle.Fg = termui.ColorBlue
	dataList.WrapText = false

	statusBar := widgets.NewParagraph()
	statusBar.TextStyle.Fg = termui.ColorGreen
	statusBar.Border = false

	// 获取终端大小
	termWidth, termHeight := termui.TerminalDimensions()

	// 设置布局
	header.SetRect(0, 0, termWidth, 2)
	dataList.SetRect(0, 2, termWidth, termHeight-1)
	statusBar.SetRect(0, termHeight-1, termWidth, termHeight)

	// 创建帮助菜单（初始不显示）
	helpMenu := showHelpMenu()
	showingHelp := false

	updateUI := func() {
		state.mu.Lock()
		defer state.mu.Unlock()

		// 更新状态栏
		modeStr := "仅显示高重要性"
		switch state.displayMode {
		case ModeAll:
			modeStr = "显示所有数据"
		case ModeWithRates:
			modeStr = "显示高重要性+利率信息"
		case ModeWithImportant:
			modeStr = "显示高重要性+重要事件"
		}

		statusBar.Text = fmt.Sprintf("模式: %s | 状态: %s | 下次刷新: %s",
			modeStr,
			map[bool]string{true: "已暂停", false: "运行中"}[state.isPaused],
			formatCountdown(state.nextRefreshTime))

		// 获取数据并更新显示
		dateStr := time.Now().Format("20060102")
		url := fmt.Sprintf("https://rl.fx678.com/date/%s.html", dateStr)
		html, err := fetcher.Fetch(url)
		if err != nil {
			logger.Error("获取数据失败", zap.Error(err))
			dataList.Rows = []string{"获取数据失败: " + err.Error()}
			if showingHelp {
				termui.Render(header, dataList, statusBar, helpMenu)
			} else {
				termui.Render(header, dataList, statusBar)
			}
			return
		}

		// 解析数据
		events, importantEvents, rates, err := parser.ParseFinancialCalendar(html)
		if err != nil {
			logger.Error("解析数据失败", zap.Error(err))
			dataList.Rows = []string{"解析数据失败: " + err.Error()}
			if showingHelp {
				termui.Render(header, dataList, statusBar, helpMenu)
			} else {
				termui.Render(header, dataList, statusBar)
			}
			return
		}

		// 根据显示模式过滤和格式化数据
		var rows []string
		rows = append(rows, "[=== 财经日历事件 ===](fg:green)")
		rows = append(rows, fmt.Sprintf("[%-*s  %-*s  %-*s  %-*s  %-*s  %s](fg:cyan)",
			config.UI.TimeWidth, "时间",
			config.UI.ImportanceWidth, "重要性",
			config.UI.ValueWidth, "前值",
			config.UI.ValueWidth, "预测",
			config.UI.ValueWidth, "公布值",
			"指标名称"))
		rows = append(rows, strings.Repeat("-", termWidth-2))

		currentTime := ""
		for _, event := range events {
			if state.displayMode == ModeHighImportance && event.Importance != "高" {
				continue
			}

			if event.Time != currentTime {
				if currentTime != "" {
					rows = append(rows, strings.Repeat("-", termWidth-2))
				}
				currentTime = event.Time
			}

			importanceColor := "white"
			if event.Importance == "高" {
				importanceColor = "red"
			} else if event.Importance == "中" {
				importanceColor = "yellow"
			}

			row := fmt.Sprintf("[%-*s](fg:cyan)  [%-*s](fg:%s)  [%-*s](fg:white)  [%-*s](fg:white)  [%-*s](fg:green)  [%s](fg:white)",
				config.UI.TimeWidth, event.Time,
				config.UI.ImportanceWidth, event.Importance,
				importanceColor,
				config.UI.ValueWidth, formatWidth(event.Previous, config.UI.ValueWidth),
				config.UI.ValueWidth, formatWidth(event.Forecast, config.UI.ValueWidth),
				config.UI.ValueWidth, formatWidth(event.Actual, config.UI.ValueWidth),
				event.Indicator)
			rows = append(rows, row)
		}

		if state.displayMode == ModeWithImportant || state.displayMode == ModeAll {
			if len(importantEvents) > 0 {
				rows = append(rows, "")
				rows = append(rows, "[=== 重要事件 ===](fg:green)")
				rows = append(rows, fmt.Sprintf("[%-*s  %-*s  %s](fg:cyan)",
					config.UI.TimeWidth, "时间",
					config.UI.ImportanceWidth, "重要性",
					"事件"))
				rows = append(rows, strings.Repeat("-", termWidth-2))

				for _, event := range importantEvents {
					if event.Importance == "高" || state.displayMode == ModeAll {
						importanceColor := "white"
						if event.Importance == "高" {
							importanceColor = "red"
						}
						row := fmt.Sprintf("[%-*s](fg:cyan)  [%-*s](fg:%s)  [%s](fg:white)",
							config.UI.TimeWidth, event.Time,
							config.UI.ImportanceWidth, event.Importance,
							importanceColor,
							event.Event)
						rows = append(rows, row)
					}
				}
			}
		}

		if state.displayMode == ModeWithRates || state.displayMode == ModeAll {
			rows = append(rows, "")
			rows = append(rows, "[=== 央行利率信息 ===](fg:green)")
			bankWidth := 20
			rateWidth := 10
			changeWidth := 8
			dateWidth := 12
			historyWidth := 20

			rows = append(rows, fmt.Sprintf("[%-*s  %-*s  %-*s  %-*s  %-*s  %-*s  %-*s](fg:cyan)",
				bankWidth, "央行/利率类型",
				rateWidth, "当前利率",
				rateWidth, "前值",
				changeWidth, "变动",
				dateWidth, "变动日期",
				historyWidth, "历史区间",
				rateWidth, "下次预测"))
			rows = append(rows, strings.Repeat("-", termWidth-2))

			for _, rate := range rates {
				parts := strings.Split(rate.LastChange, " ")
				changeDate := ""
				changeValue := ""
				if len(parts) >= 2 {
					changeDate = parts[1]
					changeValue = parts[0]
				}

				historyRange := fmt.Sprintf("%s - %s", rate.HistoryLow, rate.HistoryHigh)
				bankInfo := fmt.Sprintf("%s - %s", rate.Bank, rate.RateName)

				row := fmt.Sprintf("[%-*s](fg:yellow)  [%-*s](fg:green)  [%-*s](fg:white)  [%-*s](fg:red)  [%-*s](fg:cyan)  [%-*s](fg:white)  [%-*s](fg:white)",
					bankWidth, bankInfo,
					rateWidth, rate.CurrentRate,
					rateWidth, rate.PreviousRate,
					changeWidth, changeValue,
					dateWidth, changeDate,
					historyWidth, historyRange,
					rateWidth, rate.NextForecast)
				rows = append(rows, row)
			}
		}

		if len(rows) == 0 {
			rows = []string{"[暂无数据](fg:red)"}
		}
		dataList.Rows = rows

		// 渲染UI
		if showingHelp {
			termui.Render(header, dataList, statusBar, helpMenu)
		} else {
			termui.Render(header, dataList, statusBar)
		}
	}

	// 初始更新
	updateUI()

	// 设置定时器更新倒计时
	countdownTicker := time.NewTicker(time.Second)
	defer countdownTicker.Stop()

	// 设置数据刷新定时器
	refreshTicker := time.NewTicker(time.Duration(config.RefreshInterval) * time.Second)
	defer refreshTicker.Stop()

	// 更新下次刷新时间
	state.nextRefreshTime = time.Now().Add(time.Duration(config.RefreshInterval) * time.Second)

	for {
		select {
		case e := <-termui.PollEvents():
			switch e.ID {
			case "q", "<C-c>":
				return
			case "r":
				updateUI()
				state.nextRefreshTime = time.Now().Add(time.Duration(config.RefreshInterval) * time.Second)
			case "p":
				state.togglePause()
				termui.Render(statusBar)
			case "m":
				state.nextMode()
				updateUI()
			case "h":
				showingHelp = !showingHelp
				if showingHelp {
					termui.Render(header, dataList, statusBar, helpMenu)
				} else {
					termui.Render(header, dataList, statusBar)
				}
			case "<Escape>":
				if showingHelp {
					showingHelp = false
					termui.Render(header, dataList, statusBar)
				}
			}
		case <-countdownTicker.C:
			// 更新倒计时
			state.mu.Lock()
			statusBar.Text = fmt.Sprintf("模式: %s | 状态: %s | 下次刷新: %s",
				map[DisplayMode]string{
					ModeHighImportance: "仅显示高重要性",
					ModeAll:            "显示所有数据",
					ModeWithRates:      "显示高重要性+利率信息",
					ModeWithImportant:  "显示高重要性+重要事件",
				}[state.displayMode],
				map[bool]string{true: "已暂停", false: "运行中"}[state.isPaused],
				formatCountdown(state.nextRefreshTime))
			state.mu.Unlock()
			termui.Render(statusBar)
		case <-refreshTicker.C:
			if !state.isPaused {
				updateUI()
				state.nextRefreshTime = time.Now().Add(time.Duration(config.RefreshInterval) * time.Second)
			}
		}
	}
}

func main() {
	// 初始化日志
	logPath := filepath.Join("logs", "app.log")
	if _, err := logger.NewLogger(logPath); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Log.Sync()

	logger.Info("程序启动")

	// 加载配置
	config, err := loadConfig()
	if err != nil {
		logger.Error("加载配置失败", zap.Error(err))
		return
	}

	// 初始化应用状态
	state := &AppState{
		displayMode: DisplayMode(config.DefaultDisplayMode),
		startTime:   time.Now(),
	}

	// 初始化数据获取器
	fetcher := &htmlfetcher.DefaultFetcher{}

	// 显示数据
	displayData(fetcher, state, config)
}
