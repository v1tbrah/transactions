package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var errIDIsEmpty = errors.New("id is empty")
var errSumIsEmpty = errors.New("sum is empty")
var errInvalidID = errors.New("invalid id")
var errInvalidSum = errors.New("invalid sum")
var errInsufficientFunds = errors.New("not enough funds in the balance")

func (a *API) checkValid(c *gin.Context) {
	log.Debug().Msg("api.checkValid START")
	defer log.Debug().Msg("api.checkValid END")

	reqID := c.Param("id")
	if reqID == "" {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errIDIsEmpty.Error()))
		c.Abort()
		return
	}

	id, err := strconv.ParseInt(reqID, 10, 64)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errInvalidID.Error()))
		c.Abort()
		return
	}

	c.Set("id", id)

	reqSum := c.Param("sum")
	if reqSum == "" {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errSumIsEmpty.Error()))
		c.Abort()
		return
	}

	sum, err := strconv.ParseFloat(reqSum, 64)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errInvalidSum.Error()))
		c.Abort()
		return
	}

	c.Set("sum", sum)

}

func (a *API) receiptHandler(c *gin.Context) {
	log.Debug().Msg("api.receiptHandler START")
	defer log.Debug().Msg("api.receiptHandler END")

	idParam, ok := c.Get("id")
	if !ok {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errIDIsEmpty.Error()))
		return
	}
	id, ok := idParam.(int64)
	if !ok {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errInvalidID.Error()))
		return
	}

	sumParam, ok := c.Get("sum")
	if !ok {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errSumIsEmpty.Error()))
		return
	}
	sum, ok := sumParam.(float64)
	if !ok {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errInvalidSum.Error()))
		return
	}

	if err := a.storage.AddTx(c, id, sum); err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", nil)
		return
	}

	txQueueProcess := a.txQueueProcess(id)

	a.tryToProcessTxQueue(c, id, txQueueProcess)

	<-txQueueProcess

	c.Data(http.StatusOK, "text/plain", []byte("OK"))
}

func (a *API) withdrawHandler(c *gin.Context) {
	log.Debug().Msg("api.withdrawHandler START")
	defer log.Debug().Msg("api.withdrawHandler END")

	idParam, ok := c.Get("id")
	if !ok {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errIDIsEmpty.Error()))
		return
	}
	id, ok := idParam.(int64)
	if !ok {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errInvalidID.Error()))
		return
	}

	sumParam, ok := c.Get("sum")
	if !ok {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errSumIsEmpty.Error()))
		return
	}
	sum, ok := sumParam.(float64)
	if !ok {
		c.Data(http.StatusBadRequest, "text/plain", []byte(errInvalidSum.Error()))
		return
	}

	if err := a.storage.AddTx(c, id, -sum); err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", nil)
		return
	}

	txQueueProcess := a.txQueueProcess(id)

	a.tryToProcessTxQueue(c, id, txQueueProcess)

	<-txQueueProcess

	c.Data(http.StatusOK, "text/plain", []byte("OK"))
}
