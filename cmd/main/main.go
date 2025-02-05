package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/eiannone/keyboard"
	"gopkg.in/yaml.v3"

	"github.com/yourusername/fmt/pkg/htmlfetcher"
	"github.com/yourusername/fmt/pkg/parser"
)

// 格式化字符串到指定宽度
func formatWidth(s string, width int) string {
	if len(s) >= width {
		return s
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
	displayMode DisplayMode
	isPaused   bool
	shouldExit bool
	mu         sync.Mutex
	nextRefreshTime time.Time
	startTime  time.Time
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

func (s *AppState) toggleDisplayMode() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.displayMode = (s.displayMode + 1) % 3
}

// 显示帮助信息
func showHelp() {
	fmt.Println("\n=== 快捷键说明 ===")
	fmt.Println("q: 退出程序")
	fmt.Println("r: 强制刷新")
	fmt.Println("p: 暂停/继续刷新")
	fmt.Println("m: 切换显示模式")
	fmt.Println("h: 显示帮助")
}

// 获取终端宽度
func getTerminalWidth() int {
	// 默认宽度
	defaultWidth := 120
	
	cmd := exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return defaultWidth
	}
	
	width, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return defaultWidth
	}
	
	return width
}

// 显示状态信息在右上角
func showStatus(state *AppState, nextRefresh time.Time) {
	state.nextRefreshTime = nextRefresh
	termWidth := getTerminalWidth()
	
	modeStr := ""
	switch state.displayMode {
	case ModeHighImportance:
		modeStr = "仅显示高重要性"
	case ModeAll:
		modeStr = "显示所有数据"
	case ModeWithRates:
		modeStr = "显示高重要性+利率信息"
	case ModeWithImportant:
		modeStr = "显示高重要性+重要事件"
	}

	pauseStr := "运行中"
	if state.isPaused {
		pauseStr = "已暂停"
	}

	// 保存光标位置
	fmt.Print("\033[s")
	
	// 移动到右上角并清除该行
	fmt.Print("\033[1;1H")
	
	// 显示启动时间
	startTimeStr := fmt.Sprintf("财经数据监控系统启动 @ %s", 
		state.startTime.Format("2006-01-02 15:04:05"))
	
	// 计算倒计时
	countdown := formatCountdown(nextRefresh)
	
	// 构建状态文本
	statusText := fmt.Sprintf("模式: %s | 状态: %s | 下次刷新: %s", 
		modeStr, pauseStr, countdown)
	
	// 计算填充空格
	padding := termWidth - len(startTimeStr) - len(statusText)
	if padding < 0 {
		padding = 1
	}
	
	// 输出完整状态行
	fmt.Printf("%s%s%s\n", 
		startTimeStr,
		strings.Repeat(" ", padding),
		statusText)
	
	// 恢复光标位置
	fmt.Print("\033[u")
}

func formatCountdown(nextRefresh time.Time) string {
	duration := time.Until(nextRefresh)
	if duration < 0 {
		return "即将刷新"
	}
	seconds := int(duration.Seconds())
	return fmt.Sprintf("%02d秒", seconds)
}

// Config 配置结构
type Config struct {
	RefreshInterval    int `yaml:"refresh_interval"`
	DefaultDisplayMode int `yaml:"default_display_mode"`
	UI                struct {
		TimeWidth      int `yaml:"time_width"`
		ImportanceWidth int `yaml:"importance_width"`
		ValueWidth     int `yaml:"value_width"`
	} `yaml:"ui"`
}

// 加载配置
func loadConfig() (*Config, error) {
	// 默认配置
	config := &Config{
		RefreshInterval:    15,
		DefaultDisplayMode: 0,
		UI: struct {
			TimeWidth      int `yaml:"time_width"`
			ImportanceWidth int `yaml:"importance_width"`
			ValueWidth     int `yaml:"value_width"`
		}{
			TimeWidth:      6,
			ImportanceWidth: 4,
			ValueWidth:     12,
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
	// 创建缓冲区
	var buf bytes.Buffer
	
	// 获取当前时间
	now := time.Now()
	dateStr := now.Format("20060102")

	// 清屏，但写入缓冲区
	buf.WriteString("\033[2J")  // 清屏
	buf.WriteString("\033[H")   // 移动光标到开始位置
	
	buf.WriteString(fmt.Sprintf("财经数据监控系统启动 @ %s\n\n", now.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("尝试获取 %s 的数据：\n\n", dateStr))

	// 获取HTML内容
	url := fmt.Sprintf("https://rl.fx678.com/date/%s.html", dateStr)
	html, err := fetcher.Fetch(url)
	if err != nil {
		buf.WriteString(fmt.Sprintf("获取数据失败: %v\n", err))
		// 一次性输出缓冲区内容
		fmt.Print(buf.String())
		return
	}

	// 解析数据
	events, importantEvents, rates, err := parser.ParseFinancialCalendar(html)
	if err != nil {
		buf.WriteString(fmt.Sprintf("解析数据失败: %v\n", err))
		// 一次性输出缓冲区内容
		fmt.Print(buf.String())
		return
	}

	// 定义每列的宽度
	timeWidth := config.UI.TimeWidth
	importanceWidth := config.UI.ImportanceWidth
	valueWidth := config.UI.ValueWidth

	// 打印财经日历事件
	buf.WriteString("\n=== 财经日历事件 ===\n")
	buf.WriteString(fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-*s  %s\n",
		timeWidth, "时间",
		importanceWidth, "重要性",
		valueWidth, "前值",
		valueWidth, "预测",
		valueWidth, "公布值",
		"指标名称"))
	buf.WriteString(strings.Repeat("-", 120) + "\n")

	currentTime := ""
	lastEvent := false
	for i, event := range events {
		// 根据显示模式过滤事件
		if state.displayMode == ModeHighImportance && event.Importance != "高" {
			continue
		}

		if event.Time != currentTime {
			if currentTime != "" && !lastEvent {
				buf.WriteString(strings.Repeat("-", 120) + "\n")
			}
			currentTime = event.Time
		}

		lastEvent = (i == len(events)-1)

		// 使用color.New而不是color.Red，这样可以写入缓冲区
		c := color.New(color.FgRed)
		c.Fprintf(&buf, "%-*s  %-*s  %-*s  %-*s  %-*s  %s\n",
			timeWidth, event.Time,
			importanceWidth, event.Importance,
			valueWidth, formatWidth(event.Previous, valueWidth-2),
			valueWidth, formatWidth(event.Forecast, valueWidth-2),
			valueWidth, formatWidth(event.Actual, valueWidth-2),
			event.Indicator)
	}

	// 根据显示模式显示其他信息
	if state.displayMode == ModeWithImportant || state.displayMode == ModeAll {
		if len(importantEvents) > 0 {
			buf.WriteString("\n=== 重要事件 ===\n")
			buf.WriteString(fmt.Sprintf("%-*s  %-*s  %s\n",
				timeWidth, "时间",
				importanceWidth, "重要性",
				"事件"))
			buf.WriteString(strings.Repeat("-", 100) + "\n")

			for _, event := range importantEvents {
				if event.Importance == "高" || state.displayMode == ModeAll {
					c := color.New(color.FgRed)
					c.Fprintf(&buf, "%-*s  %-*s  %s\n",
						timeWidth, event.Time,
						importanceWidth, event.Importance,
						event.Event)
				}
			}
		}
	}

	if state.displayMode == ModeWithRates || state.displayMode == ModeAll {
		buf.WriteString("\n=== 央行利率信息 ===\n")
		
		// 定义列宽
		bankWidth := 20      // 增加央行名称列宽
		rateWidth := 10      // 调整利率列宽
		changeWidth := 8     // 调整变动列宽
		dateWidth := 12      // 保持日期列宽
		historyWidth := 20   // 历史区间列宽
		
		// 打印表头
		buf.WriteString(fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-*s  %-*s  %-*s\n",
			bankWidth, "央行/利率类型",
			rateWidth, "当前利率",
			rateWidth, "前值",
			changeWidth, "变动",
			dateWidth, "变动日期",
			historyWidth, "历史区间",
			rateWidth, "下次预测"))
		buf.WriteString(strings.Repeat("-", 120) + "\n")

		for _, rate := range rates {
			// 提取日期和变动值
			parts := strings.Split(rate.LastChange, " ")
			changeDate := ""
			changeValue := ""
			if len(parts) >= 2 {
				changeDate = parts[1]
				changeValue = parts[0]
			}

			// 格式化历史区间
			historyRange := fmt.Sprintf("%s - %s", rate.HistoryLow, rate.HistoryHigh)
			
			// 组合央行名称和利率类型
			bankInfo := fmt.Sprintf("%s - %s", rate.Bank, rate.RateName)
			
			c := color.New(color.FgYellow)
			c.Fprintf(&buf, "%-*s  %-*s  %-*s  %-*s  %-*s  %-*s  %-*s",
				bankWidth, bankInfo,
				rateWidth, rate.CurrentRate,
				rateWidth, rate.PreviousRate,
				changeWidth, changeValue,
				dateWidth, changeDate,
				historyWidth, historyRange,
				rateWidth, rate.NextForecast)
			
			// 如果有CPI信息，在行尾显示
			if rate.LatestCPI != "" {
				buf.WriteString(fmt.Sprintf("  (CPI: %s)", rate.LatestCPI))
			}
			buf.WriteString("\n")
			
			// 添加分隔线
			buf.WriteString(strings.Repeat("-", 120) + "\n")
		}
	}
	
	// 一次性输出缓冲区内容
	fmt.Print(buf.String())
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 创建应用状态
	state := &AppState{
		displayMode: DisplayMode(config.DefaultDisplayMode),
		isPaused:   false,
		shouldExit: false,
		startTime:  time.Now(),
	}

	// 创建一个新的fetcher
	fetcher := &htmlfetcher.DefaultFetcher{}

	// 设置键盘监听
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	// 创建一个channel用于接收键盘事件
	keyChan := make(chan keyboard.Key)
	go func() {
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				continue
			}
			if key != 0 {
				keyChan <- key
			} else {
				keyChan <- keyboard.Key(char)
			}
		}
	}()

	// 创建一个channel用于接收定时器事件
	timerChan := make(chan bool)
	go func() {
		for {
			time.Sleep(time.Second)
			timerChan <- true
		}
	}()

	// 主循环
	nextRefresh := time.Now().Add(time.Duration(config.RefreshInterval) * time.Second)
	for {
		if !state.isPaused {
			if time.Now().After(nextRefresh) {
				// 执行刷新
				displayData(fetcher, state, config)
				nextRefresh = time.Now().Add(time.Duration(config.RefreshInterval) * time.Second)
				state.nextRefreshTime = nextRefresh
			}
		}

		select {
		case key := <-keyChan:
			switch key {
			case 'q':
				return
			case 'r':
				nextRefresh = time.Now()
			case 'p':
				state.togglePause()
			case 'm':
				state.nextMode()
				displayData(fetcher, state, config)
			case 'h':
				showHelp()
			}
		case <-timerChan:
			// 每秒更新状态显示
			showStatus(state, nextRefresh)
		default:
			// 避免CPU占用过高
			time.Sleep(100 * time.Millisecond)
		}
	}
}
