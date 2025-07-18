package main

import (
	"allocator_service/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var ctx = context.Background()

func main() {
	godotenv.Load(".env") // Load environment variables from .env file

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

	fmt.Println("Allocator Service started...")
	for {
		// Blocking pop from the queue
		data, err := client.BRPop(ctx, 0, "customers_queue").Result()
		if err != nil {
			fmt.Println("Error popping from queue:", err)
			continue
		}

		// data[0] is the key, data[1] is the value
		var customerData model.RedisData
		err = json.Unmarshal([]byte(data[1]), &customerData)
		if err != nil {
			fmt.Println("Error unmarshalling data:", err)
			continue
		}

		// Check if the room ID and candidate ID are valid
		err = assignToAgent(customerData.RoomID, fmt.Sprintf("%d", customerData.CandidateID))
		if err == nil {
			fmt.Println("Successfully assigned agent:", customerData.CandidateID, "to room:", customerData.RoomID)
			continue
		}

		fmt.Println("Failed to assign agent:", customerData.CandidateID, "to room:", customerData.RoomID, "Error:", err)

		// If assignment failed, try to allocate and assign an agent
		err = allocateAndAssignToAgent(customerData.RoomID)
		if err == nil {
			fmt.Println("Successfully allocate assigned agent to room:", customerData.RoomID)
			continue
		}

		fmt.Println("Failed to allocate and assign agent to room:", customerData.RoomID, "Error:", err)

		// If both assignment and allocation failed, save the data back to the queue
		// Increment the retry count and check if it is less than 3
		// If retry count is less than 3, re-add to the queue for retry
		if customerData.RetryCount < 3 {
			customerData.RetryCount++
			fmt.Println("Retrying to assign agent:", customerData.CandidateID, "to room:", customerData.RoomID, "Retry count:", customerData.RetryCount)
			newData, err := json.Marshal(customerData)
			if err != nil {
				fmt.Println("Error marshalling data for retry:", err)
				continue
			}
			err = client.RPush(ctx, "customers_queue", newData).Err()
			if err != nil {
				fmt.Println("Error saving to Redis", "Error:", err)
				continue
			}
		}
	}
}

func assignToAgent(roomID string, candidateID string) error {
	baseUrl := os.Getenv("OMNI_BASE_URL")

	data := url.Values{}
	data.Set("agent_id", candidateID)
	data.Set("room_id", roomID)
	data.Set("max_agent", os.Getenv("MAX_AGENTS"))

	// Construct the API endpoint
	// Assuming the base URL is set in the environment variable OMNI_BASE_URL
	api := fmt.Sprintf("%s/api/v1/admin/service/assign_agent", baseUrl)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", api, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("Error saat buat request:", err)
		return err
	}

	// Set the necessary headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", os.Getenv("OMNI_API_KEY"))
	req.Header.Set("Qiscus-Secret-Key", os.Getenv("OMNI_API_SECRET"))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error saat request:", err)
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	// If the status code is not 200 OK, return an error
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Assign API responded with status:", resp.StatusCode)
		return fmt.Errorf("assign API responded with status %d", resp.StatusCode)
	}

	return nil
}

func allocateAndAssignToAgent(roomID string) error {
	baseUrl := os.Getenv("OMNI_BASE_URL")

	data := url.Values{}
	data.Set("room_id", roomID)

	// Construct the API endpoint
	// Assuming the base URL is set in the environment variable OMNI_BASE_URL
	api := fmt.Sprintf("%s/api/v1/admin/service/allocate_assign_agent", baseUrl)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", api, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("Error saat buat request:", err)
		return err
	}

	// Set the necessary headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", os.Getenv("OMNI_API_KEY"))
	req.Header.Set("Qiscus-Secret-Key", os.Getenv("OMNI_API_SECRET"))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error saat request:", err)
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	// If the status code is not 200 OK, return an error
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Allocate and Assign API responded with status:", resp.StatusCode)
		return fmt.Errorf("alocate and assign API responded with status %d", resp.StatusCode)
	}

	return nil
}
