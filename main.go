package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"0suyog/chirpyBootdotDev.git/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
}

type user struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileServerHits.Add(1)
		log.Printf("fileserver hit %d times", c.fileServerHits.Load())
		w.Header().Add("cache-control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "text/html")
	template := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`
	fmt.Fprintf(w, template, c.fileServerHits.Load())
}

func (c *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

	data, err := io.ReadAll(r.Body)

	if err != nil {
		respondWithError(w, 500, err.Error())
	}

	type ReqBody struct {
		Email string `json:"email"`
	}

	var reqBody ReqBody
	err = json.Unmarshal(data, &reqBody)

	if err != nil {
		respondWithError(w, 500, err.Error())
	}

	user, err := c.db.CreateUser(r.Context(), sql.NullString{String: reqBody.Email, Valid: true})
	if err != nil {
		respondWithError(w, 500, err.Error())
	}
	fmt.Println(user)

	respondWithJson(w, 200, user)

}

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {
	godotenv.Load()
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic("Couldn't connect to database")
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		db: dbQueries,
	}

	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	mux.HandleFunc("POST /api/user", apiCfg.createUser)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	log.Printf("Serving file from . to port %s ", server.Addr)

	err = server.ListenAndServe()

	if err != nil {
		return
	}
}
