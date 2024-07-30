// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"net/url"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/page"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	template2 "github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	v1 "github.com/dypflying/chime-portal/v1"
)

// Invoker contains the callback functions which are used
// in the route middleware.
type Invoker struct {
	prefix                 string
	authFailCallback       MiddlewareCallback
	permissionDenyCallback MiddlewareCallback
	conn                   db.Connection
}

// Middleware is the default auth middleware of plugins.
func Middleware(conn db.Connection) context.Handler {
	return DefaultInvoker(conn).Middleware()
}

// DefaultInvoker return a default Invoker.
func DefaultInvoker(conn db.Connection) *Invoker {
	return &Invoker{
		prefix: config.Prefix(),
		authFailCallback: func(ctx *context.Context) {
			if ctx.Request.URL.Path == config.Url(config.GetLoginUrl()) {
				return
			}
			if ctx.Request.URL.Path == config.Url("/logout") {
				ctx.Write(302, map[string]string{
					"Location": config.Url(config.GetLoginUrl()),
				}, ``)
				return
			}
			param := ""
			if ref := ctx.Referer(); ref != "" {
				param = "?ref=" + url.QueryEscape(ref)
			}

			u := config.Url(config.GetLoginUrl() + param)
			_, err := ctx.Request.Cookie(DefaultCookieKey)
			referer := ctx.Referer()

			if (ctx.Headers(constant.PjaxHeader) == "" && ctx.Method() != "GET") ||
				err != nil ||
				referer == "" {
				ctx.Write(302, map[string]string{
					"Location": u,
				}, ``)
			} else {
				msg := language.Get("login overdue, please login again")
				ctx.HTML(http.StatusOK, `<script>
	if (typeof(swal) === "function") {
		swal({
			type: "info",
			title: "`+language.Get("login info")+`",
			text: "`+msg+`",
			showCancelButton: false,
			confirmButtonColor: "#3c8dbc",
			confirmButtonText: '`+language.Get("got it")+`',
        })
		setTimeout(function(){ location.href = "`+u+`"; }, 3000);
	} else {
		alert("`+msg+`")
		location.href = "`+u+`"
    }
</script>`)
			}
		},
		permissionDenyCallback: func(ctx *context.Context) {
			if ctx.Headers(constant.PjaxHeader) == "" && ctx.Method() != "GET" {
				ctx.JSON(http.StatusForbidden, map[string]interface{}{
					"code": http.StatusForbidden,
					"msg":  language.Get(errors.PermissionDenied),
				})
			} else {
				page.SetPageContent(ctx, Auth(ctx), func(ctx interface{}) (types.Panel, error) {
					return template2.WarningPanel(errors.PermissionDenied, template2.NoPermission403Page), nil
				}, conn)
			}
		},
		conn: conn,
	}
}

// SetPrefix return the default Invoker with the given prefix.
func SetPrefix(prefix string, conn db.Connection) *Invoker {
	i := DefaultInvoker(conn)
	i.prefix = prefix
	return i
}

// SetAuthFailCallback set the authFailCallback of Invoker.
func (invoker *Invoker) SetAuthFailCallback(callback MiddlewareCallback) *Invoker {
	invoker.authFailCallback = callback
	return invoker
}

// SetPermissionDenyCallback set the permissionDenyCallback of Invoker.
func (invoker *Invoker) SetPermissionDenyCallback(callback MiddlewareCallback) *Invoker {
	invoker.permissionDenyCallback = callback
	return invoker
}

// MiddlewareCallback is type of callback function.
type MiddlewareCallback func(ctx *context.Context)

// Middleware get the auth middleware from Invoker.
func (invoker *Invoker) Middleware() context.Handler {
	return func(ctx *context.Context) {
		user, authOk, permissionOk := Filter(ctx, invoker.conn)

		if authOk && permissionOk {
			ctx.SetUserValue("user", user)
			ctx.Next()
			return
		}
		if !authOk {
			invoker.authFailCallback(ctx)
			ctx.Abort()
			return
		}
		if !permissionOk {
			ctx.SetUserValue("user", user)
			invoker.permissionDenyCallback(ctx)
			ctx.Abort()
			return
		}
	}
}

// Filter retrieve the user model from Context and check the permission
// at the same time.
func Filter(ctx *context.Context, conn db.Connection) (models.UserModel, bool, bool) {

	token := ctx.Cookie(v1.DefaultPortalCookie)
	user, ok := GetCurUser(token, conn)
	if !ok {
		return user, false, true
	}
	return user, true, CheckPermissions(user, ctx.Request.URL.RequestURI(), ctx.Method(), ctx.PostForm())
}

// GetCurUser return the user model.
func GetCurUser(sesKey string, conn db.Connection) (models.UserModel, bool) {

	var userMap map[string]any
	var err error
	user := models.User().SetConn(conn)
	if userMap, err = v1.GetUser(sesKey); err != nil {
		return user, false
	}
	user.UUID, _ = userMap["uuid"].(string)
	user.Name, _ = userMap["nick_name"].(string)
	user.UserName, _ = userMap["name"].(string)
	user.Avatar, _ = userMap["avatar"].(string)
	if user.Avatar == "" || config.GetStore().Prefix == "" {
		user.Avatar = ""
	} else {
		user.Avatar = config.GetStore().URL(user.Avatar)
	}
	user.Role, _ = userMap["role"].(int64)
	user.LevelName = "Super"
	user = user.WithMenus()
	return user, user.HasMenu()
}

// CheckPermissions check the permission of the user.
func CheckPermissions(user models.UserModel, path, method string, param url.Values) bool {
	return user.CheckPermissionByUrlMethod(path, method, param)
}
