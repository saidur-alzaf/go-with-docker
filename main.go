package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

type Item struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db"
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := initSchema(); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	mux := http.NewServeMux()

	// 1. Health check
	mux.HandleFunc("GET /health", healthHandler)

	// 2. List all items
	mux.HandleFunc("GET /items", listItemsHandler)

	// 3. Create an item
	mux.HandleFunc("POST /items", createItemHandler)

	// 4. Get a single item
	mux.HandleFunc("GET /items/{id}", getItemHandler)

	// 5. Update an item
	mux.HandleFunc("PUT /items/{id}", updateItemHandler)

	// 6. Delete an item
	mux.HandleFunc("DELETE /items/{id}", deleteItemHandler)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("listening on %s (db: %s)", addr, dbPath)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func initSchema() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func listItemsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, name, created_at FROM items ORDER BY id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Name, &it.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		items = append(items, it)
	}
	writeJSON(w, http.StatusOK, items)
}

func createItemHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}

	res, err := db.Exec(`INSERT INTO items (name) VALUES (?)`, input.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, _ := res.LastInsertId()

	var it Item
	err = db.QueryRow(`SELECT id, name, created_at FROM items WHERE id = ?`, id).
		Scan(&it.ID, &it.Name, &it.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, it)
}

func getItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var it Item
	err = db.QueryRow(`SELECT id, name, created_at FROM items WHERE id = ?`, id).
		Scan(&it.ID, &it.Name, &it.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, errors.New("item not found"))
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func updateItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}

	res, err := db.Exec(`UPDATE items SET name = ? WHERE id = ?`, input.Name, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		writeError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}

	var it Item
	err = db.QueryRow(`SELECT id, name, created_at FROM items WHERE id = ?`, id).
		Scan(&it.ID, &it.Name, &it.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	res, err := db.Exec(`DELETE FROM items WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		writeError(w, http.StatusNotFound, errors.New("item not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
