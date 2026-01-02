package externaldata

import (
	"encoding/json"
	"net/http"
)

type binanceExchangeInfoMessage struct {
	Symbols []binanceExchangeInfoSymbol `json:"symbols"`
	// Остальные поля не нужны для нашего случая
}

type binanceExchangeInfoSymbol struct {
	Symbol     string `json:"symbol"`
	BaseAsset  string `json:"baseAsset"`
	QuoteAsset string `json:"quoteAsset"`
}

func GetBinanceInfoSet(url string) (map[string]struct{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	binanceResponse := binanceExchangeInfoMessage{}

	err = json.NewDecoder(resp.Body).Decode(&binanceResponse)
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{})
	for _, symbol := range binanceResponse.Symbols {
		set[symbol.Symbol] = struct{}{}
		set[symbol.BaseAsset] = struct{}{}
		set[symbol.QuoteAsset] = struct{}{}
	}

	return set, nil
}
