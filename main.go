package main

import (
	_ "bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"functions/functions"
	_ "io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Структура для хранения конфигурации сервера.
type config struct {
	Port int64 //Сетевой порт, который будет использоваться.
}

// CustomHandler — это оболочка http.ServeMux, которая регистрирует каждый запрос..
type CustomHandler struct {
	mux *http.ServeMux //Экземпляр ServeMux, который будет обрабатывать запросы.
}

// C_path , C_Size , C_time_slap , C_time_modification;
// @ServeHTTP регистрирует детали запроса и делегирует их обернутому @ServeMux.
// Он реализует интерфейс @http.Handler.
func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Зарегистрируем детали запроса.
	log.Printf("Получен запрос от %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)
	//Делегируем завернутый @ServeMux.
	h.mux.ServeHTTP(w, r)
}

// Функция получает из запроса два параметра URL «rootValue» значение корневого параметра.
// sortValue: значение параметра «sort»
func handlerData(w http.ResponseWriter, r *http.Request) {

	rawParams := r.URL.RawQuery                   // Получаем необработанную строку запроса.
	queryParams, err := url.ParseQuery(rawParams) //Разобираем необработанную строку запроса в карту параметров запроса.
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rootParam := queryParams.Get("root") // Получаем значение «root» параметра
	sortParam := queryParams.Get("sort") // Получаем значение «sort» параметра
	fmt.Println("root value:", rootParam)
	fmt.Println("sort value:", sortParam)
	// Теперь в Root хранится экземпляр структуры Root с именем «root»
	Root := functions.Root{Name: rootParam}
	//@GetSubDirRoutine - пример метода, который возвращает данные и ошибку
	stat, err := os.Stat(Root.Name) //Получаем информацию о файле для корневого каталога.
	//Если при получении информации о файле произошла ошибка, немедленно вернем.
	if err != nil {
		return
	}
	//Включаем файловый режим, чтобы определить, обычный ли это файл или каталог.
	switch mode := stat.Mode(); {
	case mode.IsRegular():
		fmt.Printf("IS REGULAR")
	case mode.IsDir():
		//Если файл является каталогом, получаем данные подкаталога, обработаем любые ошибки,
		start := time.Now()                                // Запишиваем время начала
		filesData, err := Root.GetSubDirRoutine(Root.Name) //вызываем метод GetSubDirRoutine для объекта Root и передаем Root.Name в качестве аргумента.
		elapsed := time.Since(start).String()              // Подсчитаем прошедшее время
		if err != nil {
			panic(err)
		}
		//Отсортируем фрагмент данных файлов, используя предоставленный корень и параметры сортировки.
		data := functions.SortSlice(filesData, rootParam, sortParam)
		//Эта структура,  содержит информацию о файлах и времени, затраченном на их обработку.
		info := functions.Info{Files: data, Elapsedtime: elapsed, PathName: Root.Name}
		for _, item := range info.Files {
			fmt.Println(item.Name, item.SizeInKB, "0")
		}
		//Функция вычисляет общий размер файлов в срезе и возвращает его. Результат присваивается переменной sum.
		totalSize := functions.Sum(info.Files)
		// структура, содержит информацию о пути, затраченном времени и общем размере файлов.
		InfoPath := functions.Stat{PathName: Root.Name, ElapsedTime: elapsed, Size: totalSize}
		fmt.Println("Pathinfo:", InfoPath.PathName, InfoPath.Size, InfoPath.ElapsedTime)
		//Это примитив синхронизации, используемый для ожидания завершения выполнения набора горутин.
		var wg sync.WaitGroup // WaitGroup для ожидания завершения всех горутин
		wg.Add(2)             //Добавяем 2 в WaitGroup, чтобы учесть две горутины.
		//Запускаем две горутины
		stat := Stat{
			Field1: "value1",
			Field2: "value2",
		}
		go sendJSONResponse(w, r, info, &wg)
		go sendRequestToApache(stat, &wg)
		wg.Wait() //Подождем, пока все горутины завершатся.
	}
}

type Stat struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

// Assuming functions.Stat is defined somewhere in your codebase

func sendRequestToApache(stat Stat, wg *sync.WaitGroup) error {
	defer wg.Done()

	// Marshal the Stat struct into JSON
	jsonData, err := json.Marshal(stat)
	if err != nil {
		return err
	}

	// Create a new POST request with the JSON data
	req, err := http.NewRequest("POST", "http://localhost:80/insert.php", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Read the response body as text
		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		// Handle different status codes here
		switch resp.StatusCode {
		case http.StatusInternalServerError:
			return fmt.Errorf("internal server error: %s", bodyText)
		case http.StatusBadRequest:
			return fmt.Errorf("bad request: %s", bodyText)
		case http.StatusMethodNotAllowed:
			return fmt.Errorf("method not allowed: %s", bodyText)
		// cases for other status codes as needed
		default:
			return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, bodyText)
		}
	}

	// Read the entire response body into memory
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// If you expect a JSON response, you can unmarshal it here
	var responseData struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}

	// Check for errors in the response
	if responseData.Error != "" {
		return fmt.Errorf("server error: %s", responseData.Error)
	}

	// Print the message from the response
	if responseData.Message != "" {
		fmt.Println(responseData.Message)
	}

	return nil
}

// функция, которая отправляет клиенту ответ JSON.
func sendJSONResponse(w http.ResponseWriter, _ *http.Request, files functions.Info, wg *sync.WaitGroup) {
	defer wg.Done()
	// Маршалируем данные в @JSON
	jsonData, err := json.Marshal(files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Установите тип контента @application/json.
	w.Header().Set("Content-Type", "application/json")
	// Запишите данные @JSON в ответ
	w.Write(jsonData)

}

// @getServerPort считывает порт сервера из файла конфигурации и возвращает его в виде строки.
// Если при чтении файла или синтаксическом анализе JSON возникает ошибка, он регистрирует ошибку и возвращает ее.
func getServerPort() (string, error) {
	configData, err := os.ReadFile("ui/config.json") //Попытаем прочитать файл конфигурации.
	if err != nil {
		log.Fatalf("Не удалось прочитать файл конфигурации: %v", err)
	}
	var conf config // Объявлем переменную для хранения данных конфигурации.
	err = json.Unmarshal(configData, &conf)
	if err != nil {
		log.Fatalf("Не удалось проанализировать файл конфигурации.: %v", err)
	}
	return fmt.Sprintf(":%d", conf.Port), nil
}
func main() {
	values := map[string]string{"key": "value"}
	jsonData, err := json.Marshal(values)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Print the JSON data to the console for debugging
	fmt.Println("JSON data:", string(jsonData))

	// Create a new POST request
	url := "http://localhost/select.php" // Replace with your Apache server URL
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Print the response status and body
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(body))
	// // Создаем новый ServeMux для маршрутизации запросов
	mux := http.NewServeMux()
	// Обработчик для всех запросов
	mux.HandleFunc("/files", handlerData)
	//Загрузкаs файлы в сервер  с помощью ServeMux по пути /js/.
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
	mux.Handle("/ts/", http.StripPrefix("/ts/", http.FileServer(http.Dir("./ts"))))
	mux.Handle("/", http.FileServer(http.Dir("./templates")))
	mux.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("./templates"))))
	mux.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("./dist"))))
	//mux.Handle("/apache/", http.StripPrefix("/apache/", http.FileServer(http.Dir("./apache"))))
	// Обернем ServeMux пользовательским обработчиком.
	//path := "/Users/ismaelnvo/Desktop/"
	//Root := functions.Root{Name: path}

	handler := &CustomHandler{mux: mux}
	// Получаем номер порта севера
	portNumber, err := getServerPort()
	if err != nil {
		panic(err)
	}
	//Настраиваем сервер
	server := &http.Server{
		Addr:    portNumber, //Порт, на котором будет слушать сервер
		Handler: handler,    // Обработчик, который будет обрабатывать каждый входящий запрос.
	}
	// Создайте 2 каналa для прослушивания ошибок, исходящих от прослушивателя. Использовать
	// буферизованный канал, чтобы горутина могла завершить работу, если мы не обнаружим эту ошибку.
	serverErrors := make(chan error)
	shutdown := make(chan os.Signal, 1)
	otherErrors := make(chan error)
	// Запуска сервера в горутине
	go func() {
		fmt.Println("Сервер работает на порту 8080:") //// Создание канал для прослушивания сигнала прерывания или завершения от ОС.
		serverErrors <- server.ListenAndServe()
	}()
	// Использование буферизованный канал, поскольку этого требует пакет сигналов.
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Блокировка основного и ожидание выключения.
	select {
	//Если возникла ошибка из канала serverErrors, выведите ее и проверяем, не является ли это ошибкой закрытия сервера.
	case err := <-serverErrors:
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
		if err == http.ErrServerClosed {
			fmt.Println("Сервер успешно завершил работу:")
		}
		//Если сигнал получен на канале завершения работы, распечатайте сообщение и попытаемся корректно завершить работу сервера.
	case <-shutdown:
		fmt.Println("Начало выключения...:")
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// Просим слушателя выключиться и отключить загрузку.
		err := server.Shutdown(ctx)
		if err != nil {
			fmt.Printf("Грациозное завершение работы не завершилось %v: %v\n", timeout, err)
			err = server.Close()
		}
		if err != nil {
			fmt.Printf("Не удалось корректно остановить сервер: %v\n", err)
		}
		// Если есть ошибка из каналаotherErrors, выведим ошибку.
	case err := <-otherErrors:
		fmt.Printf("Возникла ошибка: %v\n", err)
	}

}
