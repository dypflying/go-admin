package models

import (
	"database/sql"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	"github.com/dypflying/chime-common/constant"
)

// UserModel is user model structure.
type UserModel struct {
	Base `json:"-"`

	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	UserName  string `json:"user_name"`
	Password  string `json:"password"`
	Avatar    string `json:"avatar"`
	Role      int64  `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	LevelName string `json:"level_name"`

	//no use
	Id            int64          `json:"id"`
	RememberToken string         `json:"remember_token"`
	MenuIds       []int64        `json:"menu_ids"`
	BlackMenuMap  map[string]any `json:"menu_map"`
	AllMenuMap    map[string]any `json:"menu_map"`
	Level         string         `json:"level"`
	cacheReplacer *strings.Replacer
}

// User return a default user model.
func User() UserModel {
	return UserModel{Base: Base{TableName: config.GetAuthUserTable()}}
}

func (t UserModel) SetConn(con db.Connection) UserModel {
	t.Conn = con
	return t
}

func (t UserModel) WithTx(tx *sql.Tx) UserModel {
	t.Tx = tx
	return t
}

// Find return a default user model of given id.
func (t UserModel) Find(id interface{}) UserModel {
	item, _ := t.Table(t.TableName).Find(id)
	return t.MapToModel(item)
}

// FindByUserName return a default user model of given name.
func (t UserModel) FindByUserName(username interface{}) UserModel {
	item, _ := t.Table(t.TableName).Where("username", "=", username).First()
	return t.MapToModel(item)
}

// IsEmpty check the user model is empty or not.
func (t UserModel) IsEmpty() bool {
	return t.Id == int64(0)
}

// HasMenu check the user has visitable menu or not.
func (t UserModel) HasMenu() bool {
	return len(t.MenuIds) != 0 || t.IsSuperAdmin()
}

// IsSuperAdmin check the user model is super admin or not.
func (t UserModel) IsSuperAdmin() bool {
	return t.Role == constant.ROLE_SUPER
}

func (t UserModel) GetCheckPermissionByUrlMethod(path, method string) string {
	if !t.CheckPermissionByUrlMethod(path, method, url.Values{}) {
		return ""
	}
	return path
}

func (t UserModel) Template(str string) string {
	if t.cacheReplacer == nil {
		t.cacheReplacer = strings.NewReplacer("{{.AuthId}}", strconv.Itoa(int(t.Id)),
			"{{.AuthName}}", t.Name, "{{.AuthUserName}}", t.UserName)
	}
	return t.cacheReplacer.Replace(str)
}

func (t UserModel) CheckPermissionByUrlMethod(path, method string, formParams url.Values) bool {

	if t.IsSuperAdmin() {
		return true
	}
	originalPath := strings.Split(path, "?")[0]
	if strings.Index(originalPath, config.Prefix()) == 0 {
		originalPath = originalPath[len(config.Prefix()):]
	}
	if _, ok := t.BlackMenuMap[originalPath]; ok {
		return false
	}
	return true
}

func (t UserModel) HideUserCenterEntrance() bool {
	return false
}

// UpdateAvatar update the avatar of user.
func (t UserModel) ReleaseConn() UserModel {
	t.Conn = nil
	return t
}

// UpdateAvatar update the avatar of user.
func (t UserModel) UpdateAvatar(avatar string) {
	t.Avatar = avatar
}

// WithMenus query the menu info of the user.
func (t UserModel) WithMenus() UserModel {

	var menuIdsModel []map[string]interface{}

	//initiate all menu
	if t.AllMenuMap == nil {
		t.AllMenuMap = make(map[string]any)
		allMenus, _ := t.Table("goadmin_role_menu").
			LeftJoin("goadmin_menu", "goadmin_menu.id", "=", "goadmin_role_menu.menu_id").
			Select("menu_id", "parent_id", "uri").
			All()
		for _, menu := range allMenus {
			if menu["uri"].(string) != "" {
				t.AllMenuMap[menu["uri"].(string)] = struct{}{}
			}
		}
	}

	if t.IsSuperAdmin() {
		menuIdsModel, _ = t.Table("goadmin_role_menu").
			LeftJoin("goadmin_menu", "goadmin_menu.id", "=", "goadmin_role_menu.menu_id").
			Select("menu_id", "parent_id", "role_id", "uri").
			All()
	} else {
		menuIdsModel, _ = t.Table("goadmin_role_menu").
			LeftJoin("goadmin_menu", "goadmin_menu.id", "=", "goadmin_role_menu.menu_id").
			Select("menu_id", "parent_id", "role_id", "uri").
			Where("goadmin_role_menu.role_id", "=", t.Role).
			All()
	}

	var menuIds []int64
	menuMap := make(map[string]any)
	for _, mid := range menuIdsModel {
		if parentId, ok := mid["parent_id"].(int64); ok && parentId != 0 {
			for _, mid2 := range menuIdsModel {
				if mid2["menu_id"].(int64) == mid["parent_id"].(int64) {
					menuIds = append(menuIds, mid["menu_id"].(int64))
					if mid["uri"].(string) != "" {
						menuMap[mid["uri"].(string)] = struct{}{}
					}
					break
				}
			}
		} else {
			menuIds = append(menuIds, mid["menu_id"].(int64))
			if mid["uri"].(string) != "" {
				menuMap[mid["uri"].(string)] = struct{}{}
			}
		}
	}
	t.MenuIds = menuIds

	blackMenuMap := make(map[string]any)
	for uri, _ := range t.AllMenuMap {
		if _, ok := menuMap[uri]; !ok {
			blackMenuMap[uri] = struct{}{}
		}
	}
	t.BlackMenuMap = blackMenuMap
	return t
}

// New create a user model.
func (t UserModel) New(username, password, name, avatar string) (UserModel, error) {

	id, err := t.WithTx(t.Tx).Table(t.TableName).Insert(dialect.H{
		"username": username,
		"password": password,
		"name":     name,
		"avatar":   avatar,
	})

	t.Id = id
	t.UserName = username
	t.Password = password
	t.Avatar = avatar
	t.Name = name

	return t, err
}

// Update update the user model.
func (t UserModel) Update(username, password, name, avatar string, isUpdateAvatar bool) (int64, error) {

	fieldValues := dialect.H{
		"username":   username,
		"name":       name,
		"updated_at": time.Now().Format("2006-01-02 15:04:05"),
	}

	if avatar == "" || isUpdateAvatar {
		fieldValues["avatar"] = avatar
	}

	if password != "" {
		fieldValues["password"] = password
	}

	return t.WithTx(t.Tx).Table(t.TableName).
		Where("id", "=", t.Id).
		Update(fieldValues)
}

// UpdatePwd update the password of the user model.
func (t UserModel) UpdatePwd(password string) UserModel {

	_, _ = t.Table(t.TableName).
		Where("id", "=", t.Id).
		Update(dialect.H{
			"password": password,
		})

	t.Password = password
	return t
}

// MapToModel get the user model from given map.
func (t UserModel) MapToModel(m map[string]interface{}) UserModel {
	t.Id, _ = m["id"].(int64)
	t.Name, _ = m["name"].(string)
	t.UserName, _ = m["username"].(string)
	t.Password, _ = m["password"].(string)
	t.Avatar, _ = m["avatar"].(string)
	t.RememberToken, _ = m["remember_token"].(string)
	t.CreatedAt, _ = m["created_at"].(string)
	t.UpdatedAt, _ = m["updated_at"].(string)
	return t
}
