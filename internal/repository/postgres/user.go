package postgres

import (
	"avito-test-task/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	var u domain.User
	query := `UPDATE users SET is_active = $1 WHERE id = $2 RETURNING id, username, team_name, is_active`

	err := r.db.QueryRow(ctx, query, isActive, userID).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx,
		"SELECT id, username, team_name, is_active FROM users WHERE id = $1", userID).
		Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}

func (r *UserRepo) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, username, team_name, is_active FROM users WHERE team_name = $1 AND is_active = true", teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
