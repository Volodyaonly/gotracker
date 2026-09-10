package order

import "errors"

type Order struct {
	ID          int    `json:"id"`
	Customer    string `json:"customer"`
	Address     string `json:"address"`
	IsDelivered bool   `json:"isDelivered"`
}

func (o *Order) MarkDelivered() {
	o.IsDelivered = true
}

var ErrOrderNotFound = errors.New("order not found")
