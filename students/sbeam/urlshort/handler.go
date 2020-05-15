package urlshort

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v2"
	"net/http"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(urlMap map[string]string, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		dest := urlMap[req.URL.Path]
		if dest != "" {
			http.Redirect(res, req, dest, http.StatusPermanentRedirect)
		} else {
			fallback.ServeHTTP(res, req)
		}
	})
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.

type Short struct {
	Path string
	Url  string
}

func parseYAML(yml []byte) (shorts []Short, err error) {
	err = yaml.Unmarshal(yml, &shorts)
	return
}

func makeMapFromShorts(shorts []Short) (urlMap map[string]string) {
	urlMap = make(map[string]string)
	for _, short := range shorts {
		urlMap[short.Path] = short.Url
	}
	return
}

func YAMLHandler(yml []byte, fallback http.Handler) (handler http.HandlerFunc, err error) {
	shorts, err := parseYAML(yml)
	if err != nil {
		return
	}

	fmt.Printf("loaded %d shorts from YAML config\n", len(shorts))
	handler = MapHandler(makeMapFromShorts(shorts), fallback)
	return
}

func JSONHandler(jsonConfig []byte, fallback http.Handler) (handler http.HandlerFunc, err error) {
	var shorts []Short
	err = json.Unmarshal([]byte(jsonConfig), &shorts)

	fmt.Printf("loaded %d shorts from JSON config\n", len(shorts))
	if err != nil {
		return
	}

	handler = MapHandler(makeMapFromShorts(shorts), fallback)
	return
}

func DBHandler(boltDBPath string, fallback http.Handler) (handler http.HandlerFunc, err error) {
	var shortsBucket = []byte("shortsBucket")
	var dbConn *bolt.DB
	var shorts []Short

	dbConn, err = bolt.Open(boltDBPath, 0644, &bolt.Options{ReadOnly: true})
	if err != nil {
		return
	}

	err = dbConn.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(shortsBucket)
		if bucket == nil {
			bucket, err = tx.CreateBucket(shortsBucket)
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
		}
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			shorts = append(shorts, Short{
				Path: string(k),
				Url:  string(v),
			})
		}
		fmt.Printf("loaded %d shorts from DB\n", len(shorts))
		return nil
	})

	if err != nil {
		return
	}

	handler = MapHandler(makeMapFromShorts(shorts), fallback)
	return
}
