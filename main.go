package main

import (
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "os"
    "strconv"
    "sync"
    "time"
)

type Item struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

var (
    items   []Item
    itemsMu sync.Mutex
)

func main() {
    addr := os.Getenv("ADDR")
    if addr == "" {
        if port := os.Getenv("PORT"); port != "" {
            addr = ":" + port
        } else {
            addr = ":8080"
        }
    }

    mux := http.NewServeMux()
    // Health check
    mux.HandleFunc("/health", healthHandler)
    // List items
    mux.HandleFunc("GET /items", listItemsHandler)
    // Create an item
    mux.HandleFunc("POST /items", createItemHandler)
    // Get a single item
    mux.HandleFunc("GET /items/{id}", getItemHandler)
    // Update an item
    mux.HandleFunc("PUT /items/{id}", updateItemHandler)
    // Delete an item
    mux.HandleFunc("DELETE /items/{id}", deleteItemHandler)

    log.Printf("listening on %s", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatal(err)
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func listItemsHandler(w http.ResponseWriter, r *http.Request) {
    itemsMu.Lock()
    copyItems := make([]Item, len(items))
    copy(copyItems, items)
    itemsMu.Unlock()
    writeJSON(w, http.StatusOK, copyItems)
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
    itemsMu.Lock()
    newID := int64(len(items) + 1)
    it := Item{ID: newID, Name: input.Name, CreatedAt: time.Now()}
    items = append(items, it)
    itemsMu.Unlock()
    writeJSON(w, http.StatusCreated, it)
}

func getItemHandler(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, errors.New("invalid id"))
        return
    }
    itemsMu.Lock()
    defer itemsMu.Unlock()
    for _, it := range items {
        if it.ID == id {
            writeJSON(w, http.StatusOK, it)
            return
        }
    }
    writeError(w, http.StatusNotFound, errors.New("item not found"))
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
    itemsMu.Lock()
    defer itemsMu.Unlock()
    for i, it := range items {
        if it.ID == id {
            items[i].Name = input.Name
            writeJSON(w, http.StatusOK, items[i])
            return
        }
    }
    writeError(w, http.StatusNotFound, errors.New("item not found"))
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, errors.New("invalid id"))
        return
    }
    itemsMu.Lock()
    defer itemsMu.Unlock()
    for i, it := range items {
        if it.ID == id {
            items = append(items[:i], items[i+1:]...)
            w.WriteHeader(http.StatusNoContent)
            return
        }
    }
    writeError(w, http.StatusNotFound, errors.New("item not found"))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
    writeJSON(w, status, map[string]string{"error": err.Error()})
}
