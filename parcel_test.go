package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// 1. PREPARE: Создаем подключение к БД
	// Для тестов лучше использовать ":memory:", чтобы не засорять файл tracker.db мусором от тестов.
	// Каждый запуск теста будет начинаться с чистой базы.
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Не удалось открыть соединение с БД для теста")
	defer db.Close() // Обязательно закрываем в конце теста

	// 2. ВАЖНО: Создаем таблицу вручную.
	// В отличие от реального приложения, в тестах таблица не создается сама.
	// Нам нужно явно сказать SQLite: "Создай таблицу parcel".
	createTableQuery := `CREATE TABLE IF NOT EXISTS parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`
	_, err = db.Exec(createTableQuery)
	require.NoError(t, err, "Не удалось создать таблицу в БД для теста")

	// Создаем наш сервис хранения
	store := NewParcelStore(db)

	// Готовим тестовые данные
	parcel := getTestParcel()

	// 3. ADD: Добавляем посылку
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	require.Greater(t, id, 0, "ID посылки должен быть положительным числом")

	// Обновляем ID в нашей локальной структуре, чтобы дальше сравнивать именно этот ID
	parcel.Number = id

	gotParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка при получении посылки")

	require.Equal(t, parcel.Client, gotParcel.Client, "Не совпадает Client")
	require.Equal(t, parcel.Status, gotParcel.Status, "Не совпадает Status")
	require.Equal(t, parcel.Address, gotParcel.Address, "Не совпадает Address")

	require.NotEmpty(t, gotParcel.CreatedAt, "Поле CreatedAt не должно быть пустым")

	err = store.Delete(id)
	require.NoError(t, err, "Ошибка при удалении посылки")

	_, err = store.Get(id)
	require.Error(t, err, "После удаления посылка должна быть недоступна (должна быть ошибка)")

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// 1. PREPARE: Стандартная подготовка базы в памяти
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Не удалось открыть БД")
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err, "Не удалось создать таблицу")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// 2. ADD: Сначала добавляем посылку
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	parcel.Number = id // Сохраняем полученный ID

	// 3. SET ADDRESS: Меняем адрес
	newAddress := "New Street, Building 5, City X"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Ошибка при изменении адреса")

	// 4. CHECK: Получаем посылку и проверяем, что адрес изменился
	updatedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка при получении обновленной посылки")

	require.Equal(t, newAddress, updatedParcel.Address, "Адрес не обновился!")

	// Также проверим, что остальные поля не пострадали
	require.Equal(t, parcel.Client, updatedParcel.Client, "Client изменился случайно")
	require.Equal(t, parcel.Status, updatedParcel.Status, "Status изменился случайно")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// 1. PREPARE: Создаем БД в памяти и таблицу
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Не удалось открыть БД")
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err, "Не удалось создать таблицу")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// 2. ADD: Добавляем посылку (статус по умолчанию "registered")
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	parcel.Number = id

	// 3. SET STATUS: Меняем статус на "sent"
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err, "Ошибка при изменении статуса")

	// 4. CHECK: Получаем посылку и проверяем, что статус обновился
	updatedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка получения посылки")

	require.Equal(t, newStatus, updatedParcel.Status, "Статус не обновился!")

	// Проверка, что другие поля не изменились
	require.Equal(t, parcel.Client, updatedParcel.Client, "Client изменился случайно")
	require.Equal(t, parcel.Address, updatedParcel.Address, "Address изменился случайно")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Не удалось открыть БД")
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err, "Не удалось создать таблицу")

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	// Используем ту самую мапу из заготовки
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// добавляем новую посылку в БД
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "Ошибка добавления посылки %d", i)

		// обновляем идентификатор добавленной посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map по ключу ID
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Ошибка получения списка посылок клиента")

	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.Len(t, storedParcels, 3, "Количество полученных посылок (%d) не совпадает с ожидаемым (3)", len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки.
		// Проверяем, что посылка из результата есть в нашей мапе
		expectedParcel, ok := parcelMap[parcel.Number]
		require.True(t, ok, "Посылка с ID %d не найдена в списке добавленных посылок", parcel.Number)

		// убедитесь, что значения полей полученных посылок заполнены верно
		require.Equal(t, expectedParcel.Client, parcel.Client, "Не совпадает Client")
		require.Equal(t, expectedParcel.Status, parcel.Status, "Не совпадает Status")
		require.Equal(t, expectedParcel.Address, parcel.Address, "Не совпадает Address")
	}
}
