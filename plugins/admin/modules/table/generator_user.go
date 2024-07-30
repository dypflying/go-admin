package table

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"github.com/dypflying/chime-common/client/user"
	"github.com/dypflying/chime-common/model/response"
	chimemodels "github.com/dypflying/chime-common/models"
	. "github.com/dypflying/chime-portal/v1"
)

func (s *SystemTable) GetManagerTable(ctx *context.Context) (managerTable Table) {

	managerTable = NewDefaultTable(DefaultConfig().SetPrimaryKey("uuid", db.Varchar))

	info := managerTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("Uuid", "uuid", db.Varchar).FieldHide()
	info.AddField(lg("Name"), "name", db.Varchar).FieldFilterable()
	info.AddField(lg("Nickname"), "nick_name", db.Varchar).FieldFilterable()
	info.AddField(lg("Avatar"), "avatar", db.Varchar).FieldHideForList()
	info.AddField(lg("role"), "role", db.Int).
		FieldDisplay(func(model types.FieldModel) interface{} {
			intVal, _ := strconv.Atoi(model.Value)
			return RoleDisplay[intVal]
		}).
		FieldFilterable(types.FilterType{FormType: form.SelectSingle}).
		FieldFilterOptions(types.FieldOptions{
			{Value: "0", Text: "User"},
			{Value: "1", Text: "Viewer"},
			{Value: "2", Text: "Administrator"},
			{Value: "3", Text: "Super Administrator"},
		})
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_users").
		SetTitle(lg("Managers")).
		SetDescription(lg("Managers")).
		SetDeleteFn(func(idArr []string) error {
			return nil
		})

	info.SetTitle(lg("Managers")).
		SetDescription(lg("Managers")).
		SetGetDataFn(func(param parameter.Parameters) (data []map[string]interface{}, size int) {
			queryParams := &user.ListUserParams{Context: DefaultContext}
			if param.PageInt > 0 {
				queryParams.Page = IntToInt64P(param.PageInt - 1)
			}
			queryParams.Size = IntToInt64P(param.PageSizeInt)
			queryParams.Sort = GetSortField(param)
			queryParams.Order = FieldValueString(param.SortType)

			queryParams.Name = FieldValueString(param.GetFieldValue("name"))
			queryParams.Role = FieldValueInt64(param.GetFieldValue("role"))

			obj, err := ApiClient.User.ListUser(queryParams, GetAuthToken(ctx))
			if err != nil {
				PortalLogger.Errorf("failed to list user,  err = %v\n", err)
				return EmptyResponse, 0
			}
			return ParseResponseList(obj)
		})

	info.SetDeleteFn(func(ids []string) error {
		for _, userUuid := range ids {
			if _, err := ApiClient.User.DeleteUser(&user.DeleteUserParams{
				UserUUID: TrimUuid(userUuid),
				Context:  DefaultContext,
			}, GetAuthToken(ctx)); err != nil {
				_, msg := ParseResponseError(err)
				return errors.New(msg)
			}
		}
		return nil
	})

	detail := managerTable.GetDetail()
	detail.AddField(lg("Name"), "name", db.Varchar)
	detail.AddField(lg("Avatar"), "avatar", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			if model.Value == "" || config.GetStore().Prefix == "" {
				model.Value = config.Url("/assets/dist/img/avatar04.png")
			} else {
				model.Value = config.GetStore().URL(model.Value)
			}
			return template.Default().Image().
				SetSrc(template.HTML(model.Value)).
				SetHeight("120").SetWidth("120").WithModal().GetContent()
		})
	detail.AddField(lg("Nickname"), "nick_name", db.Varchar)
	detail.AddField(lg("role"), "role", db.Int).
		FieldDisplay(func(model types.FieldModel) interface{} {
			intVal, _ := strconv.Atoi(model.Value)
			return RoleDisplay[intVal]
		})

	detail.AddField(lg("createdAt"), "created_at", db.Timestamp)
	detail.AddField(lg("updatedAt"), "updated_at", db.Timestamp)
	detail.SetGetDataFn(func(param parameter.Parameters) (data []map[string]interface{}, size int) {
		queryParams := &user.GetUserParams{
			Context:  DefaultContext,
			UserUUID: param.PK(),
		}
		obj, err := ApiClient.User.GetUser(queryParams, GetAuthToken(ctx))
		if err != nil {
			PortalLogger.Errorf("failed to get user,  err = %v\n", err)
			return EmptyResponse, 0
		}
		return ParseResponseData(obj, response.UserEntityName)
	})

	formList := managerTable.GetForm().AddXssJsFilter()

	formList.AddField("Uuid", "uuid", db.Varchar, form.Text).
		FieldNotAllowEdit().
		FieldNotAllowAdd().
		FieldInputWidth(6)

	formList.AddField(lg("Name"), "name", db.Varchar, form.Text).
		FieldMust().
		FieldInputWidth(6)

	formList.AddField(lg("Nickname"), "nick_name", db.Varchar, form.Text).
		FieldMust().
		FieldInputWidth(6)
	formList.AddField(lg("Avatar"), "avatar", db.Varchar, form.File).
		FieldInputWidth(6)
	formList.AddField("Role", "role", db.Tinyint, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "User", Value: "0"},
			{Text: "Adminitrator", Value: "1"},
			{Text: "Super Adminitrator", Value: "2"},
			{Text: "Viewer", Value: "3"},
		}).FieldDefault("0").
		FieldMust().
		FieldInputWidth(6)

	formList.AddField(lg("password"), "password", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		}).
		FieldInputWidth(6).
		FieldMust()
	formList.AddField(lg("confirm password"), "password_again", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		}).
		FieldInputWidth(6).
		FieldMust()

	formList.SetTitle(lg("Managers")).SetDescription(lg("Managers"))
	formList.SetUpdateFn(func(values form2.Values) error {
		password := values.Get("password")
		password_again := values.Get("password_again")
		if password != password_again {
			return errors.New("password does not match")
		}
		name := values.Get("name")
		nickName := values.Get("nick_name")
		req := &chimemodels.CreateUserRequest{
			Name:     &name,
			NickName: &nickName,
			Password: password,
			Avatar:   values.Get("avatar"),
		}
		req.Role = int64(GetIntValue(values.Get("role")))
		if _, err := ApiClient.User.UpdateUser(&user.UpdateUserParams{
			Body:     req,
			UserUUID: values.Get("uuid"),
			Context:  DefaultContext,
		}, GetAuthToken(ctx)); err != nil {
			_, msg := ParseResponseError(err)
			return errors.New(msg)
		}
		return nil
	})
	formList.SetInsertFn(func(values form2.Values) error {
		password := values.Get("password")
		password_again := values.Get("password_again")
		if password != password_again {
			return errors.New("password does not match")
		}
		name := values.Get("name")
		nickName := values.Get("nick_name")
		req := &chimemodels.CreateUserRequest{
			Name:     &name,
			NickName: &nickName,
			Password: password,
			Avatar:   values.Get("avatar"),
		}
		req.Role = int64(GetIntValue(values.Get("role")))
		if _, err := ApiClient.User.CreateUser(&user.CreateUserParams{
			Body:    req,
			Context: DefaultContext,
		}, GetAuthToken(ctx)); err != nil {
			_, msg := ParseResponseError(err)
			return errors.New(msg)
		}
		return nil
	})
	return
}

func (s *SystemTable) GetPersonalTable(ctx *context.Context) (personalTable Table) {

	if ctx.User() == nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}
	loginUser := ctx.User().(models.UserModel)

	personalTable = NewDefaultTable(DefaultConfig().SetPrimaryKey("uuid", db.Varchar))

	info := personalTable.GetInfo().AddXssJsFilter().HideFilterArea()
	info.AddField("Uuid", "uuid", db.Varchar).FieldHide()
	info.SetPrimaryKey("uuid", db.Varchar)
	info.AddField(lg("Name"), "name", db.Varchar).FieldFilterable()
	info.AddField(lg("Nickname"), "nick_name", db.Varchar).FieldFilterable()
	info.AddField(lg("Avatar"), "avatar", db.Varchar)
	info.AddField(lg("role"), "role", db.Int).
		FieldDisplay(func(model types.FieldModel) interface{} {
			intVal, _ := strconv.Atoi(model.Value)
			return RoleDisplay[intVal]
		}).
		FieldFilterable(types.FilterType{FormType: form.SelectSingle}).
		FieldFilterOptions(types.FieldOptions{
			{Value: "0", Text: "User"},
			{Value: "1", Text: "Administrator"},
			{Value: "2", Text: "Super Administrator"},
			{Value: "3", Text: "Viewer"},
		})
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_users").
		SetTitle(lg("Managers")).
		SetDescription(lg("Managers")).
		SetDeleteFn(func(idArr []string) error {
			return nil
		})
	info.SetTitle(lg("Managers")).
		SetDescription(lg("Managers"))

	info.SetGetDataFn(func(param parameter.Parameters) (data []map[string]interface{}, size int) {
		return nil, 0 //a placeholder function, in order to not invoke the defaul fn which based on db query
	})

	details := personalTable.GetDetailFromInfo()
	details.SetGetDataFn(func(param parameter.Parameters) (data []map[string]interface{}, size int) {
		queryParams := &user.GetUserParams{
			Context:  DefaultContext,
			UserUUID: loginUser.UUID,
		}
		obj, err := ApiClient.User.GetUser(queryParams, GetAuthToken(ctx))
		if err != nil {
			PortalLogger.Errorf("failed to get user,  err = %v\n", err)
			return EmptyResponse, 0
		}
		return ParseResponseData(obj, response.UserEntityName)
	})

	formList := personalTable.GetForm().AddXssJsFilter().
		//jump to the main board
		SetAjaxSuccessJS(`
		  	$.pjax({url:'/admin', container: '#pjax-container'});
		`).SetAjaxErrorJS(`
			swal(data.responseJSON.msg, '', 'error');
			$.pjax.reload('#pjax-container');	
		`)
	formList.SetPrimaryKey("uuid", db.Varchar)
	formList.HideBackButton()
	formList.HideContinueEditCheckBox()
	formList.HideContinueNewCheckBox()
	formList.AddField("Uuid", "uuid", db.Varchar, form.Text).
		FieldNotAllowEdit().
		FieldNotAllowAdd().
		FieldInputWidth(6)

	formList.AddField(lg("Name"), "name", db.Varchar, form.Text).
		FieldMust().
		FieldInputWidth(6)

	formList.AddField(lg("Nickname"), "nick_name", db.Varchar, form.Text).
		FieldMust().
		FieldInputWidth(6)
	formList.AddField(lg("Avatar"), "avatar", db.Varchar, form.File).
		FieldInputWidth(6)

	formList.AddField(lg("password"), "password", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		}).
		FieldInputWidth(6).
		FieldMust()
	formList.AddField(lg("confirm password"), "password_again", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		}).
		FieldInputWidth(6).
		FieldMust()

	formList.SetTitle(lg("Managers")).SetDescription(lg("Managers"))
	formList.SetUpdateFn(func(values form2.Values) error {
		password := values.Get("password")
		password_again := values.Get("password_again")
		if password != password_again {
			return errors.New("password does not match")
		}

		avatar := values.Get("avatar")
		deleteFlag := values.Get("avatar__delete_flag")
		if deleteFlag == "1" {
			avatar = ""
		} else if avatar == "" && loginUser.Avatar != "" {
			subAvatars := strings.Split(loginUser.Avatar, "/")
			avatar = subAvatars[len(subAvatars)-1]
		}
		name := values.Get("name")
		nickName := values.Get("nick_name")
		req := &chimemodels.CreateUserRequest{
			Name:     &name,
			NickName: &nickName,
			Password: password,
			Avatar:   avatar,
		}
		req.Role = loginUser.Role
		if _, err := ApiClient.User.UpdateUser(&user.UpdateUserParams{
			Body:     req,
			UserUUID: values.Get("uuid"),
			Context:  DefaultContext,
		}, GetAuthToken(ctx)); err != nil {
			_, msg := ParseResponseError(err)
			return errors.New(msg)
		}
		return nil
	})
	return
}
