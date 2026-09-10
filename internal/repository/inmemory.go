package repository

import (
	"sync"

	"gotracker/internal/order"
)

type InMemoryOrderRepo struct {
	orders map[int]order.Order
	mu     sync.Mutex
}

func NewInMemoryOrderRepo() *InMemoryOrderRepo {
	return &InMemoryOrderRepo{
		orders: make(map[int]order.Order),
	}
}

// NewInMemoryRepository добавлен под требования нового задания.
// Старый конструктор при этом продолжает работать.
func NewInMemoryRepository() *InMemoryOrderRepo {
	return NewInMemoryOrderRepo()
}

func (r *InMemoryOrderRepo) Add(o order.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[o.ID] = o

	return nil
}

func (r *InMemoryOrderRepo) GetByID(id int) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, ok := r.orders[id]
	if !ok {
		return order.Order{}, order.ErrOrderNotFound
	}

	return o, nil
}

func (r *InMemoryOrderRepo) Update(o order.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.orders[o.ID]; !ok {
		return order.ErrOrderNotFound
	}

	r.orders[o.ID] = o
	return nil
}

func (r *InMemoryOrderRepo) GetAll() []order.Order {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]order.Order, 0, len(r.orders))

	for _, o := range r.orders {
		result = append(result, o)
	}

	return result
}
