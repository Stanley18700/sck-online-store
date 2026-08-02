package order_test

import (
	"context"
	"errors"
	"fmt"
	"store-service/internal/auth"
	"store-service/internal/order"
	"store-service/internal/point"
	"store-service/internal/product"
	"store-service/internal/shipping"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CreateOrder_Input_Submitted_Order_Should_be_OrderNumber_2601069522001001(t *testing.T) {
	uid := 1
	oid := 8004359103
	var orderNumber int64 = 2601069522001001
	productPrice := 12.95
	nextSeq := 1
	fixedTime := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	datePrefix := "260106"

	expected := order.Order{
		OrderNumber: orderNumber,
	}

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, submittedOrder.BurnPoint).Return(true, nil)
	mockPointInterface.On("CalculatePoint", mock.Anything, 465.811034).Return(point.TotalPoint{Point: 9}, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    0,
		TotalPrice:       515.8110340000001,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        9,
	}

	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, nil)

	shippingInfo := order.ShippingInfo{
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
	}
	mockOrderRepository.On("CreateShipping", mock.Anything, uid, oid, shippingInfo).Return(1, nil)

	mockOrderRepository.On("CreateOrderProduct", mock.Anything, oid, submittedOrder.Cart[0].ProductID, submittedOrder.Cart[0].Quantity, productPrice).Return(nil)

	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("DeleteCart", mock.Anything, uid, submittedOrder.Cart[0].ProductID).Return(nil)

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		CartRepository:     mockCartRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_CreateOrder_Input_Burn_8_Points_Should_be_Discount_4_THB_and_Points_Deducted_Before_Order(t *testing.T) {
	uid := 1
	oid := 8004359103
	var orderNumber int64 = 2601069522001001
	productPrice := 12.95
	nextSeq := 1
	fixedTime := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	datePrefix := "260106"

	subtotalTHB := 465.811034
	discountTHB := 4.0
	totalTHB := subtotalTHB - discountTHB

	expected := order.Order{
		OrderNumber: orderNumber,
	}

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		PaymentMethodID:      1,
		BurnPoint:            8,
		SubTotalPrice:        100.00,
		DiscountPrice:        999.99,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, -8).Return(true, nil)
	mockPointInterface.On("CalculateDiscount", mock.Anything, 8, subtotalTHB).Return(point.DiscountQuote{BurnPoint: 8, Discount: discountTHB}, nil)
	mockPointInterface.On("CalculatePoint", mock.Anything, totalTHB).Return(point.TotalPoint{Point: 9}, nil)
	mockPointInterface.On("DeductPoint", mock.Anything, uid, point.SubmitedPoint{Amount: -8}).Return(point.TotalPoint{Point: 0}, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, 2).Return(product.ProductDetail{
		ID:    2,
		Name:  "43 Piece dinner Set",
		Price: productPrice,
		Stock: 1,
		Brand: "Coolkidz",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, 1).Return(shipping.ShippingMethodDetail{
		ID:   1,
		Name: "Kerry",
		Fee:  50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", 1, 1, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	// The client's DiscountPrice (999.99) must be ignored: the persisted discount
	// is the server-derived 4.00 THB for 8 burned points.
	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: 1,
		PaymentMethodID:  1,
		SubTotalPrice:    subtotalTHB,
		DiscountPrice:    discountTHB,
		TotalPrice:       totalTHB + 50,
		ShippingFee:      50,
		BurnPoint:        8,
		EarnPoint:        9,
	}
	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, nil)
	mockOrderRepository.On("CreateShipping", mock.Anything, uid, oid, order.ShippingInfo{ShippingMethodID: 1}).Return(1, nil)
	mockOrderRepository.On("CreateOrderProduct", mock.Anything, oid, 2, 1, productPrice).Return(nil)

	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("DeleteCart", mock.Anything, uid, 2).Return(nil)

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		CartRepository:     mockCartRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
	mockPointInterface.AssertCalled(t, "DeductPoint", mock.Anything, uid, point.SubmitedPoint{Amount: -8})
}

func Test_CreateOrder_Input_Odd_Burn_Point_Should_be_Return_Invalid_Burn_Error(t *testing.T) {
	uid := 1
	fixedTime := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	subtotalTHB := 465.811034

	submittedOrder := order.SubmitedOrder{
		Cart:             []order.OrderProduct{{ProductID: 2, Quantity: 1}},
		ShippingMethodID: 1,
		PaymentMethodID:  1,
		BurnPoint:        7,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, -7).Return(true, nil)
	// point-service floors 7 to the even 6 - the mismatch must reject the order.
	mockPointInterface.On("CalculateDiscount", mock.Anything, 7, subtotalTHB).Return(point.DiscountQuote{BurnPoint: 6, Discount: 3}, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, 2).Return(product.ProductDetail{ID: 2, Price: 12.95}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockCartRepository := new(mockCartRepository)
	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, 1).Return(shipping.ShippingMethodDetail{ID: 1, Fee: 50}, nil)

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		CartRepository:     mockCartRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        new(mockOrderHelper),
		Clock:              func() time.Time { return fixedTime },
	}

	_, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid burn point")
	mockOrderRepository.AssertNotCalled(t, "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
	mockPointInterface.AssertNotCalled(t, "DeductPoint", mock.Anything, mock.Anything, mock.Anything)
}

func Test_CreateOrder_Burn_Fails_Should_Fail_Order_Without_Creating_It(t *testing.T) {
	uid := 1
	var orderNumber int64 = 2601069522001001
	nextSeq := 1
	fixedTime := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	datePrefix := "260106"
	subtotalTHB := 465.811034
	discountTHB := 4.0
	totalTHB := subtotalTHB - discountTHB

	submittedOrder := order.SubmitedOrder{
		Cart:             []order.OrderProduct{{ProductID: 2, Quantity: 1}},
		ShippingMethodID: 1,
		PaymentMethodID:  1,
		BurnPoint:        8,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, -8).Return(true, nil)
	mockPointInterface.On("CalculateDiscount", mock.Anything, 8, subtotalTHB).Return(point.DiscountQuote{BurnPoint: 8, Discount: discountTHB}, nil)
	mockPointInterface.On("CalculatePoint", mock.Anything, totalTHB).Return(point.TotalPoint{Point: 9}, nil)
	mockPointInterface.On("DeductPoint", mock.Anything, uid, point.SubmitedPoint{Amount: -8}).Return(point.TotalPoint{}, fmt.Errorf("points are not enough, please try again"))

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, 2).Return(product.ProductDetail{ID: 2, Price: 12.95}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, 1).Return(shipping.ShippingMethodDetail{ID: 1, Fee: 50}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", 1, 1, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		CartRepository:     new(mockCartRepository),
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	_, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.NotNil(t, err)
	mockOrderRepository.AssertNotCalled(t, "CreateOrder", mock.Anything, mock.Anything, mock.Anything)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_Error_Points_not_Enough(t *testing.T) {
	expected := order.Order{}
	expectedErr := fmt.Errorf("points are not enough, please try again")

	uid := 1
	burnPoint := 100

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, -(burnPoint)).Return(false, fmt.Errorf("points are not enough, please try again"))

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            burnPoint,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	orderService := order.OrderService{
		PointService: mockPointInterface,
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.Equal(t, expectedErr, err)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_No_Product_in_Order_Error(t *testing.T) {
	expected := order.Order{}
	expectedErr := fmt.Errorf("There is no product in order, please try again")

	uid := 1

	submittedOrder := order.SubmitedOrder{
		Cart:                 []order.OrderProduct{},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, 0).Return(true, nil)
	mockPointInterface.On("CalculatePoint", mock.Anything, 465.811034).Return(point.TotalPoint{Point: 9}, nil)

	orderService := order.OrderService{
		PointService: mockPointInterface,
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.Equal(t, expectedErr, err)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_Create_Order_Error(t *testing.T) {
	expected := order.Order{}

	uid := 1
	oid := 8004359103
	productPrice := 12.95
	datePrefix := "260112"
	nextSeq := 32
	fixedTime := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	var orderNumber int64 = 2601129522001032

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, 0).Return(true, nil)
	mockPointInterface.On("CalculatePoint", mock.Anything, 465.811034).Return(point.TotalPoint{Point: 9}, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    0,
		TotalPrice:       515.8110340000001,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        9,
	}
	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, errors.New("CreateOrder Error"))

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_Create_Shipping_Error(t *testing.T) {
	expected := order.Order{}

	uid := 1
	oid := 8004359103
	productPrice := 12.95
	datePrefix := "261212"
	nextSeq := 80
	fixedTime := time.Date(2026, 12, 12, 0, 0, 0, 0, time.UTC)
	var orderNumber int64 = 2612129522001080

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, 0).Return(true, nil)
	mockPointInterface.On("CalculatePoint", mock.Anything, 465.811034).Return(point.TotalPoint{Point: 9}, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    0,
		TotalPrice:       515.8110340000001,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        9,
	}

	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, nil)

	shippingInfo := order.ShippingInfo{
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
	}
	mockOrderRepository.On("CreateShipping", mock.Anything, uid, oid, shippingInfo).Return(1, errors.New("CreateShipping Error"))

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_Create_Order_Product_Error(t *testing.T) {
	expected := order.Order{}

	uid := 1
	oid := 8004359103
	productPrice := 12.95
	datePrefix := "260515"
	nextSeq := 179
	fixedTime := time.Date(2026, 05, 15, 0, 0, 0, 0, time.UTC)
	var orderNumber int64 = 2605159522001179

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, submittedOrder.BurnPoint).Return(true, nil)
	mockPointInterface.On("CalculatePoint", mock.Anything, 465.811034).Return(point.TotalPoint{Point: 9}, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    0,
		TotalPrice:       515.8110340000001,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        9,
	}

	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, nil)

	shippingInfo := order.ShippingInfo{
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 à¸–.à¸¡à¸«à¸´à¸”à¸¥",
		ShippingSubDistrict:  "à¸—à¹ˆà¸²à¸¨à¸²à¸¥à¸²",
		ShippingDistrict:     "à¹€à¸¡à¸·à¸­à¸‡",
		ShippingProvince:     "à¹€à¸Šà¸µà¸¢à¸‡à¹ƒà¸«à¸¡à¹ˆ",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "à¸“à¸±à¸à¸à¸²",
		RecipientLastName:    "à¸Šà¸¸à¸•à¸´à¸šà¸¸à¸•à¸£",
		RecipientPhoneNumber: "0970809292",
	}
	mockOrderRepository.On("CreateShipping", mock.Anything, uid, oid, shippingInfo).Return(1, nil)

	mockOrderRepository.On("CreateOrderProduct", mock.Anything, oid, submittedOrder.Cart[0].ProductID, submittedOrder.Cart[0].Quantity, productPrice).Return(errors.New("CreateOrderProduct Error"))

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_OrderBurnPoint_Input_Burn_Points_100_Should_be_Return_Total_Point_50(t *testing.T) {
	expected := point.TotalPoint{
		Point: 50,
	}

	uid := 1
	burnPoint := 100
	submitedPoint := point.SubmitedPoint{
		Amount: -(burnPoint),
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("DeductPoint", mock.Anything, uid, submitedPoint).Return(point.TotalPoint{
		Point: 50,
	}, nil)

	orderService := order.OrderService{
		PointService: mockPointInterface,
	}

	actual, err := orderService.OrderBurnPoint(context.Background(), uid, burnPoint)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_OrderBurnPoint_Input_Burn_Points_100_Should_be_Return_Totol_Point_Error(t *testing.T) {
	expected := point.TotalPoint{}

	uid := 1
	burnPoint := 100
	submitedPoint := point.SubmitedPoint{
		Amount: -(burnPoint),
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("DeductPoint", mock.Anything, uid, submitedPoint).Return(point.TotalPoint{}, errors.New("DeductPoint Error"))

	orderService := order.OrderService{
		PointService: mockPointInterface,
	}

	actual, err := orderService.OrderBurnPoint(context.Background(), uid, burnPoint)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_GetOrderSummary_Should_Return_One_Product_If_OrderNumber_is_2601069522001001(t *testing.T) {
	userID := 4
	orderID := 1
	trackingNumber := "KR-443947172"
	var orderNumber int64 = 2601069522001001
	updatedTime := time.Date(2026, 2, 28, 18, 58, 44, 0, time.UTC)
	expectedUpdateTime := "01-03-2026 01:58:44"

	orderDetail := order.OrderDetailWithTrackingNumber{
		ID:               orderID,
		OrderNumber:      orderNumber,
		UserID:           userID,
		ShippingMethodID: 1,
		PaymentMethodID:  1,
		SubTotalPrice:    4314.6,
		DiscountPrice:    0,
		TotalPrice:       4364.6,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        86,
		TransactionID:    "TXN202512250934",
		Status:           "paid",
		TrackingNumber:   trackingNumber,
		Updated:          updatedTime,
	}

	orderProduct := []order.OrderProductWithPrice{
		{
			ProductBrand: "SportsFun",
			ProductName:  "Balance Training Bicycle",
			Quantity:     1,
			Price:        119.95,
		},
	}

	userDetail := auth.UserPayload{
		UserID:    userID,
		FirstName: "Noppadon",
		LastName:  "Sookwattana",
		Username:  "noppadon.s",
	}

	expected := order.OrderSummary{
		OrderNumber:    orderNumber,
		FirstName:      userDetail.FirstName,
		LastName:       userDetail.LastName,
		TrackingNumber: trackingNumber,
		ShippingMethod: "Kerry",
		PaymentMethod:  "Credit Card / Debit Card",
		OrderProductList: []order.OrderSummaryProduct{
			{
				ProductBrand:  "SportsFun",
				ProductName:   "Balance Training Bicycle",
				Quantity:      1,
				PriceTHB:      4314.6,
				TotalPriceTHB: 4314.6,
			},
		},
		SubTotalPrice:  orderDetail.SubTotalPrice,
		DiscountPrice:  orderDetail.DiscountPrice,
		TotalPrice:     orderDetail.TotalPrice,
		ShippingFee:    orderDetail.ShippingFee,
		BurnPoint:      orderDetail.BurnPoint,
		ReceivingPoint: orderDetail.EarnPoint,
		IssuedDate:     expectedUpdateTime,
	}

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderWithTrackingNumberByOrderNumber", mock.Anything, orderNumber).Return(orderDetail, nil)
	mockOrderRepository.On("GetOrderProductWithPrice", mock.Anything, orderID).Return(orderProduct, nil)

	mockUserRepository := new(mockUserRepository)
	mockUserRepository.On("FindByID", mock.Anything, userID).Return(userDetail, nil)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
		UserRepository:  mockUserRepository,
	}

	actual, err := orderService.GetOrderSummary(context.Background(), orderNumber)
	assert.Equal(t, expected, actual)
	assert.Nil(t, err)
}

func Test_GetOrderSummary_Should_Return_Two_Products_If_OrderOrderNumber_is_2601069522002002(t *testing.T) {
	userID := 5
	orderID := 2
	trackingNumber := "KR-304590466"
	var orderNumber int64 = 2601069522002002
	updatedTime := time.Date(2026, 2, 14, 1, 40, 32, 0, time.UTC)
	expectedUpdateTime := "14-02-2026 08:40:32"

	orderDetail := order.OrderDetailWithTrackingNumber{
		ID:               orderID,
		OrderNumber:      orderNumber,
		UserID:           userID,
		ShippingMethodID: 1,
		PaymentMethodID:  1,
		SubTotalPrice:    5246.22,
		DiscountPrice:    0,
		TotalPrice:       5256.22,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        104,
		TransactionID:    "TXN202512251028",
		Status:           "paid",
		TrackingNumber:   trackingNumber,
		Updated:          updatedTime,
	}

	orderProduct := []order.OrderProductWithPrice{
		{
			ProductBrand: "SportsFun",
			ProductName:  "Balance Training Bicycle",
			Quantity:     1,
			Price:        119.95,
		},
		{
			ProductBrand: "CoolKidz",
			ProductName:  "43 Piece dinner Set",
			Quantity:     2,
			Price:        12.95,
		},
	}

	userDetail := auth.UserPayload{
		UserID:    userID,
		FirstName: "Pimmida",
		LastName:  "Katethong",
		Username:  "pimmida.k",
	}

	expected := order.OrderSummary{
		OrderNumber:    orderNumber,
		FirstName:      userDetail.FirstName,
		LastName:       userDetail.LastName,
		TrackingNumber: trackingNumber,
		ShippingMethod: "Kerry",
		PaymentMethod:  "Credit Card / Debit Card",
		OrderProductList: []order.OrderSummaryProduct{
			{
				ProductBrand:  "SportsFun",
				ProductName:   "Balance Training Bicycle",
				Quantity:      1,
				PriceTHB:      4314.6,
				TotalPriceTHB: 4314.6,
			},
			{
				ProductBrand:  "CoolKidz",
				ProductName:   "43 Piece dinner Set",
				Quantity:      2,
				PriceTHB:      465.81,
				TotalPriceTHB: 931.62,
			},
		},
		SubTotalPrice:  orderDetail.SubTotalPrice,
		DiscountPrice:  orderDetail.DiscountPrice,
		TotalPrice:     orderDetail.TotalPrice,
		ShippingFee:    orderDetail.ShippingFee,
		BurnPoint:      orderDetail.BurnPoint,
		ReceivingPoint: orderDetail.EarnPoint,
		IssuedDate:     expectedUpdateTime,
	}

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderWithTrackingNumberByOrderNumber", mock.Anything, orderNumber).Return(orderDetail, nil)
	mockOrderRepository.On("GetOrderProductWithPrice", mock.Anything, orderID).Return(orderProduct, nil)

	mockUserRepository := new(mockUserRepository)
	mockUserRepository.On("FindByID", mock.Anything, userID).Return(userDetail, nil)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
		UserRepository:  mockUserRepository,
	}

	actual, err := orderService.GetOrderSummary(context.Background(), orderNumber)
	assert.Equal(t, expected, actual)
	assert.Nil(t, err)
}
