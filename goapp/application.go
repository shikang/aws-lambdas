package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"html"
)

type Page struct {
	Title string
	Body  []byte
}

type EchoJson struct {
	Payload     string		`json:"payload"`
	Timestamp   time.Time 	`json:"timestamp"`
	RequestType string 		`json:"request"`
}

type ErrorJson struct {
	ErrorMsg 	string		`json:"error"`
}

type RequestJson struct {
	Method 		string 		`json:"method"`
	URL     	string		`json:"url"`
	Body     	string		`json:"body"`
	Timestamp   time.Time 	`json:"timestamp"`
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/save/"):]
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	p.save()
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func generateErrorResponse(w http.ResponseWriter, err string) {
	errorResponse := ErrorJson{ErrorMsg: err}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(errorResponse)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	//bodyBuffer, _ := ioutil.ReadAll(r.Body)
	//log.Output(1, string(bodyBuffer))
	echoResponse := EchoJson{}

	err := json.NewDecoder(r.Body).Decode(&echoResponse)
	if err != nil {
		generateErrorResponse(w, err.Error())
		return
	}

	echoResponse.Timestamp = time.Now().Local()
	echoResponse.RequestType = r.Method

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(echoResponse)
}

func printBodyHandler(w http.ResponseWriter, r *http.Request) {
	printBodyResponse := RequestJson{}

	printBodyResponse.Timestamp = time.Now().Local()
	printBodyResponse.Method = r.Method
	printBodyResponse.URL = html.EscapeString(r.URL.Path)

	bodyBuffer, _ := ioutil.ReadAll(r.Body)
	printBodyResponse.Body = string(bodyBuffer)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(printBodyResponse)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/api/echo/", echoHandler)
	http.HandleFunc("/api/printBody/", printBodyHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
