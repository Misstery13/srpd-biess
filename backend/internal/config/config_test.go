package config

import (
	"testing"
)

func TestValidateJWTSecret(t *testing.T) {
	tests := []struct {
		name      string
		jwtSecret string
		wantErr   bool
	}{
		{
			name:      "JWT_SECRET vacío debe fallar",
			jwtSecret: "",
			wantErr:   true,
		},
		{
			name:      "JWT_SECRET por defecto (cambia-esto) debe fallar",
			jwtSecret: "cambia-esto",
			wantErr:   true,
		},
		{
			name:      "JWT_SECRET corto (<32 chars) debe fallar",
			jwtSecret: "un-secreto-demasiado-corto",
			wantErr:   true,
		},
		{
			name:      "JWT_SECRET válido (>= 32 chars y no default) debe pasar",
			jwtSecret: "12345678901234567890123456789012",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJWTSecret(tt.jwtSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWTSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
