package api

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"transactions/internal/pg"
)

type txQueuesProcesses struct {
	userProcesses map[int64]chan struct{}
	mu            sync.Mutex
}

func newTxQueuesProcesses() (newTxQueuesProcesses *txQueuesProcesses) {
	log.Debug().Msg("api.newTxQueuesProcesses START")
	defer log.Debug().Msg("api.newTxQueuesProcesses END")

	newTxQueuesProcesses = &txQueuesProcesses{}

	newTxQueuesProcesses.userProcesses = map[int64]chan struct{}{}

	newTxQueuesProcesses.mu = sync.Mutex{}

	return newTxQueuesProcesses

}

func (a *API) txQueueProcess(userID int64) (process chan struct{}) {
	log.Debug().Str("userID", fmt.Sprint(userID)).Msg("api.txQueueProcess START")
	defer log.Debug().Msg("api.newTxQueuesProcesses END")

	a.txQueuesProcesses.mu.Lock()
	defer a.txQueuesProcesses.mu.Unlock()

	process, isExists := a.txQueuesProcesses.userProcesses[userID]
	if !isExists {
		process = make(chan struct{})
		a.txQueuesProcesses.userProcesses[userID] = process
	}

	return process
}

func (a *API) tryToProcessTxQueue(ctx context.Context, userID int64, process chan struct{}) {
	log.Debug().Str("userID", fmt.Sprint(userID)).Msg("api.tryToProcessTxQueue START")
	defer log.Debug().Msg("api.tryToProcessTxQueue END")

	err := a.storage.ProcessTxQueue(ctx, userID)
	if err != nil {
		if errors.Is(err, pg.ErrInsufficientFunds) {
			log.Info().Str("userID", fmt.Sprint(userID)).Msg(errInsufficientFunds.Error())
			return
		}
		log.Warn().Err(err).Msg(fmt.Sprintf("processing txs queue: userID: %d", userID))
		return
	}

	a.txQueuesProcesses.mu.Lock()
	delete(a.txQueuesProcesses.userProcesses, userID)
	a.txQueuesProcesses.mu.Unlock()
	close(process)

}
