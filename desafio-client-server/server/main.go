package main

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

const (
	apiURL        = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	apiTimeout    = 200 * time.Millisecond
	dbTimeout     = 10 * time.Millisecond
	dbName        = "prices.db"
	dbCreateTable = "CREATE TABLE IF NOT EXISTS prices(ID INTEGER NOT NULL PRIMARY KEY, value float8)"
)

var db *sql.DB

func main() {
	var err error
	if db, err = setupDB(); err != nil {
		panic(err)
	}
	if _, err = db.Exec(dbCreateTable); err != nil {
		panic(err)
	}
	http.HandleFunc("/cotacao", priceHandler)
	slog.Info("application started")
	http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}

func setupDB() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		return
	}
	if _, err = db.Exec(dbCreateTable); err != nil {
		return
	}
	return
}

func priceHandler(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	defer func() {
		slog.Info("request completed", "status_code", statusCode)
	}()
	result, err := fetchPrice(r.Context())
	if err != nil {
		statusCode = http.StatusInternalServerError
		w.WriteHeader(statusCode)
		return
	}
	timeoutContext, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	_, err = db.ExecContext(timeoutContext, "INSERT INTO prices (value) VALUES (?)", result.USDBRL.Bid)
	if err != nil {
		statusCode = http.StatusInternalServerError
		w.Write([]byte("error saving price to the database."))
		w.WriteHeader(statusCode)
		return
	}
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(result)
}

func fetchPrice(ctx context.Context) (result Cotacao, err error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, apiURL, nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error fetching cotacao from API", err)
		return
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &result)
	return
}
