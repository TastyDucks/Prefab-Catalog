package routes

import (
	"Prefab-Catalog/lib/auth"
	"Prefab-Catalog/lib/db"
	"Prefab-Catalog/lib/web"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
adminData
*/
type adminData struct {
	HeaderData *web.HeaderData
	UserMode   int
}

/*
Admin displays the administrator tool panel.
*/
func Admin(c *gin.Context) {
	if !auth.UserCanAccess(c, 2) {
		Forbidden(c)
		return
	}
	user := db.UserGet(auth.GetLogin(c))
	data := &adminData{HeaderData: &web.HeaderData{Title: "Admin Tools"}, UserMode: user.Mode}
	c.HTML(http.StatusOK, "admin.tmpl", data)
}

/*
adminAuditLogData
*/
type adminAuditLogData struct {
	HeaderData *web.HeaderData
	UserMode   int
	List       []db.AuditLogAdmin
}

/*
AdminAuditLog displays the administrative actions audit log
*/
func AdminAuditLog(c *gin.Context) {
	if !auth.UserCanAccess(c, 2) {
		Forbidden(c)
		return
	}
	user := db.UserGet(auth.GetLogin(c))
	data := &adminAuditLogData{HeaderData: &web.HeaderData{Title: "Audit Log"}, UserMode: user.Mode, List: db.AuditLogAdminGetAll()}
	c.HTML(http.StatusOK, "adminAuditLog.tmpl", data)
}

/*
adminUserListData
*/
type adminUserListData struct {
	HeaderData *web.HeaderData
	UserMode   int
	List       []db.User
}

/*
AdminUserList displays the list of users.
*/
func AdminUserList(c *gin.Context) {
	if !auth.UserCanAccess(c, 2) {
		Forbidden(c)
		return
	}
	user := db.UserGet(auth.GetLogin(c))
	List := db.UserGetAll()
	data := &adminUserListData{HeaderData: &web.HeaderData{Title: "User List"}, UserMode: user.Mode, List: List}
	c.HTML(http.StatusOK, "adminUserList.tmpl", data)
}
