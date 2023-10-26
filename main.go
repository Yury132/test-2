package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/uploads", uploads)
	http.ListenAndServe(":8080", nil)
}

func uploads(w http.ResponseWriter, r *http.Request) {

	// Имя файла уникальное - текущее время с наносекундами
	currentTime := time.Now()
	fileName := currentTime.Format("2006-01-02 15:04:05.000000000")
	fileName = strings.Replace(fileName, " ", "_", -1)
	fileName = strings.Replace(fileName, ":", "_", -1)
	fileName = strings.Replace(fileName, ".", "_", -1)
	fileName = "./images/" + fileName + ".png"

	// Вывод всех заголовков
	//
	// for name, headers := range r.Header {
	// 	for _, h := range headers {
	// 		fmt.Fprintf(w, "%v: %v\n", name, h)
	// 	}
	// }

	// Проверка на Метод POST
	//
	// if r.Method != http.MethodPost {
	// 	fmt.Println("ошибка1")
	// 	w.WriteHeader(http.StatusMethodNotAllowed)
	// 	return
	// }

	// Читаем тело запроса
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("failed")
		panic(err)
	}
	r.Body.Close()

	// Сохраняем
	err = os.WriteFile(fileName, data, 0666)
	if err != nil {
		log.Println("failed to save your image...")
		fmt.Println(err)
	}

	// Выводим
	//fmt.Println(data)
	//w.Write(data)                // --Или-- Отображение самой картинки в postman
	//fmt.Fprintf(w, "%v\n", data) // --Или-- Вывод среза байтов картинки в postman

	// Тип файла
	//fmt.Println(r.Header["Content-Type"])
	//fmt.Fprintf(w, "%v\n", r.Header["Content-Type"])

	// Размер в байтах
	//fmt.Println(r.Header["Content-Length"])
	//fmt.Fprintf(w, "%v\n", r.Header["Content-Length"])

}
