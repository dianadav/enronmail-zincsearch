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

	// Ruta base de la carpeta 'maildir'
	//dir := "C:/Users/diana/OneDrive/Escritorio/ProyectoEnron/enron_mail_20110402/maildir/"
	dir := "C:/Users/diana/OneDrive/Escritorio/ProyectoEnron/test/" // tres carpetas de prueba
	outputFile := "enron_emails.ndjson"

	// Crear el archivo de salida NDJSON
	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creando el archivo de salida:", err)
		return
	}
	defer out.Close()

	writer := bufio.NewWriter(out)

	// Contador para asignar IDs únicos a los correos
	var emailID int

	// Recorre todas las subcarpetas y archivos dentro de 'maildir'
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}
		if !info.IsDir() { // Procesar solo archivos, no carpetas
			// Leer el contenido del archivo

			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// Procesar el contenido del archivo
			email, err := parseEmail(string(content))
			if err != nil {
				fmt.Println("Error procesando el archivo:", path, "-", err)
				return nil
			}

			// Asignar ID único al correo
			emailID++
			email.ID = emailID

			// Convertir el objeto Email a JSON
			jsonData, err := json.Marshal(email)
			if err != nil {
				return err
			}

			// Escribir la línea de metadatos de ZincSearch seguida del JSON del correo
			_, err = writer.WriteString(fmt.Sprintf("{ \"index\": { \"_index\": \"emails\" } }\n%s\n", jsonData))
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error recorriendo la carpeta:", err)
		return
	}

	writer.Flush()
	fmt.Println("Conversión a NDJSON completada. Archivo de salida:", outputFile)
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

// parseEmail extrae las etiquetas y el cuerpo de un correo
func parseEmail(content string) (*Email, error) {
	email := &Email{}
	lines := strings.Split(content, "\n")

	var bodyLines []string
	for _, line := range lines {
		// Ignorar líneas vacías
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Verificar si la línea es una etiqueta
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
			// Si no es una etiqueta, es parte del cuerpo
			bodyLines = append(bodyLines, line)
		}
	}

	// Unir las líneas del cuerpo
	email.Body = strings.Join(bodyLines, "\n")
	return email, nil
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
