package main

import (
	"fmt"
    "net/http"
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    _ "github.com/mattn/go-sqlite3"
)

func main() {

    m := martini.Classic()
    // render html templates from templates directory
  	m.Use(render.Renderer(render.Options{
  		Directory: "templates", // Specify what path to load the templates from.
  		Layout: "", // Specify a layout template. Layouts can call {{ yield }} to render the current template.
  		Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
  		Delims: render.Delims{"{{", "}}"}, // Sets delimiters to the specified strings.
  		Charset: "UTF-8", // Sets encoding for json and html content-types. Default is "UTF-8".
  		IndentJSON: true, // Output human readable JSON
	}))

	staticOptions := martini.StaticOptions{Prefix: "static"}
	m.Use(martini.Static("static", staticOptions))

  	m.Get("/", func(r render.Render) {
    	r.HTML(200, "index", "")
  	})

  	m.Get("/signin", func(r render.Render) {
  		r.HTML(200, "signin", "")
  	})

  	m.Get("/signup", func(r render.Render) {
  		r.HTML(200, "signup", "")
  	})

  	m.Get("/collections", func(r render.Render) {
  		r.HTML(200, "collections", "")
  	})

    m.Post("/login", func(r *http.Request) string {
        text := r.FormValue("username")
        fmt.Println("Username: ", text)
        return text
    })
    m.Run()
}