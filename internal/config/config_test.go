package config

import "testing"

func TestEnvironmentSettings(t *testing.T) {
	values := map[string]string{
		"SHARE_LISTEN_ADDRESS":     "127.0.0.1:9000",
		"SHARE_DATABASE_PATH":      "/tmp/fixed.db",
		"SHARE_REQUIRE_REQUEST_ID": "true",
		"SHARE_MAX_BODY_BYTES":     "4096",
	}
	settings, err := FromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatal(err)
	}
	if settings.ListenAddress != "127.0.0.1:9000" || !settings.Gateway.RequireRequestID || settings.Gateway.MaxBodyBytes != 4096 {
		t.Fatalf("settings = %#v", settings)
	}
}

func TestInvalidEnvironmentSetting(t *testing.T) {
	_, err := FromEnv(func(key string) string {
		if key == "SHARE_MAX_BODY_BYTES" {
			return "invalid"
		}
		return ""
	})
	if err == nil {
		t.Fatal("invalid setting accepted")
	}
}
