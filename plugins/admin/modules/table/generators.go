package table

import (
	"database/sql"
	"errors"
	"fmt"
	tmpl "html/template"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/ui"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/action"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"github.com/GoAdminGroup/html"
	"golang.org/x/crypto/bcrypt"
)

type SystemTable struct {
	conn db.Connection
	c    *config.Config
}

func NewSystemTable(conn db.Connection, c *config.Config) *SystemTable {
	return &SystemTable{conn: conn, c: c}
}

func (s *SystemTable) GetNormalManagerTable(ctx *context.Context) (managerTable Table) {
	managerTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := managerTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField(lg("Name"), "username", db.Varchar).FieldFilterable()
	info.AddField(lg("Nickname"), "name", db.Varchar).FieldFilterable()
	info.AddField(lg("role"), "name", db.Varchar).
		FieldJoin(types.Join{
			Table:     "goadmin_role_users",
			JoinField: "user_id",
			Field:     "id",
		}).
		FieldJoin(types.Join{
			Table:     "goadmin_roles",
			JoinField: "id",
			Field:     "role_id",
			BaseTable: "goadmin_role_users",
		}).
		FieldDisplay(func(model types.FieldModel) interface{} {
			labels := template.HTML("")
			labelTpl := label().SetType("success")

			labelValues := strings.Split(model.Value, types.JoinFieldValueDelimiter)
			for key, label := range labelValues {
				if key == len(labelValues)-1 {
					labels += labelTpl.SetContent(template.HTML(label)).GetContent()
				} else {
					labels += labelTpl.SetContent(template.HTML(label)).GetContent() + "<br><br>"
				}
			}

			if labels == template.HTML("") {
				return lg("no roles")
			}

			return labels
		})
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_users").
		SetTitle(lg("Managers")).
		SetDescription(lg("Managers")).
		SetDeleteFn(func(idArr []string) error {

			var ids = interfaces(idArr)

			_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

				deleteUserRoleErr := s.connection().WithTx(tx).
					Table("goadmin_role_users").
					WhereIn("user_id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}

				deleteUserPermissionErr := s.connection().WithTx(tx).
					Table("goadmin_user_permissions").
					WhereIn("user_id", ids).
					Delete()

				if db.CheckError(deleteUserPermissionErr, db.DELETE) {
					return deleteUserPermissionErr, nil
				}

				deleteUserErr := s.connection().WithTx(tx).
					Table("goadmin_users").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteUserErr, db.DELETE) {
					return deleteUserErr, nil
				}

				return nil, nil
			})

			return txErr
		})

	formList := managerTable.GetForm().AddXssJsFilter()

	formList.AddField("ID", "id", db.Int, form.Default).FieldDisplayButCanNotEditWhenUpdate().FieldDisableWhenCreate()
	formList.AddField(lg("Name"), "username", db.Varchar, form.Text).FieldHelpMsg(template.HTML(lg("use for login"))).FieldMust()
	formList.AddField(lg("Nickname"), "name", db.Varchar, form.Text).FieldHelpMsg(template.HTML(lg("use to display"))).FieldMust()
	formList.AddField(lg("Avatar"), "avatar", db.Varchar, form.File)
	formList.AddField(lg("password"), "password", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		})
	formList.AddField(lg("confirm password"), "password_again", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		})

	formList.SetTable("goadmin_users").SetTitle(lg("Managers")).SetDescription(lg("Managers"))
	formList.SetUpdateFn(func(values form2.Values) error {

		/*	if values.IsEmpty("name", "username") {
				return errors.New("username and password can not be empty")
			}

			user := models.UserWithId(values.Get("id")).SetConn(s.conn)

			if values.Has("permission", "role") {
				return errors.New(errs.NoPermission)
			}

			password := values.Get("password")

			if password != "" {

				if password != values.Get("password_again") {
					return errors.New("password does not match")
				}

				password = encodePassword([]byte(values.Get("password")))
			}

			avatar := values.Get("avatar")

			if values.Get("avatar__delete_flag") == "1" {
				avatar = ""
			}

			_, updateUserErr := user.Update(values.Get("username"),
				password, values.Get("name"), avatar, values.Get("avatar__change_flag") == "1")

			if db.CheckError(updateUserErr, db.UPDATE) {
				return updateUserErr
			} */

		return nil
	})
	formList.SetInsertFn(func(values form2.Values) error {
		/*	if values.IsEmpty("name", "username", "password") {
				return errors.New("username and password can not be empty")
			}

			password := values.Get("password")

			if password != values.Get("password_again") {
				return errors.New("password does not match")
			}

			if values.Has("permission", "role") {
				return errors.New(errs.NoPermission)
			}

			_, createUserErr := models.User().SetConn(s.conn).New(values.Get("username"),
				encodePassword([]byte(values.Get("password"))),
				values.Get("name"),
				values.Get("avatar"))

			if db.CheckError(createUserErr, db.INSERT) {
				return createUserErr
			} */

		return nil
	})

	return
}

func (s *SystemTable) GetOpTable(ctx *context.Context) (opTable Table) {
	opTable = NewDefaultTable(Config{
		Driver:     config.GetDatabases().GetDefault().Driver,
		CanAdd:     false,
		Editable:   false,
		Deletable:  config.GetAllowDelOperationLog(),
		Exportable: true,
		Connection: "default",
		PrimaryKey: PrimaryKey{
			Type: db.Int,
			Name: DefaultPrimaryKeyName,
		},
	})

	info := opTable.GetInfo().AddXssJsFilter().
		HideFilterArea().HideDetailButton().HideEditButton().HideNewButton()

	if !config.GetAllowDelOperationLog() {
		info = info.HideDeleteButton()
	}

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField("userID", "user_id", db.Int).FieldHide()
	info.AddField(lg("user"), "name", db.Varchar).FieldJoin(types.Join{
		Table:     config.GetAuthUserTable(),
		JoinField: "id",
		Field:     "user_id",
	}).FieldDisplay(func(value types.FieldModel) interface{} {
		return template.Default().
			Link().
			SetURL(config.Url("/info/manager/detail?__goadmin_detail_pk=") + strconv.Itoa(int(value.Row["user_id"].(int64)))).
			SetContent(template.HTML(value.Value)).
			OpenInNewTab().
			SetTabTitle("Manager Detail").
			GetContent()
	}).FieldFilterable()
	info.AddField(lg("path"), "path", db.Varchar).FieldFilterable()
	info.AddField(lg("method"), "method", db.Varchar).FieldFilterable()
	info.AddField(lg("ip"), "ip", db.Varchar).FieldFilterable()
	info.AddField(lg("content"), "input", db.Text).FieldWidth(230)
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)

	users, _ := s.table(config.GetAuthUserTable()).Select("id", "name").All()
	options := make(types.FieldOptions, len(users))
	for k, user := range users {
		options[k].Value = fmt.Sprintf("%v", user["id"])
		options[k].Text = fmt.Sprintf("%v", user["name"])
	}
	info.AddSelectBox(language.Get("user"), options, action.FieldFilter("user_id"))
	info.AddSelectBox(language.Get("method"), types.FieldOptions{
		{Value: "GET", Text: "GET"},
		{Value: "POST", Text: "POST"},
		{Value: "OPTIONS", Text: "OPTIONS"},
		{Value: "PUT", Text: "PUT"},
		{Value: "HEAD", Text: "HEAD"},
		{Value: "DELETE", Text: "DELETE"},
	}, action.FieldFilter("method"))

	info.SetTable("goadmin_operation_log").
		SetTitle(lg("operation log")).
		SetDescription(lg("operation log"))

	formList := opTable.GetForm().AddXssJsFilter()

	formList.AddField("ID", "id", db.Int, form.Default).FieldDisplayButCanNotEditWhenUpdate().FieldDisableWhenCreate()
	formList.AddField(lg("userID"), "user_id", db.Int, form.Text)
	formList.AddField(lg("path"), "path", db.Varchar, form.Text)
	formList.AddField(lg("method"), "method", db.Varchar, form.Text)
	formList.AddField(lg("ip"), "ip", db.Varchar, form.Text)
	formList.AddField(lg("content"), "input", db.Varchar, form.Text)
	formList.AddField(lg("updatedAt"), "updated_at", db.Timestamp, form.Default).FieldDisableWhenCreate()
	formList.AddField(lg("createdAt"), "created_at", db.Timestamp, form.Default).FieldDisableWhenCreate()

	formList.SetTable("goadmin_operation_log").
		SetTitle(lg("operation log")).
		SetDescription(lg("operation log"))

	return
}

func (s *SystemTable) GetSiteTable(ctx *context.Context) (siteTable Table) {
	siteTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver).
		SetOnlyUpdateForm().
		SetGetDataFun(func(params parameter.Parameters) (i []map[string]interface{}, i2 int) {
			return []map[string]interface{}{models.Site().SetConn(s.conn).AllToMapInterface()}, 1
		}))

	trueStr := lgWithConfigScore("true")
	falseStr := lgWithConfigScore("false")

	formList := siteTable.GetForm().AddXssJsFilter()
	formList.AddField("ID", "id", db.Varchar, form.Default).FieldDefault("1").FieldHide()
	/*formList.AddField(lgWithConfigScore("site off"), "site_off", db.Varchar, form.Switch).
	FieldOptions(types.FieldOptions{
		{Text: trueStr, Value: "true"},
		{Text: falseStr, Value: "false"},
	})*/
	/*	formList.AddField(lgWithConfigScore("debug"), "debug", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		}) */
	/*formList.AddField(lgWithConfigScore("env"), "env", db.Varchar, form.Default).
	FieldDisplay(func(value types.FieldModel) interface{} {
		return s.c.Env
	})*/

	langOps := make(types.FieldOptions, len(language.Langs))
	for k, t := range language.Langs {
		langOps[k] = types.FieldOption{Text: lgWithConfigScore(t, "language"), Value: t}
	}
	formList.AddField(lgWithConfigScore("language"), "language", db.Varchar, form.SelectSingle).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return language.FixedLanguageKey(value.Value)
		}).
		FieldOptions(langOps)
	/*themes := template.Themes()
	themesOps := make(types.FieldOptions, len(themes))
	for k, t := range themes {
		themesOps[k] = types.FieldOption{Text: t, Value: t}
	}

	formList.AddField(lgWithConfigScore("theme"), "theme", db.Varchar, form.SelectSingle).
	FieldOptions(themesOps).
	FieldOnChooseShow("adminlte",
		"color_scheme") */
	formList.AddField(lgWithConfigScore("title"), "title", db.Varchar, form.Text).FieldMust()
	formList.AddField(lgWithConfigScore("color scheme"), "color_scheme", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "skin-black", Value: "skin-black"},
			{Text: "skin-black-light", Value: "skin-black-light"},
			{Text: "skin-blue", Value: "skin-blue"},
			{Text: "skin-blue-light", Value: "skin-blue-light"},
			{Text: "skin-green", Value: "skin-green"},
			{Text: "skin-green-light", Value: "skin-green-light"},
			{Text: "skin-purple", Value: "skin-purple"},
			{Text: "skin-purple-light", Value: "skin-purple-light"},
			{Text: "skin-red", Value: "skin-red"},
			{Text: "skin-red-light", Value: "skin-red-light"},
			{Text: "skin-yellow", Value: "skin-yellow"},
			{Text: "skin-yellow-light", Value: "skin-yellow-light"},
		})
		/*.FieldHelpMsg(template.HTML(lgWithConfigScore("It will work when theme is adminlte")))*/
	formList.AddField(lgWithConfigScore("login title"), "login_title", db.Varchar, form.Text).FieldMust()
	//formList.AddField(lgWithConfigScore("extra"), "extra", db.Varchar, form.TextArea)
	//formList.AddField(lgWithConfigScore("logo"), "logo", db.Varchar, form.Code).FieldMust()
	//formList.AddField(lgWithConfigScore("mini logo"), "mini_logo", db.Varchar, form.Code).FieldMust()
	/*if s.c.IsNotProductionEnvironment() {
		formList.AddField(lgWithConfigScore("bootstrap file path"), "bootstrap_file_path", db.Varchar, form.Text)
		formList.AddField(lgWithConfigScore("go mod file path"), "go_mod_file_path", db.Varchar, form.Text)
	}*/
	formList.AddField(lgWithConfigScore("session life time"), "session_life_time", db.Varchar, form.Number).
		FieldMust().
		FieldHelpMsg(template.HTML(lgWithConfigScore("must bigger than 900 seconds")))
	//formList.AddField(lgWithConfigScore("custom head html"), "custom_head_html", db.Varchar, form.Code)
	//formList.AddField(lgWithConfigScore("custom foot Html"), "custom_foot_html", db.Varchar, form.Code)
	//formList.AddField(lgWithConfigScore("custom 404 html"), "custom_404_html", db.Varchar, form.Code)
	//formList.AddField(lgWithConfigScore("custom 403 html"), "custom_403_html", db.Varchar, form.Code)
	//formList.AddField(lgWithConfigScore("custom 500 Html"), "custom_500_html", db.Varchar, form.Code)
	//formList.AddField(lgWithConfigScore("footer info"), "footer_info", db.Varchar, form.Code)
	//formList.AddField(lgWithConfigScore("login logo"), "login_logo", db.Varchar, form.Code)
	/*formList.AddField(lgWithConfigScore("no limit login ip"), "no_limit_login_ip", db.Varchar, form.Switch).
	FieldOptions(types.FieldOptions{
		{Text: trueStr, Value: "true"},
		{Text: falseStr, Value: "false"},
	})*/
	formList.AddField(lgWithConfigScore("operation log off"), "operation_log_off", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("allow delete operation log"), "allow_del_operation_log", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("hide config center entrance"), "hide_config_center_entrance", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("hide app info entrance"), "hide_app_info_entrance", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	/*formList.AddField(lgWithConfigScore("hide tool entrance"), "hide_tool_entrance", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("hide plugin entrance"), "hide_plugin_entrance", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})*/
	formList.AddField(lgWithConfigScore("animation type"), "animation_type", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "", Value: ""},
			{Text: "bounce", Value: "bounce"}, {Text: "flash", Value: "flash"}, {Text: "pulse", Value: "pulse"},
			{Text: "rubberBand", Value: "rubberBand"}, {Text: "shake", Value: "shake"}, {Text: "swing", Value: "swing"},
			{Text: "tada", Value: "tada"}, {Text: "wobble", Value: "wobble"}, {Text: "jello", Value: "jello"},
			{Text: "heartBeat", Value: "heartBeat"}, {Text: "bounceIn", Value: "bounceIn"}, {Text: "bounceInDown", Value: "bounceInDown"},
			{Text: "bounceInLeft", Value: "bounceInLeft"}, {Text: "bounceInRight", Value: "bounceInRight"}, {Text: "bounceInUp", Value: "bounceInUp"},
			{Text: "fadeIn", Value: "fadeIn"}, {Text: "fadeInDown", Value: "fadeInDown"}, {Text: "fadeInDownBig", Value: "fadeInDownBig"},
			{Text: "fadeInLeft", Value: "fadeInLeft"}, {Text: "fadeInLeftBig", Value: "fadeInLeftBig"}, {Text: "fadeInRight", Value: "fadeInRight"},
			{Text: "fadeInRightBig", Value: "fadeInRightBig"}, {Text: "fadeInUp", Value: "fadeInUp"}, {Text: "fadeInUpBig", Value: "fadeInUpBig"},
			{Text: "flip", Value: "flip"}, {Text: "flipInX", Value: "flipInX"}, {Text: "flipInY", Value: "flipInY"},
			{Text: "lightSpeedIn", Value: "lightSpeedIn"}, {Text: "rotateIn", Value: "rotateIn"}, {Text: "rotateInDownLeft", Value: "rotateInDownLeft"},
			{Text: "rotateInDownRight", Value: "rotateInDownRight"}, {Text: "rotateInUpLeft", Value: "rotateInUpLeft"}, {Text: "rotateInUpRight", Value: "rotateInUpRight"},
			{Text: "slideInUp", Value: "slideInUp"}, {Text: "slideInDown", Value: "slideInDown"}, {Text: "slideInLeft", Value: "slideInLeft"}, {Text: "slideInRight", Value: "slideInRight"},
			{Text: "slideOutRight", Value: "slideOutRight"}, {Text: "zoomIn", Value: "zoomIn"}, {Text: "zoomInDown", Value: "zoomInDown"},
			{Text: "zoomInLeft", Value: "zoomInLeft"}, {Text: "zoomInRight", Value: "zoomInRight"}, {Text: "zoomInUp", Value: "zoomInUp"},
			{Text: "hinge", Value: "hinge"}, {Text: "jackInTheBox", Value: "jackInTheBox"}, {Text: "rollIn", Value: "rollIn"},
		}).FieldOnChooseHide("", "animation_duration", "animation_delay").
		FieldOptionExt(map[string]interface{}{"allowClear": true}).
		FieldHelpMsg(`see more: <a href="https://daneden.github.io/animate.css/">https://daneden.github.io/animate.css/</a>`)

	formList.AddField(lgWithConfigScore("animation duration"), "animation_duration", db.Varchar, form.Number)
	formList.AddField(lgWithConfigScore("animation delay"), "animation_delay", db.Varchar, form.Number)

	//formList.AddField(lgWithConfigScore("file upload engine"), "file_upload_engine", db.Varchar, form.Text)

	/*	formList.AddField(lgWithConfigScore("cdn url"), "asset_url", db.Varchar, form.Text).
			FieldHelpMsg(template.HTML(lgWithConfigScore("Do not modify when you have not set up all assets")))

		formList.AddField(lgWithConfigScore("info log path"), "info_log_path", db.Varchar, form.Text)
		formList.AddField(lgWithConfigScore("error log path"), "error_log_path", db.Varchar, form.Text)
		formList.AddField(lgWithConfigScore("access log path"), "access_log_path", db.Varchar, form.Text)
		formList.AddField(lgWithConfigScore("info log off"), "info_log_off", db.Varchar, form.Switch).
			FieldOptions(types.FieldOptions{
				{Text: trueStr, Value: "true"},
				{Text: falseStr, Value: "false"},
			})
		formList.AddField(lgWithConfigScore("error log off"), "error_log_off", db.Varchar, form.Switch).
			FieldOptions(types.FieldOptions{
				{Text: trueStr, Value: "true"},
				{Text: falseStr, Value: "false"},
			})
		formList.AddField(lgWithConfigScore("access log off"), "access_log_off", db.Varchar, form.Switch).
			FieldOptions(types.FieldOptions{
				{Text: trueStr, Value: "true"},
				{Text: falseStr, Value: "false"},
			})
		formList.AddField(lgWithConfigScore("access assets log off"), "access_assets_log_off", db.Varchar, form.Switch).
			FieldOptions(types.FieldOptions{
				{Text: trueStr, Value: "true"},
				{Text: falseStr, Value: "false"},
			})
		formList.AddField(lgWithConfigScore("sql log on"), "sql_log", db.Varchar, form.Switch).
			FieldOptions(types.FieldOptions{
				{Text: trueStr, Value: "true"},
				{Text: falseStr, Value: "false"},
			})
		formList.AddField(lgWithConfigScore("log level"), "logger_level", db.Varchar, form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: "Debug", Value: "-1"},
				{Text: "Info", Value: "0"},
				{Text: "Warn", Value: "1"},
				{Text: "Error", Value: "2"},
			}).FieldDisplay(defaultFilterFn("0"))

		formList.AddField(lgWithConfigScore("logger rotate max size"), "logger_rotate_max_size", db.Varchar, form.Number).
			FieldDivider(lgWithConfigScore("logger rotate")).FieldDisplay(defaultFilterFn("10", "0"))
		formList.AddField(lgWithConfigScore("logger rotate max backups"), "logger_rotate_max_backups", db.Varchar, form.Number).
			FieldDisplay(defaultFilterFn("5", "0"))
		formList.AddField(lgWithConfigScore("logger rotate max age"), "logger_rotate_max_age", db.Varchar, form.Number).
			FieldDisplay(defaultFilterFn("30", "0"))
		formList.AddField(lgWithConfigScore("logger rotate compress"), "logger_rotate_compress", db.Varchar, form.Switch).
			FieldOptions(types.FieldOptions{
				{Text: trueStr, Value: "true"},
				{Text: falseStr, Value: "false"},
			}).FieldDisplay(defaultFilterFn("false"))

		formList.AddField(lgWithConfigScore("logger rotate encoder encoding"), "logger_encoder_encoding", db.Varchar,
			form.SelectSingle).
			FieldDivider(lgWithConfigScore("logger rotate encoder")).
			FieldOptions(types.FieldOptions{
				{Text: "JSON", Value: "json"},
				{Text: "Console", Value: "console"},
			}).FieldDisplay(defaultFilterFn("console")).
			FieldOnChooseHide("Console",
				"logger_encoder_time_key", "logger_encoder_level_key", "logger_encoder_caller_key",
				"logger_encoder_message_key", "logger_encoder_stacktrace_key", "logger_encoder_name_key")

		formList.AddField(lgWithConfigScore("logger rotate encoder time key"), "logger_encoder_time_key", db.Varchar, form.Text).
			FieldDisplay(defaultFilterFn("ts"))
		formList.AddField(lgWithConfigScore("logger rotate encoder level key"), "logger_encoder_level_key", db.Varchar, form.Text).
			FieldDisplay(defaultFilterFn("level"))
		formList.AddField(lgWithConfigScore("logger rotate encoder name key"), "logger_encoder_name_key", db.Varchar, form.Text).
			FieldDisplay(defaultFilterFn("logger"))
		formList.AddField(lgWithConfigScore("logger rotate encoder caller key"), "logger_encoder_caller_key", db.Varchar, form.Text).
			FieldDisplay(defaultFilterFn("caller"))
		formList.AddField(lgWithConfigScore("logger rotate encoder message key"), "logger_encoder_message_key", db.Varchar, form.Text).
			FieldDisplay(defaultFilterFn("msg"))
		formList.AddField(lgWithConfigScore("logger rotate encoder stacktrace key"), "logger_encoder_stacktrace_key", db.Varchar, form.Text).
			FieldDisplay(defaultFilterFn("stacktrace"))

		formList.AddField(lgWithConfigScore("logger rotate encoder level"), "logger_encoder_level", db.Varchar,
			form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: lgWithConfigScore("capital"), Value: "capital"},
				{Text: lgWithConfigScore("capitalcolor"), Value: "capitalColor"},
				{Text: lgWithConfigScore("lowercase"), Value: "lowercase"},
				{Text: lgWithConfigScore("lowercasecolor"), Value: "color"},
			}).FieldDisplay(defaultFilterFn("capitalColor"))
		formList.AddField(lgWithConfigScore("logger rotate encoder time"), "logger_encoder_time", db.Varchar,
			form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: "ISO8601(2006-01-02T15:04:05.000Z0700)", Value: "iso8601"},
				{Text: lgWithConfigScore("millisecond"), Value: "millis"},
				{Text: lgWithConfigScore("nanosecond"), Value: "nanos"},
				{Text: "RFC3339(2006-01-02T15:04:05Z07:00)", Value: "rfc3339"},
				{Text: "RFC3339 Nano(2006-01-02T15:04:05.999999999Z07:00)", Value: "rfc3339nano"},
			}).FieldDisplay(defaultFilterFn("iso8601"))
		formList.AddField(lgWithConfigScore("logger rotate encoder duration"), "logger_encoder_duration", db.Varchar,
			form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: lgWithConfigScore("seconds"), Value: "string"},
				{Text: lgWithConfigScore("nanosecond"), Value: "nanos"},
				{Text: lgWithConfigScore("microsecond"), Value: "ms"},
			}).FieldDisplay(defaultFilterFn("string"))
		formList.AddField(lgWithConfigScore("logger rotate encoder caller"), "logger_encoder_caller", db.Varchar,
			form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: lgWithConfigScore("full path"), Value: "full"},
				{Text: lgWithConfigScore("short path"), Value: "short"},
			}).FieldDisplay(defaultFilterFn("full"))
	*/
	formList.HideBackButton().HideContinueEditCheckBox().HideContinueNewCheckBox()
	formList.SetTabGroups(types.NewTabGroups("id", "language",
		"title", "login_title", "session_life_time",
		"operation_log_off", "allow_del_operation_log", "hide_config_center_entrance", "hide_app_info_entrance").
		AddGroup("color_scheme", "animation_type", "animation_duration", "animation_delay")).
		/*AddGroup("access_log_off", "access_assets_log_off", "info_log_off", "error_log_off", "sql_log", "logger_level",
		"info_log_path", "error_log_path",
		"access_log_path", "logger_rotate_max_size", "logger_rotate_max_backups",
		"logger_rotate_max_age", "logger_rotate_compress",
		"logger_encoder_encoding", "logger_encoder_time_key", "logger_encoder_level_key", "logger_encoder_name_key",
		"logger_encoder_caller_key", "logger_encoder_message_key", "logger_encoder_stacktrace_key", "logger_encoder_level",
		"logger_encoder_time", "logger_encoder_duration", "logger_encoder_caller")).*/
		/*AddGroup( "logo", "mini_logo","custom_head_html", "custom_foot_html", "footer_info", "login_logo")).*/
		/*,"custom_404_html", "custom_403_html", "custom_500_html")*/
		SetTabHeaders(lgWithConfigScore("general"), lgWithConfigScore("display") /*lgWithConfigScore("log"), lgWithConfigScore("custom")*/)

	formList.SetTable("goadmin_site").
		SetTitle(lgWithConfigScore("site setting")).
		SetDescription(lgWithConfigScore("site setting"))

	formList.SetUpdateFn(func(values form2.Values) error {

		ses := values.Get("session_life_time")
		sesInt, _ := strconv.Atoi(ses)
		if sesInt < 900 {
			return errors.New("wrong session life time, must bigger than 900 seconds")
		}
		if err := checkJSON(values, "file_upload_engine"); err != nil {
			return err
		}

		//values["logo"][0] = escape(values.Get("logo"))
		//values["mini_logo"][0] = escape(values.Get("mini_logo"))
		//values["custom_head_html"][0] = escape(values.Get("custom_head_html"))
		//values["custom_foot_html"][0] = escape(values.Get("custom_foot_html"))
		//values["custom_404_html"][0] = escape(values.Get("custom_404_html"))
		//values["custom_403_html"][0] = escape(values.Get("custom_403_html"))
		//values["custom_500_html"][0] = escape(values.Get("custom_500_html"))
		//values["footer_info"][0] = escape(values.Get("footer_info"))
		//values["login_logo"][0] = escape(values.Get("login_logo"))

		var err error
		if s.c.UpdateProcessFn != nil {
			values, err = s.c.UpdateProcessFn(values)
			if err != nil {
				return err
			}
		}

		ui.GetService(services).RemoveOrShowSiteNavButton(values["hide_config_center_entrance"][0] == "true")
		ui.GetService(services).RemoveOrShowInfoNavButton(values["hide_app_info_entrance"][0] == "true")
		//ui.GetService(services).RemoveOrShowToolNavButton(values["hide_tool_entrance"][0] == "true")
		//ui.GetService(services).RemoveOrShowPlugNavButton(values["hide_plugin_entrance"][0] == "true")

		//always hide plugin and tool entrance
		ui.GetService(services).RemoveOrShowToolNavButton(true)
		ui.GetService(services).RemoveOrShowPlugNavButton(true)

		// TODO: add transaction
		err = models.Site().SetConn(s.conn).Update(values.RemoveSysRemark())
		if err != nil {
			return err
		}
		return s.c.Update(values.ToMap())
	})

	formList.EnableAjax(lgWithConfigScore("modify site config"),
		lgWithConfigScore("modify site config"),
		"",
		lgWithConfigScore("modify site config success"),
		lgWithConfigScore("modify site config fail"))

	return
}

// -------------------------
// helper functions
// -------------------------

func encodePassword(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func label() types.LabelAttribute {
	return template.Get(config.GetTheme()).Label().SetType("success")
}

func lg(v string) string {
	return language.Get(v)
}

func defaultFilterFn(val string, def ...string) types.FieldFilterFn {
	return func(value types.FieldModel) interface{} {
		if len(def) > 0 {
			if value.Value == def[0] {
				return val
			}
		} else {
			if value.Value == "" {
				return val
			}
		}
		return value.Value
	}
}

func lgWithScore(v string, score ...string) string {
	return language.GetWithScope(v, score...)
}

func lgWithConfigScore(v string, score ...string) string {
	scores := append([]string{"config"}, score...)
	return language.GetWithScope(v, scores...)
}

func link(url, content string) tmpl.HTML {
	return html.AEl().
		SetAttr("href", url).
		SetContent(template.HTML(lg(content))).
		Get()
}

func escape(s string) string {
	if s == "" {
		return ""
	}
	s, err := url.QueryUnescape(s)
	if err != nil {
		logger.Error("escape error", err)
	}
	return s
}

func checkJSON(values form2.Values, key string) error {
	v := values.Get(key)
	if v != "" && !utils.IsJSON(v) {
		return errors.New("wrong " + key)
	}
	return nil
}

func (s *SystemTable) table(table string) *db.SQL {
	return s.connection().Table(table)
}

func (s *SystemTable) connection() *db.SQL {
	return db.WithDriver(s.conn)
}

func interfaces(arr []string) []interface{} {
	var iarr = make([]interface{}, len(arr))

	for key, v := range arr {
		iarr[key] = v
	}

	return iarr
}

func addSwitchForTool(formList *types.FormPanel, head, field, def string, row ...int) {
	formList.AddField(lgWithScore(head, "tool"), field, db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: lgWithScore("show", "tool"), Value: "n"},
			{Text: lgWithScore("hide", "tool"), Value: "y"},
		}).FieldDefault(def)
	if len(row) > 0 {
		formList.FieldRowWidth(row[0])
	}
	if len(row) > 1 {
		formList.FieldHeadWidth(row[1])
	}
	if len(row) > 2 {
		formList.FieldInputWidth(row[2])
	}
}

func formTypeOptions() types.FieldOptions {
	opts := make(types.FieldOptions, len(form.AllType))
	for i := 0; i < len(form.AllType); i++ {
		v := form.AllType[i].Name()
		opts[i] = types.FieldOption{Text: v, Value: v}
	}
	return opts
}

func databaseTypeOptions() types.FieldOptions {
	opts := make(types.FieldOptions, len(db.IntTypeList)+
		len(db.StringTypeList)+
		len(db.FloatTypeList)+
		len(db.UintTypeList)+
		len(db.BoolTypeList))
	z := 0
	for _, t := range db.IntTypeList {
		text := string(t)
		v := strings.Title(strings.ToLower(text))
		opts[z] = types.FieldOption{Text: text, Value: v}
		z++
	}
	for _, t := range db.StringTypeList {
		text := string(t)
		v := strings.Title(strings.ToLower(text))
		opts[z] = types.FieldOption{Text: text, Value: v}
		z++
	}
	for _, t := range db.FloatTypeList {
		text := string(t)
		v := strings.Title(strings.ToLower(text))
		opts[z] = types.FieldOption{Text: text, Value: v}
		z++
	}
	for _, t := range db.UintTypeList {
		text := string(t)
		v := strings.Title(strings.ToLower(text))
		opts[z] = types.FieldOption{Text: text, Value: v}
		z++
	}
	for _, t := range db.BoolTypeList {
		text := string(t)
		v := strings.Title(strings.ToLower(text))
		opts[z] = types.FieldOption{Text: text, Value: v}
		z++
	}
	return opts
}

func getType(typeName string) string {
	r, _ := regexp.Compile(`\(.*?\)`)
	typeName = r.ReplaceAllString(typeName, "")
	r2, _ := regexp.Compile(`unsigned(.*)`)
	return strings.TrimSpace(strings.Title(strings.ToLower(r2.ReplaceAllString(typeName, ""))))
}
