package cart_test

import (
	"context"
	"database/sql"
	"store-service/internal/cart"
	"store-service/internal/common"
	"store-service/internal/point"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_GetCart_Should_be_Have_Data_and_Receive_Point_8(t *testing.T) {
	expected := cart.CartResult{
		Carts: []cart.CartDetail{
			{
				ID:           1,
				UserID:       1,
				ProductID:    2,
				Quantity:     1,
				Name:         "43 Piece dinner Set",
				Price:        12.95,
				PriceTHB:     436.67,
				PriceFullTHB: 436.67356,
				// qty 1 -> ConvertToThb(12.95 * 1)
				LineTotalTHB:     436.67,
				LineTotalFullTHB: 436.67356,
				Image:            "/43_Piece_dinner_Set.png",
				Stock:            10,
				Brand:            "CoolKidz",
			},
		},
		Summary: cart.CartSummary{
			TotalPrice:        12.95,
			TotalPriceTHB:     436.67,
			TotalPriceFullTHB: 436.67356,
			ReceivePoint:      8,
		},
	}

	uid := 1
	res := []cart.CartDetail{
		{
			ID:           1,
			UserID:       1,
			ProductID:    2,
			Quantity:     1,
			Name:         "43 Piece dinner Set",
			Price:        12.95,
			PriceTHB:     0,
			PriceFullTHB: 0,
			Image:        "/43_Piece_dinner_Set.png",
			Stock:        10,
			Brand:        "CoolKidz",
		},
	}
	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("GetCartDetail", mock.Anything, uid).Return(res, nil)

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CalculatePoint", mock.Anything, 436.67).Return(point.TotalPoint{Point: 8}, nil)

	cartService := cart.CartService{
		CartRepository: mockCartRepository,
		PointService:   mockPointInterface,
	}
	actual, err := cartService.GetCart(context.Background(), uid)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_GetCart_Should_be_Empty(t *testing.T) {
	expected := cart.CartResult{
		Carts:   []cart.CartDetail{},
		Summary: cart.CartSummary{},
	}
	uid := 1
	res := []cart.CartDetail{}
	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("GetCartDetail", mock.Anything, uid).Return(res, nil)

	cartService := cart.CartService{
		CartRepository: mockCartRepository,
	}
	actual, err := cartService.GetCart(context.Background(), uid)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_AddCart_Input_Submitted_First_Product_Should_be_Have_1_Quantity_and_Receive_Point_80(t *testing.T) {
	expected := cart.CartResult{
		Carts: []cart.CartDetail{
			{
				ID:           1,
				UserID:       1,
				ProductID:    1,
				Quantity:     1,
				Name:         "Balance Training Bicycle",
				Price:        119.95,
				PriceTHB:     4044.71,
				PriceFullTHB: 4044.709922,
				// qty 1 -> ConvertToThb(119.95 * 1)
				LineTotalTHB:     4044.71,
				LineTotalFullTHB: 4044.709922,
				Image:            "/Balance_Training_Bicycle.png",
				Stock:            100,
				Brand:            "SportsFun",
			},
		},
		Summary: cart.CartSummary{
			TotalPrice:        119.95,
			TotalPriceTHB:     4044.71,
			TotalPriceFullTHB: 4044.709922,
			ReceivePoint:      80,
		},
	}
	submitedCart := cart.SubmitedCart{
		ProductID: 1,
		Quantity:  1,
	}
	uid := 1
	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("GetCartByProductID", mock.Anything, uid, submitedCart.ProductID).Return(cart.Cart{}, sql.ErrNoRows)
	mockCartRepository.On("CreateCart", mock.Anything, uid, submitedCart.ProductID, submitedCart.Quantity).Return(1, nil)

	res := []cart.CartDetail{
		{
			ID:           1,
			UserID:       1,
			ProductID:    1,
			Quantity:     1,
			Name:         "Balance Training Bicycle",
			Price:        119.95,
			PriceTHB:     0,
			PriceFullTHB: 0,
			Image:        "/Balance_Training_Bicycle.png",
			Stock:        100,
			Brand:        "SportsFun",
		},
	}
	mockCartRepository.On("GetCartDetail", mock.Anything, uid).Return(res, nil)

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CalculatePoint", mock.Anything, 4044.71).Return(point.TotalPoint{Point: 80}, nil)

	cartService := cart.CartService{
		CartRepository: mockCartRepository,
		PointService:   mockPointInterface,
	}
	actual, err := cartService.AddCart(context.Background(), uid, submitedCart)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_AddCart_Input_Submitted_More_Product_Should_be_Have_2_Quantity_and_Receive_Point_161(t *testing.T) {
	expected := cart.CartResult{
		Carts: []cart.CartDetail{
			{
				ID:           1,
				UserID:       1,
				ProductID:    1,
				Quantity:     2,
				Name:         "Balance Training Bicycle",
				Price:        119.95,
				PriceTHB:     4044.71,
				PriceFullTHB: 4044.709922,
				// qty 2 -> ConvertToThb(119.95 * 2), rounded once
				LineTotalTHB:     8089.42,
				LineTotalFullTHB: 8089.419843,
				Image:            "/Balance_Training_Bicycle.png",
				Stock:            100,
				Brand:            "SportsFun",
			},
		},
		Summary: cart.CartSummary{
			TotalPrice:        239.9,
			TotalPriceTHB:     8089.42,
			TotalPriceFullTHB: 8089.419843,
			ReceivePoint:      161,
		},
	}
	submitedCart := cart.SubmitedCart{
		ProductID: 1,
		Quantity:  1,
	}
	uid := 1
	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("GetCartByProductID", mock.Anything, uid, submitedCart.ProductID).Return(cart.Cart{
		ID:        1,
		UserID:    1,
		ProductID: 1,
		Quantity:  1,
	}, nil)
	mockCartRepository.On("UpdateCart", mock.Anything, uid, submitedCart.ProductID, 2).Return(nil)

	res := []cart.CartDetail{
		{
			ID:           1,
			UserID:       1,
			ProductID:    1,
			Quantity:     2,
			Name:         "Balance Training Bicycle",
			Price:        119.95,
			PriceTHB:     0,
			PriceFullTHB: 0,
			Image:        "/Balance_Training_Bicycle.png",
			Stock:        100,
			Brand:        "SportsFun",
		},
	}
	mockCartRepository.On("GetCartDetail", mock.Anything, uid).Return(res, nil)

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CalculatePoint", mock.Anything, 8089.42).Return(point.TotalPoint{Point: 161}, nil)

	cartService := cart.CartService{
		CartRepository: mockCartRepository,
		PointService:   mockPointInterface,
	}
	actual, err := cartService.AddCart(context.Background(), uid, submitedCart)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_UpdateCart_Input_Submitted_Quantity_2_Should_be_Have_2_Quantity_and_Receive_Point_17(t *testing.T) {
	expected := cart.CartResult{
		Carts: []cart.CartDetail{
			{
				ID:           1,
				UserID:       1,
				ProductID:    2,
				Quantity:     2,
				Name:         "43 Piece dinner Set",
				Price:        12.95,
				PriceTHB:     436.67,
				PriceFullTHB: 436.67356,
				// qty 2 -> ConvertToThb(12.95 * 2), rounded once
				LineTotalTHB:     873.35,
				LineTotalFullTHB: 873.347119,
				Image:            "/43_Piece_dinner_Set.png",
				Stock:            200,
				Brand:            "CoolKidz",
			},
		},
		Summary: cart.CartSummary{
			TotalPrice:        25.9,
			TotalPriceTHB:     873.35,
			TotalPriceFullTHB: 873.347119,
			ReceivePoint:      17,
		},
	}
	submitedCart := cart.SubmitedCart{
		ProductID: 2,
		Quantity:  2,
	}
	uid := 1
	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("UpdateCart", mock.Anything, uid, submitedCart.ProductID, submitedCart.Quantity).Return(nil)

	res := []cart.CartDetail{
		{
			ID:           1,
			UserID:       1,
			ProductID:    2,
			Quantity:     2,
			Name:         "43 Piece dinner Set",
			Price:        12.95,
			PriceTHB:     0,
			PriceFullTHB: 0,
			Image:        "/43_Piece_dinner_Set.png",
			Stock:        200,
			Brand:        "CoolKidz",
		},
	}
	mockCartRepository.On("GetCartDetail", mock.Anything, uid).Return(res, nil)

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CalculatePoint", mock.Anything, 873.35).Return(point.TotalPoint{Point: 17}, nil)

	cartService := cart.CartService{
		CartRepository: mockCartRepository,
		PointService:   mockPointInterface,
	}
	actual, err := cartService.UpdateCart(context.Background(), uid, submitedCart)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_UpdateCart_Input_Submitted_Quantity_0_Should_be_Have_0_Quantity_and_Receive_Point_0(t *testing.T) {
	expected := cart.CartResult{
		Carts:   []cart.CartDetail{},
		Summary: cart.CartSummary{},
	}
	submitedCart := cart.SubmitedCart{
		ProductID: 1,
		Quantity:  0,
	}
	uid := 1
	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("DeleteCart", mock.Anything, uid, submitedCart.ProductID).Return(nil)

	res := []cart.CartDetail{}
	mockCartRepository.On("GetCartDetail", mock.Anything, uid).Return(res, nil)

	cartService := cart.CartService{
		CartRepository: mockCartRepository,
	}
	actual, err := cartService.UpdateCart(context.Background(), uid, submitedCart)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_GetCart_Line_Total_Should_be_Rounded_Once_Not_Twice(t *testing.T) {
	// Regression: the UI used to render product_price_thb * quantity, which rounds
	// 12.95 USD to 436.67 THB first and only then multiplies. That drifts a satang
	// below the order total.
	uid := 1
	res := []cart.CartDetail{
		{
			ID: 1, UserID: 1, ProductID: 5, Quantity: 3,
			Name: "Sleeping Queens Board Game", Price: 12.95, Stock: 10,
		},
	}
	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("GetCartDetail", mock.Anything, uid).Return(res, nil)

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CalculatePoint", mock.Anything, mock.Anything).Return(point.TotalPoint{Point: 26}, nil)

	cartService := cart.CartService{
		CartRepository: mockCartRepository,
		PointService:   mockPointInterface,
	}
	actual, err := cartService.GetCart(context.Background(), uid)

	assert.Equal(t, nil, err)
	assert.Equal(t, 436.67, actual.Carts[0].PriceTHB)
	assert.Equal(t, 1310.02, actual.Carts[0].LineTotalTHB)
	assert.NotEqual(t, 1310.01, actual.Carts[0].LineTotalTHB, "line total must not be the round-then-multiply value")
}

func Test_GetCart_Subtotal_Should_be_Sum_Of_Displayed_Line_Totals(t *testing.T) {
	// The subtotal has to equal what a customer gets by adding up the lines on
	// screen, otherwise the cart shows numbers that do not add up.
	uid := 1
	res := []cart.CartDetail{
		{ID: 1, UserID: 1, ProductID: 1, Quantity: 5, Name: "Balance Training Bicycle", Price: 119.95, Stock: 100},
		{ID: 2, UserID: 1, ProductID: 4, Quantity: 3, Name: "Hoppity Ball 26 inch", Price: 29.95, Stock: 12},
	}
	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("GetCartDetail", mock.Anything, uid).Return(res, nil)

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CalculatePoint", mock.Anything, mock.Anything).Return(point.TotalPoint{Point: 465}, nil)

	cartService := cart.CartService{
		CartRepository: mockCartRepository,
		PointService:   mockPointInterface,
	}
	actual, err := cartService.GetCart(context.Background(), uid)

	assert.Equal(t, nil, err)
	assert.Equal(t, 20223.55, actual.Carts[0].LineTotalTHB)
	assert.Equal(t, 3029.74, actual.Carts[1].LineTotalTHB)

	sumOfLines := 0.0
	for _, c := range actual.Carts {
		sumOfLines = sumOfLines + c.LineTotalTHB
	}
	assert.Equal(t, common.Round(sumOfLines, 2), actual.Summary.TotalPriceTHB)
	assert.Equal(t, 23253.29, actual.Summary.TotalPriceTHB)
}
