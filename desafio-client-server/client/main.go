package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	apiURL     = "http://localhost:8080/cotacao"
	apiTimeout = 300 * time.Millisecond
	fileName   = "cotacao.txt"
)

type Cotacao struct {
	Dolar float64
}

func main() {
	for {
		slog.Info("requesting new price")
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			slog.Error(err.Error())
		}
		//TODO this
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if resp.StatusCode == http.StatusInternalServerError {
				slog.Error(err.Error())
				continue
			}
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("error reading resp body", err)
			continue
		}
		var c Cotacao
		if err = json.Unmarshal(raw, &c); err != nil {
			slog.Error("error unmarshalling json", err)
			continue
		}
		if err = saveToFile(c); err != nil {
			slog.Error("error saving to file", err)
			continue
		}
		time.Sleep(1 * time.Second)
	}
}

//TODO append file instead of overwriting it
func saveToFile(c Cotacao) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	content := fmt.Sprintf("Dólar:%v", c.Dolar)
	_, err = file.WriteString(content + "\n")
	return err
}
