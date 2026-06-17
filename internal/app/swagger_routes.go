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

// documentItemRoutesSwagger documents document item endpoints.
// @Summary Get document metadata
// @Tags documents
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/documents/{id} [get]
// @Summary Download current document file
// @Tags documents
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {file} file
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/documents/{id}/file [get]
// @Summary List document versions
// @Tags versions
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/documents/{id}/versions [get]
func documentItemRoutesSwagger() {}

// documentMutationRoutesSwagger documents document mutation endpoints.
// @Summary Roll back document version
// @Tags versions
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/documents/{id}/rollback [post]
// @Summary Finalize document
// @Tags workflow
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/documents/{id}/finalize [post]
// @Summary Transition document workflow
// @Tags workflow
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/{id}/workflow/transition [post]
func documentMutationRoutesSwagger() {}

// reviewRoutesSwagger documents redaction endpoints.
// @Summary Stage redaction region
// @Tags redactions
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/{id}/redactions [post]
// @Summary List redaction proposals
// @Tags redactions
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/documents/{id}/redactions [get]
// @Summary Confirm redaction burn-in
// @Tags redactions
// @Security BearerAuth
// @Param id path string true "document id"
// @Param redaction_id path string true "redaction id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/documents/{id}/redactions/{redaction_id}/confirm [post]
func reviewRoutesSwagger() {}

// annotationRoutesSwagger documents annotation endpoints.
// @Summary Create annotation
// @Tags annotations
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/{id}/annotations [post]
// @Summary List annotations
// @Tags annotations
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/{id}/annotations [get]
// @Summary Update annotation disposition
// @Tags annotations
// @Security BearerAuth
// @Param id path string true "annotation id"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/annotations/{id}/disposition [patch]
func annotationRoutesSwagger() {}

// batesCompareRoutesSwagger documents Bates and compare endpoints.
// @Summary Apply Bates numbering
// @Tags bates
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/documents/{id}/bates [post]
// @Summary Compare document versions
// @Tags compare
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/compare [post]
func batesCompareRoutesSwagger() {}

// auditNotificationRoutesSwagger documents audit and notification endpoints.
// @Summary List audit logs
// @Tags audit
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Param user_id query string false "actor user id"
// @Param document_id query string false "document id"
// @Param action_type query string false "action type"
// @Param from query string false "start timestamp"
// @Param to query string false "end timestamp"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/audit-logs [get]
// @Summary List notifications
// @Tags notifications
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/notifications [get]
// @Summary Mark notification as read
// @Tags notifications
// @Security BearerAuth
// @Param id path string true "notification id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/notifications/{id}/read [post]
func auditNotificationRoutesSwagger() {}
