package main

import (
	"fmt"
	"os"
	"log"
	http "net/http"
	"strconv"
	"encoding/json"
	"strings"
	"flag"
)

const PORT = 8080

type apiConfig struct {
	fileserverHits int
}

var apiCfg apiConfig

var glog = log.Default()

type ErrorResponse struct {
	Error string `json:"error"`
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Add("Hits",strconv.Itoa(cfg.fileserverHits))
		w.WriteHeader(200)
		var bodyTemplate = `<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>`
		var body = fmt.Sprintf(bodyTemplate, cfg.fileserverHits)
		w.Write([]byte(body))
		//w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits)))
	default:
		w.Header().Add("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(200)
	glog.Println("Metrics reset to 0")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits += 1
		glog.Printf("Currents hits: %d", cfg.fileserverHits)
		next.ServeHTTP(w, r)
	})
}

func cleanProfane(str string) string {
	banned_words := []string{"kerfuffle","sharbert","fornax"}
	split_str := strings.Split(str, " ")
	for idx, word := range split_str {
		for _,banned := range banned_words {
			if strings.ToLower(word) == banned {
				split_str[idx] = "****"
				break
			}
		}
	}
	return strings.Join(split_str, " ")
}

func main() {
	// CLI Argument parsing
	debugFlag := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	/*
	 * Setup
	 */
	mux := http.NewServeMux()	
	server := http.Server{ 
		Addr: fmt.Sprintf("0.0.0.0:%d", PORT),
		//Addr: "0.0.0.0:8080",
		Handler: mux,
	}
	// DB Connection
	db := DiskDB{}
	err := db.InitDB()
	if err != nil {
		glog.Printf("ERROR. Can't create DB connection: %w", err)
		os.Exit(-1)
	}
	defer db.Close()
	if *debugFlag {
		db.DropDB()
	}

	/* 
	 * ROUTES configuration
	 */
	path := http.Dir(".")
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app",http.FileServer(path))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type","text/plain; charset=utf-8")
			w.WriteHeader(200)
			w.Write([]byte("OK"))
		default:
			w.Header().Add("Allow", http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/reset", apiCfg.handleReset)
	mux.HandleFunc("GET /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		idParam := r.PathValue("id")	
		if idParam == "" {
			w.WriteHeader(400)
			w.Write([]byte("Couldn't parse ID"))
			return
		}
		id, err := strconv.Atoi(idParam)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Invalid ID, must be a number"))
			return
		}
		chirp, err := db.GetChirp(id)
		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte(fmt.Sprintf("Can't retrieve chirp with ID %d", id)))
			return
		}
		chirpJson, err := json.Marshal(chirp)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error serializing Chirp"))
			return
		}	
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(200)
		w.Write(chirpJson)
		return
	})
	mux.HandleFunc("/api/chirps", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			body, err := json.MarshalIndent(db.GetChirps(), "", "  ")
			if err != nil {
				w.WriteHeader(500)
				glog.Printf("ERROR. Can't create response body: %w", err)
				return
			}
			w.WriteHeader(200)
			//fmt.Printf("%s", body)
			w.Write(body)
		case "POST":
			decoder := json.NewDecoder(r.Body)
			glog.Printf("INFO. Request: %+w", r.Body)
			chirp := Chirp{}
			err := decoder.Decode(&chirp)
			if err != nil  {
				glog.Printf("ERROR. Can't decode parameters %s", err)
				w.Header().Set("Content-Type", "application/json")
				response,err := json.Marshal(ErrorResponse{Error: "Can't decode parameters"})
				if err != nil {
					glog.Printf("ERROR. Can't marshal response %s", err)
				}
				w.WriteHeader(500)
				w.Write(response)
				return
			}

			chirp_len := len(chirp.Body)
			log.Printf("INFO: Received chirp with length: %d, %s", chirp_len, chirp.Body)
			if chirp_len > 140 {
				w.Header().Set("Content-Type", "application/json")
				response,err := json.Marshal(ErrorResponse{Error: "Chirp is too long"})
				if err != nil {
					glog.Printf("ERROR. Can't marshal response %s", err)
				}
				w.WriteHeader(400)
				w.Write(response)
				return
			}

			cleaned_body := cleanProfane(chirp.Body)
			log.Printf("DEBUG. Cleaned body of profane words: %+v", cleaned_body)
			added_chirp, err := db.AddChirp(cleaned_body)
			log.Printf("DEBUG. Added chirp: %+v", added_chirp)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				response, err := json.Marshal(ErrorResponse{Error: "Couldn't create Chirp"})
				if err != nil {
					glog.Printf("ERROR. Can't marshal response %s", err)
				}
				w.WriteHeader(500)
				w.Write(response)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			response, err := json.Marshal(added_chirp)
			if err != nil {
				glog.Printf("ERROR. Can't marshal response %s", err)
			}
			w.Write(response)
		default:
			w.Header().Set("Allow", "GET, POST")
			w.WriteHeader(405)
			w.Write([]byte("Method Not Allowed\n\n"))
		}
		return
	})
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		user := User{}
		err := decoder.Decode(&user)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error decoding request body"))
			return
		}
		created_user, err := db.AddUser(user.Email)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error adding user"))
			return
		}
		userJson, err := json.Marshal(created_user)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error creating response body, user created"))
			return
		}
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(201)
		w.Write(userJson)
	})
	mux.HandleFunc("DELETE /api/db", func(w http.ResponseWriter, r *http.Request) {
		db.DropDB()
		w.WriteHeader(200)
		w.Write([]byte("Creared all data in database"))
	})

	mux.HandleFunc("/admin/metrics", apiCfg.handleMetrics)

	/*
	  LAUNCH server 
	*/
	fmt.Printf("Server prepared to listen on port %d\n", PORT)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("ERROR. Can't bind server: %+w\n", err)
	}
	fmt.Println("Finished server")
}
