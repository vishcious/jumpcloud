package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	r := &Router{
		getNextId: getIncrementer(0),
		container: HashContainer{
			hashes: make(map[uint64]Result),
			stats:  Stats{Total: 0, Average: 0},
		},
	}

	http.ListenAndServe(":8080", r)
}

type Router struct {
	getNextId func() uint64
	container HashContainer
}

func (sr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.URL.Path == "/hash" {
		generateNewHash(w, r, sr.getNextId(), &sr.container)
		return
	}
	if r.Method == "GET" && r.URL.Path == "/stats" {
		getStats(w, r, &sr.container)
		return
	}

	if r.Method == "GET" {
		rexp := regexp.MustCompile(`^/hash/(\d+)$`)
		match := rexp.FindStringSubmatch(r.URL.Path)
		if len(match) == 2 {
			id, err := strconv.ParseUint(match[1], 10, 64)
			if err == nil {
				getHash(w, r, id, &sr.container)
				return
			}
		}
	}

	// Return 404 for every request
	http.NotFound(w, r)
}

func generateNewHash(w http.ResponseWriter, r *http.Request, id uint64, container *HashContainer) {
	start := time.Now()
	go func() {
		// This just parks the goroutine and does not consume any CPU cycles
		// https://stackoverflow.com/questions/32147421/behavior-of-sleep-and-select-in-go
		// https://xwu64.github.io/2019/02/27/Understanding-Golang-sleep-function/
		time.Sleep(5 * time.Second)
		hash, err := hashPassword(r.FormValue("password"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}
		container.setHashById(id, hash)
		container.trackTime(start)
	}()
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "%d", id)
}

func getHash(w http.ResponseWriter, r *http.Request, id uint64, container *HashContainer) {
	hash, err := container.getHashById(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if hash.err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", hash.hash)
	w.WriteHeader(http.StatusOK)
}

func getStats(w http.ResponseWriter, r *http.Request, container *HashContainer) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(container.stats)
	w.WriteHeader(http.StatusOK)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
