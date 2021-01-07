package main

import (
	"fmt"
	"html/template"
	"io"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type templates struct {
	*template.Template
}

func (t templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.ExecuteTemplate(w, name, data)
}

func initTemplates() templates {
	t := template.New("")
	t.Funcs(template.FuncMap{
		"toString": toString,
		"link":     link,
		"email": func() string {
			return "example@example.com"
		},
	})

	// parse all html files in the templates directory
	t, err := t.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	// also parse files in subdirectories of templates
	t, err = t.ParseGlob("templates/*/*.html")
	if err != nil {
		log.Fatal(err)
	}

	return templates{t}
}

func main() {
	e := echo.New()

	e.Debug = true // TODO: this line should be removed in production
	e.Renderer = initTemplates()
	e.Use(middleware.Gzip(), middleware.Secure())
	// you should also add middleware.CSRF(), once you have forms

	e.GET("/", root)
	e.GET("/foo", foo)
	e.GET("/bar", bar)
	e.Static("/dist", "./dist")
	e.Start(":3000")
}

func root(c echo.Context) error {
	return c.Render(200, "index.html", map[string]interface{}{
		"title": "Root",
		"test":  "Hello, world!",
		"slice": []int{1, 2, 3},
	})
}

func foo(c echo.Context) error {
	return c.Render(200, "foo.html", map[string]interface{}{
		"title": "Foo",
	})
}

func bar(c echo.Context) error {
	return c.Render(200, "bar.html", map[string]interface{}{
		"title": "Bar",
	})
}

// toString converts any value to string
// functions that return a string are automatically escaped by html/template
func toString(v interface{}) string {
	return fmt.Sprint(v)
}

// link returns a styled "a" tag
// functions that return a template.HTML are not escaped, so all parameters need to be escaped to avoid xss
func link(location, name string) template.HTML {
	return escSprintf(`<a class="text-blue-600 no-underline hover:underline" href="%v">%v</a>`, location, name)
}

// escSprintf is like fmt.Sprintf but uses the escaped HTML equivalent of the args
func escSprintf(format string, args ...interface{}) template.HTML {
	for i, arg := range args {
		args[i] = template.HTMLEscapeString(fmt.Sprint(arg))
	}

	return template.HTML(fmt.Sprintf(format, args...))
}
