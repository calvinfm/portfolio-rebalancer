package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/storage"
)

func AddPortfolio(ctx context.Context, r *http.Request) (models.Portfolio, int, error) {

	var p models.Portfolio
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		log.Println("Invalid request body")
		return p, http.StatusBadRequest, err
	}

	porto, err := storage.GetPortfolio(ctx, p.UserID)
	if err != nil && err.Error() != "user not found" {
		log.Println(err.Error())
		return p, http.StatusInternalServerError, err
	}
	if porto != nil {
		log.Println("User already registered")
		return p, http.StatusConflict, errors.New("User already registered")
	}

	err = storage.SavePortfolio(ctx, models.Portfolio{
		UserID:     p.UserID,
		Allocation: p.Allocation,
	})
	if err != nil {
		log.Println(err.Error())
		return p, http.StatusInternalServerError, err
	}

	return p, http.StatusCreated, nil
}

func GetPortfolio(ctx context.Context, r *http.Request) (models.Portfolio, int, error) {

	userId := r.URL.Query().Get("userId")

	porto, err := storage.GetPortfolio(ctx, userId)
	if err != nil {
		log.Println(err.Error())
		return models.Portfolio{}, http.StatusInternalServerError, err
	}
	if porto == nil {
		log.Println("User not found")
		return models.Portfolio{}, http.StatusNotFound, errors.New("User not found")
	}

	return *porto, http.StatusOK, nil
}
