package mutators

import (
	"fmt"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// consider adding: _ "net/http/pprof"?

func CreateChangeEndpoint(c *gin.Context) {
	ID, _ := strconv.Atoi(c.Param("id"))
	var change Change
	c.ShouldBind(&change)
	fmt.Printf("%d, %s\n", ID, change)
	changes = append(changes, change)

	// c1 := make(chan string)
	// c2 := make(chan *Server)
	// GetArtifactCluster(c1, change.Artifact)
	// clusterID := <-c1
	// GetArtifactServerDuo(c2, clusterID)

	// log.Info("received clusterId: ", clusterID)
	// masterHost := <-c2
	// log.Info("received master host: ", masterHost)

	change.RunChange()

	c.JSON(http.StatusOK, gin.H{"changes": changes})
}

var changes []Change

// StartWebServer starts the web API
func StartWebServer() {

	router := gin.Default()
	// people = append(people, Person{ID: "1", Firstname: "Nic", Lastname: "Raboy", Address: &Address{City: "Dublin", State: "CA"}})
	// people = append(people, Person{ID: "2", Firstname: "Maria", Lastname: "Raboy"})
	// router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	// router.HandleFunc("/people/{id}", GetPersonEndpoint).Methods("GET")
	router.POST("/change/:id", CreateChangeEndpoint)
	// router.HandleFunc("/people/{id}", DeletePersonEndpoint).Methods("DELETE")

	router.Run()

}
