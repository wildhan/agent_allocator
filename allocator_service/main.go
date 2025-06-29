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

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var ctx = context.Background()

func main() {
	godotenv.Load(".env") // Load environment variables from .env file

	redisAddr := os.Getenv("REDIS_ADDR")
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr, // Redis server address
	})

	fmt.Println("Allocator Service started...")
	for {
		// Blocking pop from the queue
		data, err := client.BRPop(ctx, 0, "customers_queue").Result()
		if err != nil {
			fmt.Println("Error popping from queue:", err)
			continue
		}

		var customerData model.RedisData
		err = json.Unmarshal([]byte(data[1]), &customerData)
		if err != nil {
			fmt.Println("Error unmarshalling data:", err)
			continue
		}

		err = assignToAgent(customerData.RoomID, fmt.Sprintf("%d", customerData.CandidateID))
		if err == nil {
			fmt.Println("Successfully assigned agent:", customerData.CandidateID, "to room:", customerData.RoomID)
			continue
		}

		fmt.Println("Failed to assign agent:", customerData.CandidateID, "to room:", customerData.RoomID, "Error:", err)

		err = allocateAndAssignToAgent(customerData.RoomID)
		if err == nil {
			fmt.Println("Successfully allocate assigned agent:", customerData.CandidateID, "to room:", customerData.RoomID)
			continue
		}

		fmt.Println("Failed to allocate and assign agent:", customerData.CandidateID, "to room:", customerData.RoomID, "Error:", err)

		err = client.RPush(ctx, "customers_queue", data).Err()
		if err != nil {
			fmt.Println("Error saving to Redis", "Error:", err)
			continue
		}
	}
}

func assignToAgent(roomID string, candidateID string) error {
	baseUrl := os.Getenv("OMNI_BASE_URL")

	data := url.Values{}
	data.Set("agent_id", candidateID)
	data.Set("room_id", roomID)
	data.Set("max_agent", os.Getenv("MAX_AGENTS"))

	api := fmt.Sprintf("%s/api/v1/admin/service/assign_agent", baseUrl)

	req, err := http.NewRequest("POST", api, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("Error saat buat request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", os.Getenv("OMNI_API_KEY"))
	req.Header.Set("Qiscus-Secret-Key", os.Getenv("OMNI_API_SECRET"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error saat request:", err)
		return err
	}
	defer resp.Body.Close()

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

	api := fmt.Sprintf("%s/api/v1/admin/service/allocate_assign_agent", baseUrl)

	req, err := http.NewRequest("POST", api, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("Error saat buat request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", os.Getenv("OMNI_API_KEY"))
	req.Header.Set("Qiscus-Secret-Key", os.Getenv("OMNI_API_SECRET"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error saat request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Allocate and Assign API responded with status:", resp.StatusCode)
		return fmt.Errorf("alocate and assign API responded with status %d", resp.StatusCode)
	}

	return nil
}
