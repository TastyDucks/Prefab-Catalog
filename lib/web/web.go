/*
Package web provides functions for setting and getting session values.
*/
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

/*
OrderItemsSet saves the current order items to a cookie.
*/
func OrderItemsSet(c *gin.Context, items Items) {
	session := sessions.Default(c)
	session.Set("orderitems", items)
	session.Save()
}

/*
OrderItemsGet retrieves saved order items.
*/
func OrderItemsGet(c *gin.Context) map[string]int {
	session := sessions.Default(c)
	var items interface{}
	if items = session.Get("orderitems"); items == nil {
		return nil
	}
	return items.(Items)
}

// Items is a map[string]int, for storing a list of parts or assemblies.
type Items map[string]int
