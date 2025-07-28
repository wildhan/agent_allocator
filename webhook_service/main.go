package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"webhook_service/model"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var ctx = context.Background()

func main() {
	godotenv.Load(".env") // Load environment variables from .env file
	r := chi.NewRouter()

	r.Use(middleware.Logger) // Log the request
	r.Use(middleware.Recoverer)

	// Initialize Redis client
	var client *redis.Client
	for {
		fmt.Println("Connecting to Redis...")

		redisAddr := os.Getenv("REDIS_ADDR")
		opt, _ := redis.ParseURL(redisAddr)
		client = redis.NewClient(opt)

		err := client.Ping(ctx).Err()
		if err == nil {
			fmt.Println("Connected to Redis successfully!")
			break // Exit the loop if Redis is reachable
		}
		fmt.Println("Failed to connect to Redis, retrying in 5 seconds...")
		fmt.Println("Error:", err)
		time.Sleep(5 * time.Second) // Wait before retrying
	}

	r.Post("/webhook", func(w http.ResponseWriter, r *http.Request) {
		var req model.Request

		// Decode the JSON request body
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		fmt.Printf("Received request: %+v\n", req)

		// Initialize the Redis data structure
		customerData := model.RedisData{
			RoomID:      req.RoomID,
			CandidateID: req.CandidateAgent.ID,
			RetryCount:  0, // Initialize retry count to 0
		}
		data, err := json.Marshal(customerData)
		if err != nil {
			http.Error(w, "Error processing request", http.StatusInternalServerError)
			return
		}

		// Add the RoomID to a Redis SET to ensure uniqueness
		// Using SAdd to add the RoomID to a set in Redis
		added, err := client.SAdd(ctx, "customers_index", customerData.RoomID).Result()
		if err != nil {
			log.Fatalf("Gagal menambahkan ke SET: %v", err)
		}

		// If the RoomID was added to the set, proceed to save the data
		// If added > 0, it means the RoomID was not already present in the set
		if added > 0 {
			// Save the data to Redis
			// Using RPush to add the data to a list in Redis
			err = client.RPush(ctx, "customers_queue", data).Err()
			if err != nil {
				http.Error(w, "Error saving to Redis", http.StatusInternalServerError)
				return
			}
			fmt.Printf("Add on queue: %+v\n", req)
		} else {
			// If the RoomID already exists, return an error
			fmt.Printf("RoomID already exist: %+v\n", req)
			http.Error(w, "RoomID already exist", http.StatusConflict)
			return
		}
	})

	// Get the port from environment variable
	port := os.Getenv("PORT")

	// Listen on the specified port
	fmt.Printf("Listening on port %s...", port)
	http.ListenAndServe(":"+port, r)
}
