package backups

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eduardolat/pgbackweb/internal/database/dbgen"
	"github.com/eduardolat/pgbackweb/internal/staticdata"
	"github.com/eduardolat/pgbackweb/internal/util/echoutil"
	"github.com/eduardolat/pgbackweb/internal/validate"
	"github.com/eduardolat/pgbackweb/internal/view/web/component"
	"github.com/eduardolat/pgbackweb/internal/view/web/respondhtmx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	nodx "github.com/nodxdev/nodxgo"
	alpine "github.com/nodxdev/nodxgo-alpine"
	htmx "github.com/nodxdev/nodxgo-htmx"
	lucide "github.com/nodxdev/nodxgo-lucide"
)

type CreateBackupFormValues struct {
	DatabaseID     string // UUID as string
	DestinationID  string // UUID as string, empty if not set
	IsLocal        string // "true" or "false"
	Name           string
	CronExpression string
	TimeZone       string
	IsActive       string // "true" or "false"
	DestDir        string
	RetentionDays  string // int16 as string
	OptDataOnly    string // "true" or "false"
	OptSchemaOnly  string // "true" or "false"
	OptClean       string // "true" or "false"
	OptIfExists    string // "true" or "false"
	OptCreate      string // "true" or "false"
	OptNoComments  string // "true" or "false"
}

func (h *handlers) createBackupHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var formData struct {
		DatabaseID     uuid.UUID `form:"database_id" validate:"required,uuid"`
		DestinationID  uuid.UUID `form:"destination_id" validate:"omitempty,uuid"`
		IsLocal        string    `form:"is_local" validate:"required,oneof=true false"`
		Name           string    `form:"name" validate:"required"`
		CronExpression string    `form:"cron_expression" validate:"required"`
		TimeZone       string    `form:"time_zone" validate:"required"`
		IsActive       string    `form:"is_active" validate:"required,oneof=true false"`
		DestDir        string    `form:"dest_dir" validate:"required"`
		RetentionDays  int16     `form:"retention_days"`
		OptDataOnly    string    `form:"opt_data_only" validate:"required,oneof=true false"`
		OptSchemaOnly  string    `form:"opt_schema_only" validate:"required,oneof=true false"`
		OptClean       string    `form:"opt_clean" validate:"required,oneof=true false"`
		OptIfExists    string    `form:"opt_if_exists" validate:"required,oneof=true false"`
		OptCreate      string    `form:"opt_create" validate:"required,oneof=true false"`
		OptNoComments  string    `form:"opt_no_comments" validate:"required,oneof=true false"`
	}
	if err := c.Bind(&formData); err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}
	if err := validate.Struct(&formData); err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	_, err := h.servs.BackupsService.CreateBackup(
		ctx, dbgen.BackupsServiceCreateBackupParams{
			DatabaseID: formData.DatabaseID,
			DestinationID: uuid.NullUUID{
				Valid: formData.IsLocal == "false", UUID: formData.DestinationID,
			},
			IsLocal:        formData.IsLocal == "true",
			Name:           formData.Name,
			CronExpression: formData.CronExpression,
			TimeZone:       formData.TimeZone,
			IsActive:       formData.IsActive == "true",
			DestDir:        formData.DestDir,
			RetentionDays:  formData.RetentionDays,
			OptDataOnly:    formData.OptDataOnly == "true",
			OptSchemaOnly:  formData.OptSchemaOnly == "true",
			OptClean:       formData.OptClean == "true",
			OptIfExists:    formData.OptIfExists == "true",
			OptCreate:      formData.OptCreate == "true",
			OptNoComments:  formData.OptNoComments == "true",
		},
	)
	if err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	return respondhtmx.Redirect(c, "/dashboard/backups")
}

func (h *handlers) createBackupFormHandler(c echo.Context) error {
	ctx := c.Request().Context()

	databases, err := h.servs.DatabasesService.GetAllDatabases(ctx)
	if err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	destinations, err := h.servs.DestinationsService.GetAllDestinations(ctx)
	if err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	return echoutil.RenderNodx(
		c, http.StatusOK, createBackupForm(databases, destinations, nil),
	)
}

func createBackupForm(
	databases []dbgen.DatabasesServiceGetAllDatabasesRow,
	destinations []dbgen.DestinationsServiceGetAllDestinationsRow,
	prefillData *CreateBackupFormValues,
) nodx.Node {
	yesNoOptions := func(currentValue string) nodx.Node {
		return nodx.Group(
			nodx.Option(nodx.Value("true"), nodx.Text("Yes"), nodx.If(currentValue == "true", nodx.Selected(""))),
			nodx.Option(nodx.Value("false"), nodx.Text("No"), nodx.If(currentValue == "false" || currentValue == "", nodx.Selected(""))),
		)
	}

	serverTZ := time.Now().Location().String()

	isLocalValue := "false"
	if prefillData != nil && prefillData.IsLocal != "" {
		isLocalValue = prefillData.IsLocal
	}

	nameValue := ""
	if prefillData != nil {
		nameValue = prefillData.Name
	}
	cronExpressionValue := ""
	if prefillData != nil {
		cronExpressionValue = prefillData.CronExpression
	}
	destDirValue := ""
	if prefillData != nil {
		destDirValue = prefillData.DestDir
	}
	retentionDaysValue := ""
	if prefillData != nil {
		retentionDaysValue = prefillData.RetentionDays
	}

	isActiveValue := "true" // Default for new forms
	if prefillData != nil {
		isActiveValue = "false" // Default for prefilled forms (duplicate task)
	}

	optDataOnlyValue := "false"
	if prefillData != nil && prefillData.OptDataOnly != "" {
		optDataOnlyValue = prefillData.OptDataOnly
	}
	optSchemaOnlyValue := "false"
	if prefillData != nil && prefillData.OptSchemaOnly != "" {
		optSchemaOnlyValue = prefillData.OptSchemaOnly
	}
	optCleanValue := "false"
	if prefillData != nil && prefillData.OptClean != "" {
		optCleanValue = prefillData.OptClean
	}
	optIfExistsValue := "false"
	if prefillData != nil && prefillData.OptIfExists != "" {
		optIfExistsValue = prefillData.OptIfExists
	}
	optCreateValue := "false"
	if prefillData != nil && prefillData.OptCreate != "" {
		optCreateValue = prefillData.OptCreate
	}
	optNoCommentsValue := "false"
	if prefillData != nil && prefillData.OptNoComments != "" {
		optNoCommentsValue = prefillData.OptNoComments
	}

	return nodx.FormEl(
		htmx.HxPost("/dashboard/backups"),
		htmx.HxDisabledELT("find button"),
		nodx.Class("space-y-2 text-base"),

		alpine.XData(fmt.Sprintf(`{ is_local: "%s" }`, isLocalValue)),

		component.InputControl(component.InputControlParams{
			Name:        "name",
			Label:       "Name",
			Placeholder: "My backup",
			Required:    true,
			Type:        component.InputTypeText,
			Children: []nodx.Node{
				nodx.If(nameValue != "", nodx.Value(nameValue)),
			},
		}),

		component.SelectControl(component.SelectControlParams{
			Name:        "database_id",
			Label:       "Database",
			Required:    true,
			Placeholder: "Select a database",
			Children: []nodx.Node{
				nodx.Map(
					databases,
					func(db dbgen.DatabasesServiceGetAllDatabasesRow) nodx.Node {
						selected := false
						if prefillData != nil && prefillData.DatabaseID != "" &&
							db.ID.String() == prefillData.DatabaseID {
							selected = true
						}
						return nodx.Option(nodx.Value(db.ID.String()), nodx.Text(db.Name), nodx.If(selected, nodx.Selected("")))
					},
				),
			},
		}),

		component.SelectControl(component.SelectControlParams{
			Name:     "is_local",
			Label:    "Local backup",
			Required: true,
			Children: []nodx.Node{
				alpine.XModel("is_local"),
				nodx.Option(nodx.Value("true"), nodx.Text("Yes"), nodx.If(isLocalValue == "true", nodx.Selected(""))),
				nodx.Option(nodx.Value("false"), nodx.Text("No"), nodx.If(isLocalValue == "false", nodx.Selected(""))),
			},
			HelpButtonChildren: localBackupsHelp(),
		}),

		alpine.Template(
			alpine.XIf("is_local == 'false'"),
			component.SelectControl(component.SelectControlParams{
				Name:        "destination_id",
				Label:       "Destination",
				Required:    true,
				Placeholder: "Select a destination",
				Children: []nodx.Node{
					nodx.Map(
						destinations,
						func(dest dbgen.DestinationsServiceGetAllDestinationsRow) nodx.Node {
							selected := false
							if prefillData != nil && prefillData.DestinationID != "" &&
								dest.ID.String() == prefillData.DestinationID {
								selected = true
							}
							return nodx.Option(nodx.Value(dest.ID.String()), nodx.Text(dest.Name), nodx.If(selected, nodx.Selected("")))
						},
					),
				},
			}),
		),

		component.InputControl(component.InputControlParams{
			Name:               "cron_expression",
			Label:              "Cron expression",
			Placeholder:        "* * * * *",
			Required:           true,
			Type:               component.InputTypeText,
			HelpText:           "The cron expression to schedule the backup",
			Pattern:            `^\S+\s+\S+\s+\S+\s+\S+\s+\S+$`,
			HelpButtonChildren: cronExpressionHelp(),
			Children: []nodx.Node{
				nodx.If(cronExpressionValue != "", nodx.Value(cronExpressionValue)),
			},
		}),

		component.SelectControl(component.SelectControlParams{
			Name:        "time_zone",
			Label:       "Time zone",
			Required:    true,
			Placeholder: "Select a time zone",
			Children: []nodx.Node{
				nodx.Map(
					staticdata.Timezones,
					func(tz staticdata.Timezone) nodx.Node {
						selected := false
						if prefillData != nil && prefillData.TimeZone != "" {
							if tz.TzCode == prefillData.TimeZone {
								selected = true
							}
						} else {
							if tz.TzCode == serverTZ {
								selected = true
							}
						}
						return nodx.Option(nodx.Value(tz.TzCode), nodx.Text(tz.Label), nodx.If(selected, nodx.Selected("")))
					},
				),
			},
			HelpButtonChildren: timezoneFilenamesHelp(),
		}),

		component.InputControl(component.InputControlParams{
			Name:               "dest_dir",
			Label:              "Destination directory",
			Placeholder:        "/path/to/backup",
			Required:           true,
			Type:               component.InputTypeText,
			HelpText:           "Relative to the base directory of the destination",
			Pattern:            `^\/\S*[^\/]$`,
			HelpButtonChildren: destinationDirectoryHelp(),
			Children: []nodx.Node{
				nodx.If(destDirValue != "", nodx.Value(destDirValue)),
			},
		}),

		component.InputControl(component.InputControlParams{
			Name:               "retention_days",
			Label:              "Retention days",
			Placeholder:        "30",
			Required:           true,
			Type:               component.InputTypeNumber,
			Pattern:            "[0-9]+",
			HelpButtonChildren: retentionDaysHelp(),
			Children: []nodx.Node{
				nodx.Min("0"),
				nodx.Max("36500"),
				nodx.If(retentionDaysValue != "", nodx.Value(retentionDaysValue)),
			},
		}),

		component.SelectControl(component.SelectControlParams{
			Name:     "is_active",
			Label:    "Activate backup",
			Required: true,
			Children: []nodx.Node{
				nodx.Option(nodx.Value("true"), nodx.Text("Yes"), nodx.If(isActiveValue == "true", nodx.Selected(""))),
				nodx.Option(nodx.Value("false"), nodx.Text("No"), nodx.If(isActiveValue == "false", nodx.Selected(""))),
			},
		}),

		nodx.Div(
			nodx.Class("pt-4"),
			nodx.Div(
				nodx.Class("flex justify-start items-center space-x-1"),
				component.H2Text("Options"),
				component.HelpButtonModal(component.HelpButtonModalParams{
					ModalTitle: "Backup options",
					Children:   pgDumpOptionsHelp(),
				}),
			),

			nodx.Div(
				nodx.Class("mt-2 grid grid-cols-2 gap-2"),

				component.SelectControl(component.SelectControlParams{
					Name:     "opt_data_only",
					Label:    "--data-only",
					Required: true,
					Children: []nodx.Node{
						yesNoOptions(optDataOnlyValue),
					},
				}),

				component.SelectControl(component.SelectControlParams{
					Name:     "opt_schema_only",
					Label:    "--schema-only",
					Required: true,
					Children: []nodx.Node{
						yesNoOptions(optSchemaOnlyValue),
					},
				}),

				component.SelectControl(component.SelectControlParams{
					Name:     "opt_clean",
					Label:    "--clean",
					Required: true,
					Children: []nodx.Node{
						yesNoOptions(optCleanValue),
					},
				}),

				component.SelectControl(component.SelectControlParams{
					Name:     "opt_if_exists",
					Label:    "--if-exists",
					Required: true,
					Children: []nodx.Node{
						yesNoOptions(optIfExistsValue),
					},
				}),

				component.SelectControl(component.SelectControlParams{
					Name:     "opt_create",
					Label:    "--create",
					Required: true,
					Children: []nodx.Node{
						yesNoOptions(optCreateValue),
					},
				}),

				component.SelectControl(component.SelectControlParams{
					Name:     "opt_no_comments",
					Label:    "--no-comments",
					Required: true,
					Children: []nodx.Node{
						yesNoOptions(optNoCommentsValue),
					},
				}),
			),
		),

		nodx.Div(
			nodx.Class("flex justify-end items-center space-x-2 pt-2"),
			component.HxLoadingMd(),
			nodx.Button(
				nodx.Class("btn btn-primary"),
				nodx.Type("submit"),
				component.SpanText("Create backup task"),
				lucide.Save(),
			),
		),
	)
}

func createBackupButton() nodx.Node {
	mo := component.Modal(component.ModalParams{
		Size:  component.SizeLg,
		Title: "Create backup task",
		Content: []nodx.Node{
			nodx.Div(
				htmx.HxGet("/dashboard/backups/create-form"),
				htmx.HxSwap("outerHTML"),
				htmx.HxTrigger("intersect once"),
				nodx.Class("p-10 flex justify-center"),
				component.HxLoadingMd(),
			),
		},
	})

	button := nodx.Button(
		mo.OpenerAttr,
		nodx.Class("btn btn-primary"),
		component.SpanText("Create backup task"),
		lucide.Plus(),
	)

	return nodx.Div(
		nodx.Class("inline-block"),
		mo.HTML,
		button,
	)
}
