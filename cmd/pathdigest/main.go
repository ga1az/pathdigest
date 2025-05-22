package main

import (
	"fmt"
	"os"

	"github.com/ga1az/pathdigest/cmd"
)

var (
	appVersion = "dev"
	// goVersion  = "unknown"
	// commitHash = "unknown"
	// buildDate  = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("pathdigest version: %s\n", appVersion)
		os.Exit(0)
	}

	cmd.SetVersionInfo(appVersion, "built with Go")

	cmd.Execute()
}
