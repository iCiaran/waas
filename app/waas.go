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
	"time"
)

func wotpp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("data")
	if err != nil {
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	f, err := ioutil.TempFile("", "wotpp")
	if err != nil {
		http.Error(w, "Error creating temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(f.Name())

	io.Copy(f, file)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "w++", f.Name())

	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		http.Error(w, "Timeout exceeded", http.StatusBadRequest)
		return
	}

	if err != nil {
		if t, ok := err.(*exec.ExitError); ok {
			http.Error(w, string(t.Stderr), http.StatusInternalServerError)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.out", header.Filename))
	fmt.Fprintf(w, "%s", out)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api", wotpp)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
