package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/getsops/sops/v3/decrypt"
)

const (
	defaultAddr = "localhost"
	defaultPort = "8181"

	headerKey = "X-Secret-Request"
	headerVal = "true"

	sopsFilePath = "/app/secrets.env"

	notFoundErr = "not found"
)

// decryptSecret fetches a specific key from the SOPS encrypted ENV file
func decryptSecret(keyToFind string) (string, error) {
	secrets := map[string]string{}

	data, err := decrypt.File(sopsFilePath, "env")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes if present
			value = strings.Trim(value, "\"")
			secrets[key] = value
		} else {
			log.Printf("Skipping a malformed line in decrypted ENV file")
		}
	}

	if value, ok := secrets[keyToFind]; !ok {
		return "", fmt.Errorf("key '%s' %s in decrypted secrets file", keyToFind, notFoundErr)
	} else {
		return value, nil
	}
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(headerKey) != headerVal {
		http.Error(w, "Missing or invalid header", http.StatusUnauthorized)
		return
	}

	key := strings.Trim(r.URL.Path, "/")
	if key == "" {
		http.Error(w, "Secret key not specified in path", http.StatusNotFound)
		return
	}

	val, err := decryptSecret(key)
	if err != nil {
		log.Printf("Failed to get secret for key '%s': %v", key, err)
		if strings.Contains(err.Error(), notFoundErr) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Failed to retrieve secret", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, val)
}

func main() {
	port := os.Getenv("SECRETS_PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/", secretHandler)

	listenAddr := fmt.Sprintf("%s:%s", defaultAddr, port)
	log.Printf("Secrets manager listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
