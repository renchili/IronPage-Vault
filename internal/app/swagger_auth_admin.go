package app

// healthSwagger documents the health endpoint.
// @Summary Health check
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /healthz [get]
func healthSwagger() {}

// loginSwagger documents local login.
// @Summary Local login
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/auth/login [post]
func loginSwagger() {}

// logoutSwagger documents logout.
// @Summary Logout
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/auth/logout [post]
func logoutSwagger() {}

// meSwagger documents current principal.
// @Summary Current principal
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/auth/me [get]
func meSwagger() {}

// createUserSwagger documents user creation.
// @Summary Create user
// @Tags admin
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/admin/users [post]
func createUserSwagger() {}

// listUsersSwagger documents user listing.
// @Summary List users
// @Tags admin
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/admin/users [get]
func listUsersSwagger() {}

// listConfigSwagger documents config listing.
// @Summary List config
// @Tags admin
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/admin/config [get]
func listConfigSwagger() {}

// patchConfigSwagger documents config patch.
// @Summary Patch Admin-managed config
// @Description Admin manages pagination.default_page_size, pagination.max_page_size, backup.schedule_enabled, and backup.interval. Pagination must satisfy 1 <= default <= max <= 100. Backup interval must be between 1m and 168h and is reloaded by the scheduler without restart. backup.local_volume is deployment-owned and read-only.
// @Tags admin
// @Security BearerAuth
// @Param key path string true "config key"
// @Param request body map[string]interface{} true "value"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/admin/config/{key} [patch]
func patchConfigSwagger() {}
