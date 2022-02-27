package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// TODO Make recognize change in snapshot
// TODO Fix parsing of datetime
// TODO Add e-mail functionality
// TODO containerize

func main() {

	db := UrlStalkerApiClient{host: "http://localhost:8000"}

	// Get resources
	resources, err := db.GetResources()
	if err != nil {
		log.Fatal(err)
	}

	// For each resource scrape URL and post result
	var wg sync.WaitGroup
	for _, resource := range *resources {
		wg.Add(1)
		resource := resource
		go ScrapeAndPost(&wg, &resource, &db)
	}
	wg.Wait()
}

func ScrapeAndPost(wg *sync.WaitGroup, resource *Resource, db *UrlStalkerApiClient) {
	defer wg.Done()

	snapshot, err := resource.CreateSnapShot()
	if err != nil {
		log.Println(err)
		return
	}
	err = db.PostSnapShot(resource, snapshot)
	if err != nil {
		log.Println(err)
		return
	}
}

type UrlStalkerApiClient struct {
	host string
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
	jsonValue, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	return err
}

type Resource struct {
	Path string `json:"path"`
	Id   int    `json:"id"`
	// Snapshots []SnapShot `json:"Snapshots"`
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
		StatusCode:   resp.StatusCode,
		DateTime:     time.Now(),
		ResponseBody: string(b),
	}
	return &snapshot, nil
}

type SnapShot struct {
	DateTime     time.Time `json:"datetime"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response"`
	Id           int       `json:"id"`
}
