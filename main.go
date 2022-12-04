package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type path struct {
	queue  string
	args   []string
	values []string
}

var qs map[string][]string

func createPath(origin string) (path, error) {
	ans := path{}
	pos := strings.Index(origin, "?")
	if pos <= -1 {
		ans.queue = origin
		return ans, nil
	}
	ans.queue = origin[:pos]
	split := strings.Split(origin[pos+1:], "&")
	for _, s := range split {
		tempos := strings.Index(s, "=")
		if tempos <= -1 {
			return ans, errors.New("Incorrect arguments")
		}
		ans.args = append(ans.args, s[:tempos])
		ans.values = append(ans.values, s[tempos+1:])
	}

	return ans, nil

}

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
	for key, m := range qs {
		fmt.Println("Queue:", key)
		for k, v := range m {
			fmt.Println("	", k, "value is", v)
		}
	}
}

func getFromQueue(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write(nil)
		if err != nil {
			return
		}
		return
	}
	req, _ := createPath(r.RequestURI[1:])
	if val, ok := qs[req.queue]; ok && len(val) > 0 {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(val[0]))
		if err != nil {
			return
		}
		qs[req.queue] = val[1:]
	} else {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write(nil)
		if err != nil {
			return
		}
	}
}

// curl -XPUT http://localhost:10000/color?v=red
func putInQueue(w http.ResponseWriter, r *http.Request) {

	req, _ := createPath(r.RequestURI[1:])
	if r.ParseForm() != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write(nil)
		if err != nil {
			return
		}
		return
	}
	if v := r.Form.Get("v"); v != "" {
		qs[req.queue] = append(qs[req.queue], v)
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

func handleRequests() {
	http.HandleFunc("/", defaultHandler)          //Регистрирует хендлер для short-hand клиента
	log.Fatal(http.ListenAndServe(":10000", nil)) //Запускает сам short-hand клиент
}

func main() {
	qs = make(map[string][]string)
	handleRequests()
}
