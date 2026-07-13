package organization

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Downtown Barbershop", "downtown-barbershop"},
		{"  Spaces  Everywhere  ", "spaces-everywhere"},
		{"Užupio Kirpykla", "uzupio-kirpykla"},
		{"123 Studio", "123-studio"},
	}

	for _, tt := range tests {
		if got := Slugify(tt.input); got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
