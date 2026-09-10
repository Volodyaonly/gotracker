-- Вставка нового заказа
INSERT INTO orders (id, customer, address) VALUES (1, 'Олег', 'ул. Ленина, 10');

-- Получение всех заказов
SELECT * FROM orders;

-- Получение всех доставленных заказов клиента "Олег"
SELECT * FROM orders WHERE customer = 'Олег' AND is_delivered = TRUE;

-- Обновление статуса заказа
UPDATE orders SET is_delivered = TRUE WHERE id = 1;

-- Удаление заказа
DELETE FROM orders WHERE id = 1;

-- Подсчёт заказов по статусу
SELECT is_delivered, COUNT(*) FROM orders GROUP BY is_delivered;