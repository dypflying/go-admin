package table

import (
	"database/sql"
	"html/template"
	"strconv"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/collection"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

func (s *SystemTable) GetMenuTable(ctx *context.Context) (menuTable Table) {
	menuTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	name := ctx.Query("__plugin_name")

	info := menuTable.GetInfo().AddXssJsFilter().HideFilterArea().Where("plugin_name", "=", name)

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField(lg("parent"), "parent_id", db.Int)
	info.AddField(lg("menu name"), "title", db.Varchar)
	info.AddField(lg("icon"), "icon", db.Varchar)
	info.AddField(lg("uri"), "uri", db.Varchar)
	info.AddField(lg("role"), "roles", db.Varchar)
	info.AddField(lg("header"), "header", db.Varchar)
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_menu").
		SetTitle(lg("Menus Manage")).
		SetDescription(lg("Menus Manage")).
		SetDeleteFn(func(idArr []string) error {

			var ids = interfaces(idArr)
			_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {
				deleteRoleMenuErr := s.connection().WithTx(tx).
					Table("goadmin_role_menu").
					WhereIn("menu_id", ids).
					Delete()

				if db.CheckError(deleteRoleMenuErr, db.DELETE) {
					return deleteRoleMenuErr, nil
				}
				deleteMenusErr := s.connection().WithTx(tx).
					Table("goadmin_menu").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteMenusErr, db.DELETE) {
					return deleteMenusErr, nil
				}
				return nil, map[string]interface{}{}
			})

			return txErr
		})

	var parentIDOptions = types.FieldOptions{
		{
			Text:  "ROOT",
			Value: "0",
		},
	}

	allMenus, _ := s.connection().Table("goadmin_menu").
		Where("parent_id", "=", 0).
		Where("plugin_name", "=", name).
		Select("id", "title").
		OrderBy("order", "asc").
		All()
	allMenuIDs := make([]interface{}, len(allMenus))

	if len(allMenuIDs) > 0 {
		for i := 0; i < len(allMenus); i++ {
			allMenuIDs[i] = allMenus[i]["id"]
		}

		secondLevelMenus, _ := s.connection().Table("goadmin_menu").
			WhereIn("parent_id", allMenuIDs).
			Where("plugin_name", "=", name).
			Select("id", "title", "parent_id").
			All()

		secondLevelMenusCol := collection.Collection(secondLevelMenus)

		for i := 0; i < len(allMenus); i++ {
			parentIDOptions = append(parentIDOptions, types.FieldOption{
				TextHTML: "&nbsp;&nbsp;┝  " + language.GetFromHtml(template.HTML(allMenus[i]["title"].(string))),
				Value:    strconv.Itoa(int(allMenus[i]["id"].(int64))),
			})
			col := secondLevelMenusCol.Where("parent_id", "=", allMenus[i]["id"].(int64))
			for i := 0; i < len(col); i++ {
				parentIDOptions = append(parentIDOptions, types.FieldOption{
					TextHTML: "&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;┝  " +
						language.GetFromHtml(template.HTML(col[i]["title"].(string))),
					Value: strconv.Itoa(int(col[i]["id"].(int64))),
				})
			}
		}
	}

	formList := menuTable.GetForm().AddXssJsFilter()
	formList.AddField("ID", "id", db.Int, form.Default).FieldDisplayButCanNotEditWhenUpdate().FieldDisableWhenCreate()
	formList.AddField(lg("parent"), "parent_id", db.Int, form.SelectSingle).
		FieldOptions(parentIDOptions).
		FieldDisplay(func(model types.FieldModel) interface{} {
			var menuItem []string

			if model.ID == "" {
				return menuItem
			}

			menuModel, _ := s.table("goadmin_menu").Select("parent_id").Find(model.ID)
			menuItem = append(menuItem, strconv.FormatInt(menuModel["parent_id"].(int64), 10))
			return menuItem
		})
	formList.AddField(lg("menu name"), "title", db.Varchar, form.Text).FieldMust()
	formList.AddField(lg("header"), "header", db.Varchar, form.Text)
	formList.AddField(lg("icon"), "icon", db.Varchar, form.IconPicker)
	formList.AddField(lg("uri"), "uri", db.Varchar, form.Text)
	formList.AddField("PluginName", "plugin_name", db.Varchar, form.Text).FieldDefault(name).FieldHide()
	formList.AddField(lg("role"), "roles", db.Int, form.Select).
		//FieldOptionsFromTable("goadmin_roles", "slug", "id").
		FieldDisplay(func(model types.FieldModel) interface{} {
			var roles []string
			if model.ID == "" {
				return roles
			}
			roleModel, _ := s.table("goadmin_role_menu").
				Select("role_id").
				Where("menu_id", "=", model.ID).
				All()

			for _, v := range roleModel {
				roles = append(roles, strconv.FormatInt(v["role_id"].(int64), 10))
			}
			return roles
		}).
		FieldOptions(types.FieldOptions{
			{Text: "User", Value: "0"},
			{Text: "Viewer", Value: "1"},
			{Text: "Administrator", Value: "2"},
		})

	formList.AddField(lg("updatedAt"), "updated_at", db.Timestamp, form.Default).FieldDisableWhenCreate()
	formList.AddField(lg("createdAt"), "created_at", db.Timestamp, form.Default).FieldDisableWhenCreate()

	formList.SetTable("goadmin_menu").
		SetTitle(lg("Menus Manage")).
		SetDescription(lg("Menus Manage"))

	return
}
