package routes

import (
	"Prefab-Catalog/lib/auth"
	"Prefab-Catalog/lib/db"
	"Prefab-Catalog/lib/web"
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
)

/*
catalogData defines the variables that may be passed to the catalog page.
*/
type catalogData struct {
	HeaderData *web.HeaderData
	UserMode   int
	Username   string
	Bracket    string
	Raceway    string
	System     string
	Device     string
}

/*
Catalog displays the catalog.
*/
func Catalog(c *gin.Context) { // TODO: Delete this!
	if !auth.UserCanAccess(c, 0) {
		Forbidden(c)
		return
	}
	user := db.UserGet(auth.GetLogin(c))
	data := &catalogData{HeaderData: &web.HeaderData{}, UserMode: user.Mode, Username: user.Username}
	c.HTML(http.StatusOK, "catalog.tmpl", data)
}

/*
All the data that may be passed to the assembly page.
*/
type assemblyData struct {
	HeaderData         *web.HeaderData
	UserMode           int
	Assembly           db.Assembly
	PartsAndQuantities map[db.Part]int
	CanDelete          bool
	Message            string
}

/*
All the data that may be passed to the assembly list page.
*/
type assemblyListData struct {
	HeaderData *web.HeaderData
	UserMode   int
	Assemblies []db.Assembly
}

var assemblyHeader *web.HeaderData = &web.HeaderData{Title: "Assembly", Stylesheets: []string{"datatables.min"}}

/*
Assembly displays the assembly description / creation / modification page.
*/
func Assembly(c *gin.Context) {
	if !auth.UserCanAccess(c, 0) {
		Forbidden(c)
		return
	}
	user := db.UserGet(auth.GetLogin(c))
	var assembly db.Assembly
	canDelete := true
	ID := c.Param("id")[1:] // We need to delete the leading "/".
	if ID != "" {
		if ID == "new" {
			if !auth.UserCanAccess(c, 2) {
				Forbidden(c)
				return
			}
			canDelete = false
			assembly = db.Assembly{ID: db.UUID(), Name: "", Description: "", BracketStyle: "", RacewayStyle: "", System: "", Device: false, Minutes: 0, Items: nil, Image: "/media/none.webp"}
		} else {
			assembly = db.AssemblyGet(ID)
			if assembly.ID == "" {
				NotFound(c)
				return
			}
		}
		parts := db.PartGetAll(false)
		partsAndQuantities := make(map[db.Part]int)
		for _, part := range parts {
			partsAndQuantities[part] = 0 // Assign a default quantity of zero.
			for p, q := range assembly.Items {
				if part.ID == p {
					partsAndQuantities[part] = q // Assign the actual quantity if there is one in the variable.
				}
			}
		}
		web.SetTarget(c, assembly.ID)
		c.HTML(http.StatusOK, "assembly.tmpl", &assemblyData{HeaderData: assemblyHeader, UserMode: user.Mode, Assembly: assembly, PartsAndQuantities: partsAndQuantities, CanDelete: canDelete, Message: ""})
		return
	}
	c.HTML(http.StatusOK, "assemblyList.tmpl", &assemblyListData{HeaderData: assemblyHeader, UserMode: user.Mode, Assemblies: db.AssemblyGetAll(false)})
}

/*
AssemblyPOST processes the assembly form.
*/
func AssemblyPOST(c *gin.Context) {
	c.Request.ParseMultipartForm(10000000) // Store at most 10 MB in memory before using a temp file.
	items := CollateItems(c.Request.PostForm)
	submitAction := c.PostForm("submit")
	if submitAction == "save" {
		name := c.PostForm("name")
		bracketStyle := c.PostForm("bracketstyle")
		racewayStyle := c.PostForm("racewaystyle")
		system := c.PostForm("system")
		device, _ := strconv.ParseBool(c.PostForm("device"))
		description := c.PostForm("description")
		minutes, _ := strconv.Atoi(c.PostForm("minutes"))
		file, _ := c.FormFile("image")
		message := ""
		canDelete := true
		var image string
		if file != nil {
			image = db.ImageSet(c, file)
		}
		err := db.AssemblySet(auth.GetLogin(c), db.Assembly{ID: web.GetTarget(c), Name: name, Description: description, BracketStyle: bracketStyle, RacewayStyle: racewayStyle, System: system, Device: device, Minutes: minutes, Items: items, Image: image})
		if err != nil {
			message = err.Error()
			canDelete = false
		} else {
			message = "Saved!"
		}
		user := db.UserGet(auth.GetLogin(c))
		assembly := db.AssemblyGet(web.GetTarget(c))
		partsAndQuantities := make(map[db.Part]int)
		for key, value := range items {
			partsAndQuantities[db.PartGet(key)] = value
		}
		c.HTML(http.StatusOK, "assembly.tmpl", &assemblyData{assemblyHeader, user.Mode, assembly, partsAndQuantities, canDelete, message})
	} else if submitAction == "delete" {
		db.AssemblyDelete(auth.GetLogin(c), web.GetTarget(c))
		c.Redirect(http.StatusSeeOther, "/assembly/")
	}
}

/*
All the data that may be passed to the Part page.
*/
type partData struct {
	HeaderData *web.HeaderData
	UserMode   int
	Part       db.Part
	CanDelete  bool   // Whether the currently viewed part can be deleted.
	Message    string // Any sort of error or notification message.
}

/*
All the data that may be passed to the Part list page.
*/
type partListData struct {
	HeaderData *web.HeaderData
	UserMode   int
	Parts      []db.Part
}

var partHeader *web.HeaderData = &web.HeaderData{Title: "Part", Stylesheets: []string{"datatables.min"}}

/*
Part displays the part description / creation / modification page.
*/
func Part(c *gin.Context) {
	if !auth.UserCanAccess(c, 0) {
		Forbidden(c)
		return
	}
	user := db.UserGet(auth.GetLogin(c))
	var part db.Part
	canDelete := true
	ID := c.Param("id")[1:] // We need to delete the leading "/".
	if ID != "" {           // A new part is being made, or a specific part is being requested.
		if ID == "new" {
			if !auth.UserCanAccess(c, 2) {
				Forbidden(c)
				return
			}
			canDelete = false
			part = db.Part{ID: db.UUID(), Number: "", Manufacturer: "", Name: "", Description: "", Unit: "quantity", CostPerUnit: 0, Image: "/media/none.webp"}
		} else {
			part = db.PartGet(ID)
			if part.ID == "" {
				NotFound(c)
				return
			}
		}
		web.SetTarget(c, part.ID)
		c.HTML(http.StatusOK, "part.tmpl", &partData{partHeader, user.Mode, part, canDelete, ""})
		return
	}
	parts := db.PartGetAll(false)
	c.HTML(http.StatusOK, "partList.tmpl", &partListData{partHeader, user.Mode, parts})
}

/*
PartPOST processes the part form.
*/
func PartPOST(c *gin.Context) {
	c.Request.ParseMultipartForm(10000000) // Store at most 10 MB in memory before using a temp file.
	submitAction := c.PostForm("submit")
	if submitAction == "save" {
		name := c.PostForm("name")
		number := c.PostForm("number")
		manufacturer := c.PostForm("manufacturer")
		description := c.PostForm("description")
		unit := c.PostForm("unit")
		costPerUnit, _ := strconv.Atoi(c.PostForm("costPerUnit")) // Ignoring if the integer is malformed here because db.PartSet() will pick that up.
		file, _ := c.FormFile("image")
		message := ""
		canDelete := true
		var image string
		if file != nil {
			image = db.ImageSet(c, file)
		}
		err := db.PartSet(auth.GetLogin(c), db.Part{ID: web.GetTarget(c), Number: number, Manufacturer: manufacturer, Name: name, Description: description, Unit: unit, CostPerUnit: costPerUnit, Image: image})
		if err != nil {
			message = err.Error()
			canDelete = false
		} else {
			message = "Saved!"
		}
		user := db.UserGet(auth.GetLogin(c))
		part := db.PartGet(web.GetTarget(c))
		c.HTML(http.StatusOK, "part.tmpl", &partData{partHeader, user.Mode, part, canDelete, message})
	} else if submitAction == "delete" {
		db.PartDelete(auth.GetLogin(c), web.GetTarget(c))
		c.Redirect(http.StatusSeeOther, "/part/")
	}
}

/*
CollateItems takes a map of keys and values, discards all keys which are not UUIDS (that is, part or assembly IDs), and discards all values less than 1. Returns a map[string]int
*/
func CollateItems(input map[string][]string) (results map[string]int) {
	results = make(map[string]int)
	for key, value := range input {
		// 32 HEX and 4 hyphens
		if len(key) == 36 {
			q, _ := strconv.Atoi(value[0])
			if q > 0 {
				results[key] = q
			}
		}
	}
	return
}
