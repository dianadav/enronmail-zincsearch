package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"zincsearch-backend/services"
)

func GetDocuments(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	index := r.URL.Query().Get("index") // El índice a consultar
	if index == "" {
		http.Error(w, "index query parameter is required", http.StatusBadRequest)
		return
	}

	// Leer parámetros de paginación
	fromStr := r.URL.Query().Get("from")
	sizeStr := r.URL.Query().Get("size")
	searchStr := r.URL.Query().Get("search")

	fmt.Println(searchStr)
	// Convertir a enteros (con valores predeterminados)
	from, err := strconv.Atoi(fromStr)
	if err != nil || from < 0 {
		from = 0
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		size = 10
	}

	// Llamar al cliente ZincSearch
	documents, err := services.FetchDocuments(index, from, size, searchStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}
