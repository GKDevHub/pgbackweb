package backups

import (
	"net/http"
	"strconv"

	"github.com/eduardolat/pgbackweb/internal/util/echoutil"
	"github.com/eduardolat/pgbackweb/internal/view/web/component"
	"github.com/eduardolat/pgbackweb/internal/view/web/respondhtmx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	nodx "github.com/nodxdev/nodxgo"
	htmx "github.com/nodxdev/nodxgo-htmx"
	lucide "github.com/nodxdev/nodxgo-lucide"
)

func (h *handlers) getDuplicateBackupFormHandler(c echo.Context) error {
	ctx := c.Request().Context()

	backupIDStr := c.Param("backupID")
	backupID, err := uuid.Parse(backupIDStr)
	if err != nil {
		return respondhtmx.ToastError(c, "Invalid backup ID")
	}

	originalBackup, err := h.servs.BackupsService.GetBackup(ctx, backupID)
	if err != nil {
		// TODO: Differentiate between not found and other errors
		// For now, a generic error is fine for this task.
		return respondhtmx.ToastError(c, "Failed to fetch original backup: "+err.Error())
	}

	prefillData := CreateBackupFormValues{
		Name:           originalBackup.Name + " (copy)",
		IsActive:       "false", // As per instructions
		DatabaseID:     originalBackup.DatabaseID.String(),
		IsLocal:        strconv.FormatBool(originalBackup.IsLocal),
		CronExpression: originalBackup.CronExpression,
		TimeZone:       originalBackup.TimeZone,
		DestDir:        originalBackup.DestDir, // Assuming DestDir is string in dbgen.Backup
		RetentionDays:  strconv.Itoa(int(originalBackup.RetentionDays)),
		OptDataOnly:    strconv.FormatBool(originalBackup.OptDataOnly),
		OptSchemaOnly:  strconv.FormatBool(originalBackup.OptSchemaOnly),
		OptClean:       strconv.FormatBool(originalBackup.OptClean),
		OptIfExists:    strconv.FormatBool(originalBackup.OptIfExists),
		OptCreate:      strconv.FormatBool(originalBackup.OptCreate),
		OptNoComments:  strconv.FormatBool(originalBackup.OptNoComments),
	}

	if originalBackup.DestinationID.Valid {
		prefillData.DestinationID = originalBackup.DestinationID.UUID.String()
	} else {
		prefillData.DestinationID = ""
	}

	databases, err := h.servs.DatabasesService.GetAllDatabases(ctx)
	if err != nil {
		return respondhtmx.ToastError(c, "Failed to fetch databases: "+err.Error())
	}

	destinations, err := h.servs.DestinationsService.GetAllDestinations(ctx)
	if err != nil {
		return respondhtmx.ToastError(c, "Failed to fetch destinations: "+err.Error())
	}

	// createBackupForm is in the same 'backups' package (create_backup.go)
	// CreateBackupFormValues is also in the same package (create_backup.go)
	formNode := createBackupForm(databases, destinations, &prefillData)
	return echoutil.RenderNodx(c, http.StatusOK, formNode)
}

func duplicateBackupButton(backupID uuid.UUID) nodx.Node {
	// Each button gets its own modal, content loaded dynamically
	mo := component.Modal(component.ModalParams{
		Size:  component.SizeLg,
		Title: "Create backup task", // The form itself is the "create" form
		Content: []nodx.Node{
			nodx.Div(
				// This div is the content area that will be populated by HTMX
				htmx.HxGet("/dashboard/backups/duplicate-form/"+backupID.String()),
				htmx.HxSwap("innerHTML"),
				htmx.HxTrigger("load"), // Load content when this div is added to the DOM (modal opened)
				nodx.Class("p-10 flex justify-center"),
				component.HxLoadingMd(), // Show a loading spinner
			),
		},
	})

	dropdownButton := component.OptionsDropdownButton(
		mo.OpenerAttr, // Attributes to open the modal
		// No HxConfirm needed, modal will show the form
		lucide.CopyPlus(),
		component.SpanText("Duplicate backup task"),
	)

	// Return both the modal structure (initially hidden) and the button that opens it
	return nodx.Group(mo.HTML, dropdownButton)
}

func (h *handlers) duplicateBackupHandler(c echo.Context) error {
	// This handler is obsolete and its route should have been removed.
	// Returning NotImplemented as a safety measure if somehow called.
	return c.String(http.StatusNotImplemented, "This functionality has been updated. Please use the new duplicate button behavior.")
}
