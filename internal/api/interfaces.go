package api

import "context"

type Config interface {
	RunAPIAddress() string
}

type Storage interface {
	AddTx(ctx context.Context, userID int64, sum float64) (err error)
	ProcessTxQueue(ctx context.Context, userID int64) (err error)
	GetUsersWithNonEmptyTxQueues(ctx context.Context) (users []int64, err error)
	Close() (err error)
}
