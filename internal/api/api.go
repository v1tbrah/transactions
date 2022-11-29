package api

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

type API struct {
	server            *http.Server
	storage           Storage
	txQueuesProcesses *txQueuesProcesses
}

func New(storage Storage, config Config) (newAPI *API, err error) {
	log.Debug().Msg("api.New START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("api.New END")
		} else {
			log.Debug().Msg("api.New END")
		}
	}()

	newAPI = &API{}

	newAPI.storage = storage

	server := newAPI.newServer(config.RunAPIAddress())
	newAPI.server = server

	newAPI.txQueuesProcesses = newTxQueuesProcesses()

	return newAPI, nil
}

func (a *API) newServer(runAPIAddress string) *http.Server {
	log.Debug().Msg("api.newServer START")
	defer log.Debug().Msg("api.newServer END")

	newServer := &http.Server{}

	newServer.Addr = runAPIAddress

	router := a.newRouter()
	newServer.Handler = router

	return newServer

}

func (a *API) newRouter() *gin.Engine {
	log.Debug().Msg("api.newRouter START")
	defer log.Debug().Msg("api.newRouter END")

	newRouter := gin.Default()

	newRouter.Use(a.checkValid)

	newRouter.POST("/:id/receipt/:sum", a.receiptHandler)
	newRouter.POST("/:id/withdraw/:sum", a.withdrawHandler)

	return newRouter
}

func (a *API) Run() {
	log.Debug().Msg("api.Run START")
	defer log.Debug().Msg("api.Run END")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	errG, ctx := errgroup.WithContext(context.Background())

	errG.Go(func() error {
		return a.startProcessingTxQueues(ctx, shutdown)
	})

	errG.Go(func() error {
		return a.startListener(ctx, shutdown)
	})

	if err := errG.Wait(); err != nil {
		log.Error().Err(err).Msg(err.Error())
		close(shutdown)
	}

	<-shutdown
	if err := a.storage.Close(); err != nil {
		log.Printf("storage closing: %v", err)
	} else {
		log.Info().Msg("storage closed")
	}
	if err := a.server.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	} else {
		log.Info().Msg("HTTP server gracefully shutdown")
	}
}

func (a *API) startProcessingTxQueues(ctx context.Context, shutdown chan os.Signal) (err error) {

	ended := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-shutdown:
			if ok {
				close(shutdown)
			}
			return
		default:
			users, errGetting := a.storage.GetUsersWithNonEmptyTxQueues(ctx)
			if errGetting != nil {
				errGetting = fmt.Errorf("getting users with non empty tx queues: %w", err)
				err = errGetting
				ended <- struct{}{}
			}
			if len(users) > 0 {
				log.Info().Msg("start processing tx queues")
			}
			for _, userID := range users {
				txQueueProcess := a.txQueueProcess(userID)
				a.tryToProcessTxQueue(ctx, userID, txQueueProcess)
				log.Info().Str("userID", fmt.Sprint(userID)).Msg("queue txs processed")
			}
			ended <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
		return err
	case _, ok := <-shutdown:
		if ok {
			close(shutdown)
		}
		return err
	case <-ended:
		return err
	}

}

func (a *API) startListener(ctx context.Context, shutdown chan os.Signal) (err error) {

	ended := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-shutdown:
			if ok {
				close(shutdown)
			}
			return
		default:
			log.Info().Str("addr", a.server.Addr).Msg("starting http server")
			err = a.server.ListenAndServe()
			ended <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
		return err
	case _, ok := <-shutdown:
		if ok {
			close(shutdown)
		}
		return err
	case <-ended:
		return err
	}

}
