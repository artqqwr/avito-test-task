package postgres

import (
	"avito-test-task/internal/domain"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepo struct {
	db *pgxpool.Pool
}

func NewTeamRepo(db *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) CreateTeamWithMembers(ctx context.Context, team domain.Team) error {
	return withTx(ctx, r.db, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", team.Name)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				return domain.ErrTeamExists
			}
			return err
		}

		batch := &pgx.Batch{}
		query := `
			INSERT INTO users (id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE 
			SET username = EXCLUDED.username, 
			    team_name = EXCLUDED.team_name, 
			    is_active = EXCLUDED.is_active`

		for _, m := range team.Members {
			batch.Queue(query, m.ID, m.Username, team.Name, m.IsActive)
		}

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		for range team.Members {
			if _, err := br.Exec(); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *TeamRepo) GetTeamByName(ctx context.Context, name string) (domain.Team, error) {
	var team domain.Team
	team.Name = name

	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", name).Scan(&exists)
	if err != nil {
		return domain.Team{}, err
	}
	if !exists {
		return domain.Team{}, domain.ErrNotFound
	}

	rows, err := r.db.Query(ctx, "SELECT id, username, team_name, is_active FROM users WHERE team_name = $1", name)
	if err != nil {
		return domain.Team{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return domain.Team{}, err
		}
		team.Members = append(team.Members, u)
	}

	if err := rows.Err(); err != nil {
		return domain.Team{}, err
	}

	return team, nil
}
