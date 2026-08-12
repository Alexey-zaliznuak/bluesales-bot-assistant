package httpapi

import (
	"net/url"
	"testing"
	"time"

	"github.com/alex/bluesales-bot-assistant/backend/internal/config"
	"github.com/alex/bluesales-bot-assistant/backend/internal/store"
)

func TestParseAdminDateRange(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 12, 18, 0, 0, 0, time.UTC)

	t.Run("defaults to last thirty Moscow days", func(t *testing.T) {
		dateRange, err := parseAdminDateRange(now, url.Values{})
		if err != nil {
			t.Fatal(err)
		}
		if got := dateRange.From.Format(adminDateLayout); got != "2026-07-14" {
			t.Fatalf("from = %s, want 2026-07-14", got)
		}
		if got := dateRange.To.Format(adminDateLayout); got != "2026-08-12" {
			t.Fatalf("to = %s, want 2026-08-12", got)
		}
	})

	t.Run("accepts inclusive explicit range", func(t *testing.T) {
		dateRange, err := parseAdminDateRange(now, url.Values{
			"from": {"2026-08-01"},
			"to":   {"2026-08-10"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if dateRange.From.Day() != 1 || dateRange.To.Day() != 10 {
			t.Fatalf("unexpected range: %+v", dateRange)
		}
	})

	t.Run("rejects reversed range", func(t *testing.T) {
		_, err := parseAdminDateRange(now, url.Values{
			"from": {"2026-08-10"},
			"to":   {"2026-08-01"},
		})
		if err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestIsAdmin(t *testing.T) {
	t.Parallel()

	server := &Server{cfg: &config.Config{SeedUserLogin: " Admin "}}
	if !server.isAdmin(&store.User{Login: "admin"}) {
		t.Fatal("seed user must be recognized case-insensitively")
	}
	if server.isAdmin(&store.User{Login: "other"}) {
		t.Fatal("regular user must not be an admin")
	}
	if server.isAdmin(nil) {
		t.Fatal("nil user must not be an admin")
	}
}
