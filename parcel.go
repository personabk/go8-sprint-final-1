package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	query := `INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`

	res, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err // Если ошибка, возвращаем 0 и ошибку
	}

	// верните идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	query := `SELECT * FROM parcel WHERE number = ?`
	// здесь из таблицы должна вернуться только одна строка
	row := s.db.QueryRow(query, number)
	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	query := `SELECT * FROM parcel WHERE client = ?`
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// заполните срез Parcel данными из таблицы
	var res []Parcel

	for rows.Next() {
		var p Parcel
		// Сканируем текущую строку в временную структуру p
		// Порядок полей тот же: number, client, status, address, created_at
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	query := `UPDATE parcel SET status = ? WHERE number = ?`

	_, err := s.db.Exec(query, status, number)

	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	queryCheck := `SELECT status FROM parcel WHERE number = ?`
	var currentStatus string
	err := s.db.QueryRow(queryCheck, number).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("посылка с номером %d не найдена", number)
		}
		return err
	}
	if currentStatus != ParcelStatusRegistered {
		return fmt.Errorf(
			"нельзя изменить адрес посылки со статусом '%s', разрешен только статус '%s'",
			currentStatus, ParcelStatusRegistered,
		)
	}
	queryUpdate := `UPDATE parcel SET address = ? WHERE number = ?`
	_, err = s.db.Exec(queryUpdate, address, number)
	if err != nil {
		return err
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	queryCheck := `SELECT status FROM parcel WHERE number = ?`
	var currentStatus string

	// Используем QueryRow, так как нам нужна ровно одна строка
	err := s.db.QueryRow(queryCheck, number).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("посылка с номером %d не найдена", number)
		}
		return err
	}

	// 2. ПРОВЕРКА: удалять можно ТОЛЬКО если статус "registered"
	if currentStatus != ParcelStatusRegistered {
		return fmt.Errorf("нельзя удалить посылку со статусом '%s', разрешен только статус '%s'", currentStatus, ParcelStatusRegistered)
	}

	// 3. Если всё ок — удаляем
	queryDelete := `DELETE FROM parcel WHERE number = ?`
	_, err = s.db.Exec(queryDelete, number)
	if err != nil {
		return err
	}

	return nil
}
