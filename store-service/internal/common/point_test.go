package common_test

import (
	"store-service/internal/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CalculatePoint_Input_0_Should_be_Return_0(t *testing.T) {
	expected := 0
	price := 0.0

	actual := common.CalculatePoint(price)

	assert.Equal(t, expected, actual)
}

func Test_CalculatePoint_Input_100_Should_be_Return_2(t *testing.T) {
	expected := 2
	price := 100.0

	actual := common.CalculatePoint(price)

	assert.Equal(t, expected, actual)
}

func Test_CalculatePoint_Input_599_Dot_999_Should_be_Return_11(t *testing.T) {
	expected := 11
	price := 599.999

	actual := common.CalculatePoint(price)

	assert.Equal(t, expected, actual)
}

func Test_CalculatePoint_Input_Minus_599_Dot_999_Should_be_Return_0(t *testing.T) {
	expected := 0
	price := -599.999

	actual := common.CalculatePoint(price)

	assert.Equal(t, expected, actual)
}

