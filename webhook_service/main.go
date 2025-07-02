package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

	redisAddr := os.Getenv("REDIS_ADDR")
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr, // Redis server address
	})

	r.Get("/webhook", func(w http.ResponseWriter, r *http.Request) {
		var req model.Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		data, err := json.Marshal(model.RedisData{
			RoomID:      req.RoomID,
			CandidateID: req.CandidateAgent.ID,
			RetryCount:  0, // Initialize retry count to 0
		})
		if err != nil {
			http.Error(w, "Error processing request", http.StatusInternalServerError)
			return
		}

		err = client.RPush(ctx, "customers_queue", data).Err()
		if err != nil {
			http.Error(w, "Error saving to Redis", http.StatusInternalServerError)
			return
		}
	})

	port := os.Getenv("PORT") // Get the port from environment variable
	fmt.Printf("Listening on port %s...", port)
	http.ListenAndServe(":"+port, r)
}
