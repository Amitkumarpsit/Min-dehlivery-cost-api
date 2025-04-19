// Go REST API to calculate minimum delivery cost

package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var (
	centers         = []string{"C1", "C2", "C3"}
	centerDistances = map[string]int{
		"C1": 10,
		"C2": 20,
		"C3": 30,
	}
	costPerKm = 2
	// Map of center to available products
	centerStock = map[string][]string{
		"C1": {"A", "B", "E", "H"},
		"C2": {"B", "C", "F", "I"},
		"C3": {"C", "D", "G", "H", "I"},
	}
)

type Order map[string]int

type CostResponse struct {
	MinimumCost int `json:"minimum_cost"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default port 8080")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/min-delivery-cost", deliveryCostHandler)
	log.Printf("Server running at http://0.0.0.0:%s\n", port)
	errListen := http.ListenAndServe("0.0.0.0:"+port, nil)
	if errListen != nil {
		log.Fatalf("Failed to start server: %v", errListen)
	}
}

func deliveryCostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	minCost := math.MaxInt
	for _, start := range centers {
		visited := make(map[string]bool)
		cost := calculateCost(order, start, visited, 0)
		if cost < minCost {
			minCost = cost
		}
	}

	respondJSON(w, CostResponse{MinimumCost: minCost})
}

func calculateCost(order Order, current string, visited map[string]bool, accumulated int) int {
	remaining := make(Order)
	for k, v := range order {
		remaining[k] = v
	}

	// Deduct available products from current center
	available := centerStock[current]
	for _, prod := range available {
		if _, ok := remaining[prod]; ok {
			delete(remaining, prod)
		}
	}

	distance := centerDistances[current]
	cost := accumulated + 2*distance*costPerKm // round trip

	if len(remaining) == 0 {
		return cost
	}

	visited[current] = true
	min := math.MaxInt
	for _, next := range centers {
		if !visited[next] {
			subCost := calculateCost(remaining, next, copyMap(visited), cost)
			if subCost < min {
				min = subCost
			}
		}
	}

	return min
}

func copyMap(orig map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range orig {
		newMap[k] = v
	}
	return newMap
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
