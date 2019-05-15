package main

import (
	"encoding/json"
	"net/http"
)

// InitHandler initializes a handler for Hello world
func InitHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		type ResponseBody struct {
			Message string `json:"message"`
		}

		if req.Method != "GET" || req.URL.Path != "/" {
			rw.WriteHeader(http.StatusNotFound)

			rw.Write([]byte("404 not found"))

			return
		}

		body := ResponseBody{
			Message: "Hello World!!",
		}

		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(body)
	})

	return mux
}

func main() {
	handler := InitHandler()

	http.ListenAndServe(":80", handler)
}
