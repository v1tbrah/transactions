package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

const queryCreateTableUsers = `
CREATE TABLE IF NOT EXISTS users
(
	id bigserial PRIMARY KEY
);
`

const (
	queryAddUser = `INSERT INTO users DEFAULT VALUES`
	queryGetUser = `SELECT id FROM users WHERE id = $1`
)

type usersStmts struct {
	stmtAddUser *sql.Stmt
	stmtGetUser *sql.Stmt
}

func prepareUserStmts(ctx context.Context, p *Pg) (err error) {
	log.Debug().Msg("pg.prepareUserStmts START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("pg.prepareUserStmts END")
		} else {
			log.Debug().Msg("pg.prepareUserStmts END")
		}
	}()

	newUsersStmts := usersStmts{}

	if newUsersStmts.stmtAddUser, err = p.db.PrepareContext(ctx, queryAddUser); err != nil {
		return fmt.Errorf("preparing `add user` stmt: %w", err)
	}

	if newUsersStmts.stmtGetUser, err = p.db.PrepareContext(ctx, queryGetUser); err != nil {
		return fmt.Errorf("preparing `get user` stmt: %w", err)
	}

	p.usersStmts = &newUsersStmts

	return nil
}

func (p *Pg) userIsExist(userID int64) (isExists bool, err error) {
	log.Debug().Str("userID", fmt.Sprint(userID)).Msg("pg.prepareUserStmts START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("pg.prepareUserStmts END")
		} else {
			log.Debug().Msg("pg.prepareUserStmts END")
		}
	}()

	var existsID int64
	err = p.usersStmts.stmtGetUser.QueryRow(userID).Scan(&existsID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("user existence check: %w", err)
	}

	return existsID != 0, nil
}
