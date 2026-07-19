package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type workflowDefinitionInput struct {
	Name    string `json:"name"`
	Mutable bool   `json:"mutable"`
}

func normalizeWorkflowDefinitions(inputs []workflowDefinitionInput) ([]workflowStatusResponse, error) {
	if len(inputs) < 5 {
		return nil, fmt.Errorf("workflow must retain Draft, Under Review, Redaction Pending, Approved, and Finalized")
	}
	definitions := make([]workflowStatusResponse, 0, len(inputs))
	seen := map[string]bool{}
	positions := map[string]int{}
	for index, input := range inputs {
		name := strings.TrimSpace(input.Name)
		if name == "" || len(name) > 100 {
			return nil, fmt.Errorf("workflow status names must contain 1 to 100 characters")
		}
		key := strings.ToLower(name)
		if seen[key] {
			return nil, fmt.Errorf("workflow status names must be unique")
		}
		seen[key] = true
		positions[name] = index
		definitions = append(definitions, workflowStatusResponse{Name: name, Position: index + 1, Mutable: input.Mutable})
	}
	if definitions[0].Name != StatusDraft || !definitions[0].Mutable {
		return nil, fmt.Errorf("Draft must remain the first mutable status")
	}
	last := definitions[len(definitions)-1]
	if last.Name != StatusFinalized || last.Mutable {
		return nil, fmt.Errorf("Finalized must remain the last immutable status")
	}
	for _, definition := range definitions[1 : len(definitions)-1] {
		if !definition.Mutable {
			return nil, fmt.Errorf("only Finalized may be immutable")
		}
	}
	requiredOrder := []string{StatusDraft, StatusUnderReview, StatusRedactionPending, StatusApproved, StatusFinalized}
	previous := -1
	for _, required := range requiredOrder {
		position, ok := positions[required]
		if !ok {
			return nil, fmt.Errorf("workflow must retain required status %s", required)
		}
		if position <= previous {
			return nil, fmt.Errorf("required workflow statuses must preserve their domain order")
		}
		previous = position
	}
	return definitions, nil
}

func (a *App) nextWorkflowDefinition(ctx context.Context, executor sqlx.ExtContext, current string) (workflowStatusResponse, error) {
	var currentPosition int
	if err := sqlx.GetContext(ctx, executor, &currentPosition, `SELECT position FROM workflow_status_definitions WHERE name=$1`, current); err != nil {
		return workflowStatusResponse{}, err
	}
	var next workflowStatusResponse
	if err := sqlx.GetContext(ctx, executor, &next, `SELECT name,position,mutable FROM workflow_status_definitions WHERE position>$1 ORDER BY position LIMIT 1`, currentPosition); err != nil {
		return workflowStatusResponse{}, err
	}
	return next, nil
}

func (a *App) replaceWorkflowStatuses(c echo.Context) error {
	p := principal(c)
	var req struct {
		Statuses []workflowDefinitionInput `json:"statuses"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_WORKFLOW_DEFINITIONS", "statuses are required")
	}
	definitions, err := normalizeWorkflowDefinitions(req.Statuses)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "INVALID_WORKFLOW_DEFINITIONS", err.Error())
	}

	tx, err := a.db.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "TX_ERROR", "could not start transaction")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(c.Request().Context(), `LOCK TABLE workflow_status_definitions IN ACCESS EXCLUSIVE MODE`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_LOCK_ERROR", "could not lock workflow definitions")
	}
	activeStatuses := []string{}
	if err := tx.SelectContext(c.Request().Context(), &activeStatuses, `SELECT DISTINCT status FROM documents`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_QUERY_ERROR", "could not inspect active document statuses")
	}
	allowed := map[string]bool{}
	for _, definition := range definitions {
		allowed[definition.Name] = true
	}
	for _, active := range activeStatuses {
		if !allowed[active] {
			return apiErr(c, http.StatusConflict, "WORKFLOW_STATUS_IN_USE", "new workflow must retain every status used by an existing document")
		}
	}
	if _, err := tx.ExecContext(c.Request().Context(), `DELETE FROM workflow_status_definitions`); err != nil {
		return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_UPDATE_ERROR", "could not replace workflow statuses")
	}
	for _, definition := range definitions {
		if _, err := tx.ExecContext(c.Request().Context(), `INSERT INTO workflow_status_definitions(name,position,mutable) VALUES($1,$2,$3)`, definition.Name, definition.Position, definition.Mutable); err != nil {
			return apiErr(c, http.StatusInternalServerError, "WORKFLOW_STATUS_UPDATE_ERROR", "could not replace workflow statuses")
		}
	}
	names := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		names = append(names, definition.Name)
	}
	if err := a.auditWithExecutor(c, tx, p.UserID, "WORKFLOW_DEFINITIONS_UPDATE", "", map[string]interface{}{"statuses": names}); err != nil {
		return apiErr(c, http.StatusInternalServerError, "AUDIT_CREATE_ERROR", "could not record workflow definition audit")
	}
	if err := tx.Commit(); err != nil {
		return apiErr(c, http.StatusInternalServerError, "COMMIT_ERROR", "could not commit workflow definitions")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": definitions})
}
