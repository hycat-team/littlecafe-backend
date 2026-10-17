package config

import (
	"testing"

	sharedconfig "github.com/hycat-team/littlecafe-backend/shared/config"
)

func TestValidatePrivateKeyByEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		key     string
		wantErr bool
	}{
		{name: "production without key", env: "production", key: "", wantErr: true},
		{name: "production with key", env: "production", key: "/run/littlecafe/certs/jwt-private.pem", wantErr: false},
		{name: "development without key", env: "development", key: "", wantErr: false},
		{name: "development with key", env: "development", key: "/tmp/jwt-private.pem", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Config:  sharedconfig.Config{App: sharedconfig.App{Env: tt.env}},
				Signing: Signing{PrivateKeyFile: tt.key},
			}
			err := cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("Validate() must fail when the production private key is missing")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
		})
	}
}
