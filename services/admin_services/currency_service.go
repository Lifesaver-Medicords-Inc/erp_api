package adminservices

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type CurrencyService struct {
}

func NewCurrencyService() *CurrencyService {
	return &CurrencyService{}
}

func (c *CurrencyService) GetCurrencyAPI(baseCode string) (currencyBody string) {
	url := "https://api.fxratesapi.com/latest?base=PHP"

	if baseCode != "" {
		url = "https://api.fxratesapi.com/latest?base=" + baseCode
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	return string(body)
}
