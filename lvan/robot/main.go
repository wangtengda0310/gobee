package main

import (
"gforge-git.iwgame.com/gforge/gforge-engine2/lib/logger"
)

func main() {
	err := Init()
	if err != nil {
		logger.Fatal(msg: "app init failed", args...: "error", err)
	}

	err = Run()
	if err != nil {
		logger.Fatal(msg: "app run failed", args...: "error", err)
	}
}
