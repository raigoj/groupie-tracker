package main

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"text/template"
)

type Error struct {
	Error  string
	Filler string
}

type Artist struct {
	Id             int                 `json:"id"`
	Image          string              `json:"image"`
	Name           string              `json:"name"`
	Members        []string            `json:"members"`
	CreationDate   int                 `json:"creationDate"`
	FirstAlbum     string              `json:"firstAlbum"`
	Relations      string              `json:"relations"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type relation struct {
	Index []struct {
		Id             int
		DatesLocations map[string][]string
	}
}

type Artists []Artist

//setting up the server
func main() {
	http.HandleFunc("/", Page)
	http.Handle("/content/", http.StripPrefix("/content", http.FileServer(http.Dir("./content"))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

//taking in the API
func Page(w http.ResponseWriter, r *http.Request) {
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		w.WriteHeader(400)
		panic(err)
	}
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		w.WriteHeader(500)
		panic(err)
	}
	var responseObject []Artist
	json.Unmarshal(responseData, &responseObject)
	if err != nil {
		w.WriteHeader(500)
		panic(err)
	}

	res, err := http.Get("https://groupietrackers.herokuapp.com/api/relation")
	if err != nil {
		w.WriteHeader(400)
		panic(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(500)
		panic(err)
	}
	var tmp relation
	json.Unmarshal(body, &tmp)
	if err != nil {
		w.WriteHeader(500)
		panic(err)
	}

	for index, value := range tmp.Index {
		responseObject[index].DatesLocations = value.DatesLocations
	}

	// main page handler
	if r.URL.Path == "/" {
		t, err := template.ParseFiles(filepath.Join("./templates/main.html"))
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}
		data := responseObject
		err = t.Execute(w, data)
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		}
	} else {
		// if unknown URL, error
		t, err := template.ParseFiles(filepath.Join("./templates/error.html"))
		data := Error{Error: "404: Page not found"}
		w.WriteHeader(404)
		err = t.Execute(w, data)
		if err != nil {
			http.Error(w, "400 - Bad Request", http.StatusBadRequest)
			return
		}
	}
}
