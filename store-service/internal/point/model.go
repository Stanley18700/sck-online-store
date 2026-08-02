package point

type SubmitedPoint struct {
	Amount int `json:"amount"`
}

type Point struct {
	UserID int `json:"userId"`
	Amount int `json:"amount"`
}

type TotalPoint struct {
	Point int `json:"point"`
}
