package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

const (
	port          = 5555
	kumis_price   = 100
	maxFileLength = 20 << 20
)

var (
	money           = 0
	totalKumisCount = 10000
)

type buyRequest struct {
	Sum int `json:"sum"`
}

type buyResponse struct {
	Change int `json:"change"`
}

func ReceiveFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseMultipartForm(maxFileLength); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()
	if _, err := io.Copy(&buf, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	out, err := os.Create(header.Filename)
	defer out.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := out.WriteString(buf.String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	router := httprouter.New()
	router.GET("/price", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if _, err := w.Write([]byte(strconv.Itoa(kumis_price))); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	router.POST("/buy", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var request buyRequest
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Sum < kumis_price {
			w.WriteHeader(http.StatusForbidden)
			if _, err := w.Write([]byte("Not enough money for kumis. Please, check the price.")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		if totalKumisCount == 0 {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Kumis is out of the stock. Sorry.")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		money += kumis_price
		totalKumisCount--
		rsp := buyResponse{Change: request.Sum - kumis_price}
		rspBytes, err := json.Marshal(rsp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(rspBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	router.POST("/upload", ReceiveFile)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
