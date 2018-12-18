package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Env contains the darksy API key & the Slack url
type Env struct {
	darksky  string
	slackURL string
}

type message struct {
	Text string `json:"text"`
}

func failOnError(err error, msg string) {
	if err == nil {
		return
	}
	log.Fatalf("%s: %s", msg, err)
	panic(fmt.Sprintf("%s: %s", msg, err))
}

func (env *Env) sendToChannel(text string) error {
	body, err := json.Marshal(message{text})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", env.slackURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (env *Env) handleFavicon(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "favicon.ico")
}

func (env *Env) handleHomeRequest(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The service is working!"))
}

func (env *Env) handlePostWeatherRequest(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// text value received from SLACK
	t := req.FormValue("text")
	if err := env.sendToChannel(t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	env := &Env{
		darksky:  getEnv("DARKSKY_API_KEY", "DARKSKY_API_KEY"),
		slackURL: getEnv("SLACK_URL", "SLACK_URL"),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", env.handleHomeRequest)
	r.Post("/weather", env.handlePostWeatherRequest)
	r.Get("/favicon.ico", env.handleFavicon)

	// Launch the Web Server
	addr := fmt.Sprintf("0.0.0.0:%s", getEnv("PORT", "3000"))
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Server run on http://" + addr)
	log.Fatal(srv.ListenAndServe())
}
