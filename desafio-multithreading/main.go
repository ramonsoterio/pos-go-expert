package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	urlViaCep    = "https://viacep.com.br/ws/%s/json"
	urlBrasilAPI = "https://brasilapi.com.br/api/cep/v1/%s"
	resultMsg    = "Melhor resultado API %s: %s"
)

func main() {
	chanViaCep := make(chan string)
	chanBrasilAPI := make(chan string)
	for _, cep := range os.Args[1:] {
		go fetchCEP(urlViaCep, cep, chanViaCep)
		go fetchCEP(urlBrasilAPI, cep, chanBrasilAPI)
	}

	select {
	case resultViaCep := <-chanViaCep:
		fmt.Printf(resultMsg, "Via Cep", resultViaCep)
	case resultBrasilAPI := <-chanBrasilAPI:
		fmt.Println(resultBrasilAPI)
		fmt.Printf(resultMsg, "Brasil API", resultBrasilAPI)
	case <-time.After(1 * time.Second):
		fmt.Println("Timeout")
	}
}

func fetchCEP(baseURL, cep string, resultChan chan<- string) {
	url := fmt.Sprintf(baseURL, cep)
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("error fetching cep in %s", url)
	}
	defer resp.Body.Close()
	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error decoding response in %s", url)
	}
	resultChan <- string(rawResp)
}
