package parser

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CalendarEvent 表示一个财经日历事件
type CalendarEvent struct {
	Time        string  // 时间
	Region      string  // 地区
	Indicator   string  // 指标
	Previous    string  // 前值
	Forecast    string  // 预测值
	Actual      string  // 公布值
	Importance  string  // 重要性
	Impact      string  // 利多利空
	Description string  // 解读
}

// ImportantEvent 表示一个重要事件
type ImportantEvent struct {
	Time       string // 时间
	Region     string // 国家地区
	Location   string // 地点
	Importance string // 重要性
	Event      string // 事件内容
}

// CentralBankRate 表示央行利率信息
type CentralBankRate struct {
	Bank           string    // 央行名称
	RateName       string    // 利率名称
	CurrentRate    string    // 当前值
	PreviousRate   string    // 前次值
	LastChange     string    // 最近非0变动基点
	HistoryHigh    string    // 历史峰值
	HistoryLow     string    // 历史最低
	NextForecast   string    // 下次预测值
	LatestCPI      string    // CPI最新值
	LastUpdateTime time.Time // 最后更新时间
}

// ParseFinancialCalendar 解析财经日历页面
func ParseFinancialCalendar(html string) ([]CalendarEvent, []ImportantEvent, []CentralBankRate, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, nil, nil, err
	}

	var events []CalendarEvent
	var importantEvents []ImportantEvent
	var rates []CentralBankRate

	// 解析财经日历事件
	doc.Find("table.cjsj_tab tr").Each(func(i int, tr *goquery.Selection) {
		// 跳过表头
		if i == 0 {
			return
		}

		cells := tr.Find("td")
		if cells.Length() >= 7 {
			event := CalendarEvent{
				Time:        strings.TrimSpace(cells.Eq(0).Text()),
				Region:      strings.TrimSpace(cells.Eq(1).Text()),
				Indicator:   strings.TrimSpace(cells.Eq(2).Text()),
				Previous:    strings.TrimSpace(cells.Eq(3).Text()),
				Forecast:    strings.TrimSpace(cells.Eq(4).Text()),
				Actual:      strings.TrimSpace(cells.Eq(5).Text()),
				Importance:  strings.TrimSpace(cells.Eq(6).Text()),
				Impact:      strings.TrimSpace(cells.Eq(7).Text()),
				Description: strings.TrimSpace(cells.Eq(8).Text()),
			}
			if event.Time != "" && event.Indicator != "" {
				events = append(events, event)
			}
		}
	})

	// 解析重要事件
	doc.Find("table.cjsj_tab2 tr").Each(func(i int, tr *goquery.Selection) {
		// 跳过表头和无效行
		if i <= 1 || !strings.Contains(tr.Text(), "|") {
			return
		}

		cells := tr.Find("td")
		if cells.Length() >= 5 {
			event := ImportantEvent{
				Time:       strings.TrimSpace(cells.Eq(0).Text()),
				Region:     strings.TrimSpace(cells.Eq(1).Text()),
				Location:   strings.TrimSpace(cells.Eq(2).Text()),
				Importance: strings.TrimSpace(cells.Eq(3).Text()),
				Event:      strings.TrimSpace(cells.Eq(4).Text()),
			}
			if event.Time != "" && event.Event != "" {
				importantEvents = append(importantEvents, event)
			}
		}
	})

	// 解析央行利率信息
	doc.Find("table.cjsj_tab2").Last().Find("tr").Each(func(i int, tr *goquery.Selection) {
		// 跳过表头
		if i == 0 {
			return
		}

		cells := tr.Find("td")
		if cells.Length() >= 9 {
			rate := CentralBankRate{
				Bank:         strings.TrimSpace(cells.Eq(0).Text()),
				RateName:     strings.TrimSpace(cells.Eq(1).Text()),
				CurrentRate:  strings.TrimSpace(cells.Eq(2).Text()),
				PreviousRate: strings.TrimSpace(cells.Eq(3).Text()),
				LastChange:   strings.TrimSpace(cells.Eq(4).Text()),
				HistoryHigh:  strings.TrimSpace(cells.Eq(5).Text()),
				HistoryLow:   strings.TrimSpace(cells.Eq(6).Text()),
				NextForecast: strings.TrimSpace(cells.Eq(7).Text()),
				LatestCPI:   strings.TrimSpace(cells.Eq(8).Text()),
			}
			if rate.Bank != "" && rate.RateName != "" {
				rates = append(rates, rate)
			}
		}
	})

	return events, importantEvents, rates, nil
}
