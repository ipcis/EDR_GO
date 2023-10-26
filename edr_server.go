package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type MyData struct {
	Field1      string                 `json:"field1"`
	Field2      int                    `json:"field2"`
	ExtraFields map[string]interface{} `json:"extra_fields"`
}

func main() {
	http.HandleFunc("/receive-json", receiveJSON)
	serverAddr := "localhost:8080" // Hier den gewünschten Server-Adresse und Port angeben

	fmt.Printf("Server läuft auf https://%s\n", serverAddr)

	err := http.ListenAndServeTLS(serverAddr, "server.crt", "server.key", nil)
	if err != nil {
		fmt.Println("Fehler beim Starten des Servers:", err)
	}
}

func prettyPrintJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Fehler beim Konvertieren in JSON:", err)
		return
	}

	fmt.Println(string(jsonData))
}

func receiveJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Nur POST-Anfragen sind erlaubt", http.StatusMethodNotAllowed)
		return
	}

	//var requestData json.RawMessage

	// Lesen Sie den Request-Body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Fehler beim Lesen des Request-Bodys", http.StatusBadRequest)
		return
	}

	// Setzen Sie requestData auf das ungeparsste JSON
	//requestData = json.RawMessage(body)

	// Jetzt können Sie requestData als Text ausgeben
	//textData := string(requestData)
	//fmt.Printf("JSON-Daten als Text: %s\n", textData)

	// JSON-Rohdaten in eine json.RawMessage einlesen
	var requestDataJ json.RawMessage
	if err := json.Unmarshal(body, &requestDataJ); err != nil {
		fmt.Println("Fehler beim Deserialisieren von JSON:", err)
		return
	}

	// JSON-Daten in einem leserlichen Format ausgeben
	prettyPrintJSON(requestDataJ)

	w.WriteHeader(http.StatusOK)
}
