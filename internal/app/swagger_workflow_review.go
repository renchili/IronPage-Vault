package app

// rollbackVersionSwagger documents rollback.
// @Summary Roll back document version
// @Tags versions
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/documents/{id}/rollback [post]
func rollbackVersionSwagger() {}

// finalizeDocumentSwagger documents finalization.
// @Summary Finalize document
// @Tags workflow
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/documents/{id}/finalize [post]
func finalizeDocumentSwagger() {}

// transitionDocumentSwagger documents workflow transition.
// @Summary Transition document workflow
// @Tags workflow
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/{id}/workflow/transition [post]
func transitionDocumentSwagger() {}

// proposeRedactionSwagger documents redaction proposal.
// @Summary Stage redaction region
// @Tags redactions
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/{id}/redactions [post]
func proposeRedactionSwagger() {}

// listRedactionsSwagger documents redaction listing.
// @Summary List redaction proposals
// @Tags redactions
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/documents/{id}/redactions [get]
func listRedactionsSwagger() {}

// confirmRedactionSwagger documents redaction confirmation.
// @Summary Confirm redaction burn-in
// @Tags redactions
// @Security BearerAuth
// @Param id path string true "document id"
// @Param redaction_id path string true "redaction id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/documents/{id}/redactions/{redaction_id}/confirm [post]
func confirmRedactionSwagger() {}
