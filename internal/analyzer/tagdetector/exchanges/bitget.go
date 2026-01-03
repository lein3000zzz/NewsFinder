package exchanges

import (
	"encoding/json"
	"net/http"
)

type bitgetInfoMessage struct {
	Data []bitgetData `json:"data"`
	// Остальные поля не нужны для нашего случая
}

type bitgetData struct {
	Symbol    string `json:"symbol"`
	BaseCoin  string `json:"baseCoin"`
	QuoteCoin string `json:"quoteCoin"`
}

func GetBitgetInfoSet(url string) (map[string]struct{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	binanceResponse := bitgetInfoMessage{}

	err = json.NewDecoder(resp.Body).Decode(&binanceResponse)
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{})
	for _, symbol := range binanceResponse.Data {
		set[symbol.Symbol] = struct{}{}
		set[symbol.BaseCoin] = struct{}{}
		set[symbol.QuoteCoin] = struct{}{}
	}

	return set, nil
}
