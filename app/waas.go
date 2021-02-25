package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func writeLog(status int, message string, r *http.Request) {
	ip := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]

	// If the X-Forwarded-For header is not filled fall back to remote address
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	log.Printf("[%d] - %-15s - %s", status, ip, message)
}

func wotpp(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != "POST" {
		writeLog(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), r)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("data")
	if err != nil {
		writeLog(http.StatusBadRequest, "No data field in request", r)
		http.Error(w, "No data field in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	f, err := ioutil.TempFile("", "wotpp")
	if err != nil {
		writeLog(http.StatusInternalServerError, "Error creating temporary file", r)
		http.Error(w, "Error creating temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(f.Name())

	io.Copy(f, file)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "w++", "-i", f.Name())

	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		writeLog(http.StatusBadRequest, "Timeout exceeded", r)
		http.Error(w, "Timeout exceeded", http.StatusBadRequest)
		return
	}

	if err != nil {
		if t, ok := err.(*exec.ExitError); ok {
			writeLog(http.StatusBadRequest, strings.ReplaceAll(string(t.Stderr), "\n", " "), r)
			http.Error(w, string(t.Stderr), http.StatusBadRequest)
		} else {
			writeLog(http.StatusInternalServerError, "w++ failed with non ExitError", r)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	stat, _ := f.Stat()
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.out", header.Filename))
	writeLog(http.StatusOK, fmt.Sprintf("%s %db %s", f.Name(), stat.Size(), time.Since(startTime)), r)
	fmt.Fprintf(w, "%s", out)
}

func main() {
	// Set up logging
	currentTime := time.Now().Format("06-01-02_150405")
	file, err := os.OpenFile(fmt.Sprintf("%s%s%s", "log/", currentTime, ".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	log.SetOutput(file)
	defer log.Print("Server stopped")
	log.Print("Server started")

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api", wotpp)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
