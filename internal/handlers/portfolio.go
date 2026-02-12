package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/kafka"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/services"
	"time"
)

// HandlePortfolio handles new portfolio creation requests (feel free to update the request parameter/model)
// Sample Request (POST /portfolio):
//
//	{
//	    "user_id": "1",
//	    "allocation": {"stocks": 60, "bonds": 30, "gold": 10}
//	}
func HandlePortfolio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := http.StatusInternalServerError
	var response interface{}
	var err error

	switch r.Method {
	case http.MethodPost:
		var rsp models.Portfolio
		rsp, status, err = services.AddPortfolio(ctx, r)
		response = rsp
	case http.MethodGet:
		var rsp models.Portfolio
		rsp, status, err = services.GetPortfolio(ctx, r)
		response = rsp
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		response = err
	} else {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// HandleRebalance handles portfolio rebalance requests from 3rd party provider (feel free to update the request parameter/model)
// Sample Request (POST /rebalance):
//
//	{
//	    "user_id": "1",
//	    "new_allocation": {"stocks": 70, "bonds": 20, "gold": 10}
//	}
func HandleRebalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UpdatedPortfolio
	json.NewDecoder(r.Body).Decode(&req)
	req.CreatedAt = time.Now()

	log.Println("HandleRebalance==", req)

	reqByte, err := json.Marshal(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: Add Logic here
	err = kafka.PublishMessage(ctx, reqByte)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Success Rebalance User Portfolio")
}

func HandleRebalanceConsume(ctx context.Context, value []byte) {
	var req models.UpdatedPortfolio
	err := json.Unmarshal(value, &req)
	if err != nil {
		log.Println(err.Error())
		return
	}

	req.CreatedAt = time.Now()

	log.Println("HandleRebalance==", req)

	err = services.RebalanceTransaction(ctx, req)
	if err != nil {
		err = kafka.PublishMessage(ctx, value)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}

func HandleGetRebalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := http.StatusInternalServerError
	var response interface{}
	var err error

	switch r.Method {
	case http.MethodGet:
		var rsp []models.RebalanceTransaction
		rsp, status, err = services.GetRebalanceTransaction(ctx, r)
		response = rsp
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		response = err
	} else {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
