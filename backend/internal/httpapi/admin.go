package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

const adminDateLayout = "2006-01-02"

var moscowLocation = time.FixedZone("Europe/Moscow", 3*60*60)

type adminDateRange struct {
	From time.Time
	To   time.Time
}

type adminDashboardResponse struct {
	From        string                  `json:"from"`
	To          string                  `json:"to"`
	Timezone    string                  `json:"timezone"`
	TotalTokens int64                   `json:"totalTokens"`
	Daily       []store.DailyTokenUsage `json:"daily"`
	Users       []store.UserTokenUsage  `json:"users"`
}

func parseAdminDateRange(now time.Time, query url.Values) (adminDateRange, error) {
	fromValue := strings.TrimSpace(query.Get("from"))
	toValue := strings.TrimSpace(query.Get("to"))

	today := now.In(moscowLocation)
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, moscowLocation)
	if fromValue == "" && toValue == "" {
		return adminDateRange{From: today.AddDate(0, 0, -29), To: today}, nil
	}
	if fromValue == "" || toValue == "" {
		return adminDateRange{}, fmt.Errorf("укажите обе даты: from и to")
	}

	from, err := time.ParseInLocation(adminDateLayout, fromValue, moscowLocation)
	if err != nil {
		return adminDateRange{}, fmt.Errorf("дата from должна быть в формате YYYY-MM-DD")
	}
	to, err := time.ParseInLocation(adminDateLayout, toValue, moscowLocation)
	if err != nil {
		return adminDateRange{}, fmt.Errorf("дата to должна быть в формате YYYY-MM-DD")
	}
	if from.After(to) {
		return adminDateRange{}, fmt.Errorf("дата from не может быть позже to")
	}
	return adminDateRange{From: from, To: to}, nil
}

func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	dateRange, err := parseAdminDateRange(time.Now(), r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	endExclusive := dateRange.To.AddDate(0, 0, 1)
	dailyFromStore, err := s.store.DailyTokenUsage(r.Context(), dateRange.From, endExclusive)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	now := time.Now().In(moscowLocation)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, moscowLocation)
	users, err := s.store.UserTokenUsage(r.Context(), monthStart, monthStart.AddDate(0, 1, 0))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	dailyByDate := make(map[string]int64, len(dailyFromStore))
	for _, item := range dailyFromStore {
		dailyByDate[item.Date] = item.TotalTokens
	}

	daily := make([]store.DailyTokenUsage, 0, int(dateRange.To.Sub(dateRange.From).Hours()/24)+1)
	var totalTokens int64
	for day := dateRange.From; !day.After(dateRange.To); day = day.AddDate(0, 0, 1) {
		date := day.Format(adminDateLayout)
		tokens := dailyByDate[date]
		totalTokens += tokens
		daily = append(daily, store.DailyTokenUsage{Date: date, TotalTokens: tokens})
	}

	writeJSON(w, http.StatusOK, adminDashboardResponse{
		From:        dateRange.From.Format(adminDateLayout),
		To:          dateRange.To.Format(adminDateLayout),
		Timezone:    "Europe/Moscow",
		TotalTokens: totalTokens,
		Daily:       daily,
		Users:       users,
	})
}
