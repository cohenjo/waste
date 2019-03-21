package http

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"

	"github.com/cohenjo/waste/go/scheduler"

	"github.com/cohenjo/waste/go/config"
	"github.com/cohenjo/waste/go/logic"
	"github.com/cohenjo/waste/go/mutators"
	"github.com/cohenjo/waste/go/types"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var router *gin.Engine

// @title WASTE Swagger docs
// @version 1.0
// @description This is waste server
// @termsOfService http://swagger.io/terms/

// @contact.name cohenjo
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host waste.cohenjo.io
// @BasePath /v1

var (
	rxURL = regexp.MustCompile(`^/regexp\d*`)
)

// Serve is the main entry point to start serving the web api.
func Serve() {

	router := gin.Default()
	router.LoadHTMLGlob("resources/templates/*")

	initializeRoutes(router)
	// router.GET("/", showIndexPage)
	if !config.Config.Debug { // change this when rolling to prod
		gin.SetMode(gin.ReleaseMode)
	}
	router.Use(logger.SetLogger())

	router.Use(logger.SetLogger(logger.Config{
		Logger:         &log.Logger,
		UTC:            true,
		SkipPath:       []string{"/skip"},
		SkipPathRegexp: rxURL,
	}))

	router.Run("127.0.0.1:8080")
}

func initializeRoutes(router *gin.Engine) {

	// Handle the index route
	// router.GET("/", showIndexPage)
	userRoutes := router.Group("/u")
	{
		userRoutes.GET("/login", showLoginPage)
		userRoutes.POST("/login", performLogin)
		userRoutes.GET("/logout", logout)
		userRoutes.GET("/register", showRegistrationPage)
		userRoutes.POST("/register", register)
	}
	// queueRoutes := router.Group("/queue")
	// {
	// 	queueRoutes.GET("/needreview", needReview)
	// 	queueRoutes.GET("/reviewed", reviewed)
	// 	// userRoutes.POST("/login", performLogin)
	// 	queueRoutes.GET("/scheduled", scheduled)
	// 	queueRoutes.GET("/completed", completed)
	// 	// userRoutes.POST("/register", register)
	// }
	// router.GET("/v1/cluster/view/:cluster_id", getCluster)
	router.GET("/change", getChange)
	router.POST("/change", createChangeEndpoint)

	router.GET("/tasks", getTasks)
}

// swagger:route POST /change change
//
// Changes a table
//
// This will show all available pets by default.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       default: Change
//       200: Change
func createChangeEndpoint(c *gin.Context) {
	var change mutators.Change
	c.ShouldBind(&change)
	fmt.Printf(" %+v\n", change)
	// change.RunChange()
	_ = logic.CM.MangeChange(change)

	c.JSON(http.StatusOK, gin.H{"change": change})
}

// func needReview(c *gin.Context) {
// 	var clusters []cluster
// 	clusters = getAllClusters()
// 	render(c, gin.H{
// 		"title":   "Need Review",
// 		"payload": clusters}, "index.html")
// }
// func reviewed(c *gin.Context) {
// 	var clusters []cluster
// 	clusters = getAllClusters()
// 	render(c, gin.H{
// 		"title":   "Reviewed",
// 		"payload": clusters}, "index.html")
// }
// func scheduled(c *gin.Context) {
// 	var clusters []cluster
// 	clusters = getAllClusters()
// 	render(c, gin.H{
// 		"title":   "scheduled",
// 		"payload": clusters}, "index.html")
// }
// func completed(c *gin.Context) {
// 	var clusters []cluster
// 	clusters = getAllClusters()
// 	render(c, gin.H{
// 		"title":   "completed",
// 		"payload": clusters}, "index.html")
// }

// func showIndexPage(c *gin.Context) {
// 	var clusters []cluster
// 	clusters = getAllClusters()
// 	render(c, gin.H{
// 		"title":   "Home Page",
// 		"payload": clusters}, "index.html")
// }

// func getCluster(c *gin.Context) {
// 	// Check if the article ID is valid
// 	if articleID, err := strconv.Atoi(c.Param("cluster_id")); err == nil {
// 		// Check if the article exists
// 		if cluster, err := getClusterByID(articleID); err == nil {
// 			// Call the HTML method of the Context to render a template
// 			render(c, gin.H{
// 				"title":   cluster.Name,
// 				"payload": cluster}, "cluster.html")

// 		} else {
// 			// If the article is not found, abort with an error
// 			c.AbortWithError(http.StatusNotFound, err)
// 		}

// 	} else {
// 		// If an invalid article ID is specified in the URL, abort with an error
// 		c.AbortWithStatus(http.StatusNotFound)
// 	}
// }

func getChange(c *gin.Context) {
	// Check if the article ID is valid
	render(c, gin.H{"title": "change"}, "change.html")

}

func showRegistrationPage(c *gin.Context) {
	render(c, gin.H{
		"title": "Register"}, "register.html")
}

func showLoginPage(c *gin.Context) {
	render(c, gin.H{
		"title": "Login",
	}, "login.html")
}

func performLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if types.IsUserValid(username, password) {
		token := generateSessionToken()
		c.SetCookie("token", token, 3600, "", "", false, true)

		render(c, gin.H{
			"title": "Successful Login"}, "login-successful.html")

	} else {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "Invalid credentials provided"})
	}
}

func logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "", "", false, true)

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func generateSessionToken() string {
	// We're using a random 16 character string as the session token
	// This is NOT a secure way of generating session tokens
	// DO NOT USE THIS IN PRODUCTION
	return strconv.FormatInt(rand.Int63(), 16)
}

func register(c *gin.Context) {
	// Obtain the POSTed username and password values
	username := c.PostForm("username")
	password := c.PostForm("password")

	if _, err := types.RegisterNewUser(username, password); err == nil {
		// If the user is created, set the token in a cookie and log the user in
		token := generateSessionToken()
		c.SetCookie("token", token, 3600, "", "", false, true)
		c.Set("is_logged_in", true)

		render(c, gin.H{
			"title": "Successful registration & Login"}, "login-successful.html")

	} else {
		// If the username/password combination is invalid,
		// show the error message on the login page
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"ErrorTitle":   "Registration Failed",
			"ErrorMessage": err.Error()})

	}
}

func render(c *gin.Context, data gin.H, templateName string) {

	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		// Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}

}

func getTasks(c *gin.Context) {
	// change.RunChange()
	tasks, _ := scheduler.WS.GetTasks()

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
