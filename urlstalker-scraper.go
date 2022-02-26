package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// TODO Make recognize change in snapshot
// TODO Fix parsing of datetime
// TODO Add e-mail functionality
// TODO Implement concurrency

type SnapShot struct {
	DateTime   time.Time `json:"datetime"`
	StatusCode int       `json:"status_code"`
	Response   string    `json:"response"`
	Id         int       `json:"id"`
}

type Resource struct {
	Path      string     `json:"path"`
	Id        int        `json:"id"`
	Snapshots []SnapShot `json:"Snapshots"`
}

type UrlStalkerApiClient struct {
	host string
}

func main() {

	client := UrlStalkerApiClient{host: "http://localhost:8000"}

	// Get resources
	resources, err := client.GetResources()
	if err != nil {
		log.Fatal(err)
	}

	// For each resource scrape URL and post result
	for _, resource := range *resources {
		snapshot, err := resource.CreateSnapShot()
		if err != nil {
			log.Println(err)
			continue
		}
		err = client.PostSnapShot(&resource, snapshot)
		if err != nil {
			log.Println(err)
			continue
		}
	}

}

func (c UrlStalkerApiClient) GetResources() (*[]Resource, error) {
	// Make HTTP call to API
	resp, err := http.Get(c.host + "/resource")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Parse response
	var resources []Resource
	err = json.NewDecoder(resp.Body).Decode(&resources)
	return &resources, err
}

func (c UrlStalkerApiClient) PostSnapShot(resource *Resource, snapshot *SnapShot) error {
	log.Printf(`Posting snapshot for resource with path "%v"`, resource.Path)
	url := c.host + fmt.Sprintf("/resource/%v/snapshot", resource.Id)
	jsonValue, _ := json.Marshal(snapshot)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	return err
}

func (resource Resource) CreateSnapShot() (*SnapShot, error) {
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
		StatusCode: resp.StatusCode,
		DateTime:   time.Now(),
		Response:   string(b),
	}
	return &snapshot, nil
}
