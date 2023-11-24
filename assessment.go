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
	Requestchannel   chan map[string]string
	Convertedchannel chan models.Converted
)

func main() {
	Requestchannel = make(chan map[string]string)
	Convertedchannel = make(chan models.Converted)
	go worker()

	router := http.NewServeMux()
	router.HandleFunc("/process", ProcessHandler)
	server := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}
	fmt.Println("Server listening on :8000")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
func ProcessHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string]string

	decoder := json.NewDecoder(r.Body)
	r.Header.Add("Content-Type", "application/json")
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	Requestchannel <- req
	json.NewEncoder(w).Encode(<-Convertedchannel)

}
func Convert(m map[string]string) {
	ConvertRequest := new(models.Converted)
	ConvertRequest.Event = m["ev"]
	ConvertRequest.EventType = m["et"]
	ConvertRequest.AppID = m["id"]
	ConvertRequest.UserID = m["uid"]
	ConvertRequest.MessageID = m["mid"]
	ConvertRequest.PageTitle = m["t"]
	ConvertRequest.PageURL = m["p"]
	ConvertRequest.BrowserLanguage = m["l"]
	ConvertRequest.ScreenSize = m["cs"]
	ConvertRequest.Attributes = make(map[string]models.Attribute)
	ConvertRequest.UserTraits = make(map[string]models.Attribute)
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
			ConvertRequest.Attributes[value] = atr
		}
		if search1.MatchString(key) {
			str := strings.Split(key, "uatrk")
			v := "uatrv" + str[1]
			t := "uatrt" + str[1]
			var atr models.Attribute
			atr.Value = m[v]
			atr.Type = m[t]
			ConvertRequest.UserTraits[value] = atr
		}
	}
	Convertedchannel <- *ConvertRequest
}

func worker() {
	for Req := range Requestchannel {
		Convert(Req)
	}
}
