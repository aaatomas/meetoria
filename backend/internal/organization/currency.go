package organization

import "strings"

const DefaultCurrency = "EUR"
const DefaultTimezone = "Europe/Vilnius"

func NormalizeCurrency(currency string) string {
	code := strings.TrimSpace(strings.ToUpper(currency))
	if len(code) != 3 {
		return DefaultCurrency
	}
	return code
}
