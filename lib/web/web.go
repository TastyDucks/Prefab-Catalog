package web

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

/*
HeaderData defines the variables which may be passed to a page's header.
*/
type HeaderData struct {
	Title       string   // The title of a page.
	Stylesheets []string // Any additional CSS documents to include when rendering the page.
}

/*
SetTarget sets the "target" cookie.
*/
func SetTarget(c *gin.Context, target string) {
	session := sessions.Default(c)
	session.Set("target", target)
	session.Save()
}

/*
GetTarget gets the currently targeted uuid.
*/
func GetTarget(c *gin.Context) string {
	session := sessions.Default(c)
	target := session.Get("target")
	if target == nil {
		return ""
	}
	return target.(string)
}
