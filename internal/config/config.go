package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddress    string
	DatabaseURL    string
	JWTSecret      string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	LabName        string
	Timezone       string
	StoragePath    string
	MaxUploadBytes int64
}

func Load() (Config, error) {
	c := Config{HTTPAddress: env("HTTP_ADDRESS", ":18080"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: env("JWT_SECRET", "development-secret-change-me"), AccessTTL: 15 * time.Minute, RefreshTTL: 7 * 24 * time.Hour, LabName: env("LAB_NAME", "东海材料检测实验室"), Timezone: env("LAB_TIMEZONE", "Asia/Tokyo"), StoragePath: env("STORAGE_PATH", "./data/uploads"), MaxUploadBytes: 10 << 20}
	if raw := os.Getenv("ACCESS_TTL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return c, fmt.Errorf("ACCESS_TTL: %w", err)
		}
		c.AccessTTL = d
	}
	if raw := os.Getenv("REFRESH_TTL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return c, fmt.Errorf("REFRESH_TTL: %w", err)
		}
		c.RefreshTTL = d
	}
	if raw := os.Getenv("MAX_UPLOAD_BYTES"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return c, fmt.Errorf("MAX_UPLOAD_BYTES: %w", err)
		}
		c.MaxUploadBytes = n
	}
	if c.JWTSecret == "" {
		return c, fmt.Errorf("JWT_SECRET required")
	}
	return c, nil
}
func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
