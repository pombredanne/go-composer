package main
import (
  "log"
  "net/http"
  "net/url"
  "io"
  "os"
  "flag"
  "fmt"
  "path/filepath"
  "io/ioutil"
  "encoding/json"
  "strconv"
  "bytes"   // http://stackoverflow.com/questions/1760757/how-to-efficiently-concatenate-strings-in-go
  "time"
)

// http://openmymind.net/Golang-Hot-Configuration-Reload/

// filter
// calculate
// chunk
// serialize


// api_ver, serialization_type, filter_func, sortedby, catid

// {"data": [{"id": 12, "order": [1, 2, 3]}, {"id": 3, "order": [2, 1, 3]}]}
type SortingJson struct {
    Data []SingleSortingData
}
type SingleSortingData struct {
    Id int
    Order []int
}

type VodJson struct {
    Id int
    Title string
    Categories []int
}



type Composer struct {
    Serializer Serializer
    Sorting string
    Context string
    Page string
    PerPage string
}
func (this *Composer) render() string {
    file, err := ioutil.ReadFile("./static/sorting.json")
    if err != nil {
        log.Printf("Couldn't load sorting.json: %v\n", err)
        return ""
    }
    //log.Printf("sorting.json loaded: %s\n", string(file))
    var sorting_json SortingJson
    e := json.Unmarshal(file, &sorting_json)
    if e != nil {
        log.Printf("Couldn't unmarshal sorting.json: %v\n", e)
        return ""
    }

    var ordering []int
    for i := 0; i < len(sorting_json.Data); i++ {
        if strconv.Itoa(sorting_json.Data[i].Id) == this.Sorting {
            ordering = sorting_json.Data[i].Order
            break
        }
    }
    //log.Printf("Selected ordering: %v", ordering)



    // next step is to load vods and filter them out
    var vods []VodJson
    for i := 0; i < len(ordering); i++ {
        file, err := ioutil.ReadFile(fmt.Sprintf("./static/vods/%d.json", ordering[i]))
        if err != nil {
            log.Printf("Couldn't load vods/%d.json: %v\n", ordering[i], err)
            continue
        }
        var vod_json VodJson
        e := json.Unmarshal(file, &vod_json)
        if e != nil {
            log.Printf("Couldn't unmarshal vods/%d.json: %v\n", ordering[i], e)
            continue
        }

        // filter out
        if this.Context != "" && contains_int(this.Context, vod_json.Categories) {
            vods = append(vods, vod_json)
        }
    }
    //log.Printf("vods: %v", vods)
    return this.Serializer.render(vods)
}

func contains_int(what string, list []int) bool {
    for j := 0; j < len(list); j++ {
        if what == strconv.Itoa(list[j]) {
            return true
        }
    }
    return false
}

type Serializer struct {
    Name string
}
func (this *Serializer) render(vods []VodJson) string {
    switch this.Name {
    case "tv/12":
        return this.tv12(vods)
    case "pc/85":
        return this.pc85(vods)
    }
    return ""
}
func (this *Serializer) tv12(vods []VodJson) string {
    var buffer bytes.Buffer
    buffer.WriteString("<vods>")
    for i := 0; i < len(vods); i++ {
        buffer.WriteString(fmt.Sprintf("<vod id=\"%d\" title=\"%s\" />", vods[i].Id, vods[i].Title))
    }
    buffer.WriteString("</vods>")
    return buffer.String()
}
func (this *Serializer) pc85(vods []VodJson) string {
    return ""
}

type Conclusion struct {
    Form url.Values

}
func (this *Conclusion) getComposer() Composer {
    serializer := this.Form.Get("serializer")   // serialization: tv, pc, mobile
    sorting := this.Form.Get("sorting") // sort: need only to load vods in right order
    context := this.Form.Get("context") // filter: need to know which category
    page := this.Form.Get("page")
    per_page := this.Form.Get("perpage")

    //log.Printf("Serializer:%s Sort:%s Ctx:%s", serializer, sorting, context)

    switch serializer {
    case "tv/12":
        return Composer{Serializer: Serializer{Name: "tv/12"}, Sorting: sorting, Context: context, Page: page, PerPage: per_page}
    case "pc/85":
        return Composer{Serializer: Serializer{Name: "pc/85"}, Sorting: sorting, Context: context, Page: page, PerPage: per_page}
    }
    return Composer{Serializer: Serializer{Name: "default"}, Sorting: sorting, Context: context, Page: page, PerPage: per_page}
}


func compose(res http.ResponseWriter, req *http.Request) {
    start := time.Now()

    req.ParseForm()
    //log.Printf("Params: %s", req.Form)

    res.Header().Set(
        "Content-Type",
        "application/xml",
    )

    conclusion := Conclusion{Form: req.Form}
    composer := conclusion.getComposer()

    io.WriteString(
        res,
        composer.render(),
    )

    log.Printf("\"%s %s\" %s \"%s\" %s",
            req.Method,
            req.URL.Path,
            req.Proto,
            req.UserAgent(),
            time.Since(start))
}


func main() {
  //arguments := os.Args
  //fmt.Println(arguments)
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
