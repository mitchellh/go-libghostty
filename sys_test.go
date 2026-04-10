package libghostty

import "testing"

func TestSysLogLevelString(t *testing.T) {
	tests := []struct {
		level SysLogLevel
		want  string
	}{
		{SysLogLevelError, "error"},
		{SysLogLevelWarning, "warning"},
		{SysLogLevelInfo, "info"},
		{SysLogLevelDebug, "debug"},
		{SysLogLevel(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("SysLogLevel(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestSysSetLogNil(t *testing.T) {
	// Clearing with nil should always succeed.
	if err := SysSetLog(nil); err != nil {
		t.Fatalf("SysSetLog(nil) = %v, want nil", err)
	}
}

func TestSysSetLogStderr(t *testing.T) {
	if err := SysSetLogStderr(); err != nil {
		t.Fatalf("SysSetLogStderr() = %v, want nil", err)
	}

	// Clean up.
	if err := SysSetLog(nil); err != nil {
		t.Fatalf("SysSetLog(nil) = %v, want nil", err)
	}
}

func TestSysSetLogCallback(t *testing.T) {
	// Install a Go callback and verify it can be set and cleared.
	called := false
	if err := SysSetLog(func(level SysLogLevel, scope string, message string) {
		called = true
	}); err != nil {
		t.Fatalf("SysSetLog(fn) = %v, want nil", err)
	}

	// We can't easily trigger an internal log message, but we can
	// verify the callback was installed by clearing it.
	_ = called

	if err := SysSetLog(nil); err != nil {
		t.Fatalf("SysSetLog(nil) = %v, want nil", err)
	}
}

func TestSysSetDecodePngNil(t *testing.T) {
	// Clearing with nil should always succeed.
	if err := SysSetDecodePng(nil); err != nil {
		t.Fatalf("SysSetDecodePng(nil) = %v, want nil", err)
	}
}

func TestSysSetDecodePngCallback(t *testing.T) {
	// Install a Go decode-PNG callback and verify it can be set and cleared.
	if err := SysSetDecodePng(func(data []byte) (*SysImage, error) {
		return &SysImage{Width: 1, Height: 1, Data: []byte{0, 0, 0, 255}}, nil
	}); err != nil {
		t.Fatalf("SysSetDecodePng(fn) = %v, want nil", err)
	}

	// Clean up.
	if err := SysSetDecodePng(nil); err != nil {
		t.Fatalf("SysSetDecodePng(nil) = %v, want nil", err)
	}
}
