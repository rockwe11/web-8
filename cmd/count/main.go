package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	if r.Method == "POST" {
		a, err := strconv.Atoi(r.FormValue("count"))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("это не число"))
			return
		}
		err = h.dbProvider.AddCount(a)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.Write([]byte("OK!"))
		return
	} else if r.Method == "GET" {
		value, _ := h.dbProvider.GetCount()
		w.Write([]byte(strconv.Itoa(value)))
		return
	}
	w.Write([]byte("Разрешен только метод POST и GET!"))
}

func (dp *DatabaseProvider) GetCount() (int, error) {
	var value int

	// Получаем одно сообщение из таблицы hello, отсортированной в случайном порядке
	row := dp.db.QueryRow("SELECT COALESCE(count, 0) FROM count WHERE name=$1", "key1")
	err := row.Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (dp *DatabaseProvider) AddCount(a int) error {
	_, err := dp.db.Exec("INSERT INTO count (name, count) VALUES ($2, $1) ON CONFLICT (name) DO UPDATE SET count = count.count + $1", a, "key1")
	if err != nil {
		return err
	}

	return nil
}

func main() {
	address := flag.String("address", "127.0.0.1:3333", "адрес для запуска сервера")
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

	http.HandleFunc("/count", h.Handler)
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
