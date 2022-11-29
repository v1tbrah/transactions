package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

const queryCreateTableBalance = `
CREATE TABLE IF NOT EXISTS balance
(
	id             bigserial PRIMARY KEY,
	user_id        bigint REFERENCES users(id) ON DELETE CASCADE,
	sum            double precision NOT NULL CHECK (NOT(sum < 0))
);
`

const queryCreateStartingBalance = `INSERT INTO balance (user_id, sum) VALUES ($1, 0)`

const queryChangeBalance = `UPDATE balance SET sum = sum + $2 WHERE user_id=$1`

type balanceStmts struct {
	stmtCreateStartingBalance *sql.Stmt
	stmtChangeBalance         *sql.Stmt
}

func prepareBalanceStmts(ctx context.Context, p *Pg) (err error) {
	log.Debug().Msg("pg.prepareBalanceStmts START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("pg.prepareBalanceStmts END")
		} else {
			log.Debug().Msg("pg.prepareBalanceStmts END")
		}
	}()

	newBalanceStmts := balanceStmts{}

	if newBalanceStmts.stmtCreateStartingBalance, err = p.db.PrepareContext(ctx, queryCreateStartingBalance); err != nil {
		return fmt.Errorf("preparing `create starting user balance` stmt: %w", err)
	}

	if newBalanceStmts.stmtChangeBalance, err = p.db.PrepareContext(ctx, queryChangeBalance); err != nil {
		return fmt.Errorf("preparing `change balance` stmt: %w", err)
	}

	p.balanceStmts = &newBalanceStmts

	return nil
}
