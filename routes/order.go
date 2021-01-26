package routes

import (
	"Prefab-Catalog/lib/auth"
	"Prefab-Catalog/lib/db"
	"Prefab-Catalog/lib/mail"
	"Prefab-Catalog/lib/web"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

/*
All the data that may be passed to the Order page.
*/
type orderData struct {
	HeaderData *web.HeaderData
	UserMode   int
	Order      db.AuditLogOrder
	Message    string // Any sort of error or notification message.
	New        bool   // If this is a new order.
}

/*
The data for the printable / offline order form
*/
type orderFormData struct {
	Order     db.AuditLogOrder
	LogoData  string       // Embedded logo.
	StyleData template.CSS // Embedded stylesheet.
	IconData  string       // Embedded favicon.
}

type orderListData struct {
	HeaderData *web.HeaderData
	UserMode   int
	Orders     []db.AuditLogOrder
}

var orderHeader *web.HeaderData = &web.HeaderData{Title: "Order Review", Stylesheets: []string{"datatables.min"}}

/*
Order displays the Order review / list page.
*/
func Order(c *gin.Context) {
	if !auth.UserCanAccess(c, 1) {
		Forbidden(c)
		return
	}
	user := db.UserGet(auth.GetLogin(c))
	var order db.AuditLogOrder
	ID := c.Param("id")[1:] // We need to delete the leading "/".
	if ID != "" {           // A specific order is being requested.
		if ID == "new" { // There isn't really a /order/new page, but I'm making this redirect to / for consistency with /user/new, /assembly/new, and /part/new.
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		// Requesting a previous order.
		order = db.AuditLogOrderGet(ID)
		if order.ID == "" {
			NotFound(c)
			return
		}
		// Download portable order form.
		if c.Query("download") == "true" {
			c.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", "Order Form.html"))
			bytes, _ := ioutil.ReadFile("static/media/logo_color.png")
			logo := base64.StdEncoding.EncodeToString(bytes)
			bytes, _ = ioutil.ReadFile("static/css/orderform.css")
			css := template.CSS(string(bytes))
			bytes, _ = ioutil.ReadFile("static/media/favicon.ico")
			icon := base64.StdEncoding.EncodeToString(bytes)
			c.HTML(http.StatusCreated, "orderForm.tmpl", &orderFormData{
				Order:     order,
				LogoData:  logo,
				StyleData: css,
				IconData:  icon,
			})
			return
		}
		c.HTML(http.StatusOK, "order.tmpl", &orderData{HeaderData: orderHeader, UserMode: user.Mode, Order: order, Message: "", New: false})
		return
	}
	// Order list.
	if !auth.UserCanAccess(c, 2) {
		Forbidden(c)
		return
	}
	orders := db.AuditLogOrderGetAll()
	c.HTML(http.StatusOK, "orderList.tmpl", &orderListData{HeaderData: orderHeader, UserMode: user.Mode, Orders: orders})
}

/*
OrderPOST displays the new order review page.
*/
func OrderPOST(c *gin.Context) {
	c.Request.ParseForm()
	orderID := db.AuditLogOrderIncrementID()
	web.SetTarget(c, orderID)
	user := db.UserGet(auth.GetLogin(c))
	items := CollateItems(c.Request.PostForm)
	web.OrderItemsSet(c, items)
	order := db.AuditLogOrder{
		ID:      orderID,
		Time:    "[Generated upon submission]",
		User:    user.ID,
		Items:   items,
		Project: db.Project{Foreman: db.Foreman{Name: user.FirstName + " " + user.LastName, Contact: db.Contact{Email: user.Contact.Email, Phone: user.Contact.Phone}}}, // Fill in the information for the Foreman (assume that the Foreman is the one making the order).
	}
	c.HTML(http.StatusOK, "order.tmpl", &orderData{HeaderData: orderHeader, UserMode: user.Mode, Order: order, Message: "", New: true})
	return
}

/*
OrderFinishPOST displays the order confirmation page.
*/
func OrderFinishPOST(c *gin.Context) {
	c.Request.ParseForm()
	order := db.AuditLogOrder{
		ID:              web.GetTarget(c),
		Time:            time.Now().Format("2006-01-02 15:04:05"),
		User:            auth.GetLogin(c),
		Items:           web.OrderItemsGet(c),
		Project:         db.Project{Number: c.PostForm("Project.Number"), Name: c.PostForm("Project.Name"), ShippingAddress: c.PostForm("Project.ShippingAddress"), DeliveryLocation: c.PostForm("Project.DeliveryLocation"), Foreman: db.Foreman{Name: c.PostForm("Foreman.Name"), Contact: db.Contact{Email: c.PostForm("Foreman.Contact.Email"), Phone: c.PostForm("Foreman.Contact.Phone"), Other: ""}}},
		PackagingMethod: c.PostForm("PackagingMethod"),
		PurchasingAgent: db.PurchasingAgent{Name: c.PostForm("PurchasingAgent.Name"), Contact: db.Contact{Email: c.PostForm("PurchasingAgent.Contact.Email"), Phone: c.PostForm("PurchasingAgent.Contact.Phone"), Other: ""}},
		PrefabDirector:  db.PrefabDirector{Name: c.PostForm("PrefabDirector.Name"), Contact: db.Contact{Email: c.PostForm("PrefabDirector.Contact.Email"), Phone: c.PostForm("PrefabDirector.Contact.Phone"), Other: ""}},
		DeliveryReceiver: db.DeliveryReceiver{
			Name: c.PostForm("DeliveryReceiver.Name"),
			Contact: db.Contact{
				Email: c.PostForm("DeliveryReceiver.Contact.Email"),
				Phone: c.PostForm("DeliveryReceiver.Contact.Phone"),
				Other: "",
			},
		},
		LaborTaskCode: c.PostForm("LaborTaskCode"),
		Notes:         c.PostForm("Notes"),
	}
	if err := db.AuditLog(order); err != nil {
		order.Time = "[Generated upon submission]"
		c.HTML(http.StatusOK, "order.tmpl", &orderData{HeaderData: orderHeader, UserMode: db.UserGet(auth.GetLogin(c)).Mode, Order: order, Message: err.Error(), New: true})
		return
	}
	emails := []string{c.PostForm("Foreman.Contact.Email"), c.PostForm("PurchasingAgent.Contact.Email"), c.PostForm("PrefabDirector.Contact.Email"), c.PostForm("DeliveryReceiver.Contact.Email")}
	mail.Send(emails, web.GetTarget(c))
	c.Redirect(http.StatusSeeOther, "/order/"+order.ID)
	return
}

/*
CalculateBOM takes a [assembly]quantity and returns a [part]quantity
*/
func CalculateBOM(assemblies map[string]int) (bom map[string]int) {
	bom = make(map[string]int)
	for assembly, quantity := range assemblies {
		for part, q := range db.AssemblyGet(assembly).Items {
			bom[part] = bom[part] + q*quantity
		}
	}
	return
}
