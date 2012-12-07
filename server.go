package main

import (
        "fmt"
        "github.com/gokyle/webshell"
        "github.com/russross/blackfriday"
        "html/template"
        "io/ioutil"
        "net/http"
        "os"
        "path/filepath"
        "regexp"
)

var (
        server_port = "8080"
        page_tpl = webshell.MustCompileTemplate("templates/page.html")
        htmlToMd = regexp.MustCompile("^(.+)\\.html$")
        extRegex = regexp.MustCompile("^(.+)\\.(\\w+)$")
        slash_replace = regexp.MustCompile("/")
)

type Page struct {
        HomeActive bool
        AboutActive bool
        Title string
        Body template.HTML
}

func init() {
        port := os.Getenv("PORT")
        if port != "" {
                server_port = port
        }
}

func servePage(w http.ResponseWriter, r *http.Request) {
        file := r.URL.Path[1:]
        if file == "" {
                file = "index.html"
        }

        var md_file string
        if extRegex.MatchString(file) {
                if ext := extRegex.ReplaceAllString(file, "$2"); ext == "html" {
                        md_file = "pages/" + htmlToMd.ReplaceAllString(file,
                                     "$1.md")
                }
        } else {
                md_file = "pages/" + file + ".md"
        }
        out, err := loadMarkdown(md_file)
        if err != nil {
                fmt.Printf("saw request for %s, tried %s\n", file, md_file)
                webshell.Error404("Page not found.", "text/plain", w, r)
                return
        }
        title := extRegex.ReplaceAllString(filepath.Base(file), "$1")
        title = slash_replace.ReplaceAllString(title, " ")
        page := Page{false, false, title, template.HTML(string(out))}
        active := extRegex.ReplaceAllString(filepath.Base(file), "$1")
        if active == "index" {
                page.HomeActive = true
        } else if active == "about" {
                page.AboutActive = true
        }
        body, err := webshell.BuildTemplate(page_tpl, page)
        if err != nil {
                webshell.Error500(err.Error(), "text/plain", w, r)
                return
        }
        w.Write(body)
}

func loadMarkdown(file string) (out []byte, err error) {
        var in []byte
        in, err = ioutil.ReadFile(file)
        if err != nil {
                return
        }
        out = blackfriday.MarkdownCommon(in)
        return
}

func main() {
        app := webshell.NewApp("webshell tutorial", "", server_port)
        app.AddRoute("/", servePage)
        app.StaticRoute("/assets/", "assets/")
        app.StaticRoute("/examples/", "examples/")
        fmt.Println("[!] ", app.Serve())
}
