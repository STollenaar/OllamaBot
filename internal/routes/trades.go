package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stollenaar/ollamabot/internal/database"
)

// RegisterTradeRoutes registers trade-related routes to the given router group.
func RegisterTradeRoutes(mux *http.ServeMux) {
	mux.HandleFunc(fmt.Sprintf("GET %s/trades/:id", API), GetTrade)
	mux.HandleFunc(fmt.Sprintf("GET %s/trades/platform/:id", API), ListTrades)
	mux.HandleFunc(fmt.Sprintf("POST %s/trades/:id", API), UpdateTrade)
}

// GetTrade returns a specific trade by ID.
//
//	@Summary		Get trade by ID
//	@Description	Get a specific trade by its ID
//	@Tags			trades
//	@Produce		json
//	@Param			id	path		string	true	"Trade ID"
//	@Success		200	{object}	database.Transaction
//	@Failure		404	{object}	map[string]string
//	@Router			/trades/{id} [get]
func GetTrade(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	id := params.Get("id")
	tx, err := database.GetTransactionByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Transaction not found"})
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

// ListTrades returns all trades.
//
//	@Summary		List all trades by platform id
//	@Description	Get a list of all trades
//	@Tags			trades
//	@Produce		json
//	@Param			id	path		string	true	"Platform ID"
//	@Success		200	{array}		database.Transaction
//	@Failure		500	{object}	map[string]string
//	@Router			/trades/platform/{id} [get]
func ListTrades(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	id := params.Get("id")
	tx, err := database.GetTransactionByPlatformID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Transaction not found"})
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

// UpdateTrade updates the status of a transaction by ID.
//
//	@Summary		Update trade status
//	@Description	Update the status of a trade by ID and platform ID
//	@Tags			trades
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Trade ID"
//	@Param			body	body		routes.UpdateTrade.request	true	"Update payload"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/trades/{id} [post]
func UpdateTrade(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	id := params.Get("id")

	type request struct {
		PlatformID string `json:"platform_id" binding:"required"`
		Status     string `json:"status" binding:"required"`
	}

	var req request
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	tx, err := database.GetTransactionByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Transaction not found"})
		return
	}

	if tx.PlatformID != req.PlatformID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Platform ID does not match"})
		return
	}

	tx.Status = req.Status
	if err := database.UpdateTransaction(*tx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update transaction"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Transaction updated"})
}
