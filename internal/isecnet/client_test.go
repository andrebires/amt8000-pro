package isecnet

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckAuthResponse(t *testing.T) {
	tests := []struct {
		name    string
		frame   Frame
		wantErr string
	}{
		{
			name:  "success",
			frame: Frame{Command: cmdAuth, Payload: []byte{0x00}},
		},
		{
			name:    "invalid password",
			frame:   Frame{Command: cmdAuth, Payload: []byte{0x01}},
			wantErr: ErrInvalidAuth.Error(),
		},
		{
			name:    "unexpected code",
			frame:   Frame{Command: cmdAuth, Payload: []byte{0x7f}},
			wantErr: "panel rejected authentication: code=0x7f",
		},
		{
			name:    "unexpected command",
			frame:   Frame{Command: cmdStatus, Payload: []byte{0x00}},
			wantErr: "unexpected auth response command 0x0b4a",
		},
		{
			name:    "empty payload",
			frame:   Frame{Command: cmdAuth},
			wantErr: "empty auth response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkAuthResponse(tt.frame)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("checkAuthResponse returned error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("checkAuthResponse returned nil error, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("checkAuthResponse error = %q, want containing %q", err.Error(), tt.wantErr)
			}
			if tt.wantErr == ErrInvalidAuth.Error() && !errors.Is(err, ErrInvalidAuth) {
				t.Fatalf("checkAuthResponse error = %v, want ErrInvalidAuth", err)
			}
		})
	}
}

func TestStatusPayloadTooShortError(t *testing.T) {
	err := statusPayloadTooShortError(Frame{Command: cmdStatus, Payload: []byte{0x01}})
	if err == nil {
		t.Fatal("statusPayloadTooShortError returned nil")
	}
	for _, want := range []string{"command=0x0b4a", "code=0x01", "got 1 want at least 143"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("statusPayloadTooShortError = %q, want containing %q", err.Error(), want)
		}
	}
}
