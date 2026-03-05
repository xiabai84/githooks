package buildinfo

import (
	"encoding/json"
	"testing"
)

func TestGetBuildInfo(t *testing.T) {
	bi := GetBuildInfo()
	if bi.Version == "" {
		t.Error("expected non-empty Version")
	}
	if bi.GoOs == "" {
		t.Error("expected non-empty GoOs")
	}
	if bi.GoArch == "" {
		t.Error("expected non-empty GoArch")
	}
	if bi.BuildDate == "" {
		t.Error("expected non-empty BuildDate")
	}
}

func TestBuildInfoString(t *testing.T) {
	bi := GetBuildInfo()
	s := bi.String()
	if s == "" {
		t.Error("expected non-empty string from String()")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		t.Errorf("String() output is not valid JSON: %v", err)
	}

	if parsed["version"] != bi.Version {
		t.Errorf("expected version %q, got %q", bi.Version, parsed["version"])
	}
}
