package point

import (
	"context"
	"fmt"
	"log/slog"
)

type PointInterface interface {
	TotalPoint(ctx context.Context, uid int) (TotalPoint, error)
	DeductPoint(ctx context.Context, uid int, submitedPoint SubmitedPoint) (TotalPoint, error)
	CheckBurnPoint(ctx context.Context, uid int, amount int) (bool, error)
	CalculatePoint(ctx context.Context, amount float64) (TotalPoint, error)
	CalculateDiscount(ctx context.Context, points int, subtotal float64) (DiscountQuote, error)
}

type PointService struct {
	PointGateway PointGatewayInterface
}

type PointGatewayInterface interface {
	GetPoints(ctx context.Context, uid int) ([]Point, error)
	CreatePoint(ctx context.Context, uid int, body Point) (Point, error)
	CalculatePoint(ctx context.Context, amount float64) (int, error)
	CalculateDiscount(ctx context.Context, points int, subtotal float64) (DiscountQuote, error)
}

func (pointService PointService) TotalPoint(ctx context.Context, uid int) (TotalPoint, error) {
	points, err := pointService.PointGateway.GetPoints(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "PointGateway.GetPoints failed",
			"log_type", "error", "error_code", "POINT_GATEWAY_FAILED", "error_message", err.Error(), "user_id", uid)
	}

	total := 0
	for _, point := range points {
		total += point.Amount
	}
	return TotalPoint{
		Point: total,
	}, err
}

func (pointService PointService) DeductPoint(ctx context.Context, uid int, submitedPoint SubmitedPoint) (TotalPoint, error) {
	_, err := pointService.CheckBurnPoint(ctx, uid, submitedPoint.Amount)
	if err != nil {
		return TotalPoint{}, err
	}

	point := Point{
		OrgID:  1,
		UserID: 1,
		Amount: submitedPoint.Amount,
	}
	_, err_ := pointService.PointGateway.CreatePoint(ctx, uid, point)
	if err_ != nil {
		slog.ErrorContext(ctx, "PointGateway.CreatePoint failed",
			"log_type", "error", "error_code", "POINT_CREATE_FAILED", "error_message", err_.Error(),
			"user_id", uid, "amount", submitedPoint.Amount)
		return TotalPoint{}, err_
	}
	return pointService.TotalPoint(ctx, uid)
}

func (pointService PointService) CalculatePoint(ctx context.Context, amount float64) (TotalPoint, error) {
	receivedPoint, err := pointService.PointGateway.CalculatePoint(ctx, amount)
	if err != nil {
		slog.ErrorContext(ctx, "PointGateway.CalculatePoint failed",
			"log_type", "error", "error_code", "POINT_GATEWAY_FAILED", "error_message", err.Error(), "amount", amount)
		return TotalPoint{}, err
	}
	return TotalPoint{
		Point: receivedPoint,
	}, nil
}

func (pointService PointService) CalculateDiscount(ctx context.Context, points int, subtotal float64) (DiscountQuote, error) {
	quote, err := pointService.PointGateway.CalculateDiscount(ctx, points, subtotal)
	if err != nil {
		slog.ErrorContext(ctx, "PointGateway.CalculateDiscount failed",
			"log_type", "error", "error_code", "POINT_GATEWAY_FAILED", "error_message", err.Error(),
			"points", points, "subtotal", subtotal)
		return DiscountQuote{}, err
	}
	return quote, nil
}

func (pointService PointService) CheckBurnPoint(ctx context.Context, uid int, amount int) (bool, error) {
	total, err := pointService.TotalPoint(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "PointService.TotalPoint failed",
			"log_type", "error", "error_code", "POINT_CHECK_FAILED", "error_message", err.Error(), "user_id", uid)
		return false, err
	}
	if amount+total.Point < 0 {
		return false, fmt.Errorf("points are not enough, please try again")
	}
	return true, nil
}
