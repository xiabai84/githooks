package buildinfo

import (
	"encoding/json"
	"log"
	"runtime"
	"time"
)

var (
	version   = "local build"
	gitCommit = ""
	buildDate = time.Now().UTC().Format(time.RFC3339)
	goos      = runtime.GOOS
	goarch    = runtime.GOARCH
)

type BuildInfo struct {
	Version   string `json:"version,omitempty"`
	GitCommit string `json:"gitCommit,omitempty"`
	BuildDate string `json:"buildDate,omitempty"`
	GoOs      string `json:"goOs,omitempty"`
	GoArch    string `json:"goArch,omitempty"`
}

func GetBuildInfo() BuildInfo {
	return BuildInfo{
		version,
		gitCommit,
		buildDate,
		goos,
		goarch,
	}
}

func (v BuildInfo) String() string {
	j, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}
	return string(j)
}
