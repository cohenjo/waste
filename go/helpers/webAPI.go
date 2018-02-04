package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// consider adding: _ "net/http/pprof"?

var Config CLIOptions

func CreateChangeEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var change Change
	_ = json.NewDecoder(req.Body).Decode(&change)
	ID := params["id"]
	fmt.Printf("%s, %s\n", ID, change)
	changes = append(changes, change)

	c1 := make(chan string)
	c2 := make(chan *Server)
	GetArtifactCluster(c1, change.Artifact)
	clusterID := <-c1
	GetArtifactServerDuo(c2, clusterID)

	fmt.Println("received clusterId: ", clusterID)
	masterHost := <-c2
	fmt.Println("received master host: ", masterHost)

	change.RunChange(masterHost)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(changes); err != nil {
		panic(err)
	}
}

var changes []Change

// StartWebServer starts the web API
func StartWebServer() {
	router := mux.NewRouter()
	// people = append(people, Person{ID: "1", Firstname: "Nic", Lastname: "Raboy", Address: &Address{City: "Dublin", State: "CA"}})
	// people = append(people, Person{ID: "2", Firstname: "Maria", Lastname: "Raboy"})
	// router.HandleFunc("/people", GetPeopleEndpoint).Methods("GET")
	// router.HandleFunc("/people/{id}", GetPersonEndpoint).Methods("GET")
	router.HandleFunc("/change/{id}", CreateChangeEndpoint).Methods("POST")
	// router.HandleFunc("/people/{id}", DeletePersonEndpoint).Methods("DELETE")
	log.Fatal(http.ListenAndServe(Config.WebAddress, router))
}
