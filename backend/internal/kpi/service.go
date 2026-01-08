package kpi

import (
	"context"
	"time"
)

// DailyKPIResponse represents the response for daily KPI endpoint
type DailyKPIResponse struct {
	FreshnessTimestamp time.Time    `json:"freshnessTimestamp"`
	Range              string       `json:"range"`
	Totals             *KPITotals   `json:"totals"`
	ByChannel          []KPISummary `json:"byChannel"`
	ByDaypart          []KPISummary `json:"byDaypart"`
}

// Service handles KPI business logic
type Service struct {
	store *Store
}

// NewService creates a new KPI service
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// GetDailyKPIs retrieves KPIs for a date range with channel/daypart breakdowns
func (s *Service) GetDailyKPIs(ctx context.Context, startDate, endDate time.Time, rangeLabel string) (*DailyKPIResponse, error) {
	// Get totals
	totals, err := s.store.GetTotals(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get by channel
	byChannel, err := s.store.GetByChannel(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get by daypart
	byDaypart, err := s.store.GetByDaypart(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate percentages for breakdowns
	if totals.Revenue > 0 {
		for i := range byChannel {
			byChannel[i].Revenue = roundTo2(byChannel[i].Revenue)
			byChannel[i].COGS = roundTo2(byChannel[i].COGS)
			byChannel[i].GrossMargin = roundTo2(byChannel[i].GrossMargin)
			byChannel[i].LaborCost = roundTo2(byChannel[i].LaborCost)
			byChannel[i].Opex = roundTo2(byChannel[i].Opex)
			byChannel[i].NetProfit = roundTo2(byChannel[i].NetProfit)
			byChannel[i].AvgCheck = roundTo2(byChannel[i].AvgCheck)
		}
		for i := range byDaypart {
			byDaypart[i].Revenue = roundTo2(byDaypart[i].Revenue)
			byDaypart[i].COGS = roundTo2(byDaypart[i].COGS)
			byDaypart[i].GrossMargin = roundTo2(byDaypart[i].GrossMargin)
			byDaypart[i].LaborCost = roundTo2(byDaypart[i].LaborCost)
			byDaypart[i].Opex = roundTo2(byDaypart[i].Opex)
			byDaypart[i].NetProfit = roundTo2(byDaypart[i].NetProfit)
			byDaypart[i].AvgCheck = roundTo2(byDaypart[i].AvgCheck)
		}
	}

	// Round totals
	totals.Revenue = roundTo2(totals.Revenue)
	totals.COGS = roundTo2(totals.COGS)
	totals.GrossMargin = roundTo2(totals.GrossMargin)
	totals.LaborCost = roundTo2(totals.LaborCost)
	totals.LaborPct = roundTo2(totals.LaborPct)
	totals.Opex = roundTo2(totals.Opex)
	totals.NetProfit = roundTo2(totals.NetProfit)
	totals.AvgCheck = roundTo2(totals.AvgCheck)

	return &DailyKPIResponse{
		FreshnessTimestamp: totals.FreshnessTimestamp,
		Range:              rangeLabel,
		Totals:             totals,
		ByChannel:          byChannel,
		ByDaypart:          byDaypart,
	}, nil
}

// ParseDateRange converts a range string to start/end dates
func ParseDateRange(rangeStr string, referenceDate time.Time) (start, end time.Time) {
	// Use Brisbane timezone
	loc, _ := time.LoadLocation("Australia/Brisbane")
	ref := referenceDate.In(loc)

	// End is always end of reference date
	end = time.Date(ref.Year(), ref.Month(), ref.Day(), 23, 59, 59, 0, loc)

	switch rangeStr {
	case "30d":
		start = end.AddDate(0, 0, -30)
	case "ytd":
		start = time.Date(ref.Year(), 1, 1, 0, 0, 0, 0, loc)
	case "trailing12m":
		start = end.AddDate(-1, 0, 0)
	default:
		// Default to 30 days
		start = end.AddDate(0, 0, -30)
	}

	// Start from beginning of day
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)

	return start, end
}

func roundTo2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
