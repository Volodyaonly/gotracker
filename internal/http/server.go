package http

import (
	"encoding/json"
	"errors"
	"log"
	stdhttp "net/http"
	"strconv"
	"strings"

	"gotracker/internal/order"
	"gotracker/internal/service"
)

type Server struct {
	service *service.OrderService
}

func NewServer(svc *service.OrderService) *Server {
	return &Server{
		service: svc,
	}
}

// createOrderRequest описывает JSON для POST /orders.
//
// Мы не меняем внутреннюю структуру order.Order.
// HTTP-слой принимает name/status и преобразует их
// в Customer/IsDelivered.
type createOrderRequest struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

// updateOrderRequest описывает JSON для PUT /orders/{id}.
type updateOrderRequest struct {
	Address string `json:"address"`
	Status  string `json:"status"`
}

// orderResponse описывает JSON, возвращаемый клиенту.
type orderResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

// ServeHTTP делает Server реализацией интерфейса net/http.Handler.
func (s *Server) ServeHTTP(
	w stdhttp.ResponseWriter,
	r *stdhttp.Request,
) {
	path := r.URL.Path

	switch {
	case path == "/":
		if r.Method != stdhttp.MethodGet {
			writeError(
				w,
				stdhttp.StatusMethodNotAllowed,
				"Метод не поддерживается",
			)
			return
		}

		s.handleRoot(w)

	case path == "/orders":
		switch r.Method {
		case stdhttp.MethodGet:
			s.handleListOrders(w)

		case stdhttp.MethodPost:
			s.handleCreateOrder(w, r)

		default:
			writeError(
				w,
				stdhttp.StatusMethodNotAllowed,
				"Метод не поддерживается",
			)
		}

	case strings.HasPrefix(path, "/orders/"):
		switch r.Method {
		case stdhttp.MethodGet:
			s.handleGetOrder(w, r)

		case stdhttp.MethodPut:
			s.handleUpdateOrder(w, r)

		default:
			writeError(
				w,
				stdhttp.StatusMethodNotAllowed,
				"Метод не поддерживается",
			)
		}

	default:
		writeError(
			w,
			stdhttp.StatusNotFound,
			"Маршрут не найден",
		)
	}
}

func (s *Server) handleRoot(w stdhttp.ResponseWriter) {
	w.Header().Set(
		"Content-Type",
		"text/plain; charset=utf-8",
	)
	w.WriteHeader(stdhttp.StatusOK)

	if _, err := w.Write(
		[]byte("Добро пожаловать в GoTracker API"),
	); err != nil {
		log.Printf("ошибка записи ответа: %v", err)
	}
}

func (s *Server) handleListOrders(w stdhttp.ResponseWriter) {
	orders, err := s.service.ListOrders()
	if err != nil {
		writeError(
			w,
			stdhttp.StatusInternalServerError,
			"Не удалось получить список заказов",
		)
		return
	}

	response := make([]orderResponse, 0, len(orders))

	for _, currentOrder := range orders {
		response = append(
			response,
			toOrderResponse(currentOrder),
		)
	}

	writeJSON(w, stdhttp.StatusOK, response)
}

func (s *Server) handleCreateOrder(
	w stdhttp.ResponseWriter,
	r *stdhttp.Request,
) {
	var request createOrderRequest

	if err := decodeJSON(r, &request); err != nil {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Некорректный JSON: "+err.Error(),
		)
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Address = strings.TrimSpace(request.Address)
	request.Status = strings.ToLower(
		strings.TrimSpace(request.Status),
	)

	if request.ID <= 0 {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"ID должен быть больше нуля",
		)
		return
	}

	if request.Name == "" {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Имя клиента обязательно",
		)
		return
	}

	if request.Address == "" {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Адрес обязателен",
		)
		return
	}

	if !service.IsValidStatus(request.Status) {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Статус должен быть pending или delivered",
		)
		return
	}

	newOrder := order.Order{
		ID:          request.ID,
		Customer:    request.Name,
		Address:     request.Address,
		IsDelivered: request.Status == service.StatusDelivered,
	}

	createdOrder, err := s.service.CreateOrder(newOrder)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(
		w,
		stdhttp.StatusCreated,
		toOrderResponse(createdOrder),
	)
}

func (s *Server) handleGetOrder(
	w stdhttp.ResponseWriter,
	r *stdhttp.Request,
) {
	id, err := parseOrderID(r.URL.Path)
	if err != nil {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Некорректный ID",
		)
		return
	}

	currentOrder, err := s.service.GetByID(id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(
		w,
		stdhttp.StatusOK,
		toOrderResponse(currentOrder),
	)
}

func (s *Server) handleUpdateOrder(
	w stdhttp.ResponseWriter,
	r *stdhttp.Request,
) {
	id, err := parseOrderID(r.URL.Path)
	if err != nil {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Некорректный ID",
		)
		return
	}

	var request updateOrderRequest

	if err := decodeJSON(r, &request); err != nil {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Некорректный JSON: "+err.Error(),
		)
		return
	}

	request.Address = strings.TrimSpace(request.Address)
	request.Status = strings.ToLower(
		strings.TrimSpace(request.Status),
	)

	if request.Address == "" {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Адрес обязателен",
		)
		return
	}

	if !service.IsValidStatus(request.Status) {
		writeError(
			w,
			stdhttp.StatusBadRequest,
			"Статус должен быть pending или delivered",
		)
		return
	}

	updatedOrder, err := s.service.UpdateOrder(
		id,
		request.Address,
		request.Status,
	)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(
		w,
		stdhttp.StatusOK,
		toOrderResponse(updatedOrder),
	)
}

func parseOrderID(path string) (int, error) {
	idString := strings.TrimPrefix(path, "/orders/")

	if idString == path ||
		idString == "" ||
		strings.Contains(idString, "/") {
		return 0, errors.New("invalid order ID")
	}

	id, err := strconv.Atoi(idString)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid order ID")
	}

	return id, nil
}

func decodeJSON(
	r *stdhttp.Request,
	destination any,
) error {
	decoder := json.NewDecoder(r.Body)

	// Запрещаем неизвестные поля.
	decoder.DisallowUnknownFields()

	return decoder.Decode(destination)
}

func toOrderResponse(currentOrder order.Order) orderResponse {
	status := service.StatusPending

	if currentOrder.IsDelivered {
		status = service.StatusDelivered
	}

	return orderResponse{
		ID:      currentOrder.ID,
		Name:    currentOrder.Customer,
		Address: currentOrder.Address,
		Status:  status,
	}
}

func handleServiceError(
	w stdhttp.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrInvalidOrder):
		writeError(
			w,
			stdhttp.StatusBadRequest,
			err.Error(),
		)

	case errors.Is(err, order.ErrOrderNotFound):
		writeError(
			w,
			stdhttp.StatusNotFound,
			"Заказ не найден",
		)

	default:
		log.Printf("внутренняя ошибка: %v", err)

		writeError(
			w,
			stdhttp.StatusInternalServerError,
			"Внутренняя ошибка сервера",
		)
	}
}

func writeJSON(
	w stdhttp.ResponseWriter,
	status int,
	data any,
) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("ошибка сериализации JSON: %v", err)

		stdhttp.Error(
			w,
			"Внутренняя ошибка сервера",
			stdhttp.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		log.Printf("ошибка отправки ответа: %v", err)
	}
}

func writeError(
	w stdhttp.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		w,
		status,
		map[string]string{
			"error": message,
		},
	)
}
