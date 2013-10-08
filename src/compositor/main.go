package main

import (
	"bytes" // http://stackoverflow.com/questions/1760757/how-to-efficiently-concatenate-strings-in-go
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// http://blog.repustate.com/migrating-code-from-python-to-golang-what-you-need-to-know/2013/04/23/
// http://openmymind.net/Golang-Hot-Configuration-Reload/
// http://synflood.at/tmp/golang-slides/mrmcd2012.html#52
// https://code.google.com/p/go-wiki/wiki/SliceTricks
// http://golang.org/doc/effective_go.html
// http://www.golang-book.com/
// http://jan.newmarch.name/go/
// https://gobyexample.com/
// http://tip.golang.org/ref/spec

type Vod struct {
	Id         int
	Title      string
	Categories []int
}

// {"data": [{"id": 12, "order": [1, 2, 3]}, {"id": 3, "order": [2, 1, 3]}]}
type SortingJson struct {
	Data []SingleSortingData
}
type SingleSortingData struct {
	Id    int
	Order []int
}

// Serializer interface
type Serializer interface {
	render(sorting int, category int, page int, per_page int) string
	mimetype() string
}

// Default Serializer
type DefaultSerializer struct{}

// Load sorting order
func (this DefaultSerializer) load_sorting(sorting int) []int {
	// Czytamy plik
	if file, err := ioutil.ReadFile("./static/sorting.json"); err == nil {
		// Ladujemy JSONa
		var sorting_json SortingJson
		if err := json.Unmarshal(file, &sorting_json); err == nil {

			// Przeszukujemy dostepne w pliku sortowania w poszukiwaniu naszego
			for i := range sorting_json.Data {
				if sorting_json.Data[i].Id == sorting {
					return sorting_json.Data[i].Order
				}
			}
			log.Printf("ERROR Sorting %v not found", sorting)
		} else {
			log.Printf("ERROR %v", err)
		}
	} else {
		log.Printf("ERROR %v", err)
	}
	return []int{}
}

// Load vods
func (this DefaultSerializer) load_vods(vod_ids []int, category int, page int, per_page int) []Vod {
	start := (page - 1) * per_page
	stop := start + per_page
	if stop > len(vod_ids) {
		stop = len(vod_ids)
	}

	vods := []Vod{}
	for i := range vod_ids {
		if file, err := ioutil.ReadFile(fmt.Sprintf("./static/vods/%d.json", vod_ids[i])); err == nil {
			var vod Vod
			if err := json.Unmarshal(file, &vod); err == nil {
				if this.valid_vod(vod, category) {
					vods = append(vods, vod)
					if len(vods) >= stop {
						return vods
					}
				}
			} else {
				log.Printf("ERROR %v", err)
			}
		} else {
			log.Printf("ERROR %v", err)
		}
	}

	if start > len(vods) {
		return []Vod{}
	}
	if stop > len(vods) {
		stop = len(vods)
	}

	return vods[start:stop]
}

// Check if vod is valid
func (this DefaultSerializer) valid_vod(vod Vod, category int) bool {
	return true
}
func (this DefaultSerializer) render(sorting int, category int, page int, per_page int) string {
	return "Empty"
}
func (this DefaultSerializer) mimetype() string {
	return "text/plain"
}

// XML Serializer
type XMLSerializer struct {
	DefaultSerializer
}

func (this XMLSerializer) render(sorting int, category int, page int, per_page int) string {
	return "<vods />"
}
func (this XMLSerializer) mimetype() string {
	return "application/xml"
}

// JSON Serializer
type JSONSerializer struct {
	DefaultSerializer
}

func (this JSONSerializer) render(sorting int, category int, page int, per_page int) string {
	return "{\"vods\": []}"
}
func (this JSONSerializer) mimetype() string {
	return "application/json"
}

// Ipla
type Ipla300Serializer struct {
	XMLSerializer
}

func (this Ipla300Serializer) render(sorting int, category int, page int, per_page int) string {
	vod_ids := this.load_sorting(sorting)
	vods := this.load_vods(vod_ids, category, page, per_page)
	var buffer bytes.Buffer
	buffer.WriteString("<vods>")
	for i := range vods {
		buffer.WriteString(fmt.Sprintf("<vod id=\"%d\" title=\"%s\" />", vods[i].Id, vods[i].Title))
	}
	buffer.WriteString("</vods>")
	return buffer.String()
}

// Samsung
type Samsung20Serializer struct {
	JSONSerializer
}

func (this Samsung20Serializer) render(sorting int, category int, page int, per_page int) string {
	return "{\"vods\": [{\"id\": 3}]}"
}

func get_serializer(client_name string, client_build int) Serializer {
	if contains_str(client_name, []string{"ipla"}) {
		switch {
		case client_build > 300:
			return Ipla300Serializer{}
		default:
			return DefaultSerializer{}
		}
	} else if contains_str(client_name, []string{"tv_samsung"}) {
		return Samsung20Serializer{}
	}
	return DefaultSerializer{}
}

func contains_str(what string, list []string) bool {
	for j := 0; j < len(list); j++ {
		if what == list[j] {
			return true
		}
	}
	return false
}

func compose_params(Form url.Values) (string, int, int, int, int, int) {
	client_name := Form.Get("client_name")
	client_build, _ := strconv.Atoi(Form.Get("client_build")) // defaults to 0
	sorting, _ := strconv.Atoi(Form.Get("sorting"))
	category, _ := strconv.Atoi(Form.Get("category"))
	page, _ := strconv.Atoi(Form.Get("page"))
	per_page, _ := strconv.Atoi(Form.Get("per_page"))
	// log.Printf("DEBUG client_name=%v client_build=%v sorting=%v category=%v page=%v per_page=%v",
	// 	client_name, client_build, sorting, category, page, per_page)
	return client_name, client_build, sorting, category, page, per_page
}

func log_request(start time.Time, request *http.Request) {
	log.Printf("\"%s %s\" %s \"%s\" %s",
		request.Method,
		request.URL.Path,
		request.Proto,
		request.UserAgent(),
		time.Since(start),
	)
}

func compose(res http.ResponseWriter, req *http.Request) {
	defer log_request(time.Now(), req)

	req.ParseForm()
	client_name, client_build, sorting, category, page, per_page := compose_params(req.Form)
	serializer := get_serializer(client_name, client_build) // pass conf

	res.Header().Set("Content-Type", serializer.mimetype())
	io.WriteString(res, serializer.render(sorting, category, page, per_page))
}

func main() {
	port_ptr := flag.Int("port", 9000, "Port number")
	path_ptr := flag.String("path", "static", "Path to static folder")
	flag.Parse()

	static_path, error := filepath.Abs(*path_ptr)
	if error != nil {
		fmt.Printf(error.Error())
		os.Exit(2)
	}

	http.HandleFunc("/compose", compose)

	log.Printf("Static path: %s", static_path)
	log.Printf("Starting server on :%d", *port_ptr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port_ptr), nil))
	os.Exit(1)
}
