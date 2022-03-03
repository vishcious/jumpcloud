package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := &Router{
		getNextId:     GetIncrementer(0),
		container:     *CreateHashProcessor(),
		stopListening: cancel,
	}

	server := &http.Server{Addr: ":8080", Handler: r}
	fmt.Println("Starting server")
	go server.ListenAndServe()
	go r.container.StartAcceptingWork()
	// Wait for cancel to be called on http server context
	<-ctx.Done()
	// Stop listening to http requests
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
	// Wait for workers to be shutdown
	// worker currently processing will be waited on to finish
	r.container.worker.Wait()
	fmt.Println("Shutting down server")
}

type Router struct {
	getNextId     func() uint64
	container     HashContainer
	stopListening context.CancelFunc
}

func (sr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Here goes the simplest router I came up with without using external packages
	if r.Method == "POST" && r.URL.Path == "/hash" {
		generateNewHash(w, r, sr.getNextId(), &sr.container)
		return
	}
	if r.Method == "GET" && r.URL.Path == "/stats" {
		getStats(w, r, &sr.container)
		return
	}
	if r.Method == "POST" && r.URL.Path == "/shutdown" {
		// Stop listening on web server
		sr.stopListening()
		// Stop the worker from accepting any more jobs
		// The web server should be stopped before stopping workers to ensure jobs are not added to a closed channel
		sr.container.StopAcceptingWork()
		w.WriteHeader(http.StatusOK)
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
	container.DoWork(&HashRequest{id: id, password: r.FormValue("password")})
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "%d", id)
}

func getHash(w http.ResponseWriter, r *http.Request, id uint64, container *HashContainer) {
	hash, err := container.GetHashById(id)
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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(container.stats)
}
