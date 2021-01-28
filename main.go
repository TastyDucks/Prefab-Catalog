package main

import (
	"Prefab-Catalog/lib/config"
	"Prefab-Catalog/lib/db"
	"Prefab-Catalog/lib/lumberjack"
	"Prefab-Catalog/lib/web"
	"Prefab-Catalog/routes"
	"encoding/gob"
	"html/template"

	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	nice "github.com/ekyoung/gin-nice-recovery"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var configuration config.YAML
var router *gin.Engine

var log *lumberjack.Lumberjack

func main() {
	configuration = config.Load() // TODO: errors that happen during config.Load() aren't saved to the proper LogPath -- because that might not be loaded yet.
	lumberjack.Start(configuration.LogPath, configuration.Verbosity)
	log = lumberjack.New("Main")
	db.TouchBase(configuration.DatabaseURI, configuration.DatabaseTimeout) // Set up databases and cache things
	gob.Register(web.Items{})
	go start(configuration.Port)
	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Warn("Prefab Catalog shutting down.")
}

func start(port int) {
	if port <= 0 || port > 65535 {
		log.Warn("Port specified out of valid range (1-65535), set to 80.")
		port = 80
	}
	address := fmt.Sprintf(":%d", port)
	if configuration.Verbosity < 1 {
		gin.SetMode(gin.ReleaseMode)
	}

	// TODO BEGIN: Integrate Gin log messages under their appropriate logLevels with *lumberjack*

	ginLogFile, errOpen := os.OpenFile(configuration.LogPath+"access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errOpen != nil {
		fmt.Fprintf(os.Stderr, "Unable to create log file: %s", errOpen.Error())
		os.Exit(1)
	}
	defer ginLogFile.Close()
	outlog := io.MultiWriter(ginLogFile, os.Stdout)
	errlog := io.MultiWriter(ginLogFile, os.Stderr)
	gin.DefaultWriter = outlog
	gin.DefaultErrorWriter = errlog

	// TODO END

	router = gin.New()
	router.Use(gin.Logger())
	router.Use(gzip.Gzip(gzip.BestCompression)) // Gzip compression
	router.RedirectTrailingSlash = true         // TODO: This doesn't seem to be working. Not critical but would be good to have. Fix with NGNIX in front of go? Would be good for production, mitigating DDOS, spambots, etc
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("session", store))
	router.Static("/css", "./static/css") // TODO: Make URL access to /css/, /js/, /media/, /templates/, and /upload/ use the proper NoRoute() handler and NOT http.404, as it currently just returns a blank page.
	router.Static("/js", "./static/js")
	router.Static("/media", "./static/media")
	router.Static("/templates", "./static/templates")
	router.Static("/upload", "./upload")
	router.SetFuncMap(template.FuncMap{"AssemblyGetName": db.AssemblyGetName, "PartGetName": db.PartGetName, "CalculateBOM": routes.CalculateBOM}) // Functions that may be called from within templates. This needs to be run before router.LoadHTMLGlob()
	router.LoadHTMLGlob("./static/templates/*")                                                                                                    // Load templates.
	// TODO: Dynamically load routes from files in "/routes/" instead of hard-coding them.
	router.NoRoute(routes.NotFound)                       // 404.
	router.NoMethod(routes.MethodNotAllowed)              // 405.
	router.Use(nice.Recovery(routes.InternalServerError)) // 500.
	// Admin
	router.GET("/admin", routes.Admin)
	router.GET("/adminAuditLog", routes.AdminAuditLog)
	router.GET("/adminMailLog", routes.AdminMailLog)
	router.GET("/adminUserList", routes.AdminUserList)
	// User
	router.GET("/profile/*id", routes.Profile)
	router.POST("/profile/", routes.ProfilePOST)
	router.GET("/logout", routes.Logout)
	// Catalog
	router.GET("/", routes.Index)
	router.HEAD("/", routes.Index) // For status checks (e.g. via "curl -I URL").
	router.POST("/", routes.Login)
	router.GET("/assembly/*id", routes.Assembly)
	router.POST("/assembly", routes.AssemblyPOST)
	router.GET("/part/*id", routes.Part)
	router.POST("/part/", routes.PartPOST)
	router.GET("/order/*id", routes.Order)
	router.POST("/order", routes.OrderPOST)             // Order review page
	router.POST("/orderFinish", routes.OrderFinishPOST) // After order is completed.
	// Statistics
	router.GET("/stats")
	build := config.Build()
	log.Infof("Prefab Catalog started. Version: %s", build)

	log.Fatal(router.Run(address), "Main thread error!")
}
