package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

const queryCreateTableTxQueues = `
CREATE TABLE IF NOT EXISTS tx_queues
(
	id             bigserial PRIMARY KEY,
	user_id        bigint REFERENCES users(id) ON DELETE CASCADE,
	sum            double precision NOT NULL
);
`

const (
	queryAddTx                        = `INSERT INTO tx_queues (user_id, sum) VALUES ($1, $2)`
	queryGetTxsByUser                 = `SELECT id, sum FROM tx_queues WHERE user_id = $1`
	queryDeleteTxsByIds               = `DELETE FROM tx_queues WHERE id = any($1)`
	queryGetUsersWithNonEmptyTxQueues = `SELECT DISTINCT user_id FROM tx_queues`
)

type txQueuesStmts struct {
	stmtAddTx                        *sql.Stmt
	stmtGetTxsByUser                 *sql.Stmt
	stmtDeleteTxsByIds               *sql.Stmt
	stmtGetUsersWithNonEmptyTxQueues *sql.Stmt
}

func prepareTxStmts(ctx context.Context, p *Pg) (err error) {
	log.Debug().Msg("pg.prepareTxStmts START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("pg.prepareTxStmts END")
		} else {
			log.Debug().Msg("pg.prepareTxStmts END")
		}
	}()

	newTxQueuesStmts := txQueuesStmts{}

	if newTxQueuesStmts.stmtAddTx, err = p.db.PrepareContext(ctx, queryAddTx); err != nil {
		return fmt.Errorf("preparing `add tx` stmt: %w", err)
	}

	if newTxQueuesStmts.stmtGetTxsByUser, err = p.db.PrepareContext(ctx, queryGetTxsByUser); err != nil {
		return fmt.Errorf("preparing `get txs by user` stmt: %w", err)
	}

	if newTxQueuesStmts.stmtDeleteTxsByIds, err = p.db.PrepareContext(ctx, queryDeleteTxsByIds); err != nil {
		return fmt.Errorf("preparing `delete txs by ids` stmt: %w", err)
	}

	if newTxQueuesStmts.stmtGetUsersWithNonEmptyTxQueues, err = p.db.PrepareContext(ctx, queryGetUsersWithNonEmptyTxQueues); err != nil {
		return fmt.Errorf("preparing `get users with non empty txs queues` stmt: %w", err)
	}

	p.txQueuesStmts = &newTxQueuesStmts

	return nil
}
