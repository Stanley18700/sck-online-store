package common_test

import (
	"store-service/internal/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Decimal struct {
	ShortDecimal float64 `json:"short_digit"`
	LongDecimal  float64 `json:"long_digit"`
}

func Test_ConvertToThb_Input_123_Should_be_4147_Point_56_and_Point_555818(t *testing.T) {
	expected := Decimal{
		ShortDecimal: 4147.56,
		LongDecimal:  4147.555818,
	}

	actual := common.ConvertToThb(123)

	assert.Equal(t, expected.ShortDecimal, actual.ShortDecimal)
	assert.Equal(t, expected.LongDecimal, actual.LongDecimal)
}

func Test_ConvertToThb_Input_0_Should_be_0(t *testing.T) {
	expected := Decimal{
		ShortDecimal: 0,
		LongDecimal:  0,
	}

	actual := common.ConvertToThb(0)

	assert.Equal(t, expected.ShortDecimal, actual.ShortDecimal)
	assert.Equal(t, expected.LongDecimal, actual.LongDecimal)
}

func Test_ConvertToThb_Input_Minus_123_Should_be_Minus_4147_Point_56_and_Point_555818(t *testing.T) {
	expected := Decimal{
		ShortDecimal: -4147.56,
		LongDecimal:  -4147.555818,
	}

	actual := common.ConvertToThb(-123)

	assert.Equal(t, expected.ShortDecimal, actual.ShortDecimal)
	assert.Equal(t, expected.LongDecimal, actual.LongDecimal)
}
