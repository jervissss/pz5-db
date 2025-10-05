package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:admin@localhost:5432/todo?sslmode=disable"
	}

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("openDB error: %v", err)
	}
	defer db.Close()

	repo := NewRepo(db)

	// Базовые операции из оригинального задания
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Вставка нескольких задач
	titles := []string{"Сделать ПЗ №5", "Купить кофе", "Проверить отчёты"}
	for _, title := range titles {
		id, err := repo.CreateTask(ctx, title)
		if err != nil {
			log.Printf("CreateTask error: %v", err)
		} else {
			log.Printf("Inserted task id=%d (%s)", id, title)
		}
	}

	// 2. Вывод всех задач
	tasks, err := repo.ListTasks(ctx)
	if err != nil {
		log.Fatalf("ListTasks error: %v", err)
	}

	fmt.Println("=== Все задачи ===")
	for _, t := range tasks {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}

	// ПРОВЕРОЧНЫЕ ЗАДАНИЯ:

	// Задание 1: ListDone
	fmt.Println("\n=== Невыполненные задачи (ListDone) ===")
	undoneTasks, err := repo.ListDone(ctx, false)
	if err != nil {
		log.Printf("ListDone error: %v", err)
	} else {
		for _, t := range undoneTasks {
			fmt.Printf("#%d | %s\n", t.ID, t.Title)
		}
	}

	// Задание 2: FindByID
	fmt.Println("\n=== Поиск по ID (FindByID) ===")
	task, err := repo.FindByID(ctx, 1)
	if err != nil {
		log.Printf("FindByID error: %v", err)
	} else {
		fmt.Printf("Найдена: #%d | %s | done=%v\n", task.ID, task.Title, task.Done)
	}

	// Задание 3: CreateMany
	fmt.Println("\n=== Массовое добавление (CreateMany) ===")
	newTasks := []string{"Массовая задача 1", "Массовая задача 2"}
	err = repo.CreateMany(ctx, newTasks)
	if err != nil {
		log.Printf("CreateMany error: %v", err)
	} else {
		fmt.Println("Успешно добавлены задачи одной транзакцией")

		// Покажем обновленный список
		updatedTasks, _ := repo.ListTasks(ctx)
		fmt.Println("Обновленный список:")
		for _, t := range updatedTasks {
			fmt.Printf("#%d | %s\n", t.ID, t.Title)
		}
	}
}
