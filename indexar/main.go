package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

func main() {
	// Inicialización del perfil de rendimiento
	cpu, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer cpu.Close()
	pprof.StartCPUProfile(cpu)
	defer pprof.StopCPUProfile()

	// Habilitar el perfil de bloqueos
	runtime.SetBlockProfileRate(1)
	block, err := os.Create("block.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer block.Close()

	// Habilitar el perfil de mutex
	runtime.SetMutexProfileFraction(1)
	mutex, err := os.Create("mutex.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer mutex.Close()

	// Ruta base de la carpeta 'maildir'
	dir := "C:/Users/diana/OneDrive/Escritorio/ProyectoEnron/test/" // Tres carpetas de prueba
	outputFile := "enron_emails.ndjson"

	// Crear el archivo de salida NDJSON
	out, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creando el archivo de salida: %v", err)
	}
	defer out.Close()

	// Buffer de escritor de 1MB
	writer := bufio.NewWriterSize(out, 1<<20) // 1MB

	// Canal para pasar los archivos a procesar
	filesChan := make(chan string, runtime.NumCPU()*4)
	// Canal para recibir los correos procesados
	resultsChan := make(chan *Email, runtime.NumCPU()*4)

	// Grupo de espera para sincronizar las goroutines
	var wg sync.WaitGroup

	// Pool de workers
	numWorkers := runtime.NumCPU() * 4 // Ajusta según necesidades
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(filesChan, resultsChan, &wg)
	}

	// Goroutine para cerrar el canal de resultados cuando los workers terminen
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Goroutine para escribir los resultados al archivo en lotes
	go func() {
		batchSize := 1000 // Tamaño del lote
		var batch []*Email
		var emailID int

		for email := range resultsChan {
			emailID++
			email.ID = emailID
			batch = append(batch, email)

			if len(batch) >= batchSize {
				writeBatch(writer, batch)
				batch = nil
			}
		}

		// Escribir el último lote si queda algo pendiente
		if len(batch) > 0 {
			writeBatch(writer, batch)
		}

		writer.Flush()
		fmt.Println("Conversión a NDJSON completada. Archivo de salida:", outputFile)
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

	// Crear el perfil de goroutines
	goroutines, err := os.Create("goroutines.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer goroutines.Close()
	pprof.Lookup("goroutine").WriteTo(goroutines, 0)

	// Crear el perfil de bloqueos
	pprof.Lookup("block").WriteTo(block, 0)

	// Crear el perfil de mutex
	pprof.Lookup("mutex").WriteTo(mutex, 0)

	////Fin proceso de rendimiento de la aplicación/////////////
}

// Función worker que procesa los archivos
func worker(filesChan <-chan string, resultsChan chan<- *Email, wg *sync.WaitGroup) {
	defer wg.Done()

	// Pool para reutilizar estructuras Email
	emailPool := sync.Pool{
		New: func() interface{} {
			return &Email{}
		},
	}

	for path := range filesChan {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Error leyendo archivo %s: %v", path, err)
			continue
		}

		email := emailPool.Get().(*Email)
		email, err = parseEmail(string(content), email)
		if err != nil {
			log.Printf("Error procesando archivo %s: %v", path, err)
			emailPool.Put(email) // Devolver al pool si hay error
			continue
		}

		resultsChan <- email
	}
}

// parseEmail extrae las etiquetas y el cuerpo de un correo
func parseEmail(content string, email *Email) (*Email, error) {
	// Reiniciar el struct Email
	*email = Email{}

	lines := strings.Split(content, "\n")
	var bodyLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
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

// Escribe un lote de correos en el archivo NDJSON
func writeBatch(writer *bufio.Writer, batch []*Email) {
	for _, email := range batch {
		jsonData, err := json.Marshal(email)
		if err == nil {
			_, _ = writer.WriteString(fmt.Sprintf("{ \"index\": { \"_index\": \"emails\" } }\n%s\n", jsonData))
		}
	}
}

//go tool pprof -http=:8091 cpu.prof
//go tool pprof -http=:8092 memory.prof
//go tool pprof -http=:8093 goroutines.prof
//go tool pprof -http=:8094 block.prof
//go tool pprof -http=:8095 mutex.prof

//go tool pprof -http=:8080 --nodefraction=0.01 --edgefraction=0.01 cpu.prof
//go tool pprof -svg cpu.prof > cpu_graph.svg
//go tool pprof -pdf cpu.prof > cpu_graph.pdf
