package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
)

// TODO: Set to false in production
const debug = true

type templates struct {
	*template.Template
}

func (t templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.ExecuteTemplate(w, name, data)
}

func initTemplates() templates {
	t := template.New("")
	t.Funcs(template.FuncMap{
		"timeString": func(t time.Time) string {
			return t.Format("15:04")
		},
	})

	// parse all html files in the templates directory
	t, err := t.ParseGlob("templates/*.html")
	if err != nil {
		panic(err)
	}

	return templates{t}
}

// stringRender render a html/template to a string
func stringRender(c echo.Context, name string, data interface{}) string {
	b := bytes.NewBuffer(nil)
	err := c.Echo().Renderer.Render(b, "message-other-stream", data, c)
	if err != nil {
		panic(err)
	}

	return b.String()
}

type Message struct {
	Text string
	Date time.Time
	User string
}

var state struct {
	id       int
	messages map[int]Message

	sync.RWMutex
}

func init() {
	state.messages = make(map[int]Message)
}

func addMiddleware(e *echo.Echo) {
	if debug {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}, latency_human=${latency_human}\n",
		}))
	}

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.Use(middleware.Gzip(), middleware.Secure())

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))
}

func routes(e *echo.Echo) {
	e.Static("/dist", "./dist")

	e.GET("/", root)
	e.GET("/recieve", recieveMessages)

	e.POST("/send", sendMessage)
}

func main() {
	e := echo.New()
	e.Debug = debug
	e.Renderer = initTemplates()

	addMiddleware(e)
	routes(e)

	e.Start(":3000")
}

func root(c echo.Context) error {
	sess, _ := session.Get("session", c)

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if sess.Values["user"] == nil {
		sess.Values["user"] = uuid.New().String()
	}

	sess.Save(c.Request(), c.Response())

	state.RLock()
	defer state.RUnlock()

	return c.Render(200, "index.html", map[string]interface{}{
		"title":    "Chat",
		"messages": state.messages,
		"user":     userID(c),
		"csrf":     csrfToken(c),
	})
}

func sendMessage(c echo.Context) error {
	msg := c.FormValue("message")
	if msg == "" {
		return fmt.Errorf("empty message not allowed")
	}

	state.Lock()
	defer state.Unlock()
	state.id++

	message := Message{
		Text: msg,
		Date: time.Now(),
		User: userID(c),
	}

	state.messages[state.id] = message

	if isTurbo(c) {
		return c.Render(200, "message-self-stream", message)
	}

	return root(c)
}

func userID(c echo.Context) string {
	sess, _ := session.Get("session", c)

	return sess.Values["user"].(string)
}

func recieveMessages(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		state.RLock()
		lastid := state.id
		state.RUnlock()

		for {
			func() {
				state.RLock()
				defer state.RUnlock()
				if state.id == lastid {
					return
				}

				lastid = state.id

				msg := state.messages[state.id]

				if msg.User == userID(c) {
					return
				}

				rMsg := stringRender(c, "message-other-stream", msg)

				websocket.Message.Send(ws, rMsg)
			}()

			time.Sleep(1 * time.Millisecond)
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

func csrfToken(c echo.Context) string {
	return c.Get("csrf").(string)
}

func isTurbo(c echo.Context) bool {
	accept := c.Request().Header.Get("Accept")
	if !strings.Contains(accept, "turbo-stream") {
		return false
	}

	c.Response().Header().Set("Content-Type", "text/html; turbo-stream; charset=utf-8")

	return true
}
