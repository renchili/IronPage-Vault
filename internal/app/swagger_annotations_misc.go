package app

// createAnnotationSwagger documents annotation creation.
// @Summary Create annotation
// @Tags annotations
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/{id}/annotations [post]
func createAnnotationSwagger() {}

// listAnnotationsSwagger documents annotation listing.
// @Summary List annotations
// @Tags annotations
// @Security BearerAuth
// @Param id path string true "document id"
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/{id}/annotations [get]
func listAnnotationsSwagger() {}

// updateAnnotationDispositionSwagger documents annotation disposition update.
// @Summary Update annotation disposition
// @Tags annotations
// @Security BearerAuth
// @Param id path string true "annotation id"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/annotations/{id}/disposition [patch]
func updateAnnotationDispositionSwagger() {}

// applyBatesVersionSwagger documents Bates numbering.
// @Summary Apply Bates numbering
// @Tags bates
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/documents/{id}/bates [post]
func applyBatesVersionSwagger() {}

// auditLogsSwagger documents audit log listing.
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
func auditLogsSwagger() {}

// notificationsSwagger documents notification listing.
// @Summary List notifications
// @Tags notifications
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/notifications [get]
func notificationsSwagger() {}

// readNotificationSwagger documents mark-as-read.
// @Summary Mark notification as read
// @Tags notifications
// @Security BearerAuth
// @Param id path string true "notification id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/notifications/{id}/read [post]
func readNotificationSwagger() {}
