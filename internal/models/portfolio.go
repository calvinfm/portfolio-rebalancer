package models

import "time"

type Portfolio struct {
	UserID     string             `json:"user_id"`
	Allocation map[string]float64 `json:"allocation"` // Current user allocation in percentage terms
}

type UpdatedPortfolio struct {
	UserID        string             `json:"user_id"`
	NewAllocation map[string]float64 `json:"new_allocation"` // Updated user allocation from provider in percentage terms
	CreatedAt     time.Time          `json:"created_at"`
}

type RebalanceTransaction struct {
	Id               string    `json:"id"`
	UserID           string    `json:"user_id"`
	RebalancePercent float64   `json:"rebalance_percent "`
	Action           string    `json:"action"` // BUY/SELL
	Asset            string    `json:"asset"`  // STOCKS/BONDS/GOLD
	CreatedAt        time.Time `json:"created_at"`
}

type RebalancePortfolio struct {
	UserID     string             `json:"user_id"`
	Allocation map[string]float64 `json:"allocation"` // Current user allocation in percentage terms
	CreatedAt  time.Time          `json:"created_at"`
}
