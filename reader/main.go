package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	randomString string
	timestamp    string
	mutex        sync.Mutex
	db           *sqlx.DB
)

type Counter struct {
	Counter int `db:"counter"`
}

func main() {
	log.Println("Application Started")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	db, err = sqlx.Connect("postgres", databaseURL)
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for t := range ticker.C {
			mutex.Lock()
			timestamp = t.Format(time.RFC3339)
			randomString = uuid.New().String()
			mutex.Unlock()
			fmt.Printf("%s: %s\n", timestamp, randomString)
		}
	}()

	go func() {
		http.HandleFunc("/", homeHandler)
		port := "8080"
		log.Printf("Server started on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	go func() {
		http.HandleFunc("/healthz", health)
		port := "3541"
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed to start healthz endpoint: %v", err)
		}
	}()

	select {}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	logEntry := fmt.Sprintf("%s: %s\n", timestamp, randomString)

	var counter Counter
	err := db.Get(&counter, "SELECT counter FROM counter LIMIT 1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching counter from database: %v", err), http.StatusInternalServerError)
		return
	}

	fileContent, err := reader("/config/information.txt")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading HTTP response: %v", err), http.StatusInternalServerError)
		return
	}
	messageEnvVar := os.Getenv("MESSAGE")
	if messageEnvVar == "" {
		messageEnvVar = "MESSAGE environment variable not set"
	}

	fmt.Fprintln(w, "file content: "+fileContent+"\n"+"env variable: MESSAGE="+messageEnvVar+"\n"+logEntry+"\n"+fmt.Sprintf("Counter: %d", counter.Counter))
}

func reader(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return lastLine, nil
}

func health(w http.ResponseWriter, r *http.Request) {
	err := db.Ping()
	if err != nil {
		http.Error(w, "Failed to connect to database.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
