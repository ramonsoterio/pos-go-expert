package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	apiURL          = "http://localhost:8080/cotacao"
	apiTimeout      = 300 * time.Millisecond
	fileName        = "cotacao.txt"
	requestInterval = 1 * time.Second
)

type APIResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	for {
		time.Sleep(requestInterval)
		slog.Info("requesting new price")
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			slog.Error("error creating request.", "error", err)
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("error making request.", "error", err)
			continue
		}
		defer resp.Body.Close()
		if resp != nil {
			if resp.StatusCode == http.StatusInternalServerError {
				slog.Error("request failed.", "status", resp.StatusCode)
				continue
			}
		}
		result, err := decodeResponse(resp)
		if err != nil {
			slog.Error("error decoding response.", "error", err)
			continue
		}
		fetchedValue, err := strconv.ParseFloat(result.USDBRL.Bid, 64)
		if err != nil {
			slog.Error("error converting fetched dolar.", "error", err)
			continue
		}
		if err = saveToFile(fetchedValue); err != nil {
			slog.Error("error saving to file.", "error", err)
			continue
		}
	}
}

func decodeResponse(resp *http.Response) (result APIResponse, err error) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &result)
	return
}

func saveToFile(dolar float64) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	content := fmt.Sprintf("Dólar:%v", dolar)
	_, err = file.WriteString(content + "\n")
	return err
}
