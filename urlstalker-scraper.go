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

func main() {

	host := "http://localhost:8000"

	// Get resources
	resources := GetResources(host)
	for _, resource := range *resources {
		log.Println(resource.Path)
	}

	// For each resource scrape URL
	snapshots := GetSnapShots(resources)

	// For each resource post result
	PostSnapShots(host, snapshots)

}

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

func GetResources(host string) *[]Resource {
	resp := getResponse(host + "/resource")
	defer resp.Body.Close()
	return parseResponse(resp)
}

func getResponse(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
	}
	return resp
}

func parseResponse(resp *http.Response) *[]Resource {
	var resources []Resource
	err := json.NewDecoder(resp.Body).Decode(&resources)
	if err != nil {
		log.Print(err)
	}
	return &resources
}

func GetSnapShots(resources *[]Resource) *map[int]SnapShot {
	snapshots := make(map[int]SnapShot)
	for _, resource := range *resources {
		resp, err := http.Get(resource.Path)
		if err != nil {
			log.Print(err)
			continue
		}
		snapshots[resource.Id] = *createSnapShot(resp)
	}
	return &snapshots
}

func createSnapShot(resp *http.Response) *SnapShot {
	b, _ := io.ReadAll(resp.Body)
	snapshot := SnapShot{
		StatusCode: resp.StatusCode,
		DateTime:   time.Now(),
		Response:   string(b),
	}
	return &snapshot
}

func PostSnapShots(host string, snapshots *map[int]SnapShot) {
	for i, snapshot := range *snapshots {
		url := host + fmt.Sprintf("/resource/%v/snapshot", i)
		jsonValue, _ := json.Marshal(snapshot)
		_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			log.Print(err)
		}
	}
}
