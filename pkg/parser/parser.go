package parser

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"time"
)

type FinancialData struct {
	PreviousValue string
	ForecastValue string
	ActualValue  string
	Timestamp    string
}

func ParseHTML(html string) (*FinancialData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	data := &FinancialData{}
	doc.Find(".data-row").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			data.PreviousValue = s.Text()
		case 1:
			data.ForecastValue = s.Text()
		case 2:
			data.ActualValue = s.Text()
		}
	})
	data.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	return data, nil
}
