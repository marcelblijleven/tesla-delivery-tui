package data

import "testing"

func TestLookupStoreName_ZeroID(t *testing.T) {
	if got := GetStoreName("0"); got != "N/A" {
		t.Errorf("LookupStoreName(%q) = %q, want %q", "0", got, "N/A")
	}
}

func TestLookupStoreName_Passthrough(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{"human-readable name", "Tilburg-Asteriastraat", "Tilburg-Asteriastraat"},
		{"numeric ID", "999999", "999999"},
		{"delivery center name", "München Freiham", "München Freiham"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStoreName(tt.id); got != tt.want {
				t.Errorf("LookupStoreName(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
