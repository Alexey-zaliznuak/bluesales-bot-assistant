package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv      string
	APIPort     string
	DatabaseURL string

	CORSAllowedOrigins []string

	SessionCookieName   string
	SessionCookieSecure bool
	SessionTTL          time.Duration

	SeedUserLogin    string
	SeedUserPassword string

	OpenRouter OpenRouter
	Upload     Upload
}

type OpenRouter struct {
	APIKey          string
	BaseURL         string
	Model           string
	ReasoningEffort string
	ProxyURL        string
	Timeout         time.Duration
	// CacheMode: explicit — только помеченные блоки участвуют в кэше,
	// auto — кэш-брейкпоинт ставит провайдер, off — кэширование не запрашиваем.
	CacheMode   string
	CacheTTL    string
	AppTitle    string
	HTTPReferer string
}

type Upload struct {
	MaxSizeBytes      int64
	MaxFiles          int
	AllowedExtensions []string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:      env("APP_ENV", "development"),
		APIPort:     env("API_PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", ""),

		CORSAllowedOrigins: splitList(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),

		SessionCookieName:   env("SESSION_COOKIE_NAME", "bsa_session"),
		SessionCookieSecure: envBool("SESSION_COOKIE_SECURE", false),
		SessionTTL:          time.Duration(envInt("SESSION_TTL_HOURS", 720)) * time.Hour,

		SeedUserLogin:    env("SEED_USER_LOGIN", ""),
		SeedUserPassword: env("SEED_USER_PASSWORD", ""),

		OpenRouter: OpenRouter{
			APIKey:          env("OPENROUTER_API_KEY", ""),
			BaseURL:         strings.TrimRight(env("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"), "/"),
			Model:           env("OPENROUTER_MODEL", "openai/gpt-5.6-luna"),
			ReasoningEffort: env("OPENROUTER_REASONING_EFFORT", "max"),
			ProxyURL:        strings.TrimSpace(env("OPENROUTER_PROXY_URL", "")),
			Timeout:         time.Duration(envInt("OPENROUTER_TIMEOUT_SECONDS", 300)) * time.Second,
			CacheMode:       env("OPENROUTER_CACHE_MODE", "explicit"),
			CacheTTL:        env("OPENROUTER_CACHE_TTL", "30m"),
			AppTitle:        env("OPENROUTER_APP_TITLE", "BlueSales Bot Assistant"),
			HTTPReferer:     env("OPENROUTER_HTTP_REFERER", ""),
		},

		Upload: Upload{
			MaxSizeBytes:      int64(envInt("MAX_UPLOAD_SIZE_MB", 5)) * 1024 * 1024,
			MaxFiles:          envInt("MAX_UPLOAD_FILES", 10),
			AllowedExtensions: normalizeExtensions(splitList(env("ALLOWED_UPLOAD_EXTENSIONS", ".txt,.md,.csv,.json,.yaml,.yml,.log,.xml"))),
		},
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL не задан")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return n
}

func envBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return b
}

func splitList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func normalizeExtensions(exts []string) []string {
	out := make([]string, 0, len(exts))
	for _, e := range exts {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		out = append(out, e)
	}
	return out
}
