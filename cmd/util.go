package cmd

import (
	"log"
	"os"
	"path/filepath"
)

// bindPath bind '~' to home directory.
func bindPath(path string) string {
	if path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("error bind path:", err)
		}

		path = filepath.Join(homeDir, path[1:])
	}

	return path
}
