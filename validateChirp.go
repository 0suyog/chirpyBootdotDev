package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {
	maxChirpLen := 140
	profaneList := [3]string{"kerfuffle", "sharbert", "formax"}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("%v", err)
		respondWithError(w, 400, err.Error())
		return
	}

	var chirp struct {
		Body string `json:"body"`
	}

	err = json.Unmarshal(data, &chirp)
	if err != nil {
		log.Printf("%v", err)
		respondWithError(w, 400, err.Error())
		return
	}

	if l := len(chirp.Body); l > maxChirpLen {
		log.Printf("Chirp too long. Chirp: %s MaxLength: %d", chirp.Body, maxChirpLen)
		respondWithError(w, 400, "Chirp too long")
	}

	splittedChirp := strings.Split(chirp.Body, " ")
	for _, profaneWord := range profaneList {
		for i, word := range splittedChirp {
			if strings.EqualFold(word, profaneWord) {
				splittedChirp[i] = "****"
			}
		}
	}
	cleanedBody := strings.Join(splittedChirp, " ")
	respondWithJson(w, http.StatusAccepted, map[string]string{"cleaned_word": cleanedBody})

}
