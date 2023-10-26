package main

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/google/uuid" // Importieren der uuid-Bibliothek
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"

	"net/http"
)

type JSONData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Struct für die Prozessinformationen
type ProcessInfo struct {
	Pid       int32  `json:"pid"`
	Name      string `json:"name"`
	ExePath   string `json:"exePath"`
	Md5Sum    string `json:"md5Sum"`
	StartTime int64  `json:"startTime"`
}

// Struct für Netzwerkverbindungen
type NetworkInfo struct {
	LocalAddress  string `json:"localAddress"`
	RemoteAddress string `json:"remoteAddress"`
	Status        string `json:"status"`
}

// Struct für die gesamte Information
type SystemInfo struct {
	ComputerName string        `json:"computerName"`
	UUID         string        `json:"uuid"`
	Processes    []ProcessInfo `json:"processes"`
	Connections  []NetworkInfo `json:"connections"`
}

// Map zur Speicherung der gecachten MD5-Summen
var md5SumCache = make(map[int32]string)

func main() {
	interval := 10 * time.Second // Intervall für das Hochladen

	for {
		// Sammle Prozessinformationen
		processInfo := getProcessInfo()

		// Sammle Netzwerkverbindungen
		networkInfo := getNetworkInfo()

		// Sammle Hostinformationen
		hostInfo, err := host.Info()
		if err != nil {
			fmt.Println("Fehler beim Abrufen von Hostinformationen:", err)
		}

		systemInfo := SystemInfo{
			ComputerName: hostInfo.Hostname,
			UUID:         generateUUID(), // Setze die UUID
			Processes:    processInfo,
			Connections:  networkInfo,
		}

		// Konvertiere das SystemInfo-Objekt in JSON
		jsonData, err := json.Marshal(systemInfo)
		if err != nil {
			fmt.Println("Fehler beim Konvertieren in JSON:", err)
			continue
		}

		fmt.Println(string(jsonData)) // Ausgabe der JSON-Daten

		//http.DefaultTransport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		serverURL := "https://localhost:8080/receive-json" // Ersetzen Sie dies durch die tatsächliche Server-URL

		dataToSend := string(jsonData)

		// Senden Sie die JSON-Daten an den Server
		err = uploadJSON(serverURL, string(dataToSend))
		if err != nil {
			fmt.Println("Fehler beim Hochladen der JSON-Daten:", err)
		}

		// Alle 10 Sekunden erneut ausführen
		time.Sleep(interval)
	}
}

func uploadJSON(serverURL string, jsonData string) error {
	// Erstellen Sie einen HTTP-Request mit den JSON-Daten als Text
	req, err := http.NewRequest("POST", serverURL, strings.NewReader(jsonData))
	if err != nil {
		return err
	}

	// Setzen Sie den Content Type-Header auf "application/json"
	req.Header.Set("Content-Type", "application/json")

	// Erstellen Sie einen HTTP-Client mit angepasstem Timeout (optional)
	httpClient := &http.Client{
		Timeout: 10 * time.Second, // Timeout nach 10 Sekunden (kann angepasst werden)
	}

	// Senden Sie den HTTP-Request an den Server
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Überprüfen Sie die Antwort des Servers (optional)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fehlerhafte Antwort vom Server: %s", resp.Status)
	}

	return nil
}

// Funktion zum Generieren einer UUID
func generateUUID() string {
	id := uuid.New()
	return id.String()
}

// Funktion, um Prozessinformationen zu sammeln
func getProcessInfo() []ProcessInfo {
	pids, err := process.Pids()
	if err != nil {
		fmt.Println("Fehler beim Abrufen von PIDs:", err)
		return nil
	}

	var processInfo []ProcessInfo
	for _, pid := range pids {
		proc, err := process.NewProcess(pid)
		if err != nil {
			fmt.Println("Fehler beim Abrufen von Prozessinformationen:", err)
			continue
		}

		exePath, _ := proc.Exe()
		name, _ := proc.Name()
		createTime, _ := proc.CreateTime()

		// Überprüfe, ob die MD5-Summe bereits gecacht ist
		md5Sum, ok := md5SumCache[pid]
		if !ok {
			// Wenn nicht, berechne die MD5-Summe und speichere sie in der Cache-Map
			md5Sum, err := calculateMD5(exePath)
			if err != nil {
				//fmt.Println("Fehler beim Berechnen der MD5-Summe:", err)
			} else {
				md5SumCache[pid] = md5Sum
			}
		}

		processInfo = append(processInfo, ProcessInfo{
			Pid:       pid,
			Name:      name,
			ExePath:   exePath,
			Md5Sum:    md5Sum,
			StartTime: createTime,
		})
	}

	return processInfo
}

// Funktion, um Netzwerkverbindungen zu sammeln
func getNetworkInfo() []NetworkInfo {
	connections, err := net.Connections("all")
	if err != nil {
		fmt.Println("Fehler beim Abrufen von Netzwerkverbindungen:", err)
		return nil
	}

	var networkInfo []NetworkInfo
	for _, conn := range connections {
		networkInfo = append(networkInfo, NetworkInfo{
			LocalAddress:  conn.Laddr.IP,
			RemoteAddress: conn.Raddr.IP,
			Status:        conn.Status,
		})
	}

	return networkInfo
}

// Funktion zum Berechnen der MD5-Summe einer Datei
func calculateMD5(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:]), nil
}
