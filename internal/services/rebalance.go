package services

import (
	"context"
	"errors"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/storage"
	"time"
)

func CalculateRebalance(
	updatedAllocation, currentAlloctaion map[string]float64, userId string, createdAt time.Time,
) []models.RebalanceTransaction {
	log.Println("Start Calculate Rebalancing")
	var result []models.RebalanceTransaction

	// TODO: create rebalance transactions and update portfolio
	for k := range updatedAllocation {
		if updatedAllocation[k] > currentAlloctaion[k] {
			result = append(result,
				models.RebalanceTransaction{
					UserID:           userId,
					RebalancePercent: updatedAllocation[k] - currentAlloctaion[k],
					Action:           "SELL",
					Asset:            k,
					CreatedAt:        createdAt,
				})
		} else if updatedAllocation[k] < currentAlloctaion[k] {
			result = append(result,
				models.RebalanceTransaction{
					UserID:           userId,
					RebalancePercent: currentAlloctaion[k] - updatedAllocation[k],
					Action:           "BUY",
					Asset:            k,
					CreatedAt:        createdAt,
				})
		} else {
			continue
		}
	}

	return result
}

func RebalanceTransaction(ctx context.Context, request models.UpdatedPortfolio) error {
	log.Println("Start Rebalancing")
	porto, err := storage.GetPortfolio(ctx, request.UserID)
	if err != nil && err.Error() != "user not found" {
		log.Println(err.Error())
		return errors.New("Server Error")
	}
	if porto == nil {
		log.Println("User Not Found")
		return errors.New("User Not Found")
	}

	rebalanceTransactons := CalculateRebalance(request.NewAllocation, porto.Allocation, porto.UserID, request.CreatedAt)

	err = storage.SaveRebalanceTransactionBulk(ctx, rebalanceTransactons)
	if err != nil && err.Error() != "user not found" {
		log.Println(err.Error())
		return errors.New("Server Error")
	}

	return nil
}

func GetRebalanceTransaction(ctx context.Context, r *http.Request) ([]models.RebalanceTransaction, int, error) {

	userId := r.URL.Query().Get("userId")

	rebalanceTransactions, err := storage.GetTransactionsByUserID(ctx, userId)
	if err != nil {
		log.Println(err.Error())
		return []models.RebalanceTransaction{}, http.StatusInternalServerError, err
	}
	if rebalanceTransactions == nil {
		log.Println("User not found")
		return []models.RebalanceTransaction{}, http.StatusNotFound, errors.New("User not found")
	}

	return rebalanceTransactions, http.StatusOK, nil
}
