package main

import (
	"assessment/models"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	RequestChan   chan map[string]string
	Convertedchan chan models.Converted
)

func main() {
	RequestChan = make(chan map[string]string)
	Convertedchan = make(chan models.Converted)
	go worker()

	router := http.NewServeMux()
	router.HandleFunc("/convert", Handler)
	server := &http.Server{
		Addr:    ":8100",
		Handler: router,
	}
	fmt.Println("Server listening on :8100")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
func Handler(w http.ResponseWriter, r *http.Request) {
	var req map[string]string

	decoder := json.NewDecoder(r.Body)
	r.Header.Add("Content-Type", "application/json")
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	RequestChan <- req
	json.NewEncoder(w).Encode(<-Convertedchan)

}
func Convert(m map[string]string) {
	ConRequest := new(models.Converted)

	ConRequest.Attributes = make(map[string]models.Attribute)
	ConRequest.UserTraits = make(map[string]models.Attribute)
	pattern := "^atrk.*"
	pattern1 := "^uatrk.*"
	search, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return
	}
	search1, err := regexp.Compile(pattern1)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return
	}
	for key, value := range m {
		if search.MatchString(key) {
			str := strings.Split(key, "atrk")
			v := "atrv" + str[1]
			t := "atrt" + str[1]
			var atr models.Attribute
			atr.Value = m[v]
			atr.Type = m[t]
			ConRequest.Attributes[value] = atr
		}
		if search1.MatchString(key) {
			str := strings.Split(key, "uatrk")
			v := "uatrv" + str[1]
			t := "uatrt" + str[1]
			var atr models.Attribute
			atr.Value = m[v]
			atr.Type = m[t]
			ConRequest.UserTraits[value] = atr
		}
	}
	ConRequest.Event = m["ev"]
	ConRequest.EventType = m["et"]
	ConRequest.AppID = m["id"]
	ConRequest.UserID = m["uid"]
	ConRequest.MessageID = m["mid"]
	ConRequest.PageTitle = m["t"]
	ConRequest.PageURL = m["p"]
	ConRequest.BrowserLanguage = m["l"]
	ConRequest.ScreenSize = m["cs"]
	Convertedchan <- *ConRequest
}

func worker() {
	for Req := range RequestChan {
		Convert(Req)
	}
}
