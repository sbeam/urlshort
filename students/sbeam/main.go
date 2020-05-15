package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gophercises/urlshort/students/sbeam/urlshort"
)

func main() {
	var (
		yamlFile = flag.String("yaml", "", "path to YAML URL mappings")
		jsonFile = flag.String("json", "", "path to JSON URL mappings")
	)
	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	handler := urlshort.MapHandler(pathsToUrls, mux)

	if *yamlFile != "" {
		yamlconfig, err := ioutil.ReadFile(*yamlFile)

		if err != nil {
			log.Fatalln("Couldn't open the yaml config file", err)
		}

		handler, err = urlshort.YAMLHandler([]byte(yamlconfig), handler)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if *jsonFile != "" {
		jsondata, err := ioutil.ReadFile(*jsonFile)

		if err != nil {
			log.Fatalln("Couldn't open the JSON config file", err)
		}

		handler, err = urlshort.JSONHandler([]byte(jsondata), handler)
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", handler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
