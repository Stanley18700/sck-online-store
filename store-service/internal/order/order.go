package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"store-service/internal/auth"
	"store-service/internal/cart"
	"store-service/internal/common"
	"store-service/internal/metrics"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"store-service/internal/point"
	"store-service/internal/product"
	"store-service/internal/shipping"
	"time"
)

type OrderInterface interface {
	CreateOrder(ctx context.Context, uid int, submitedOrder SubmitedOrder) (Order, error)
	OrderBurnPoint(ctx context.Context, uid int, burn int) (point.TotalPoint, error)
	GetOrderSummary(ctx context.Context, orderNumber int64) (OrderSummary, error)
	GeneratePDFFromData(orderDetail OrderSummary) ([]byte, error)
}

type OrderService struct {
	CartRepository     cart.CartRepository
	OrderRepository    OrderRepository
	PointService       point.PointInterface
	ProductRepository  product.ProductRepository
	ShippingRepository shipping.ShippingRepository
	UserRepository     auth.UserRepository
	PDFHelper          PDFHelper
	OrderHelper        OrderHelperInterface
	Clock              func() time.Time
}

type CartRepository interface {
	DeleteCart(userID int, productID int)
}
type PointService interface {
	DeductPoint(uid int, submitedPoint point.SubmitedPoint) (point.TotalPoint, error)
}

type ProductRepository interface {
	GetProductByID(id int) (product.ProductDetail, error)
}

type ShippingRepository interface {
	GetShippingMethodByID(id int) (shipping.ShippingMethodDetail, error)
}

var PaymentMethod = map[int]string{
	1: "Credit Card / Debit Card",
	2: "Line Pay",
}

var ShippingMethod = map[int]string{
	1: "Kerry",
	2: "Thai Post",
	3: "Lineman",
}

var ErrOrderNotFound = errors.New("Order not found")

func (orderService OrderService) CreateOrder(ctx context.Context, uid int, submitedOrder SubmitedOrder) (Order, error) {
	_, err := orderService.PointService.CheckBurnPoint(ctx, uid, -(submitedOrder.BurnPoint))
	if err != nil {
		return Order{}, err
	}

	if len(submitedOrder.Cart) == 0 {
		return Order{}, fmt.Errorf("There is no product in order, please try again")
	}

	subtotalPrice := 0.0
	for _, productSelected := range submitedOrder.Cart {
		product, _ := orderService.ProductRepository.GetProductByID(ctx, productSelected.ProductID)
		subtotalPrice = subtotalPrice + (product.Price * float64(productSelected.Quantity))
	}

	subtotalPriceTHB := common.ConvertToThb(subtotalPrice).LongDecimal

	// The discount is derived server-side from burn_point via point-service
	// (2 points = 1.00 THB). The client-supplied discount_price is never trusted.
	discountPriceTHB := 0.0
	if submitedOrder.BurnPoint > 0 {
		quote, err := orderService.PointService.CalculateDiscount(ctx, submitedOrder.BurnPoint, subtotalPriceTHB)
		if err != nil {
			slog.ErrorContext(ctx, "PointService.CalculateDiscount failed",
				"log_type", "error", "error_code", "POINT_DISCOUNT_FAILED", "error_message", err.Error(), "user_id", uid)
			return Order{}, err
		}
		if quote.BurnPoint != submitedOrder.BurnPoint {
			return Order{}, fmt.Errorf("invalid burn point: points burn in pairs and the discount cannot exceed the subtotal")
		}
		discountPriceTHB = quote.Discount
	}
	totalPriceTHB := subtotalPriceTHB - discountPriceTHB

	shippingDetail, _ := orderService.ShippingRepository.GetShippingMethodByID(ctx, submitedOrder.ShippingMethodID)
	shippingFeeTHB := shippingDetail.Fee

	now := orderService.Clock()
	datePrefix := now.Format("060102")

	seq, err := orderService.OrderRepository.GetNextSequence(ctx, datePrefix, uid)
	if err != nil {
		slog.ErrorContext(ctx, "OrderRepository.GetNextSequence failed",
			"log_type", "error", "error_code", "ORDER_SEQ_FAILED", "error_message", err.Error(), "user_id", uid)
		return Order{}, err
	}

	orderNumber, err := orderService.OrderHelper.GenerateOrderNumber(submitedOrder.PaymentMethodID, submitedOrder.ShippingMethodID, uid, seq, now)
	if err != nil {
		slog.ErrorContext(ctx, "OrderHelper.GenerateOrderNumber failed",
			"log_type", "error", "error_code", "ORDER_NUMBER_FAILED", "error_message", err.Error(), "user_id", uid)
		return Order{}, err
	}

	earnPoint, err := orderService.PointService.CalculatePoint(ctx, totalPriceTHB)
	if err != nil {
		slog.ErrorContext(ctx, "PointService.CalculatePoint failed",
			"log_type", "error", "error_code", "POINT_CALCULATE_FAILED", "error_message", err.Error(), "user_id", uid)
		return Order{}, err
	}

	orderDetail := OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submitedOrder.ShippingMethodID,
		PaymentMethodID:  submitedOrder.PaymentMethodID,
		SubTotalPrice:    subtotalPriceTHB,
		DiscountPrice:    discountPriceTHB,
		TotalPrice:       totalPriceTHB + shippingFeeTHB,
		ShippingFee:      shippingFeeTHB,
		BurnPoint:        submitedOrder.BurnPoint,
		EarnPoint:        earnPoint.Point,
	}

	// Burn BEFORE the order is written: a failed deduction must fail the order,
	// otherwise the discount would be granted while the points stay untouched.
	if submitedOrder.BurnPoint > 0 {
		if _, err := orderService.OrderBurnPoint(ctx, uid, submitedOrder.BurnPoint); err != nil {
			slog.ErrorContext(ctx, "OrderBurnPoint failed",
				"log_type", "error", "error_code", "POINT_BURN_FAILED", "error_message", err.Error(),
				"user_id", uid, "burn_point", submitedOrder.BurnPoint)
			return Order{}, err
		}
	}
	// If persisting the order fails after the burn, credit the points back (best effort).
	refundBurnedPoints := func() {
		if submitedOrder.BurnPoint <= 0 {
			return
		}
		if _, err := orderService.PointService.DeductPoint(ctx, uid, point.SubmitedPoint{Amount: submitedOrder.BurnPoint}); err != nil {
			slog.ErrorContext(ctx, "Compensating point refund failed",
				"log_type", "error", "error_code", "POINT_REFUND_FAILED", "error_message", err.Error(),
				"user_id", uid, "burn_point", submitedOrder.BurnPoint)
		}
	}

	orderID, err := orderService.OrderRepository.CreateOrder(ctx, uid, orderDetail)
	if err != nil {
		slog.ErrorContext(ctx, "OrderRepository.CreateOrder failed",
			"log_type", "error", "error_code", "ORDER_INSERT_FAILED", "error_message", err.Error(), "user_id", uid)
		refundBurnedPoints()
		return Order{}, err
	}

	shippingInfo := ShippingInfo{
		ShippingMethodID:     submitedOrder.ShippingMethodID,
		ShippingAddress:      submitedOrder.ShippingAddress,
		ShippingSubDistrict:  submitedOrder.ShippingSubDistrict,
		ShippingDistrict:     submitedOrder.ShippingDistrict,
		ShippingProvince:     submitedOrder.ShippingProvince,
		ShippingZipCode:      submitedOrder.ShippingZipCode,
		RecipientFirstName:   submitedOrder.RecipientFirstName,
		RecipientLastName:    submitedOrder.RecipientLastName,
		RecipientPhoneNumber: submitedOrder.RecipientPhoneNumber,
	}
	_, err = orderService.OrderRepository.CreateShipping(ctx, uid, orderID, shippingInfo)
	if err != nil {
		slog.ErrorContext(ctx, "OrderRepository.CreateShipping failed",
			"log_type", "error", "error_code", "SHIPPING_INSERT_FAILED", "error_message", err.Error(), "user_id", uid)
		refundBurnedPoints()
		return Order{}, err
	}

	for _, selectedProduct := range submitedOrder.Cart {
		product, err := orderService.ProductRepository.GetProductByID(ctx, selectedProduct.ProductID)
		err = orderService.OrderRepository.CreateOrderProduct(ctx, orderID, selectedProduct.ProductID, selectedProduct.Quantity, product.Price)
		if err != nil {
			slog.ErrorContext(ctx, "OrderRepository.CreateOrderProduct failed",
				"log_type", "error", "error_code", "ORDER_PRODUCT_FAILED", "error_message", err.Error(), "user_id", uid,
				"product_id", selectedProduct.ProductID)
			refundBurnedPoints()
			return Order{}, err
		}

		orderService.CartRepository.DeleteCart(ctx, uid, selectedProduct.ProductID)
	}

	if metrics.OrdersCreated != nil {
		metrics.OrdersCreated.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("status", "success"),
				attribute.String("payment_method", PaymentMethod[submitedOrder.PaymentMethodID]),
				attribute.String("shipping_method", ShippingMethod[submitedOrder.ShippingMethodID]),
			),
		)
		metrics.OrderRevenue.Add(ctx, orderDetail.TotalPrice,
			metric.WithAttributes(
				attribute.String("payment_method", PaymentMethod[submitedOrder.PaymentMethodID]),
			),
		)
		metrics.OrderItemsCount.Record(ctx, int64(len(submitedOrder.Cart)))
		metrics.OrderTotalPrice.Record(ctx, orderDetail.TotalPrice)
	}

	slog.InfoContext(ctx, "Order completed",
		"log_type", "business",
		"event", "order_completed",
		"entity_type", "order",
		"entity_id", orderNumber,
		"actor_id", uid,
		slog.Any("metadata", map[string]any{
			"subtotal_thb":     orderDetail.SubTotalPrice,
			"discount_thb":     orderDetail.DiscountPrice,
			"shipping_fee_thb": orderDetail.ShippingFee,
			"total_price_thb":  orderDetail.TotalPrice,
			"burn_point":       orderDetail.BurnPoint,
			"earn_point":       orderDetail.EarnPoint,
			"payment_method":   PaymentMethod[submitedOrder.PaymentMethodID],
			"shipping_method":  ShippingMethod[submitedOrder.ShippingMethodID],
			"item_count":       len(submitedOrder.Cart),
		}),
	)

	return Order{
		OrderNumber: orderNumber,
	}, nil
}

func (orderService OrderService) OrderBurnPoint(ctx context.Context, uid int, burn int) (point.TotalPoint, error) {
	submit := point.SubmitedPoint{
		Amount: -(burn),
	}

	totalPoint, err := orderService.PointService.DeductPoint(ctx, uid, submit)
	if err != nil {
		return point.TotalPoint{}, err
	}
	return totalPoint, nil
}

func (orderService OrderService) GetOrderSummary(ctx context.Context, orderNumber int64) (OrderSummary, error) {
	orderDetail, err := orderService.OrderRepository.GetOrderWithTrackingNumberByOrderNumber(ctx, orderNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.ErrorContext(ctx, "Order not found",
				"log_type", "error", "error_code", "ORDER_NOT_FOUND", "error_message", err.Error(),
				"user_id", 0, "order_number", orderNumber)
			return OrderSummary{}, ErrOrderNotFound
		}
		slog.ErrorContext(ctx, "OrderRepository.GetOrderWithTrackingNumberByOrderNumber failed",
			"log_type", "error", "error_code", "ORDER_QUERY_FAILED", "error_message", err.Error(),
			"user_id", 0, "order_number", orderNumber)
		return OrderSummary{}, err
	}

	orderedProducts, err := orderService.OrderRepository.GetOrderProductWithPrice(ctx, orderDetail.ID)
	if err != nil {
		slog.ErrorContext(ctx, "OrderRepository.GetOrderProductWithPrice failed",
			"log_type", "error", "error_code", "ORDER_PRODUCT_QUERY_FAILED", "error_message", err.Error(),
			"user_id", 0, "order_number", orderNumber)
		return OrderSummary{}, err
	}

	var productList []OrderSummaryProduct
	for _, orderProduct := range orderedProducts {
		totalPrice := orderProduct.Price * float64(orderProduct.Quantity)

		totalPriceTHB := common.ConvertToThb(totalPrice)
		priceTHB := common.ConvertToThb(orderProduct.Price)
		product := OrderSummaryProduct{
			ProductBrand:  orderProduct.ProductBrand,
			ProductName:   orderProduct.ProductName,
			Quantity:      orderProduct.Quantity,
			PriceTHB:      priceTHB.ShortDecimal,
			TotalPriceTHB: totalPriceTHB.ShortDecimal,
		}
		productList = append(productList, product)
	}

	paymentMethod := PaymentMethod[orderDetail.PaymentMethodID]
	shippingMethod := ShippingMethod[orderDetail.ShippingMethodID]

	userDetail, err := orderService.UserRepository.FindByID(ctx, orderDetail.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "UserRepository.FindByID failed",
			"log_type", "error", "error_code", "USER_QUERY_FAILED", "error_message", err.Error(),
			"user_id", orderDetail.UserID)
		return OrderSummary{}, err
	}

	factor2 := math.Pow(10, 2)
	subTotal := math.Round(orderDetail.SubTotalPrice*factor2) / factor2
	totalPrice := math.Round(orderDetail.TotalPrice*factor2) / factor2
	shippingFee := math.Round(orderDetail.ShippingFee*factor2) / factor2
	discountPrice := math.Round(orderDetail.DiscountPrice*factor2) / factor2

	bangkok, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		slog.ErrorContext(ctx, "Could not load timezone",
			"log_type", "error", "error_code", "TIMEZONE_LOAD_FAILED", "error_message", err.Error(), "user_id", 0)
		return OrderSummary{}, err
	}

	issuedDate := orderDetail.Updated.In(bangkok).Format("02-01-2006 15:04:05")

	orderSummary := OrderSummary{
		OrderNumber:      orderDetail.OrderNumber,
		FirstName:        userDetail.FirstName,
		LastName:         userDetail.LastName,
		TrackingNumber:   orderDetail.TrackingNumber,
		ShippingMethod:   shippingMethod,
		PaymentMethod:    paymentMethod,
		OrderProductList: productList,
		SubTotalPrice:    subTotal,
		DiscountPrice:    discountPrice,
		TotalPrice:       totalPrice,
		ShippingFee:      shippingFee,
		BurnPoint:        orderDetail.BurnPoint,
		ReceivingPoint:   orderDetail.EarnPoint,
		IssuedDate:       issuedDate,
	}

	return orderSummary, nil
}

func (orderService OrderService) GeneratePDFFromData(orderSummary OrderSummary) ([]byte, error) {
	pdfBytes, err := orderService.PDFHelper.GenerateOrderSummaryPDF(orderSummary)
	if err != nil {
		slog.Error("PDFHelper.GenerateOrderSummaryPDF failed",
			"log_type", "error", "error_code", "PDF_GENERATION_FAILED", "error_message", err.Error(), "user_id", 0)
		return []byte(""), err
	}

	return pdfBytes, nil
}
