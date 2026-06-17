package app

// healthSwagger documents the health endpoint.
// @Summary Health check
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /healthz [get]
func healthSwagger() {}

// authRoutesSwagger documents auth endpoints.
// @Summary Local login
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/auth/login [post]
// @Summary Logout
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/auth/logout [post]
// @Summary Current principal
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/auth/me [get]
func authRoutesSwagger() {}

// adminRoutesSwagger documents admin endpoints.
// @Summary Create user
// @Tags admin
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/users [post]
// @Summary List users
// @Tags admin
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/users [get]
// @Summary List config
// @Tags admin
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/config [get]
// @Summary Patch config
// @Tags admin
// @Security BearerAuth
// @Param key path string true "config key"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/config/{key} [patch]
func adminRoutesSwagger() {}

// adminWorkflowRoutesSwagger documents workflow and template admin endpoints.
// @Summary List workflow statuses
// @Tags admin
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/workflow-statuses [get]
// @Summary List notification templates
// @Tags admin
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/notification-templates [get]
// @Summary Patch notification template
// @Tags admin
// @Security BearerAuth
// @Param key path string true "template key"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/notification-templates/{key} [patch]
func adminWorkflowRoutesSwagger() {}

// backupRoutesSwagger documents backup endpoints.
// @Summary Run backup
// @Tags backup
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/backup/run [post]
// @Summary List backup jobs
// @Tags backup
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/backup/jobs [get]
// @Summary Restore backup
// @Tags backup
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/backup/restore [post]
func backupRoutesSwagger() {}

// documentRoutesSwagger documents document collection endpoints.
// @Summary List documents
// @Tags documents
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/documents [get]
// @Summary Upload document
// @Tags documents
// @Security BearerAuth
// @Accept multipart/form-data
// @Param title formData string false "title"
// @Param file formData file true "PDF file"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents [post]
// @Summary Batch upload documents
// @Tags documents
// @Security BearerAuth
// @Accept multipart/form-data
// @Param files formData file true "PDF files"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/batch [post]
func documentRoutesSwagger() {}
