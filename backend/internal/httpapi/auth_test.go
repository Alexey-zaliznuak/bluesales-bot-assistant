package httpapi

import "testing"

func TestValidateRegistration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		login    string
		password string
		wantErr  bool
	}{
		{name: "valid", login: "new-user", password: "strong-password", wantErr: false},
		{name: "unicode credentials", login: "алексей", password: "пароль-секрет", wantErr: false},
		{name: "short login", login: "ab", password: "strong-password", wantErr: true},
		{name: "long login", login: string(make([]rune, 65)), password: "strong-password", wantErr: true},
		{name: "short password", login: "new-user", password: "short", wantErr: true},
		{name: "bcrypt byte limit", login: "new-user", password: "яяяяяяяяяяяяяяяяяяяяяяяяяяяяяяяяяяяяя", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateRegistration(tt.login, tt.password); (err != nil) != tt.wantErr {
				t.Fatalf("validateRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
