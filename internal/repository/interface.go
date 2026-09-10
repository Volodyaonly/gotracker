package repository

import "gotracker/internal/order"

type Repository interface {
	Add(order.Order) error
	GetByID(id int) (order.Order, error)
	Update(order.Order) error
	GetAll() []order.Order
}
