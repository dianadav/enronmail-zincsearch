package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
)

// Estructura del correo
type Email struct {
	ID                      int    `json:"id"`
	MessageID               string `json:"message_id"`
	Date                    string `json:"date"`
	From                    string `json:"from"`
	To                      string `json:"to"`
	Subject                 string `json:"subject"`
	MimeVersion             string `json:"mime_version"`
	ContentType             string `json:"content_type"`
	ContentTransferEncoding string `json:"content_transfer_encoding"`
	XFrom                   string `json:"x_from"`
	XTo                     string `json:"x_to"`
	XCc                     string `json:"x_cc"`
	XBcc                    string `json:"x_bcc"`
	XFolder                 string `json:"x_folder"`
	XOrigin                 string `json:"x_origin"`
	XFileName               string `json:"x_file_name"`
	Cc                      string `json:"cc"`
	Body                    string `json:"body"`
}

const (
	zincBaseURL  = "http://localhost:4080" // Base URL de ZincSearch
	zincUser     = "admin"                 // Usuario
	zincPassword = "Complexpass#123"       // Contraseña
	index        = "emails2"
	outputFile   = "enron_emails.ndjson"
	dir          = "C:/Users/diana/OneDrive/Escritorio/ProyectoEnron/enron_mail_20110402/maildir/"
	//dir = "C:/Users/diana/OneDrive/Escritorio/ProyectoEnron/test/" // tres carpetas de prueba
)

func main() {
	////Prceso de rendimiento de la aplicación/////////////
	cpu, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer cpu.Close() // Cierra el archivo al finalizar
	pprof.StartCPUProfile(cpu)
	defer pprof.StopCPUProfile()

	// Habilitar el perfil de bloqueos (captura contenciones de sincronización)
	runtime.SetBlockProfileRate(1) // Habilita la captura de bloqueos (1 = captura todos los eventos)
	block, err := os.Create("block.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer block.Close() // Cierra el archivo del perfil de bloqueos al finalizar

	// Habilitar el perfil de mutex (captura contenciones de bloqueos de mutex)
	runtime.SetMutexProfileFraction(1) // Habilita la captura de mutex (1 = captura todos los eventos)
	mutex, err := os.Create("mutex.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer mutex.Close() // Cierra el archivo del perfil de mutex al finalizar

	////////Fin prceso de rendimiento de la aplicación/////////

	// Crear el archivo de salida NDJSON
	out, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creando el archivo de salida: %v", err)
	}
	defer out.Close()

	writer := bufio.NewWriterSize(out, 1<<20) // 1MB

	// Canal para pasar los archivos a procesar
	filesChan := make(chan string, runtime.NumCPU()*2)
	// Canal para recibir los correos procesados
	resultsChan := make(chan *Email, runtime.NumCPU()*2)

	// Grupo de espera para sincronizar las goroutines
	var wg sync.WaitGroup

	// Lanzar workers
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(filesChan, resultsChan, &wg)
	}

	// Goroutine para cerrar el canal de resultados cuando los workers terminen
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Lanzar la goroutine para escribir los resultados al archivo
	go func() {
		var emailID int
		for email := range resultsChan {
			emailID++
			email.ID = emailID
			// Convertir el objeto Email a JSON y escribirlo en el archivo
			jsonData, err := json.Marshal(email)
			if err == nil {
				//postEmail(jsonData)
				_, _ = writer.WriteString(fmt.Sprintf("{ \"index\": { \"_index\": \"emails\" } }\n%s\n", jsonData))
			}
		}
		writer.Flush()
		fmt.Println("Se logró completar indexar todos los archivos")
	}()

	// Recorrer los archivos y enviarlos al canal
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filesChan <- path
		}
		return nil
	})
	close(filesChan) // Cerrar el canal para indicar que no habrá más archivos

	if err != nil {
		log.Fatalf("Error recorriendo la carpeta: %v", err)
	}
	////Proceso de rendimiento de la aplicación/////////////
	runtime.GC()
	mem, err := os.Create("memory.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer mem.Close()
	if err := pprof.WriteHeapProfile(mem); err != nil {
		log.Fatal(err)
	}
	// Crear el perfil de goroutines (captura el estado actual de las goroutines activas)
	goroutines, err := os.Create("goroutines.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer goroutines.Close()                         // Cierra el archivo del perfil de goroutines al finalizar
	pprof.Lookup("goroutine").WriteTo(goroutines, 0) // Guarda los datos de las goroutines activas

	// Crear el perfil de bloqueos y escribirlo al archivo
	pprof.Lookup("block").WriteTo(block, 0) // Guarda los datos de los bloqueos en el programa

	// Crear el perfil de mutex y escribirlo al archivo
	pprof.Lookup("mutex").WriteTo(mutex, 0) // Guarda los datos de contención de mutex

	////Fin proceso de rendimiento de la aplicación/////////////

}

// Función worker que procesa los archivos
func worker(filesChan <-chan string, resultsChan chan<- *Email, wg *sync.WaitGroup) {
	defer wg.Done()
	for path := range filesChan {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Error leyendo archivo %s: %v", path, err)
			continue
		}

		email, err := parseEmail(string(content))
		if err != nil {
			log.Printf("Error procesando archivo %s: %v", path, err)
			continue
		}

		resultsChan <- email
	}
}

// parseEmail extrae las etiquetas y el cuerpo de un correo
func parseEmail(content string) (*Email, error) {
	email := &Email{}
	lines := strings.Split(content, "\n")

	var bodyLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			bodyLines = append(bodyLines, line)
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Mapear las etiquetas conocidas al struct Email
			switch key {
			case "Message-ID":
				if email.MessageID == "" {
					email.MessageID = value
				}

			case "Date":
				if email.Date == "" {
					email.Date = value
				}
			case "From":
				if email.From == "" {
					email.From = value
				}

			case "To":
				if email.To == "" {
					email.To = value
				}

			case "Subject":
				if email.Subject == "" {
					email.Subject = value
				}

			case "Mime-Version":
				if email.MimeVersion == "" {
					email.MimeVersion = value
				}

			case "Content-Type":
				if email.ContentType == "" {
					email.ContentType = value
				}

			case "Content-Transfer-Encoding":
				if email.ContentTransferEncoding == "" {
					email.ContentTransferEncoding = value
				}

			case "X-From":
				if email.XFrom == "" {
					email.XFrom = value
				}

			case "X-To":
				if email.XTo == "" {
					email.XTo = value
				}

			case "X-cc":
				if email.XCc == "" {
					email.XCc = value
				}

			case "X-bcc":
				if email.XBcc == "" {
					email.XBcc = value
				}

			case "X-Folder":
				if email.XFolder == "" {
					email.XFolder = value
				}

			case "X-Origin":
				if email.XOrigin == "" {
					email.XOrigin = value
				}

			case "X-FileName":
				if email.XFileName == "" {
					email.XFileName = value
				}

			case "Cc":
				if email.Cc == "" {
					email.Cc = value
				}
			default:
				bodyLines = append(bodyLines, line)
			}
		} else {
			bodyLines = append(bodyLines, line)
		}
	}

	email.Body = strings.Join(bodyLines, "\n")
	return email, nil
}

// Esta función realiza una petición y envía los datos.
func postEmail(emailJson []byte) {
	auth := zincUser + ":" + zincPassword
	bas64encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	zincUrl := zincBaseURL + "/api/" + index + "/_doc"

	req, err := http.NewRequest("POST", zincUrl, bytes.NewBuffer(emailJson))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}
	// Set headers

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+bas64encodedAuth)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

}

//go tool pprof -http=:8091 cpu.prof
//go tool pprof -http=:8092 memory.prof
//go tool pprof -http=:8093 goroutines.prof
//go tool pprof -http=:8094 block.prof
//go tool pprof -http=:8095 mutex.prof

//go tool pprof -http=:8080 --nodefraction=0.01 --edgefraction=0.01 cpu.prof
//go tool pprof -svg cpu.prof > cpu_graph.svg
//go tool pprof -pdf cpu.prof > cpu_graph.pdf

//curl http://localhost:4080/api/_bulk -i -u admin:Complexpass#123 --data-binary "@C:/Users/diana/OneDrive/Escritorio/ProyectoEnron/indexar2/enron_emails.ndjson"
//curl http://localhost:4080/api/_bulk -i -u admin:Complexpass#123 --data-binary "@C:/Users/diana/OneDrive/Escritorio/ProyectoEnron/indexar2/enron_emails.ndjson"
//C:\dev\zinc-search\release

//curl http://localhost:4080/api/_bulk -i -u admin:Complexpass#123 --data-binary "@C:/Users/diana/OneDrive/Documentos/GitHub/EnronMail-ZincSearch/indexar/emails1.ndjson"
//C:\dev\zinc-search\release

//curl http://localhost:4080/api/_bulk -i -u admin:Complexpass#123 --data-binary "@/home/diana/Documents/indexar/emails2.ndjson"
