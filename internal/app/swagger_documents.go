package app

// listDocumentsSwagger documents document listing.
// @Summary List documents
// @Tags documents
// @Security BearerAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/documents [get]
func listDocumentsSwagger() {}

// uploadDocumentSwagger documents document upload.
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
func uploadDocumentSwagger() {}

// batchUploadDocumentsSwagger documents batch upload.
// @Summary Batch upload documents
// @Tags documents
// @Security BearerAuth
// @Accept multipart/form-data
// @Param files formData file true "PDF files"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/batch [post]
func batchUploadDocumentsSwagger() {}

// compareVersionsSwagger documents version compare.
// @Summary Compare document versions
// @Tags compare
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/documents/compare [post]
func compareVersionsSwagger() {}

// getDocumentSwagger documents metadata fetch.
// @Summary Get document metadata
// @Tags documents
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/documents/{id} [get]
func getDocumentSwagger() {}

// downloadDocumentSwagger documents PDF download.
// @Summary Download current document file
// @Tags documents
// @Security BearerAuth
// @Param id path string true "document id"
// @Success 200 {file} file
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/documents/{id}/file [get]
func downloadDocumentSwagger() {}

// listVersionsSwagger documents version listing.
// @Summary List document versions
// @Tags versions
// @Security BearerAuth
// @Param id path string true "document id"
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/documents/{id}/versions [get]
func listVersionsSwagger() {}
