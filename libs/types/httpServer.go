package htypes

type OrderParams struct {
	OrderId string  `json:"order_id"`
	UserId string	`json:"user_id"`
	Market string	`json:"market"`
	Side string		`json:"side"`
	Amount string	`json:"amount"`
	Price string	`json:"price"`
	TakeFee string	`json:"take_fee"`
	MakerFee string	`json:"make_fee"`
	Source string	`json:"source"`
}

type ResponseDataFormat struct {
	Status int `json:"status"`
	Msg string `json:msg`
	Data []interface{} `json:data,omitempty`
}