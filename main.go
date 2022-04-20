package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"log"
	"mime"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
)

func taskHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/api/v1/remote-execution" {
		if req.Method == http.MethodPost {
			createTaskHandler(w, req)
		} else {
			http.Error(w, fmt.Sprintf("expect method POST at /task/, got %v", req.Method), http.StatusMethodNotAllowed)
			return
		}
	} else {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func WinShellExe(strCommand string) (out string, err string) {
	argsCommand := strings.Split(strCommand, " ")
	cmd := exec.Command("powershell", argsCommand...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, _ := cmd.Output()
	d := charmap.CodePage866.NewDecoder()
	decodeOut, _ := d.Bytes(stdout)
	out = string(decodeOut)
	return
}

//func LinuxShellExe(strCommand string) (out string) {
//	argsCommand := strings.Split(strCommand, " ")
//	cmd := exec.Command("bash", argsCommand...)
//	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
//	stdout, _ := cmd.Output()
//	d := charmap.CodePage866.NewDecoder()
//	decodeOut, _ := d.Bytes(stdout)
//	out = string(decodeOut)
//	return
//}

func createTaskHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling task create at %s\n", req.URL.Path)

	type RequestTask struct {
		Cmd   string `json:"cmd"`
		Os    string `json:"os"`
		Stdin string `json:"stdin"`
	}

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var rt []RequestTask
	if err := dec.Decode(&rt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type Answer struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
	}
	var answer Answer
	if rt[0].Os == "windows" {
		answer.Stdout, answer.Stderr = WinShellExe(rt[0].Cmd)
	} else {
		answer.Stderr = "Incorrect OS"
	}
	js, _ := json.Marshal(answer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	cert, _ := tls.LoadX509KeyPair("localhost.crt", "localhost.key")
	s := &http.Server{
		Addr:    ":8085",
		Handler: nil,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
	log.Printf("Server started")
	http.HandleFunc("/api/v1/remote-execution", taskHandler)
	log.Fatal(s.ListenAndServeTLS("", ""))
}
