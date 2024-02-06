package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	MyToken = "access_token"
)

type DataArray struct {
	Data []DataStruct `json:"data"`
}

type DataStruct struct {
	ID            int    `json:"address_id"`
	Town          string `json:"town"`
	Address       string `json:"address"`
	Apartment     int    `json:"apartment"`
	House_type    string `json:"house_type"`
	Meter_reading int    `json:"meter_reading"`
	Meter_ID      int    `json:"meter_id"`
	Meter_status  string `json:"meter_status"`
}

// структура с переменной для стоимости кубометра воды
type Result struct {
	Meter_ID   int     `json:"meter_id"`
	Address_ID int     `json:"address_id"`
	Price      float64 `json:"price"`
	Token      string  `json:"token"`
}

// Функция для генерации случайной стоимости кубометра воды от 40 до 60
func randomPrice() float64 {
	minPrice := 40.0
	maxPrice := 60.0

	time.Sleep(5 * time.Second) // Задержка на 5 секунд
	rand.Seed(time.Now().UnixNano())
	return minPrice + rand.Float64()*(maxPrice-minPrice)
}

// Функция для отправки стоимости в отдельной горутине
func SendPrice(meter_id int, address_id int, url string, data DataStruct) {
	price := randomPrice()

	dataResult := Result{Meter_ID: meter_id, Address_ID: address_id, Price: price, Token: MyToken}
	_, err := performPUTRequest(url, dataResult)
	if err != nil {
		fmt.Println("Error sending price:", err)
		return
	}
	fmt.Println("Price sent successfully for meter_id:", meter_id, "and address_id:", address_id, "-", price)
}

func performPUTRequest(url string, data Result) (*http.Response, error) {
	// Сериализация структуры в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Создание PUT-запроса
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}

func main() {
	// Создание роутера Gin
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"Origin", "Content-Type"}
	config.AllowCredentials = true

	r.Use(cors.New(config))

	r.POST("/price/", func(c *gin.Context) {
		var dataArray DataArray

		if err := c.ShouldBindJSON(&dataArray); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		var wg sync.WaitGroup
		for _, data := range dataArray.Data {
			wg.Add(1)
			go func(data DataStruct) {
				defer wg.Done()
				SendPrice(data.Meter_ID, data.ID, "http://127.0.0.1:8000/manytomany/async/put/", data)
			}(data)
		}

		wg.Wait()

		c.JSON(http.StatusOK, gin.H{"message": "Status update initiated for all IDs"})

	})
	// Запуск сервера
	r.Run("localhost:9000")

}
