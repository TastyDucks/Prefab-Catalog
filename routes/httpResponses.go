package routes

import (
	"Prefab-Catalog/lib/lumberjack"
	"Prefab-Catalog/lib/web"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
Unauthorized (401).
*/
func Unauthorized(c *gin.Context) {
	HeaderData := &web.HeaderData{Title: "401: Unauthorized", Stylesheets: []string{"error"}}
	c.HTML(http.StatusUnauthorized, "401.tmpl", gin.H{"HeaderData": HeaderData})
}

/*
Forbidden (403).
*/
func Forbidden(c *gin.Context) {
	HeaderData := &web.HeaderData{Title: "403: Forbidden", Stylesheets: []string{"error"}}
	c.HTML(http.StatusForbidden, "403.tmpl", gin.H{"HeaderData": HeaderData})
}

/*
NotFound (404).
*/
func NotFound(c *gin.Context) {
	HeaderData := &web.HeaderData{Title: "404: Resource not found", Stylesheets: []string{"error"}}
	c.HTML(http.StatusNotFound, "404.tmpl", gin.H{"HeaderData": HeaderData})
}

/*
MethodNotAllowed (405).
*/
func MethodNotAllowed(c *gin.Context) {
	HeaderData := &web.HeaderData{Title: "405: Method not allowed", Stylesheets: []string{"error"}}
	c.HTML(http.StatusMethodNotAllowed, "405.tmpl", gin.H{"HeaderData": HeaderData})
}

/*
InternalServerError (500).
*/
func InternalServerError(c *gin.Context, err interface{}) {
	log := lumberjack.New("Router")
	type internalServerErrorData struct {
		HeaderData *web.HeaderData
		Error      interface{}
	}
	HeaderData := &web.HeaderData{Title: "500: Internal server error", Stylesheets: []string{"error"}}
	data := &internalServerErrorData{HeaderData: HeaderData, Error: err}
	c.HTML(http.StatusInternalServerError, "500.tmpl", data)
	e := fmt.Errorf("%s", err)
	log.Error(e, "Internal server error")
}

/*
ServiceUnavailable (503).
*/
func ServiceUnavailable(c *gin.Context) {
	HeaderData := &web.HeaderData{Title: "503: Service unavailable", Stylesheets: []string{"error"}}
	c.HTML(http.StatusServiceUnavailable, "503.tmpl", gin.H{"HeaderData": HeaderData})
}
