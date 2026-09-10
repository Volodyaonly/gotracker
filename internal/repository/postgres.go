package repository

import (
	"database/sql"
	"fmt"

	"gotracker/internal/order"

	"github.com/jmoiron/sqlx"
)

type PostgresOrderRepo struct {
	db *sqlx.DB
}

func NewPostgresOrderRepo(db *sqlx.DB) *PostgresOrderRepo {
	return &PostgresOrderRepo{
		db: db,
	}
}

func (r *PostgresOrderRepo) Add(o order.Order) error {
	query := `
		INSERT INTO orders (
			id,
			customer,
			address,
			is_delivered
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(
		query,
		o.ID,
		o.Customer,
		o.Address,
		o.IsDelivered,
	)

	return err
}

func (r *PostgresOrderRepo) GetByID(id int) (order.Order, error) {
	var o order.Order

	query := `
		SELECT id, customer, address, is_delivered
		FROM orders
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&o.ID,
		&o.Customer,
		&o.Address,
		&o.IsDelivered,
	)

	if err == sql.ErrNoRows {
		return order.Order{}, order.ErrOrderNotFound
	}

	if err != nil {
		return order.Order{}, err
	}

	return o, nil
}

func (r *PostgresOrderRepo) Update(o order.Order) error {
	query := `
		UPDATE orders
		SET customer = $1,
		    address = $2,
		    is_delivered = $3
		WHERE id = $4
	`

	result, err := r.db.Exec(
		query,
		o.Customer,
		o.Address,
		o.IsDelivered,
		o.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return order.ErrOrderNotFound
	}

	return nil
}

func (r *PostgresOrderRepo) GetAll() []order.Order {
	query := `
		SELECT id, customer, address, is_delivered
		FROM orders
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		fmt.Println("ошибка получения заказов:", err)
		return nil
	}
	defer rows.Close()

	orders := make([]order.Order, 0)

	for rows.Next() {
		var o order.Order

		if err := rows.Scan(
			&o.ID,
			&o.Customer,
			&o.Address,
			&o.IsDelivered,
		); err != nil {
			fmt.Println("ошибка чтения заказа:", err)
			continue
		}

		orders = append(orders, o)
	}

	return orders
}
