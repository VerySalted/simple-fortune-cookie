package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var BACKEND_DNS = getEnv("BACKEND_DNS", "localhost")
var BACKEND_PORT = getEnv("BACKEND_PORT", "9000")

type fortune struct {
	ID      string `json:"id" redis:"id"`
	Message string `json:"message" redis:"message"`
}

type newFortune struct {
	Message string `json:"message"`
}

// use a custom client, because we don't do blocking operations wihout timeouts
var myClient = &http.Client{Timeout: 10 * time.Second}

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := io.WriteString(w, "healthy")
	if err != nil {
		log.Fatalln(err)
		fmt.Fprint(w, err)
	}
}

func main() {

	http.HandleFunc("/healthz", HealthzHandler)

	http.HandleFunc("/api/random", func(w http.ResponseWriter, r *http.Request) {
		resp, err := myClient.Get(fmt.Sprintf("http://%s:%s/fortunes/random", BACKEND_DNS, BACKEND_PORT))
		if err != nil {
			log.Fatalln(err)
			fmt.Fprint(w, err)
			return
		}

		f := new(fortune)
		err = json.NewDecoder(resp.Body).Decode(f)
		if err != nil {
			panic(err)
		}

		fmt.Fprint(w, f.Message)
	})

	http.HandleFunc("/api/all", func(w http.ResponseWriter, r *http.Request) {
		resp, err := myClient.Get(fmt.Sprintf("http://%s:%s/fortunes", BACKEND_DNS, BACKEND_PORT))
		if err != nil {
			log.Fatalln(err)
			fmt.Fprint(w, err)
			return
		}

		fortunes := new([]fortune)
		err = json.NewDecoder(resp.Body).Decode(fortunes)
		if err != nil {
			panic(err)
		}

		tmpl, err := template.ParseFiles("./templates/fortunes.html")

		if err != nil {
			log.Fatalln(err)
			fmt.Fprint(w, err)
			return
		}

		err = tmpl.Execute(w, fortunes)
		if err != nil {
			panic(err)
		}
	})

	http.HandleFunc("/api/add", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "Use POST", http.StatusMethodNotAllowed)
			return
		}

		f := new(newFortune)
		err := json.NewDecoder(r.Body).Decode(f)
		if err != nil {
			panic(err)
		}

		var postUrl = fmt.Sprintf("http://%s:%s/fortunes", BACKEND_DNS, BACKEND_PORT)
		var jsonStr = []byte(fmt.Sprintf(`{"id": "%d", "message": "%s"}`, rand.Intn(10000), f.Message))

		_, err = myClient.Post(postUrl, "application/json", bytes.NewBuffer(jsonStr))
		if err != nil {
			log.Fatalln(err)
			fmt.Fprint(w, err)
			return
		}

		fmt.Fprint(w, "Cookie added!")
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
