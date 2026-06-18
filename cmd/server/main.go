package main

import "ironpage-vault/internal/app"

// @title IronPage Vault API
// @version 1.0
// @description Offline legal PDF lifecycle management backend API for air-gapped legal and compliance environments.
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := app.LoadConfig()
	app.MustRun(cfg)
}
