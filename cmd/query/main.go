package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "r0ckwe11"
	password = "postgres1"
	dbname   = "mydatabase"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

func (h *Handlers) Handler(w http.ResponseWriter, r *http.Request) {
	// Добавляем заголовки CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")                   // Разрешить доступ всем источникам (можно указать конкретный домен вместо "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS") // Разрешённые методы
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")       // Разрешённые заголовки

	// Обрабатываем OPTIONS-запросы
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.URL.Query().Has("name") {
		err := h.dbProvider.InsertHello(r.URL.Query().Get("name"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.Write([]byte("Hello," + r.URL.Query().Get("name") + "!"))
	} else {
		msg, err := h.dbProvider.SelectHello()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		w.Write([]byte("Hello," + msg + "!"))
	}
}

func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string

	// Получаем одно сообщение из таблицы hello, отсортированной в случайном порядке
	row := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}

func (dp *DatabaseProvider) InsertHello(msg string) error {
	query := `SELECT EXISTS(SELECT 1 FROM hello WHERE message = $1);`
	var exists bool
	err := dp.db.QueryRow(query, msg).Scan(&exists)

	if err != nil {
		return fmt.Errorf("ошибка при запросе: %v", err)
	}
	if !exists {
		_, err := dp.db.Exec("INSERT INTO hello (message) VALUES ($1)", msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	address := flag.String("address", "127.0.0.1:9000", "адрес для запуска сервера")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}

	h := Handlers{dbProvider: dp}

	http.HandleFunc("/api/user", h.Handler)

	err = http.ListenAndServe(*address, nil)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
