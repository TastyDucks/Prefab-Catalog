package routes

import (
	"Prefab-Catalog/lib/auth"
	"Prefab-Catalog/lib/config"
	"Prefab-Catalog/lib/db"
	"Prefab-Catalog/lib/web"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
indexData defines the variables that may be passed to the index page.
*/
type indexData struct {
	HeaderData *web.HeaderData
	UserMode   int
	Build      string
	LoggedIn   bool
}

var indexHeader *web.HeaderData = &web.HeaderData{Title: "Prefab Catalog - Login", Stylesheets: []string{"index"}}

/*
Index shows the index.
*/
func Index(c *gin.Context) {
	if auth.UserGetMode(c) == -1 {
		c.HTML(http.StatusOK, "index.tmpl", &indexData{indexHeader, -1, config.Build(), true})
	} else {
		c.Redirect(http.StatusSeeOther, "/assembly")
	}
}

/*
Login attempts to log a user in.
*/
func Login(c *gin.Context) {
	c.Request.ParseForm()
	username := c.PostForm("username")
	password := c.PostForm("password")
	loggedIn := db.UserLogin(username, password)
	if loggedIn {
		userData := db.UserGet(username)
		auth.SetLogin(c, userData.ID)
		c.Redirect(http.StatusSeeOther, "/assembly")
	} else {
		c.HTML(http.StatusOK, "index.tmpl", &indexData{indexHeader, -1, config.Build(), false})
	}
}
