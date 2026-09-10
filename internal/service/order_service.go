package service

import (
	"errors"
	"fmt"
	"strings"

	"gotracker/internal/cache"
	"gotracker/internal/order"
	"gotracker/internal/queue"
	"gotracker/internal/repository"
)

const (
	StatusPending   = "pending"
	StatusDelivered = "delivered"
)

var ErrInvalidOrder = errors.New("invalid order")

type OrderService struct {
	repo  repository.Repository
	cache *cache.RedisCache
}

func NewOrderService(
	repo repository.Repository,
	cache *cache.RedisCache,
) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

func IsValidStatus(status string) bool {
	return status == StatusPending || status == StatusDelivered
}

// CreateOrder keeps the service API expected by the HTTP layer.
// After a successful PostgreSQL insert it invalidates the cache and publishes
// an order-created event to Kafka.
func (s *OrderService) CreateOrder(o order.Order) (order.Order, error) {
	o.Customer = strings.TrimSpace(o.Customer)
	o.Address = strings.TrimSpace(o.Address)

	if o.ID <= 0 || o.Customer == "" || o.Address == "" {
		return order.Order{}, ErrInvalidOrder
	}

	if err := s.repo.Add(o); err != nil {
		return order.Order{}, err
	}

	if s.cache != nil {
		if err := s.cache.Delete(o.ID); err != nil {
			fmt.Println("[Cache] Не удалось очистить кэш:", err)
		}
	}

	// Kafka is an auxiliary side effect in this training project: the order is
	// already stored in PostgreSQL, so a Kafka failure is logged rather than
	// turning the successful create into an HTTP 500.
	if err := queue.SendOrderCreated(o); err != nil {
		fmt.Println("[Kafka] Не удалось отправить событие:", err)
	}

	return o, nil
}

// AddOrder is retained as a convenience wrapper for the naming used in the
// current module materials.
func (s *OrderService) AddOrder(o order.Order) error {
	_, err := s.CreateOrder(o)
	return err
}

func (s *OrderService) ListOrders() ([]order.Order, error) {
	return s.repo.GetAll(), nil
}

func (s *OrderService) GetAll() []order.Order {
	return s.repo.GetAll()
}

func (s *OrderService) GetByID(id int) (order.Order, error) {
	if id <= 0 {
		return order.Order{}, ErrInvalidOrder
	}

	if s.cache != nil {
		o, err := s.cache.Get(id)
		if err == nil {
			fmt.Println("[Cache] Найден заказ в Redis")
			return o, nil
		}
	}

	fmt.Println("[Cache] Читаем заказ из PostgreSQL")

	o, err := s.repo.GetByID(id)
	if err != nil {
		return order.Order{}, err
	}

	if s.cache != nil {
		if err := s.cache.Set(o); err != nil {
			fmt.Println("[Cache] Ошибка записи:", err)
		}
	}

	return o, nil
}

func (s *OrderService) UpdateOrder(
	id int,
	address string,
	status string,
) (order.Order, error) {
	address = strings.TrimSpace(address)
	status = strings.ToLower(strings.TrimSpace(status))

	if id <= 0 || address == "" || !IsValidStatus(status) {
		return order.Order{}, ErrInvalidOrder
	}

	currentOrder, err := s.repo.GetByID(id)
	if err != nil {
		return order.Order{}, err
	}

	currentOrder.Address = address
	currentOrder.IsDelivered = status == StatusDelivered

	if err := s.repo.Update(currentOrder); err != nil {
		return order.Order{}, err
	}

	if s.cache != nil {
		if err := s.cache.Delete(id); err != nil {
			fmt.Println("[Cache] Не удалось очистить кэш:", err)
		}
	}

	return currentOrder, nil
}

// Update is retained for compatibility with the repository-shaped API used in
// the module solution.
func (s *OrderService) Update(o order.Order) error {
	if o.ID <= 0 || strings.TrimSpace(o.Customer) == "" || strings.TrimSpace(o.Address) == "" {
		return ErrInvalidOrder
	}

	if err := s.repo.Update(o); err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.Delete(o.ID); err != nil {
			fmt.Println("[Cache] Не удалось очистить кэш:", err)
		}
	}

	return nil
}
