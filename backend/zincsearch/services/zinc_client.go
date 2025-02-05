package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"zincsearch-backend/models"
)

const (
	zincBaseURL  = "http://localhost:4080/es" // Base URL de ZincSearch
	zincUser     = "admin"                    // Usuario
	zincPassword = "Complexpass#123"          // Contraseña
)

// FetchDocuments consulta los documentos de un índice en ZincSearch.
// FetchDocuments consulta los documentos de un índice en ZincSearch con paginación.
func FetchDocuments(index string, from, size int, searchText string) (*models.ZincResponse, error) {
	url := fmt.Sprintf("%s/%s/_search", zincBaseURL, index)

	// Crear la consulta básica
	var query interface{}

	if searchText == "" {
		// Crear la consulta básica
		query = models.ZincRequest{
			From: from,
			Size: size,
			Sort: []models.SortField{
				{
					Id: models.Order{
						Order: "asc",
					},
				},
			},
		}
	} else {

		query = models.ZincBodyRequest{
			From: from,
			Size: size,
			Sort: []models.SortField{
				{
					Id: models.Order{
						Order: "asc",
					},
				},
			},
			Query: models.MatchPhraseQuery{
				MatchPhrase: struct {
					Body string `json:"body"`
				}{
					Body: "*" + searchText + "*",
				},
			},
		}
	}
	// Convertir a JSON
	body, err := json.Marshal(query)
	fmt.Println(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	// Realizar la solicitud HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	auth := zincUser + ":" + zincPassword
	bas64encoded_creds := base64.StdEncoding.EncodeToString([]byte(auth))

	//req.SetBasicAuth(zincUser, zincPassword)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+bas64encoded_creds)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from ZincSearch: %s , %v", resp.Status, data)
	}

	// Leer la respuesta
	//data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Decodificar la respuesta JSON
	var zincResponse models.ZincResponse
	if err := json.Unmarshal(data, &zincResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &zincResponse, nil
}
