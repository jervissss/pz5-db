package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Task — модель для сканирования результатов SELECT
type Task struct {
	ID        int
	Title     string
	Done      bool
	CreatedAt time.Time
}

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo { return &Repo{DB: db} }

// CreateTask — параметризованный INSERT с возвратом id
func (r *Repo) CreateTask(ctx context.Context, title string) (int, error) {
	var id int
	const q = `INSERT INTO tasks (title) VALUES ($1) RETURNING id;`
	err := r.DB.QueryRowContext(ctx, q, title).Scan(&id)
	return id, err
}

// ListTasks — базовый SELECT всех задач (демо для занятия)
func (r *Repo) ListTasks(ctx context.Context) ([]Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks ORDER BY id;`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ListDone возвращает задачи по статусу выполнения (true - выполненные, false - невыполненные)
func (r *Repo) ListDone(ctx context.Context, done bool) ([]Task, error) {
	// SQL-запрос с фильтром по полю done
	const q = `SELECT id, title, done, created_at FROM tasks WHERE done = $1 ORDER BY id;`

	// Выполняем запрос с параметром
	rows, err := r.DB.QueryContext(ctx, q, done)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		// Сканируем данные из строки в структуру Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	// Проверяем ошибки после цикла
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// FindByID находит задачу по её ID
func (r *Repo) FindByID(ctx context.Context, id int) (*Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks WHERE id = $1;`

	var task Task
	// QueryRowContext используется для запросов, которые возвращают максимум одну строку
	err := r.DB.QueryRowContext(ctx, q, id).Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Если задача не найдена, возвращаем nil и ошибку
			return nil, fmt.Errorf("задача с ID %d не найдена", id)
		}
		return nil, err
	}

	return &task, nil
}

// CreateMany добавляет несколько задач одной транзакцией
func (r *Repo) CreateMany(ctx context.Context, titles []string) error {
	// Начинаем транзакцию
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}

	// Гарантируем откат транзакции в случае ошибки
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	// Подготавливаем SQL-запрос
	const q = `INSERT INTO tasks (title) VALUES ($1);`
	stmt, err := tx.PrepareContext(ctx, q)
	if err != nil {
		return fmt.Errorf("не удалось подготовить запрос: %v", err)
	}
	defer stmt.Close()

	// Выполняем вставку для каждого заголовка
	for _, title := range titles {
		_, err := stmt.ExecContext(ctx, title)
		if err != nil {
			return fmt.Errorf("не удалось вставить задачу '%s': %v", title, err)
		}
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("не удалось подтвердить транзакцию: %v", err)
	}

	return nil
}
