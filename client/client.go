package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	var cotacao string = getCotacaoFromServer()
	f, _ := os.Create("cotacao.txt")

	fmt.Fprintf(f, "Dolar: %s", cotacao)

	defer f.Close()
}

func getCotacaoFromServer() string {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Printf("req timed out")
		}
		return ""
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	return string(body)
}
