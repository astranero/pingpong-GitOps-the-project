package main


import (
	"fmt"
	"log"
	"os"
	"sync"
	"net/http"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)


var (
	db *sqlx.DB 
	mutex sync.Mutex
)


type Counter struct {
    Counter int `db:"counter"`
}



func main(){
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
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS counter (
		counter INTEGER NOT NULL DEFAULT 0
	)`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create counter table: %v", err)
	}

	ensureCounterExistsQuery := `
	INSERT INTO counter (counter)
	SELECT 0
	WHERE NOT EXISTS (SELECT 1 FROM counter)`
	_, err = db.Exec(ensureCounterExistsQuery)
	if err != nil {
		log.Fatalf("Failed to ensure counter exists: %v", err)
	}

	fmt.Println("Counter table is ready")

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully connected")
	}

	go func(){
		http.HandleFunc("/pingpong", pingHandler)

		port := "8081"
		log.Printf("Server started on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	go func(){
		http.HandleFunc("/healthz", health)
		port := "3541"
		if err:= http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed to start server:%v", err)
		}
	}()

	select{}
}

func pingHandler(w http.ResponseWriter, r *http.Request){
	mutex.Lock()
	defer mutex.Unlock()
	incrementQuery := `UPDATE counter SET counter = counter + 1 RETURNING counter`
	var counter Counter
	err := db.Get(&counter, incrementQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error incrementing counter: %v", err), http.StatusInternalServerError)
		return 
	}

	fmt.Fprintf(w,"New value for counter: %d\n", counter.Counter)
}

func health(w http.ResponseWriter, r *http.Request){
	var err error
	err = db.Ping()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"OK")
}