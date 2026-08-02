package common

import "math"

type Decimal struct {
	ShortDecimal float64 `json:"short_digit"`
	LongDecimal  float64 `json:"long_digit"`
}

func ConvertToThb(amount float64) Decimal {
	rate := 33.719966
	result := amount * rate
	factor2 := math.Pow(10, 2)
	factor6 := math.Pow(10, 6)

	return Decimal{
		ShortDecimal: math.Round(result*factor2) / factor2,
		LongDecimal:  math.Round(result*factor6) / factor6,
	}
}

// Round returns amount rounded to the given number of decimal places.
func Round(amount float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(amount*factor) / factor
}

// LineTotal converts a USD unit price into the THB total for quantity items.
//
// The multiplication happens at full precision and the result is rounded once, so
// callers must never multiply an already-rounded unit price themselves — doing so
// rounds twice and drifts a satang away from the order total.
func LineTotal(unitPrice float64, quantity int) Decimal {
	return ConvertToThb(unitPrice * float64(quantity))
}

// SumLineTotals adds up line totals that were already rounded by LineTotal, so a
// displayed invoice always adds up to its own lines. Summing the rounded lines
// (rather than rounding the exact sum) is what keeps "subtotal" equal to what the
// customer can add up by hand; the extra rounding here only absorbs the float
// drift introduced by the additions themselves.
func SumLineTotals(lineTotals []Decimal) Decimal {
	shortTotal := 0.0
	longTotal := 0.0
	for _, lineTotal := range lineTotals {
		shortTotal = shortTotal + lineTotal.ShortDecimal
		longTotal = longTotal + lineTotal.LongDecimal
	}

	return Decimal{
		ShortDecimal: Round(shortTotal, 2),
		LongDecimal:  Round(longTotal, 6),
	}
}
