package phone_test

import (
	"testing"

	"github.com/meetoria/meetoria/backend/pkg/phone"
)

func TestNormalizeE164(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "lt display format", input: "+370 123 12345", want: "+37012312345"},
		{name: "lt dashed format", input: "+370-123-12345", want: "+37012312345"},
		{name: "lt e164 format", input: "+37012312345", want: "+37012312345"},
		{name: "local leading zero", input: "012312345", want: "+37012312345"},
		{name: "uk number", input: "+44 20 7946 0958", want: "+442079460958"},
		{name: "us number", input: "+1 202 555 0123", want: "+12025550123"},
		{name: "too short", input: "+1234", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := phone.NormalizeE164(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
