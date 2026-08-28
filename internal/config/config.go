package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"example.com/temporary-share-gateway/internal/model"
)

type Settings struct {
	ListenAddress string
	DatabasePath  string
	Gateway       model.GatewayConfig
}

func Defaults() Settings {
	return Settings{
		ListenAddress: ":8080",
		DatabasePath:  "/tmp/temporary-share-gateway.db",
		Gateway:       model.GatewayConfig{Name: "temporary-share-gateway", MaxBodyBytes: 1 << 20, RequestTimeout: 5 * time.Second, AuditRetention: 1000, RequireRequestID: false},
	}
}

func FromEnv(getenv func(string) string) (Settings, error) {
	settings := Defaults()
	if value := getenv("SHARE_LISTEN_ADDRESS"); value != "" {
		settings.ListenAddress = value
	}
	if value := getenv("SHARE_DATABASE_PATH"); value != "" {
		settings.DatabasePath = value
	}
	if value := getenv("SHARE_REQUIRE_REQUEST_ID"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return settings, fmt.Errorf("parse request id setting: %w", err)
		}
		settings.Gateway.RequireRequestID = parsed
	}
	if value := getenv("SHARE_MAX_BODY_BYTES"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed < 1 {
			return settings, fmt.Errorf("parse max body setting")
		}
		settings.Gateway.MaxBodyBytes = parsed
	}
	if err := settings.Validate(); err != nil {
		return settings, err
	}
	return settings, nil
}

func (s Settings) Validate() error {
	if s.ListenAddress == "" || s.DatabasePath == "" {
		return fmt.Errorf("listen address and database path are required")
	}
	if err := s.Gateway.Validate(); err != nil {
		return err
	}
	return nil
}

func Environment() (Settings, error) { return FromEnv(os.Getenv) }
