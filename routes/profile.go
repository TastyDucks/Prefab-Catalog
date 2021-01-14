package routes

import (
	"Prefab-Catalog/lib/auth"
	"Prefab-Catalog/lib/db"
	"Prefab-Catalog/lib/web"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

/*
profileData defines the variables that may be passed to the profile page.
*/
type profileData struct {
	HeaderData *web.HeaderData
	UserMode   int
	Image      string
	ID         string
	Username   string
	FirstName  string
	LastName   string
	Contact    string
	Mode       int    // This is the Mode of the user who is being viewed.
	CanDelete  bool   // Whether the currently viewed profile can be deleted. Basically just so the "delete profile" button is disabled in /profile/new/
	Message    string // Error or success message.
}

var profileHeader *web.HeaderData = &web.HeaderData{Title: "Profile"}

/*
Profile shows the profile page.
*/
func Profile(c *gin.Context) {
	if !auth.UserCanAccess(c, 0) {
		Forbidden(c)
		return
	}
	user := db.UserGet(auth.GetLogin(c))
	userMode := user.Mode
	canDelete := true
	ID := c.Param("id")[1:] // We need to delete the leading "/".
	if ID != "" {           // A new profile is being made, or a specific profile is requested.
		if !auth.UserCanAccess(c, 2) {
			Forbidden(c)
			return
		}
		if ID == "new" {
			user = &db.User{ID: db.UUID(), Mode: 1, Image: "/media/none.webp"}
			canDelete = false
		} else {
			user = db.UserGet(ID)
			if user == nil {
				NotFound(c)
				return
			}
		}
	}
	// Users can only delete a targeted user with a lower Mode
	if user.Mode >= userMode {
		canDelete = false
	}
	web.SetTarget(c, user.ID)
	c.HTML(http.StatusOK, "profile.tmpl", &profileData{profileHeader, userMode, user.Image, user.ID, user.Username, user.FirstName, user.LastName, user.Contact, user.Mode, canDelete, ""})
}

/*
ProfilePOST processes the user profile form.
*/
func ProfilePOST(c *gin.Context) {
	c.Request.ParseMultipartForm(10000000) // Store at most 10 MB in memory before using a temp file.
	submitAction := c.PostForm("submit")
	if submitAction == "save" {
		mode, _ := strconv.Atoi(c.PostForm("mode"))
		username := c.PostForm("username")
		password := c.PostForm("password")
		firstname := c.PostForm("firstname")
		lastname := c.PostForm("lastname")
		contact := c.PostForm("contact")
		file, _ := c.FormFile("image")
		message := ""
		var image string
		if file != nil {
			image = db.ImageSet(c, file)
		}
		canDelete := true
		err := db.UserSet(auth.GetLogin(c), &db.User{ID: web.GetTarget(c), Mode: mode, Username: username, Password: password, FirstName: firstname, LastName: lastname, Contact: contact, Image: image})
		if err != nil {
			message = err.Error()
			canDelete = false
		} else {
			message = "Saved!"
		}
		user := db.UserGet(auth.GetLogin(c))
		targetedUser := db.UserGet(web.GetTarget(c))
		c.HTML(http.StatusOK, "profile.tmpl", &profileData{profileHeader, user.Mode, targetedUser.Image, web.GetTarget(c), targetedUser.Username, targetedUser.FirstName, targetedUser.LastName, targetedUser.Contact, targetedUser.Mode, canDelete, message})
	} else if submitAction == "delete" {
		if !auth.UserCanAccess(c, 2) {
			Forbidden(c)
			return
		}
		callerID := auth.GetLogin(c)
		targetID := web.GetTarget(c)
		err := db.UserDelete(callerID, targetID)
		if err != nil {
			user := db.UserGet(web.GetTarget(c))
			c.HTML(http.StatusOK, "profile.tmpl", &profileData{profileHeader, user.Mode, user.Image, web.GetTarget(c), user.Username, user.FirstName, user.LastName, user.Contact, user.Mode, true, err.Error()})
			return
		}
		c.Redirect(http.StatusSeeOther, "/adminUserList")
	}
}
