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
// @Description Uses a durable local lifecycle journal, preserves the requesting Admin as acting user, and reconciles interrupted terminal persistence before startup continues.
// @Tags backup
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/admin/backup/restore [post]
func restoreBackupSwagger() {}
