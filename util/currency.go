package util

const (
	USD string = "USD"
	EUR string = "EUR"
	GBP string = "GBP"
	KZT string = "KZT"
	RUB string = "RUB"
	UAH string = "UAH"
	BYN string = "BYN"
)

var supportedCurrencies = map[string]struct{}{
	USD: {},
	EUR: {},
	GBP: {},
	KZT: {},
	RUB: {},
	UAH: {},
	BYN: {},
}

func IsCurrencySupported(currency string) bool {
	_, exists := supportedCurrencies[currency]
	return exists
}
