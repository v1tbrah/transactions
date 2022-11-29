package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var ErrDBIsNilPointer = errors.New("database is nil pointer")

type Pg struct {
	db            *sql.DB
	usersStmts    *usersStmts
	balanceStmts  *balanceStmts
	txQueuesStmts *txQueuesStmts
}

func New(pgConn string) (newPg *Pg, err error) {
	log.Debug().Msg("Pg.New START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.New END")
		} else {
			log.Debug().Msg("Pg.New END")
		}
	}()

	newPg = &Pg{}

	db, err := sql.Open("pgx", pgConn)
	if err != nil {
		return nil, fmt.Errorf("opening connection with postgres db: %w", err)
	}
	newPg.db = db

	newPg.db.SetMaxOpenConns(20)
	newPg.db.SetMaxIdleConns(20)
	newPg.db.SetConnMaxIdleTime(time.Second * 30)
	newPg.db.SetConnMaxLifetime(time.Minute * 2)

	ctx := context.Background()

	if err = initTables(ctx, newPg); err != nil {
		return nil, fmt.Errorf("initializing tables: %w", err)
	}

	if err = prepareUserStmts(ctx, newPg); err != nil {
		return nil, fmt.Errorf("preparing user stmts: %w", err)
	}

	if err = prepareBalanceStmts(ctx, newPg); err != nil {
		return nil, fmt.Errorf("preparing balance stmts: %w", err)
	}

	if err = prepareTxStmts(ctx, newPg); err != nil {
		return nil, fmt.Errorf("preparing tx queues stmts: %w", err)
	}

	return newPg, nil
}

func initTables(ctx context.Context, p *Pg) (err error) {
	log.Debug().Msg("pg.initTables START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("pg.initTables END")
		} else {
			log.Debug().Msg("pg.initTables END")
		}
	}()

	if p.db == nil {
		return ErrDBIsNilPointer
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, queryCreateTableUsers)
	if err != nil {
		return fmt.Errorf("creating table `users`: %w", err)
	}

	_, err = tx.ExecContext(ctx, queryCreateTableBalance)
	if err != nil {
		return fmt.Errorf("creating table `balance`: %w", err)
	}

	_, err = tx.ExecContext(ctx, queryCreateTableTxQueues)
	if err != nil {
		return fmt.Errorf("creating table `tx_queues`: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (p *Pg) InitFirstFiveUsersIfNotExistsForTestingApp() (err error) {
	log.Debug().Msg("Pg.InitFirstFiveUsersIfNotExistsForTestingApp START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.InitFirstFiveUsersIfNotExistsForTestingApp END")
		} else {
			log.Debug().Msg("Pg.InitFirstFiveUsersIfNotExistsForTestingApp END")
		}
	}()

	usersExist, err := p.userIsExist(1)
	if err != nil {
		return fmt.Errorf("users existence check: %w", err)
	}

	if usersExist {
		return nil
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for id := 1; id <= 5; id++ {

		_, err = tx.Stmt(p.usersStmts.stmtAddUser).Exec()
		if err != nil {
			return fmt.Errorf("creating user: %w", err)
		}

		_, err = tx.Stmt(p.balanceStmts.stmtCreateStartingBalance).Exec(id)
		if err != nil {
			return fmt.Errorf("creating user start balance: %w", err)
		}

	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *Pg) ChangeBalance(ctx context.Context, userID int64, sum float64) (err error) {
	log.Debug().Msg("Pg.ChangeBalance START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.ChangeBalance END")
		} else {
			log.Debug().Msg("Pg.ChangeBalance END")
		}
	}()

	_, err = p.balanceStmts.stmtChangeBalance.ExecContext(ctx, userID, sum)
	if err != nil {
		return fmt.Errorf("user balance change: userID: %d: %w", userID, err)
	}

	return nil
}

func (p *Pg) AddTx(ctx context.Context, userID int64, sum float64) (err error) {
	log.Debug().Msg("Pg.AddTx START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.AddTx END")
		} else {
			log.Debug().Msg("Pg.AddTx END")
		}
	}()

	_, err = p.txQueuesStmts.stmtAddTx.ExecContext(ctx, userID, sum)
	if err != nil {
		return fmt.Errorf("adding a transaction to the user's queue: userID: %d: %w", userID, err)
	}

	return nil
}

func (p *Pg) GetUsersWithNonEmptyTxQueues(ctx context.Context) (users []int64, err error) {
	log.Debug().Msg("Pg.GetUsersWithNonEmptyTxQueues START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.GetUsersWithNonEmptyTxQueues END")
		} else {
			log.Debug().Msg("Pg.GetUsersWithNonEmptyTxQueues END")
		}
	}()

	rows, err := p.txQueuesStmts.stmtGetUsersWithNonEmptyTxQueues.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting users with non empty tx queues: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var currUser int64
		if err = rows.Scan(&currUser); err != nil {
			return nil, fmt.Errorf("reading users with non empty tx queues: %w", err)
		}
		users = append(users, currUser)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("reading users with non empty tx queues: %w", err)
	}

	return users, nil
}

func (p *Pg) ProcessTxQueue(ctx context.Context, userID int64) (err error) {
	log.Debug().Msg("Pg.ProcessTxQueue START")
	defer func() {
		if err != nil {
			if errors.Is(err, ErrInsufficientFunds) {
				log.Info().Err(err).Msg("Pg.ProcessTxQueue END")
			} else {
				log.Error().Err(err).Msg("Pg.ProcessTxQueue END")
			}
		} else {
			log.Debug().Msg("Pg.ProcessTxQueue END")
		}
	}()

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txRows, err := tx.StmtContext(ctx, p.txQueuesStmts.stmtGetTxsByUser).QueryContext(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting transactions by user: userID: %d: %w", userID, err)
	}
	defer txRows.Close()

	txsFromDB := []int64{}
	var sumInQueue float64
	for txRows.Next() {
		var currTxFromDB int64
		var currSum float64
		if err = txRows.Scan(&currTxFromDB, &currSum); err != nil {
			return fmt.Errorf("getting transactions by user: userID: %d: %w", userID, err)
		}
		txsFromDB = append(txsFromDB, currTxFromDB)
		sumInQueue += currSum
	}

	if err = txRows.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("reading transactions by user: userID: %d: %w", userID, err)
	}

	_, err = tx.StmtContext(ctx, p.balanceStmts.stmtChangeBalance).ExecContext(ctx, userID, sumInQueue)
	if err != nil {
		tx.Rollback()
		if pgError, ok := err.(*pgconn.PgError); ok &&
			pgError.Code == pgerrcode.CheckViolation &&
			pgError.ConstraintName == "balance_sum_check" {
			return fmt.Errorf("changing user balance: userID: %d: %w", userID, ErrInsufficientFunds)
		}
		return fmt.Errorf("changing user balance: userID: %d: %w", userID, err)
	}

	_, err = p.txQueuesStmts.stmtDeleteTxsByIds.ExecContext(ctx, pq.Array(txsFromDB))
	if err != nil {
		return fmt.Errorf("clearing txs from queue: userID: %d: %w", userID, err)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil

}

func (p *Pg) Close() (err error) {
	log.Debug().Msg("Pg.Close START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.Close END")
		} else {
			log.Debug().Msg("Pg.Close END")
		}
	}()

	err = p.db.Close()
	if err != nil {
		return fmt.Errorf("close connection with db: %w", err)
	}

	return nil
}
