package app

// workflowStatusesSwagger documents workflow status listing.
// @Summary List workflow statuses
// @Tags admin
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/workflow-statuses [get]
func workflowStatusesSwagger() {}

// replaceWorkflowStatusesSwagger documents complete ordered-chain replacement.
// @Summary Replace workflow status definitions
// @Description Replaces the ordered workflow chain. Draft remains first and mutable; Finalized remains last and immutable; statuses used by existing documents must be retained.
// @Tags admin
// @Security BearerAuth
// @Param request body map[string]interface{} true "ordered workflow statuses"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/workflow-statuses [put]
func replaceWorkflowStatusesSwagger() {}

// notificationTemplatesSwagger documents template listing.
// @Summary List notification templates
// @Tags admin
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/notification-templates [get]
func notificationTemplatesSwagger() {}

// patchNotificationTemplateSwagger documents template patch.
// @Summary Patch notification template
// @Tags admin
// @Security BearerAuth
// @Param key path string true "template key"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/notification-templates/{key} [patch]
func patchNotificationTemplateSwagger() {}

// runBackupSwagger documents backup run.
// @Summary Run backup
// @Description Acquires the exclusive application mutation barrier across the PostgreSQL dump and filesystem snapshot so both artifacts form one application recovery boundary.
// @Tags backup
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/backup/run [post]
func runBackupSwagger() {}

// backupJobsSwagger documents backup jobs.
// @Summary List backup jobs
// @Tags backup
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/admin/backup/jobs [get]
func backupJobsSwagger() {}

// restoreBackupSwagger documents the explicit restore lifecycle.
// @Summary Restore backup
// @Description Enters exclusive HTTP maintenance, drains application requests, blocks new reads and mutations, and uses a durable encrypted lifecycle journal. A crash before a durable platform outcome creates Interrupted rather than falsely asserting Failed.
// @Tags backup
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /api/admin/backup/restore [post]
func restoreBackupSwagger() {}

// resolveInterruptedRestoreSwagger documents operator attestation for an unknown restore outcome.
// @Summary Resolve interrupted restore
// @Description Records an Admin-verified Completed or Failed conclusion for an Interrupted restore. This endpoint does not rerun restore; verification_note must describe the external evidence used.
// @Tags backup
// @Security BearerAuth
// @Param id path string true "restore id"
// @Param request body map[string]interface{} true "status and verification_note"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/backup/restore/{id}/resolve [post]
func resolveInterruptedRestoreSwagger() {}
