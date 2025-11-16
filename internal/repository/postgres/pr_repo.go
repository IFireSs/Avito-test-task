package postgres

import (
	"avito-autumn2025-internship/internal/api"
	"avito-autumn2025-internship/internal/repository"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type prRepository struct {
	pool *pgxpool.Pool
}

func NewPRRepository(pool *pgxpool.Pool) repository.PRRepository {
	return &prRepository{pool: pool}
}

func (r *prRepository) Create(ctx context.Context, pr *api.PullRequest) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		pr.PullRequestId,
		pr.PullRequestName,
		pr.AuthorId,
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return err
	}

	if len(pr.AssignedReviewers) > 0 {
		batch := &pgx.Batch{}
		for _, reviewerID := range pr.AssignedReviewers {
			batch.Queue(`
				INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
				VALUES ($1, $2)
			`, pr.PullRequestId, reviewerID)
		}
		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *prRepository) GetByID(ctx context.Context, prID string) (*api.PullRequest, error) {
	var pr api.PullRequest
	var status string
	err := r.pool.QueryRow(ctx, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`, prID).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	pr.Status = api.PullRequestStatus(status)

	reviewers, err := r.ListReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *prRepository) SetMerged(ctx context.Context, prID string, mergedAt time.Time) (*api.PullRequest, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE pull_requests
		SET status = 'MERGED',
		    merged_at = COALESCE(merged_at, $2)
		WHERE pull_request_id = $1
	`, prID, mergedAt)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, prID)
}

func (r *prRepository) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE pull_request_reviewers
		SET reviewer_id = $3
		WHERE pull_request_id = $1 AND reviewer_id = $2
	`, prID, oldReviewerID, newReviewerID)
	return err
}

func (r *prRepository) ListReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
		ORDER BY reviewer_id
	`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		res = append(res, id)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

func (r *prRepository) SetReviewers(ctx context.Context, prID string, reviewers []string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1
	`, prID); err != nil {
		return err
	}

	if len(reviewers) > 0 {
		batch := &pgx.Batch{}
		for _, reviewerID := range reviewers {
			batch.Queue(`
				INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
				VALUES ($1, $2)
			`, prID, reviewerID)
		}
		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *prRepository) ListShortByReviewer(ctx context.Context, reviewerID string) ([]api.PullRequestShort, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pr.pull_request_id,
		       pr.pull_request_name,
		       pr.author_id,
		       pr.status
		FROM pull_requests pr
		JOIN pull_request_reviewers r
		  ON pr.pull_request_id = r.pull_request_id
		WHERE r.reviewer_id = $1
		ORDER BY pr.created_at, pr.pull_request_id
	`, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []api.PullRequestShort
	for rows.Next() {
		var item api.PullRequestShort
		var status string
		if err := rows.Scan(
			&item.PullRequestId,
			&item.PullRequestName,
			&item.AuthorId,
			&status,
		); err != nil {
			return nil, err
		}
		item.Status = api.PullRequestShortStatus(status)
		result = append(result, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return result, nil
}

func (r *prRepository) GetReviewerAssignmentsStats(ctx context.Context) ([]repository.ReviewerAssignmentsStat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT reviewer_id, COUNT(*) AS cnt
		FROM pull_request_reviewers
		GROUP BY reviewer_id
		ORDER BY reviewer_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []repository.ReviewerAssignmentsStat
	for rows.Next() {
		var s repository.ReviewerAssignmentsStat
		if err := rows.Scan(&s.UserID, &s.Count); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}
