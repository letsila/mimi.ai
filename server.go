package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/omidnikta/logrus"
)

// {"object": "page", "entry": [{"messaging": [{"message": "TEST_MESSAGE"}]}]}

// BodyMsg message messenger
type BodyMsg struct {
	Object string      `json:"object"`
	Entry  []messaging `json:"entry"`
}

type messaging struct {
	ID        string        `json:"id"`
	Messaging []messagingEl `json:"messaging"`
	Time      int           `json:"time"`
}

type messagingEl struct {
	Message   string    `json:"message"`
	Recipient recipient `json:"recipient"`
	Sender    sender    `json:"sender"`
	Timestamp int       `json:"timestamp"`
}

type message struct {
	Mid  string `json:"mid"`
	Seq  int    `json:"seq"`
	Text string `json:"text"`
}

type recipient struct {
	ID string `json:"id"`
}

type sender struct {
	ID string `json:"id"`
}

type PrivacyData struct {
	Domain   string
	Business string
	City     string
	Country  string
}

func hello(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Greetings from mi AI %s!", r.URL.Path[1:])
}

func verifyHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	verifyToken := os.Getenv("VERIFY_TOKEN")

	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode != "" && token != "" {
		if mode == "subscribe" && token == verifyToken {
			fmt.Println("WEBHOOK_VERIFIED")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", challenge)
		} else {
			http.Error(w, "Challenge failure", http.StatusBadRequest)
		}
	}
}

func apiHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body := BodyMsg{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logrus.Errorf("Failed to decode JSON: %v.", err)
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if body.Object == "page" {
		for _, entry := range body.Entry {
			webhookEvent := entry.Messaging[0]

			fmt.Println(webhookEvent)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", "EVENT_RECEIVED")
	} else {
		http.Error(w, "Message error", http.StatusBadRequest)
	}
}

func privacyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	privacyData := PrivacyData{
		os.Getenv("DOMAIN"),
		os.Getenv("BUSINESS"),
		os.Getenv("CITY"),
		os.Getenv("COUNTRY"),
	}

	renderTemplate(w, "template/privacy_policy", privacyData)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data PrivacyData) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, data)
}

// Handles messages events
func handleMessage(sender_psid string, received_message string) {

}

// Handles messaging_postbacks events
func handlePostback(sender_psid string, received_postback string) {

}

// Sends response messages via the Send API
func callSendAPI(sender_psid string, response string) {

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	certPath := os.Getenv("CERT_PATH")
	certKeyPath := os.Getenv("CERT_KEY_PATH")

	router := httprouter.New()
	router.GET("/", hello)
	router.GET("/privacy", privacyHandler)
	router.GET("/webhook", verifyHook)
	router.POST("/webhook", apiHook)

	error := http.ListenAndServeTLS(":443", certPath, certKeyPath, router)
	if error != nil {
		log.Fatal(http.ListenAndServe(":8080", router))
	}
}
