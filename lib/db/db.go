/*
Package db provides tools for creating and interacting with the SQLite databases.
*/
package db

import (
	"Prefab-Catalog/lib/config"
	"Prefab-Catalog/lib/lumberjack"
	"Prefab-Catalog/lib/web"
	"strconv"

	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/raja/argon2pw"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
CONSTANTS, STRUCTS, AND OTHER VARIABLES
*/

var log = lumberjack.New("DB")
var dbClient *mongo.Client
var db *mongo.Database
var dbContext context.Context
var collections []string = []string{"assemblies", "audit-logs", "orders", "parts", "users"} // TODO: get from config

// DATABASE CACHES

// Assemblies cache
var Assemblies []Assembly

// AssembliesAndNames is a list of assembly IDs and their names.
var AssembliesAndNames map[string]string = make(map[string]string)

// Parts cache
var Parts []Part

// PartsAndNames is a list of part IDs and their names.
var PartsAndNames map[string]string = make(map[string]string)

/*
User describes all the elements of a User.
*/
type User struct {
	ID        string  `bson:"id"`   // The internal ID.
	Mode      int     `bson:"mode"` // The user's mode. 0 = Read (read only), 1 = User (read+order), 2 = Admin (read+order+write), 3+ = Sysadmin (previous plus special rendering)
	Username  string  `bson:"username"`
	Password  string  `bson:"password"`
	FirstName string  `bson:"firstname"`
	LastName  string  `bson:"lastname"`
	Contact   Contact `bson:"contact"`
	Image     string  `bson:"image"`
}

/*
Part describes all the elements of a Part.
*/
type Part struct {
	ID           string `bson:"id"`           // The internal ID.
	Number       string `bson:"number"`       // The part number as provided by the manufacturer, if any.
	Manufacturer string `bson:"manufacturer"` // The manufacturer of the part.
	Name         string `bson:"name"`
	Description  string `bson:"description"`
	Unit         string `bson:"unit"`        // "quantity" if unit of measurement is quantity. "length" if unit of measurement is feet.
	CostPerUnit  int    `bson:"costperunit"` // Cost per unit in cents, if any.
	Image        string `bson:"image"`
}

/*
Assembly describes all the elements of an Assembly.
*/
type Assembly struct {
	ID           string         `bson:"id"`   // The internal ID.
	Name         string         `bson:"name"` // The name of the Assembly.
	Description  string         `bson:"description"`
	BracketStyle string         `bson:"bracketstyle"` // The style of bracket. "Floor Stand", "Single Box Adjustable", "Multi Box Adjustable", "Single Box Nearest Stud"
	RacewayStyle string         `bson:"racewaystyle"` // The style of raceway. "Conduit", "MC cable", "Both"
	System       string         `bson:"system"`       // If the Assembly is part of a larger control system. "Generic", "Lighting Control", "Fire Alarm", "Power"
	Device       bool           `bson:"device"`       // If the Assembly has a device (switch, etc).
	Minutes      int            `bson:"minutes"`      // The estimated amount of time, in minutes, to build the assembly from its constituent parts.
	Items        map[string]int `bson:"items"`        // The parts the assembly contains
	Image        string         `bson:"image"`
}

/*
AuditLogAdmin describes an audit log administrative entry.
*/
type AuditLogAdmin struct {
	Time   string `bson:"time"`   // The time of the event.
	User   string `bson:"user"`   // Which user performed the action (this is a user.ID).
	Action string `bson:"action"` // What kind of action was performed. "update", or "delete"
	Kind   string `bson:"kind"`   // What kind of object was affected. "profile", "part", or "assembly"
	Target string `bson:"target"` // The targeted object ID
}

/*
AuditLogOrder describes an audit log order entry.
*/
type AuditLogOrder struct {
	ID               string           `bson:"id"` // The order number.
	Time             string           `bson:"time"`
	User             string           `bson:"user"`  // Requesting user.
	Items            map[string]int   `bson:"items"` // The assemblies which were ordered. item:quantity
	Project          Project          `bson:"project"`
	PackagingMethod  string           `bson:"packagingmethod"`
	PurchasingAgent  PurchasingAgent  `bson:"purchasingagent"`
	PrefabDirector   PrefabDirector   `bson:"prefabdirector"`
	DeliveryReceiver DeliveryReceiver `bson:"deliveryreceiver"`
	LaborTaskCode    string           `bson:"labortaskcode"`
	Notes            string           `bson:"notes"` // Notes about the request.
}

// Project is a child of AuditLogOrder.
type Project struct {
	Number           string  `bson:"number"`
	Name             string  `bson:"name"`
	ShippingAddress  string  `bson:"shippingaddress"`
	DeliveryLocation string  `bson:"deliverylocaton"` // Where on the project site the materials are to be delivered
	DeliveryDate     string  `bson:"deliverydate"`
	Foreman          Foreman `bson:"foreman"`
}

// Foreman is a child of AuditLogOrder.Project.
type Foreman struct { // In nearly all cases, the foreman is the user requesting the order -- so this information will be autofilled when a new order request is created.
	Name    string  `bson:"name"`
	Contact Contact `bson:"contact"`
}

// PurchasingAgent is a child of AuditLogOrder.
type PurchasingAgent struct {
	Name    string  `bson:"name"`
	Contact Contact `bson:"contact"`
}

// PrefabDirector is a child of AuditLogOrder.
type PrefabDirector struct {
	Name    string  `bson:"name"`
	Contact Contact `bson:"contact"`
}

// DeliveryReceiver is a child of AuditLogOrder.
type DeliveryReceiver struct {
	Name    string  `bson:"name"`
	Contact Contact `bson:"contact"`
}

// Contact is a struct containing various contact info fields.
type Contact struct {
	Email string `bson:"email"`
	Phone string `bson:"phone"`
	Other string `bson:"other"`
}

/*
MISC DB FUNCTIONS
*/

/*
Cache loads the Assemblies and Parts databases into memory so they can be served as quickly as possible.
*/
func Cache() {
	Assemblies = AssemblyGetAll(true)
	AssemblyGetName("cache")
	Parts = PartGetAll(true)
	PartGetName("cache")
}

/*
TouchBase creates all databases used by the server if they do not exist, along with the default SysAdmin user.
It is required to be run in order to connect to the database.
Its name is a play on the GNU program "touch", the idiom "[to] touch base", and the word "database". The author is rather proud of this.
*/
func TouchBase(dbURI string, dbTimeout int) {
	dbContext = context.TODO()
	dbClient, _ = mongo.Connect(dbContext, options.Client().ApplyURI(dbURI))
	db = dbClient.Database("prefab-catalog")
	for _, collectionName := range collections { // Create all collections if they do not exist.
		if createError := db.CreateCollection(dbContext, collectionName); createError != nil {
			log.Debugf("Skipping creation of database collection %q as it already exists.", collectionName)
		} else {
			log.Warnf("Database collection %q created.", collectionName)
		}
	}
	// root violates the laws of causality by creating itself if it does not exist :)
	user := UserGet("00000000-0000-0000-0000-000000000000")
	if user.ID == "" {
		result := UserSet("00000000-0000-0000-0000-000000000000", &User{ID: "00000000-0000-0000-0000-000000000000", Mode: 3, Username: "root", Password: "root", FirstName: "JC", LastName: "Denton", Contact: Contact{Email: "admin@pkre.co"}})
		if result != nil {
			log.Fatal(result, "Unable to create default system administrator account.")
		} else {
			log.Warn("Default system administrator account created. Username: \"root\". Password: \"root\".")
			log.Warn("IMPORTANT: Change the password for this account before making your production server public!!!")
		}
	}
	Cache() // Cache databases
	return
}

/*
EncryptPassword encrypts the given plaintext. Returns a string. FATAL if this fails for some reason.
*/
func EncryptPassword(password string) string {
	hashedPassword, err := argon2pw.GenerateSaltedHash(password)
	if err != nil {
		log.Error(err, "Password hash failed?!")
	}
	return hashedPassword
}

/*
UUID is a shorthand for uuid.New().String().
*/
func UUID() string {
	return uuid.New().String()
}

/*
USER FUNCTIONS
*/

/*
UserDelete deletes a user from the database. Returns nil. Returns error if something goes wrong.
*/
func UserDelete(callerID string, id string) error {
	// Users can only delete users with lower permission.
	if UserGet(callerID).Mode <= UserGet(id).Mode {
		log.Warnf("User %s attempted to delete user %s but had insufficient permission -- this should not be possible via profile.tmpl as the delete button should be disabled.", callerID, id)
		return errors.New("insufficient permission")
	}
	if _, err := db.Collection("users").DeleteOne(dbContext, bson.M{"id": id}); err != nil {
		log.Error(err, "database delete error")
		return err
	}
	log.Debugf("User %s deleted user %s", callerID, id)
	AuditLog(&AuditLogAdmin{Time: time.Now().Format("2006-01-02 15:04:05.000000"), User: callerID, Action: "delete", Kind: "profile", Target: id})
	return nil
}

/*
UserGet takes a user's ID or Username returns the user's information as a User struct. Returns nil if the user could not be found.
User.password is the hashed password.
*/
func UserGet(usernameOrID string) (user User) {
	if err := db.Collection("users").FindOne(dbContext, bson.M{"$or": []interface{}{bson.M{"username": usernameOrID}, bson.M{"id": usernameOrID}}}).Decode(&user); err != nil {
		log.Debugf("Failed to get user data for %q: %s", usernameOrID, err.Error())
		return User{}
	}
	return
}

/*
UserGetAll returns all users.
*/
func UserGetAll() (users []User) {
	cursor, err := db.Collection("users").Find(dbContext, bson.M{})
	if err != nil {
		log.Error(err, "database read error")
		return nil
	}
	if err := cursor.All(dbContext, &users); err != nil {
		log.Error(err, "database read error")
		return nil
	}
	return
}

/*
UserLogin returns true if the username and password exist in the users database. Otherwise returns false.
*/
func UserLogin(username, password string) (loggedIn bool) {
	if password == "" || username == "" {
		log.Debugf("Failed to log in user %q: %s", username, "username or password is null")
		return false
	}
	var user User
	if err := db.Collection("users").FindOne(dbContext, bson.M{"username": username}).Decode(&user); err != nil {
		log.Debugf("Failed to log in user %q: %s", username, err.Error())
		return false
	}
	valid, err := argon2pw.CompareHashWithPassword(user.Password, password)
	if err != nil {
		log.Error(err, "password hash failure")
	}
	if !valid {
		log.Debugf("Failed to log in user %q: %s", username, "password mismatch.")
		return false
	}
	log.Debugf("Logged in user %q.", username)
	return true
}

/*
UserSet modifies an existing user account or creates one if it does not exist. Returns nil. Returns error if something goes wrong. "Password" is not modified if it is blank.
*/
func UserSet(callerID string, user *User) error {
	// Username can not be set to ID, as that may break UserGet(). Better safe than sorry.
	if user.ID == user.Username {
		return errors.New("username cannot be the same as ID")
	}
	if user.Username == "" {
		return errors.New("username cannot be null")
	}
	// If target user currently exists...
	if existingUser := UserGet(user.ID); existingUser.ID != "" {
		// Don't overwrite an existing image with null.
		if user.Image == "" && existingUser.Image != "" {
			user.Image = existingUser.Image
		}
		// Forbid a user from changing their own Mode.
		if callerID == user.ID {
			user.Mode = existingUser.Mode
		} else { // Forbid a user from modifying a different user with an equal or higher Mode.
			if UserGet(callerID).Mode <= existingUser.Mode {
				return errors.New("insufficient permission")
			}
		}
	} else if user.Image == "" {
		// If no image is provided, use the "none" image.
		user.Image = "/media/none.webp"
	}
	upsert := true
	opt := options.UpdateOptions{Upsert: &upsert}
	if _, err := db.Collection("users").UpdateOne(dbContext, bson.M{"id": user.ID}, bson.M{"$set": bson.M{"id": user.ID, "mode": user.Mode, "username": user.Username, "firstname": user.FirstName, "lastname": user.LastName, "contact": user.Contact, "image": user.Image}}, &opt); err != nil {
		return err
	}
	// Only update password if a new value was set.
	if user.Password != "" {
		password := EncryptPassword(user.Password)
		if _, err := db.Collection("users").UpdateOne(dbContext, bson.M{"id": user.ID}, bson.M{"$set": bson.M{"password": password}}, &opt); err != nil { // If we were to write "User{Password: password}", all of the other struct fields would be set to "", deleting the rest of the entry.
			return err
		}
	}
	AuditLog(&AuditLogAdmin{Time: time.Now().Format("2006-01-02 15:04:05.000000"), User: callerID, Action: "update", Kind: "profile", Target: user.ID})
	return nil
}

/*
SYSADMIN UTILITY FUNCTIONS
*/

// TODO: MongoDB Charts

/*
CATALOG FUNCTIONS
*/

/*
AssemblyDelete deletes an assembly from the database. Returns nil. Returns error if something goes wrong.
*/
func AssemblyDelete(callerID string, ID string) error {
	if _, err := db.Collection("assemblies").DeleteOne(dbContext, bson.M{"id": ID}); err != nil {
		log.Error(err, "database delete error")
		return err
	}
	return nil
}

/*
AssemblyGet returns the Assembly{} specified by the ID. Returns nil if no such assembly exists.
*/
func AssemblyGet(ID string) (result Assembly) {
	for _, assembly := range Assemblies {
		if ID == assembly.ID {
			return assembly
		}
	}
	if err := db.Collection("assemblies").FindOne(dbContext, bson.M{"id": ID}).Decode(&result); err != nil {
		log.Debugf("Failed to get assembly data for %q: %s", ID, err.Error())
		return Assembly{}
	}
	Assemblies = append(Assemblies, result)
	return
}

/*
AssemblyGetAll returns all assemblies.
*/
func AssemblyGetAll(fromDB bool) (results []Assembly) {
	if fromDB {
		log.Debugf("Loading assemblies from database.")
		cursor, err := db.Collection("assemblies").Find(dbContext, bson.M{})
		if err != nil {
			log.Error(err, "database read error")
			return nil
		}
		if err := cursor.All(dbContext, &results); err != nil {
			log.Error(err, "database read error")
			return nil
		}
	} else {
		return Assemblies
	}
	return
}

/*
AssemblyGetName is a utility function for Order to easily get an assembly's name from its ID.
*/
func AssemblyGetName(id string) (name string) {
	// Load cache of assemblies and names
	if id == "cache" {
		assemblies := AssemblyGetAll(false)
		for _, assembly := range assemblies {
			AssembliesAndNames[assembly.ID] = assembly.Name
		}
		return
	}
	name = AssembliesAndNames[id]
	if name == "" {
		name = AssemblyGet(id).Name
		if name == "" {
			return "NAME UNKNOWN"
		}
	}
	AssembliesAndNames[id] = name
	return
}

/*
AssemblySet modifies and existing assembly or creates one if it does not exist. Returns nil. Returns error if something goes wrong.
*/
func AssemblySet(callerID string, assembly Assembly) error {
	if assembly.Name == "" {
		return errors.New("name cannot be null")
	}
	if existingAssembly := AssemblyGet(assembly.ID); existingAssembly.ID != "" {
		if assembly.Image == "" && existingAssembly.Image != "" {
			assembly.Image = existingAssembly.Image
		}
	}
	if assembly.Image == "" {
		assembly.Image = "/media/none.webp"
	}
	upsert := true
	opt := options.UpdateOptions{Upsert: &upsert}
	if _, err := db.Collection("assemblies").UpdateOne(dbContext, bson.M{"id": assembly.ID}, bson.M{"$set": bson.M{"id": assembly.ID, "name": assembly.Name, "description": assembly.Description, "bracketstyle": assembly.BracketStyle, "racewaystyle": assembly.RacewayStyle, "system": assembly.System, "device": assembly.Device, "minutes": assembly.Minutes, "items": assembly.Items, "image": assembly.Image}}, &opt); err != nil {
		return err
	}
	log.Debugf("User %s set assembly %s", callerID, assembly.ID)
	for _, originalAssembly := range Assemblies {
		if originalAssembly.ID == assembly.ID {
			originalAssembly = assembly
		}
	}
	AuditLog(&AuditLogAdmin{Time: time.Now().Format("2006-01-02 15:04:05.000000"), User: callerID, Action: "update", Kind: "assembly", Target: assembly.ID})
	return nil
}

/*
PartDelete deletes a part from the database. Returns nil. Returns error if something goes wrong.
*/
func PartDelete(callerID string, ID string) error {
	if _, err := db.Collection("parts").DeleteOne(dbContext, bson.M{"id": ID}); err != nil {
		log.Error(err, "database delete error")
		return err
	}
	log.Debugf("User %s deleted part %s", callerID, ID)
	AuditLog(&AuditLogAdmin{Time: time.Now().Format("2006-01-02 15:04:05.000000"), User: callerID, Action: "delete", Kind: "part", Target: ID})
	return nil
}

/*
PartGet returns the Part{} specified by the ID. Returns nil if no such part exists.
*/
func PartGet(ID string) (result Part) {
	for _, part := range Parts {
		if ID == part.ID {
			return part
		}
	}
	if err := db.Collection("parts").FindOne(dbContext, bson.M{"id": ID}).Decode(&result); err != nil {
		log.Debugf("Failed to get part data for %q: %s", ID, err.Error())
		return Part{}
	}
	Parts = append(Parts, result)
	return
}

/*
PartGetAll returns all parts.
*/
func PartGetAll(fromDB bool) (results []Part) {
	if fromDB {
		log.Debugf("Loading parts from database.")
		cursor, err := db.Collection("parts").Find(dbContext, bson.M{})
		if err != nil {
			log.Error(err, "database read error")
			return nil
		}
		if err := cursor.All(dbContext, &results); err != nil {
			log.Error(err, "database read error")
			return nil
		}
	} else {
		return Parts
	}
	return
}

/*
PartGetName is a utility function for AssemblyList to easily get a part's name from its ID.
*/
func PartGetName(id string) (name string) {
	// Load cache of parts and names
	if id == "cache" {
		parts := PartGetAll(false)
		for _, part := range parts {
			PartsAndNames[part.ID] = part.Name
		}
		return
	}
	name = PartsAndNames[id]
	if name == "" {
		name = PartGet(id).Name
		if name == "" {
			return "NAME UNKNOWN"
		}
	}
	PartsAndNames[id] = name
	return
}

/*
PartSet modifies an existing part or creates one if it does not exist. Returns nil. Returns error if something goes wrong.
*/
func PartSet(callerID string, part Part) error {
	if part.Name == "" {
		return errors.New("name cannot be null")
	}
	if existingPart := PartGet(part.ID); existingPart.ID != "" {
		if part.Image == "" && existingPart.Image != "" {
			part.Image = existingPart.Image
		}
	}
	if part.Image == "" {
		part.Image = "/media/none.webp"
	}
	upsert := true
	opt := options.UpdateOptions{Upsert: &upsert}
	if _, err := db.Collection("parts").UpdateOne(dbContext, bson.M{"id": part.ID}, bson.M{"$set": bson.M{"id": part.ID, "number": part.Number, "manufacturer": part.Manufacturer, "name": part.Name, "description": part.Description, "unit": part.Unit, "costperunit": part.CostPerUnit, "image": part.Image}}, &opt); err != nil {
		return err
	}
	PartsAndNames[part.ID] = part.Name // Update caches.
	for _, originalPart := range Parts {
		if originalPart.ID == part.ID {
			originalPart = part
		}
	}
	AuditLog(&AuditLogAdmin{Time: time.Now().Format("2006-01-02 15:04:05.000000"), User: callerID, Action: "update", Kind: "part", Target: part.ID})
	return nil
}

/*
AUDIT LOG FUNCTIONS
*/

/*
AuditLog takes AuditLogAdmin{} or AuditLogOrder{}. Returns nil. Returns error if an error occurs.
AuditLog is for recording modifications to the databases -- not for general system output. That is what lib/log is for!
*/
func AuditLog(auditLog interface{}) error {
	switch entry := auditLog.(type) {
	case AuditLogAdmin:
		if _, err := db.Collection("audit-logs").InsertOne(dbContext, entry); err != nil {
			return err
		}
	case AuditLogOrder:
		// Check for mandatory fields.
		// TODO: Get "friendly name" via struct tags!
		if entry.Project.Number == "" {
			return errors.New("project number required")
		} else if entry.Project.Name == "" {
			return errors.New("project name required")
		} else if entry.Project.ShippingAddress == "" {
			return errors.New("shipping address required")
		} else if entry.Project.DeliveryLocation == "" {
			return errors.New("delivery location required")
		} else if entry.Project.Foreman.Name == "" {
			return errors.New("foreman name required")
		} else if entry.Project.Foreman.Contact.Email == "" {
			return errors.New("foreman email required")
		} else if entry.PrefabDirector.Name == "" {
			return errors.New("prefab director name required")
		} else if entry.PrefabDirector.Contact.Email == "" {
			return errors.New("prefab director email required")
		} else if entry.PurchasingAgent.Name == "" {
			return errors.New("purchasing agent name required")
		} else if entry.PurchasingAgent.Contact.Email == "" {
			return errors.New("purchasing agent email required")
		} else if entry.DeliveryReceiver.Name == "" {
			return errors.New("delivery receiver name required")
		} else if entry.DeliveryReceiver.Contact.Email == "" {
			return errors.New("delivery receiver email required")
		} else if entry.PackagingMethod == "" {
			return errors.New("packaging method required")
		}
		// Check important fields.
		if len(entry.Items) == 0 {
			log.Warnf("Order %s contains no items, attempting to save anyway.") // TODO: Present a front-end warning requiring confirmation when saving an order with zero items.
		}
		// Finish.
		if _, err := db.Collection("orders").InsertOne(dbContext, entry); err != nil {
			return err
		}
	default:
		err := errors.New("function AuditLog only accepts arguments of type AuditLogAdmin or AuditLogOrder")
		log.Error(err, "improper type of argument passed to AuditLog")
		return err
	}
	return nil
}

/*
AuditLogAdminGetAll returns all administrative actions in the AuditLog.
*/
func AuditLogAdminGetAll() (results []AuditLogAdmin) {
	cursor, err := db.Collection("audit-logs").Find(dbContext, bson.M{})
	if err != nil {
		log.Error(err, "database read error")
		return nil
	}
	if err := cursor.All(dbContext, &results); err != nil {
		log.Error(err, "database read error")
		return nil
	}
	return
}

/*
AuditLogOrderGet returns the AuditLogOrder{} specified by order number. Returns nil if no such order is found.
*/
func AuditLogOrderGet(ID string) (result AuditLogOrder) {
	if err := db.Collection("orders").FindOne(dbContext, bson.M{"id": ID}).Decode(&result); err != nil {
		log.Debugf("Failed to get order data for %q: %s", ID, err.Error())
		return AuditLogOrder{}
	}
	return
}

/*
AuditLogOrderGetAll returns all orders in the AuditLog.
*/
func AuditLogOrderGetAll() (results []AuditLogOrder) {
	cursor, err := db.Collection("orders").Find(dbContext, bson.M{})
	if err != nil {
		log.Error(err, "database read error")
		return nil
	}
	if err := cursor.All(dbContext, &results); err != nil {
		log.Error(err, "database read error")
		return nil
	}
	return
}

/*
AuditLogOrderIncrementID returns the next order ID.
*/
func AuditLogOrderIncrementID() (id string) {
	var order AuditLogOrder
	opt := &options.FindOneOptions{Sort: bson.M{"id": -1}}
	db.Collection("orders").FindOne(dbContext, bson.M{}, opt).Decode(&order)
	log.Debugf("Last order ID was %s, incrementing by one.", order.ID)
	if order.ID == "" {
		return "0"
	}
	i, _ := strconv.Atoi(order.ID)
	return strconv.Itoa(i + 1)
}

/*
MEDIA FUNCTIONS
*/

/*
ImageSet saves the image in the appropriate location. Context needs to have web.SetTarget() updated before this.
*/
func ImageSet(c *gin.Context, image *multipart.FileHeader) (url string) {
	path := config.Load().UploadPath + web.GetTarget(c)
	if string(path[0]) == "." {
		url = path[1:]
	}
	if err := c.SaveUploadedFile(image, path); err != nil {
		log.Error(err, "failed to save an image")
	} else {
		log.Debugf("Saved image %q", path)
	}
	return
}
