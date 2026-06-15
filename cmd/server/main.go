package main

import "ironpage-vault/internal/app"

func main() {
	cfg := app.LoadConfig()
	app.MustRun(cfg)
}
