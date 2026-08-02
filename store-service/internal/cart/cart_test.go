package cart_test

import (
	"context"
	"database/sql"
	"store-service/internal/cart"
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
				Image:        "/43_Piece_dinner_Set.png",
				Stock:        10,
				Brand:        "CoolKidz",
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
				Image:        "/Balance_Training_Bicycle.png",
				Stock:        100,
				Brand:        "SportsFun",
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
				Image:        "/Balance_Training_Bicycle.png",
				Stock:        100,
				Brand:        "SportsFun",
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
				Image:        "/43_Piece_dinner_Set.png",
				Stock:        200,
				Brand:        "CoolKidz",
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
