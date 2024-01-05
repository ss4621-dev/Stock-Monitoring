package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var stocksData []Stock
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []Stock)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Stock struct {
	Symbol          string  `json:"symbol"`
	OpenPrice       float64 `json:"openPrice"`
	CurrentPrice    float64 `json:"currentPrice"`
	RefreshInterval int     `json:"refreshInterval"`
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Define API routes
	router.GET("/api/stocks", getStocksHandler)

	// Define WebSocket route
	router.GET("/ws", func(c *gin.Context) {
		handleWebSocketConnections(c)
	})

	// Start handling WebSocket messages
	go handleMessages()

	// Start fetching and updating stocks data in the background
	go fetchStocks()

	// Start the server
	port := getPort()
	go func() {
		if err := router.Run(":" + port); err != nil {
			log.Fatal(err)
		}
	}()

	// Handle OS signals for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdown
		log.Println("Shutting down gracefully...")
		// Add any cleanup or finalization code here
		os.Exit(0)
	}()

	// Wait for shutdown signal
	<-shutdown
}

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Handler for fetching stocks
func getStocksHandler(c *gin.Context) {
	c.JSON(http.StatusOK, stocksData)
}

// Function to fetch and update stocks data
func fetchStocks() {
	apiKey := os.Getenv("POLYGON_API_KEY")
	if apiKey == "" {
		log.Fatal("Polygon API key is not set")
	}

	for {
		log.Println("Fetching ticker symbols from Polygon.io API...")

		// Move Result and APIResponse type declarations here
		type Result struct {
			Ticker         string `json:"ticker"`
			Name           string `json:"name"`
			Exchange       string `json:"primary_exchange"`
			Type           string `json:"type"`
			IsActive       bool   `json:"active"`
			LastUpdatedUTC string `json:"last_updated_utc"`
			CIK            string `json:"cik"`
			CompositeFIGI  string `json:"composite_figi"`
			CurrencyName   string `json:"currency_name"`
			DelistedUTC    string `json:"delisted_utc"`
			Locale         string `json:"locale"`
			Market         string `json:"market"`
			ShareClassFIGI string `json:"share_class_figi"`
		}

		type APIResponse struct {
			Count     int      `json:"count"`
			NextURL   string   `json:"next_url"`
			RequestID string   `json:"request_id"`
			Results   []Result `json:"results"`
		}

		// Make a request to the Polygon.io API to get a list of ticker symbols
		response, err := http.Get("https://api.polygon.io/v3/reference/tickers?type=all&active=true&apiKey=" + apiKey)
		if err != nil {
			log.Println("Error fetching ticker symbols:", err)
			sleepAndRetry(1 * time.Minute)
			continue
		}

		if response.StatusCode != http.StatusOK {
			log.Printf("Non-OK response from Polygon.io API: %s\n", response.Status)
			sleepAndRetry(1 * time.Minute)
			continue
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println("Error reading response body:", err)
			sleepAndRetry(1 * time.Minute)
			continue
		}

		var apiResponse APIResponse
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			log.Println("Error unmarshalling JSON:", err)
			sleepAndRetry(1 * time.Minute)
			continue
		}

		if apiResponse.Count == 0 {
			log.Println("No stock data available")
			sleepAndRetry(1 * time.Hour) // Retry after an hour
			continue
		}

		// Process the data as needed
		var stocks []Stock
		for _, result := range apiResponse.Results {
			stocks = append(stocks, Stock{
				Symbol:          result.Ticker,
				OpenPrice:       rand.Float64() * 100,
				CurrentPrice:    rand.Float64() * 100,
				RefreshInterval: rand.Intn(5) + 1,
			})
		}

		stocksData = stocks
		writeToFile("stocks.json", stocks)

		log.Println("Stocks data fetched and stored successfully")

		// Fetch new stocks every hour
		sleepAndRetry(1 * time.Hour)
	}
}

// Function to get the port from environment variables or default to 3001
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	return port
}

// Function to write data to a file
func writeToFile(filename string, data interface{}) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println("Error marshalling data:", err)
		return
	}

	if err := ioutil.WriteFile(filename, file, 0644); err != nil {
		log.Println("Error writing to file:", err)
		return
	}
}

// Helper function to sleep for a duration and retry
func sleepAndRetry(duration time.Duration) {
	log.Printf("Sleeping for %s and retrying...\n", duration.String())
	time.Sleep(duration)
}

// Function to broadcast updated stocks to all connected clients
func handleWebSocketConnections(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg []Stock
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			delete(clients, ws)
			break
		}

		broadcast <- msg
	}
}

// Function to continuously broadcast updates to connected clients
func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error writing JSON: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
