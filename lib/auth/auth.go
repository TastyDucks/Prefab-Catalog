/*
Package auth handles user authentication.
*/
package auth

import (
	"Prefab-Catalog/lib/db"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

/*
SetLogin sets a login cookie.
*/
func SetLogin(c *gin.Context, ID string) {
	session := sessions.Default(c)
	session.Set("ID", ID)
	session.Save()
}

/*
GetLogin gets the currently logged in uuid.
*/
func GetLogin(c *gin.Context) string {
	session := sessions.Default(c)
	ID := session.Get("ID")
	if ID == nil {
		return ""
	}
	return ID.(string)
}

/*
UserGetMode is a simple function to get the Mode of a User. If user is not logged in, returns -1.
*/
func UserGetMode(c *gin.Context) int {
	uuid := GetLogin(c)
	user := db.UserGet(uuid)
	if user != nil {
		return user.Mode
	}
	return -1
}

/*
UserCanAccess takes an integer of the required permission level to view a page. If the user's level is below the required one, they are shown 401 FORBIDDEN.
*/
func UserCanAccess(c *gin.Context, level int) bool {
	mode := UserGetMode(c)
	if mode < level {
		return false
	}
	return true
}
