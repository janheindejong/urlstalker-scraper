/*
Runs webscraper
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// TODO Add e-mail functionality
// TODO containerize

func main() {

	host := os.Args[1]

	db := UrlStalkerDbApi{host: host}

	// Get resources
	resources, err := db.GetResources()
	if err != nil {
		log.Fatal(err)
	}

	// Inspect each resource and save a snapshot to the DB if changed
	var wg sync.WaitGroup
	for _, resource := range *resources {
		wg.Add(1)
		resource := resource
		go SnapAndSaveIfChanged(&wg, &resource, &db)
	}
	wg.Wait()
}

func SnapAndSaveIfChanged(wg *sync.WaitGroup, resource *Resource, db *UrlStalkerDbApi) {
	defer wg.Done()

	// Create a new snapshot of the resource
	newSnapShot, err := resource.Snap()
	if err != nil {
		log.Println(err)
		return
	}

	// Append snapshot and post if has changed
	if newSnapShot.ResponseBody != resource.MostRecentSnapShot().ResponseBody {
		resource.Snapshots = append(resource.Snapshots, *newSnapShot)
		err = db.SaveSnapShot(newSnapShot)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Printf(`Resource with path "%v" has not changed`, resource.Path)
	}

}

type UrlStalkerDbApi struct {
	// UrlStalkerDbApi is used for communicating with the URL stalker
	// database (currently implemented as a REST API)
	host string
}

func (api UrlStalkerDbApi) GetResources() (*[]Resource, error) {
	// Gets all the resources in the database
	log.Printf(`Getting resources from database at "%v"`, api.host)

	// Make HTTP call to API
	resp, err := http.Get(api.host + "/resource")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Parse response
	var resources []Resource
	err = json.NewDecoder(resp.Body).Decode(&resources)
	return &resources, err
}

func (api UrlStalkerDbApi) SaveSnapShot(snapshot *SnapShot) error {
	// Post a new snapshot for a given resource
	log.Printf(`Posting snapshot for resource with path "%v"`, snapshot.Resource.Path)
	url := api.host + fmt.Sprintf("/resource/%v/snapshot", snapshot.Resource.Id)
	jsonValue, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	return err
}

type Resource struct {
	Path      string     `json:"path"`
	Id        int        `json:"id"`
	Snapshots []SnapShot `json:"snapshots"`
}

func (resource Resource) Snap() (*SnapShot, error) {
	log.Printf(`Creating snapshot for resource with path "%v"`, resource.Path)
	// Make call to resource path
	resp, err := http.Get(resource.Path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Read response body
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Create new snapshot object
	snapshot := SnapShot{
		StatusCode:   resp.StatusCode,
		DateTime:     time.Now(),
		ResponseBody: string(b),
		Resource:     &resource,
	}
	return &snapshot, nil
}

func (resource Resource) MostRecentSnapShot() *SnapShot {
	return &resource.Snapshots[len(resource.Snapshots)-1]
}

type SnapShot struct {
	DateTime     time.Time `json:"datetime"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response"`
	Id           int       `json:"id"`
	Resource     *Resource
}
