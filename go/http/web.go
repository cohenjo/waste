package http

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/cohenjo/waste/go/types"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func Serve() {
	router := gin.Default()
	router.LoadHTMLGlob("resources/templates/*")

	initializeRoutes(router)
	router.GET("/", showIndexPage)

	router.Run()
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
	queueRoutes := router.Group("/queue")
	{
		queueRoutes.GET("/needreview", needReview)
		queueRoutes.GET("/reviewed", reviewed)
		// userRoutes.POST("/login", performLogin)
		queueRoutes.GET("/scheduled", scheduled)
		queueRoutes.GET("/completed", completed)
		// userRoutes.POST("/register", register)
	}
	router.GET("/cluster/view/:cluster_id", getCluster)
}

func needReview(c *gin.Context) {
	var clusters []cluster
	clusters = getAllClusters()
	render(c, gin.H{
		"title":   "Need Review",
		"payload": clusters}, "index.html")
}
func reviewed(c *gin.Context) {
	var clusters []cluster
	clusters = getAllClusters()
	render(c, gin.H{
		"title":   "Reviewed",
		"payload": clusters}, "index.html")
}
func scheduled(c *gin.Context) {
	var clusters []cluster
	clusters = getAllClusters()
	render(c, gin.H{
		"title":   "scheduled",
		"payload": clusters}, "index.html")
}
func completed(c *gin.Context) {
	var clusters []cluster
	clusters = getAllClusters()
	render(c, gin.H{
		"title":   "completed",
		"payload": clusters}, "index.html")
}

func showIndexPage(c *gin.Context) {
	var clusters []cluster
	clusters = getAllClusters()
	render(c, gin.H{
		"title":   "Home Page",
		"payload": clusters}, "index.html")
}

func getCluster(c *gin.Context) {
	// Check if the article ID is valid
	if articleID, err := strconv.Atoi(c.Param("cluster_id")); err == nil {
		// Check if the article exists
		if cluster, err := getClusterByID(articleID); err == nil {
			// Call the HTML method of the Context to render a template
			render(c, gin.H{
				"title":   cluster.Name,
				"payload": cluster}, "cluster.html")

		} else {
			// If the article is not found, abort with an error
			c.AbortWithError(http.StatusNotFound, err)
		}

	} else {
		// If an invalid article ID is specified in the URL, abort with an error
		c.AbortWithStatus(http.StatusNotFound)
	}
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
