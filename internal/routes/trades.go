package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stollenaar/ollamabot/internal/database"
)

// RegisterTradeRoutes registers trade-related routes to the given router group.
func RegisterTradeRoutes(rg *gin.RouterGroup) {
	trades := rg.Group("/trades")
	{
		trades.GET("/:id", GetTrade)
		trades.GET("/platform/:id", ListTrades)
		trades.POST("/:id", UpdateTrade)
	}
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
func GetTrade(c *gin.Context) {
	id := c.Param("id")
	tx, err := database.GetTransactionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	c.JSON(http.StatusOK, tx)
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
func ListTrades(c *gin.Context) {
	id := c.Param("id")
	tx, err := database.GetTransactionByPlatformID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transactions not found"})
		return
	}
	c.JSON(http.StatusOK, tx)

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
func UpdateTrade(c *gin.Context) {
	id := c.Param("id")
	type request struct {
		PlatformID string `json:"platform_id" binding:"required"`
		Status     string `json:"status" binding:"required"`
	}
	var req request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	tx, err := database.GetTransactionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if tx.PlatformID != req.PlatformID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Platform ID does not match"})
		return
	}

	tx.Status = req.Status
	if err := database.UpdateTransaction(*tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated"})
}
