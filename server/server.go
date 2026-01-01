package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

func main() {
	http.HandleFunc("/cotacao", cotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func cotacaoHandler(rw http.ResponseWriter, rq *http.Request) {
	var cotacao Cotacao = getCotacao()

	if err := criarTabelaNoBanco(); err != nil {
		http.Error(rw, "Erro ao criar tabela", http.StatusInternalServerError)
		return
	}
	if err := salvarCotacaoNoBanco(&cotacao); err != nil {
		http.Error(rw, "Erro ao salvar cotacao", http.StatusInternalServerError)
		return
	}

	bid, _ := strconv.ParseFloat(cotacao.USDBRL.Bid, 64)

	json.NewEncoder(rw).Encode(bid)
}

func criarTabelaNoBanco() error {
	dsn := "file:cotacoes.db?cache=shared&mode=rwc&_foreign_keys=on"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(
		`create table if not exists cotacoes (
			id integer primary key autoincrement, 
			code text, 
			codein text, 
			name text,
			high text,
			low text,
			varBid text,
			pctChange text,
			bid text,
			ask text,
			timestamp text,
			create_date text
		);
		`)
	return err
}

func salvarCotacaoNoBanco(cotacao *Cotacao) error {
	dsn := "file:cotacoes.db?cache=shared&mode=rwc&_foreign_keys=on"

	db, _ := sql.Open("sqlite3", dsn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx,
		`insert into cotacoes (
			code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date
		) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cotacao.USDBRL.Code,
		cotacao.USDBRL.Codein,
		cotacao.USDBRL.Name,
		cotacao.USDBRL.High,
		cotacao.USDBRL.Low,
		cotacao.USDBRL.VarBid,
		cotacao.USDBRL.PctChange,
		cotacao.USDBRL.Bid,
		cotacao.USDBRL.Ask,
		cotacao.USDBRL.Timestamp,
		cotacao.USDBRL.CreateDate,
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Printf("db insert timed out")
		}
	}

	return err
}

func getCotacao() Cotacao {
	// Cria um contexto com timeout de 200ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	// Cria a requisição HTTP com o contexto
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}
	// Executa a requisição HTTP
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Printf("req timed out")
			panic(err)
		}
	}
	defer res.Body.Close()

	// Lê o corpo da resposta
	body, _ := io.ReadAll(res.Body)

	// Faz o unmarshal(serialização) do JSON para a struct Cotacao
	var cotacao Cotacao
	json.Unmarshal(body, &cotacao)

	return cotacao
}
