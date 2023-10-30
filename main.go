package main

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nfnt/resize"
)

var js jetstream.JetStream
var ctx context.Context
var cancel context.CancelFunc
var fileName = ""
var fileNameMini = ""

// Воркер
func worker(id int, jobs <-chan string) {
	// Ожидаем получения данных для работы
	// Если данных нет в канале - блокировка
	for path := range jobs {
		fmt.Println("worker", id, "начал создание миниатюры по пути: ", path)
		// Создаем миниатюру, сохраняем данные в бд и прочее
		createMini(path)
		time.Sleep(time.Second * 1)
		fmt.Println("worker", id, "создал миниатюру по пути: ", path)
	}
}

func main() {
	// Каналы для воркера
	jobs := make(chan string, 100)

	// Сразу запускаем воркеров в горутинах
	// Они будут ожидать получения данных для работы
	for w := 1; w <= 3; w++ {
		go worker(w, jobs)
	}

	// Адрес сервера nats
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	// Подключаемся к серверу
	nc, err := nats.Connect(url)
	if err != nil {
		fmt.Println("err")
		fmt.Println("1")
	}
	defer nc.Drain()

	js, err = jetstream.New(nc)
	if err != nil {
		fmt.Println("err")
		fmt.Println("2")
	}

	cfg := jetstream.StreamConfig{
		Name: "EVENTS",
		// Очередь
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{"events.>"},
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Создаем поток
	stream, err := js.CreateStream(ctx, cfg)
	if err != nil {
		fmt.Println("err")
		fmt.Println("3")
	}
	fmt.Println("Создали поток")

	// Создаем получателя
	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "processor-1",
	})
	if err != nil {
		fmt.Println("err")
		fmt.Println("4")
	}

	// В горутине получатель беспрерывно ждет входящих сообщений
	// При получении сообщений, передает пути к изображениям (задачи) воркерам
	go func() {
		cons.Consume(func(msg jetstream.Msg) {
			// Печатаем полученные данные
			fmt.Println("Получатель получил сообщение - ", string(msg.Data()))
			// Заполняем канал данными
			// Воркеры начнут работать
			jobs <- string(msg.Data())
			// Подтверждаем получение сообщения
			msg.DoubleAck(ctx)
			//msg.Ack()
		})
	}()

	// Слушаем запрос
	http.HandleFunc("/uploads", uploads)
	http.ListenAndServe(":8080", nil)

	// Завершение программы по Ctrl+C
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT)
	<-shutdown
}

func uploads(w http.ResponseWriter, r *http.Request) {

	// Имя файла уникальное - текущее время с наносекундами
	currentTime := time.Now()
	fileName = currentTime.Format("2006-01-02 15:04:05.000000000")
	fileName = strings.Replace(fileName, " ", "_", -1)
	fileName = strings.Replace(fileName, ":", "_", -1)
	fileName = strings.Replace(fileName, ".", "_", -1)

	// Путь к миниатюре картинки
	fileNameMini = "./miniature/" + fileName + ".png"
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
		fmt.Println("5")
		panic(err)
	}
	r.Body.Close()

	// Сохраняем картинку
	err = os.WriteFile(fileName, data, 0666)
	if err != nil {
		log.Println("failed to save your image...")
		fmt.Println(err)
	}

	// Здесь ошибка---------через раз-------------------------------------------------------context deadline exceeded------------------
	_, err = js.Publish(ctx, "events.us.page_loaded", []byte(fileName))
	if err != nil {
		fmt.Println(err)
		fmt.Println("66")
	}

}

// Создание миниатюры воркерами
func createMini(fileName string) {

	// Открываем ранее сохраненную картинку
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("6")
		log.Fatal(err)
	}

	// Размер исходного изображения
	info, err := file.Stat()
	if err != nil {
		fmt.Println("failed to get info...")
		log.Fatal(err)
	}
	fmt.Println("Размер Изображения в байтах: = ", info.Size())

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

	// Закрываем файл
	err = file.Close()
	if err != nil {
		fmt.Println("failed to close file...")
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
		fmt.Println("7")
		log.Fatal(err)
	}
	defer imgfile.Close()

	// Сохраняем миниатюру в формате PNG
	err = png.Encode(imgfile, newImage)
	if err != nil {
		fmt.Println("8")
		log.Fatal(err)
	}

	// Получаем размер миниатюры
	miniInfo, err := os.Stat(fileNameMini)
	if err != nil {
		fmt.Println("9")
		log.Fatal(err)
	}

	// Размер
	miniSize := miniInfo.Size()
	fmt.Println("Размер Миниатюры в байтах: = ", miniSize)
}
