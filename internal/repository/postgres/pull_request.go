package postgres

import (
	"avito-test-task/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PRRepo struct {
	db *pgxpool.Pool
}

func NewPRRepo(db *pgxpool.Pool) *PRRepo {
	return &PRRepo{db: db}
}

func (r *PRRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	return withTx(ctx, r.db, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO pull_requests (id, name, author_id, status, created_at)
			VALUES ($1, $2, $3, $4, $5)`,
			pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrPRExists
			}
			return err
		}

		if len(pr.Reviewers) > 0 {
			batch := &pgx.Batch{}
			for _, rID := range pr.Reviewers {
				batch.Queue("INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)", pr.ID, rID)
			}
			br := tx.SendBatch(ctx, batch)
			defer br.Close()
			for range pr.Reviewers {
				if _, err := br.Exec(); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *PRRepo) GetByID(ctx context.Context, id string) (domain.PullRequest, error) {
	var pr domain.PullRequest
	err := r.db.QueryRow(ctx, `
		SELECT id, name, author_id, status, created_at, merged_at 
		FROM pull_requests WHERE id = $1`, id).
		Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PullRequest{}, domain.ErrNotFound
		}
		return domain.PullRequest{}, err
	}

	rows, err := r.db.Query(ctx, "SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1", id)
	if err != nil {
		return domain.PullRequest{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var revID string
		if err := rows.Scan(&revID); err != nil {
			return domain.PullRequest{}, err
		}
		pr.Reviewers = append(pr.Reviewers, revID)
	}

	return pr, rows.Err()
}

func (r *PRRepo) Merge(ctx context.Context, id string) (domain.PullRequest, error) {
	var pr domain.PullRequest
	err := withTx(ctx, r.db, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE pull_requests 
			SET status = 'MERGED', merged_at = NOW() 
			WHERE id = $1 AND status = 'OPEN'`, id)
		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, `
			SELECT id, name, author_id, status, created_at, merged_at 
			FROM pull_requests WHERE id = $1`, id).
			Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	})

	if err != nil {
		return domain.PullRequest{}, err
	}

	rows, err := r.db.Query(ctx, "SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1", id)
	if err != nil {
		return domain.PullRequest{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var rid string
		rows.Scan(&rid)
		pr.Reviewers = append(pr.Reviewers, rid)
	}

	return pr, nil
}

func (r *PRRepo) UpdateReviewer(ctx context.Context, prID, oldID, newID string) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE pr_reviewers 
		SET reviewer_id = $1 
		WHERE pull_request_id = $2 AND reviewer_id = $3`,
		newID, prID, oldID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotAssigned
	}
	return nil
}

func (r *PRRepo) GetByReviewerID(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	query := `
		SELECT pr.id, pr.name, pr.author_id, pr.status 
		FROM pull_requests pr
		JOIN pr_reviewers rev ON pr.id = rev.pull_request_id
		WHERE rev.reviewer_id = $1`

	rows, err := r.db.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}
