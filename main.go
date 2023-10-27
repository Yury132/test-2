package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

// Для функции calculateRatioFit
// const DEFAULT_MAX_WIDTH float64 = 320
// const DEFAULT_MAX_HEIGHT float64 = 240

// Рассчитываем размер изображения после масштабирования
// func calculateRatioFit(srcWidth, srcHeight int) (int, int) {
// 	ratio := math.Min(DEFAULT_MAX_WIDTH/float64(srcWidth), DEFAULT_MAX_HEIGHT/float64(srcHeight))
// 	return int(math.Ceil(float64(srcWidth) * ratio)), int(math.Ceil(float64(srcHeight) * ratio))
// }

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

	// Путь к миниатюре картинки
	fileNameMini := "./miniature/" + fileName + ".png"
	// Путь к исходной картинке
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

	// Читаем тело запроса, получаем срез байтов
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("failed")
		panic(err)
	}
	r.Body.Close()

	// Сохраняем картинку
	err = os.WriteFile(fileName, data, 0666)
	if err != nil {
		log.Println("failed to save your image...")
		fmt.Println(err)
	}

	// Открываем ранее сохраненную картинку
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("123")
		log.Fatal(err)
	}

	// Размер исходного изображения
	info, err := file.Stat()
	if err != nil {
		fmt.Println("failed to get info...")
		log.Fatal(err)
	}
	fmt.Println("Размер Изображения в байтах: = ", info.Size())

	defer file.Close()
	// // Получаем размеры картинки и формат через DecodeConfig
	// config, format, err := image.DecodeConfig(file)
	// if err != nil {
	// 	fmt.Println("456")
	// 	log.Fatal(err)
	// }
	// fmt.Println("DecodeConfig:", format, config.Height, config.Width)

	// // После image.DecodeConfig и до image.Decode необходимо поставить указатель на начало считываемого файла
	// file.Seek(0, 0)

	// Получаем image.Image
	imageData, imageType, err := image.Decode(file)
	if err != nil {
		fmt.Println("failed to decode...")
		log.Fatal(err)
	}
	// Массив байтов
	//fmt.Println("Decode:", imageData)
	fmt.Println("Тип Изображения: = ", imageType)
	b := imageData.Bounds()
	widthDecode := b.Max.X
	heightDecode := b.Max.Y
	fmt.Println("Ширина Изображения: = ", widthDecode)
	fmt.Println("Высота Изображения: = ", heightDecode)

	// Масштабируем изображение - Расчет
	//width, height := calculateRatioFit(config.Width, config.Height)

	//fmt.Println("width = ", config.Width, " height = ", config.Height)
	//fmt.Println("width = ", width, "height = ", height)

	// Создаем миниатюру
	newImage := resize.Thumbnail(100, 100, imageData, resize.Lanczos3)
	fmt.Println("Создание миниатюры...")
	newImageWidth := newImage.Bounds().Max.X
	newImageHeight := newImage.Bounds().Max.Y
	fmt.Println("Ширина Миниатюры: = ", newImageWidth)
	fmt.Println("Высота Миниатюры: = ", newImageHeight)

	// Файл для сохранения миниатюры
	imgfile, err := os.Create(fileNameMini)
	if err != nil {
		fmt.Println("789")
		log.Fatal(err)
	}
	defer imgfile.Close()

	// Сохраняем миниатюру в формате PNG
	err = png.Encode(imgfile, newImage)
	if err != nil {
		fmt.Println("10,11,12")
		log.Fatal(err)
	}

	// Получаем размер миниатюры
	miniInfo, err := os.Stat(fileNameMini)
	if err != nil {
		fmt.Println("789,123,431")
		log.Fatal(err)
	}

	// Размер
	miniSize := miniInfo.Size()
	fmt.Println("Размер Миниатюры в байтах: = ", miniSize)
}
