package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

// Возможно следовало написать кастомную структуру содержащую и канал и реальную очередь
// В таком случае функционала было бы значительно больше, но поскольку в задании не указанно подобного
// То обошелся обычным каналом строк
var qs map[string]chan string

var defaultTimeout int = 30

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		println("Received GET request")
		getFromQueue(w, r)
	case "PUT":
		println("Received PUT request")
		putInQueue(w, r)
	default:
		println("Received unsupported request")
		_, err := fmt.Fprintf(w, "Method not supported")
		if err != nil {
			println(err)
			return
		}
	}

}

func getFromQueue(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write(nil)
		//Probably redundant code, but who cares
		if err != nil {
			return
		}
		return
	}

	timeout, _ := strconv.Atoi(r.Form.Get("timeout"))
	if timeout == 0 || timeout > 600 {
		timeout = defaultTimeout
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	queueName := path.Base(r.URL.Path)
	if _, exists := qs[queueName]; !exists {
		qs[queueName] = make(chan string)
	}

	select {
	//Wait until timeout
	case <-ctxTimeout.Done():
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write(nil)
		if err != nil {
			return
		}
	//Parse data from channel
	case result := <-qs[queueName]:
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(result))
		if err != nil {
			return
		}
	}

}

func putInQueue(w http.ResponseWriter, r *http.Request) {

	//Check for correct request form
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write(nil)
		if err != nil {
			return
		}
		return
	}

	queueName := path.Base(r.URL.Path)

	//Check for correct request args
	if v := r.Form.Get("v"); v != "" {

		//Run in goroutine because channel doesn't allow func to end
		go func() {
			if q, exists := qs[queueName]; exists {
				q <- v
			} else {
				//Else create queue
				qs[queueName] = make(chan string)
				qs[queueName] <- v
			}
		}()
		//If queue already exists just put new elem
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(nil)
		if err != nil {
			return
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write(nil)
		if err != nil {
			return
		}
	}

}

func handleRequests(port string) {
	if port == "" {
		port = ":10000"
	} else {
		port = ":" + port
	}
	println("Started with port:", port[1:])
	http.HandleFunc("/", defaultHandler)      //Регистрирует хендлер для short-hand клиента
	log.Fatal(http.ListenAndServe(port, nil)) //Запускает сам short-hand клиент
}

func main() {
	qs = make(map[string]chan string)
	if len(os.Args) > 1 {
		handleRequests(os.Args[1])
	} else {
		handleRequests("")
	}

}
