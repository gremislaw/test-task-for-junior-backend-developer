package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (
			title, description, status, due_date,
			is_recurrence, recurrence_parent_id, recurrence_type, recurrence_config,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, title, description, status, due_date, is_recurrence, 
		          recurrence_parent_id, recurrence_type, recurrence_config, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.DueDate,
		task.IsRecurrence,
		task.RecurrenceParentID,
		task.RecurrenceType,
		task.RecurrenceConfig,
		task.CreatedAt,
		task.UpdatedAt,
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, due_date, is_recurrence,
		       recurrence_parent_id, recurrence_type, recurrence_config, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			due_date = $4,
			is_recurrence = $5,
			recurrence_parent_id = $6,
			recurrence_type = $7,
			recurrence_config = $8,
			updated_at = $9
		WHERE id = $10
		RETURNING id, title, description, status, due_date, is_recurrence,
		          recurrence_parent_id, recurrence_type, recurrence_config, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.DueDate,
		task.IsRecurrence,
		task.RecurrenceParentID,
		task.RecurrenceType,
		task.RecurrenceConfig,
		task.UpdatedAt,
		task.ID,
	)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if err := r.DeleteInstancesByParent(ctx, id); err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) DeleteInstancesByParent(ctx context.Context, parentID int64) error {
	const query = `DELETE FROM tasks WHERE recurrence_parent_id = $1`
	_, err := r.pool.Exec(ctx, query, parentID)
	return err
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, due_date, is_recurrence,
		       recurrence_parent_id, recurrence_type, recurrence_config, created_at, updated_at
		FROM tasks
		ORDER BY due_date ASC, id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) ListRecurrenceParents(ctx context.Context, from, to time.Time) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, due_date, is_recurrence,
		       recurrence_parent_id, recurrence_type, recurrence_config, created_at, updated_at
		FROM tasks
		WHERE is_recurrence = true
		  AND due_date <= $1
		ORDER BY id
	`
	rows, err := r.pool.Query(ctx, query, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parents := make([]taskdomain.Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		parents = append(parents, *t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parents, nil
}

func (r *Repository) GetRecurrenceInstanceDates(ctx context.Context, parentID int64, from, to time.Time) ([]time.Time, error) {
	const query = `
		SELECT due_date FROM tasks
		WHERE recurrence_parent_id = $1 AND due_date >= $2 AND due_date <= $3
	`
	rows, err := r.pool.Query(ctx, query, parentID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		dates = append(dates, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dates, nil
}

func (r *Repository) BatchCreate(ctx context.Context, tasks []*taskdomain.Task) error {
	const batchSize = 500
	for i := 0; i < len(tasks); i += batchSize {
		end := i + batchSize
		if end > len(tasks) {
			end = len(tasks)
		}
		if err := r.batchCreateChunk(ctx, tasks[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) batchCreateChunk(ctx context.Context, chunk []*taskdomain.Task) error {
	if len(chunk) == 0 {
		return nil
	}

	vals := make([]string, 0, len(chunk))
	args := make([]any, 0, len(chunk)*10)

	for i, t := range chunk {
		base := i*10 + 1
		vals = append(vals, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base, base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9,
		))
		args = append(args,
			t.Title, t.Description, t.Status, t.DueDate, t.IsRecurrence,
			t.RecurrenceParentID, t.RecurrenceType, t.RecurrenceConfig,
			t.CreatedAt, t.UpdatedAt,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO tasks (
			title, description, status, due_date, is_recurrence,
			recurrence_parent_id, recurrence_type, recurrence_config,
			created_at, updated_at
		) VALUES %s
		ON CONFLICT (recurrence_parent_id, due_date)
		WHERE recurrence_parent_id IS NOT NULL
		DO NOTHING
	`, strings.Join(vals, ","))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *Repository) ListByRange(ctx context.Context, from, to time.Time, cursorDate *time.Time, cursorID *int64, limit int) ([]*taskdomain.Task, bool, error) {
	query := `
		SELECT id, title, description, status, due_date, is_recurrence,
           recurrence_parent_id, recurrence_type, recurrence_config,
           created_at, updated_at
    FROM tasks
    WHERE due_date >= $1 AND due_date <= $2
	`

	var args []any
	if cursorDate != nil {
		query += ` AND ((due_date > $3) OR (due_date = $3 AND id < $4))`
		query += ` ORDER BY due_date ASC, id DESC LIMIT $5`
		args = []any{from, to, cursorDate, cursorID, limit + 1}
	} else {
		query += ` ORDER BY due_date ASC, id DESC LIMIT $3`
		args = []any{from, to, limit + 1}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*taskdomain.Task, 0, limit)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, false, err
		}
		tasks = append(tasks, t)
	}

	hasMore := false
	if len(tasks) > limit {
		tasks = tasks[:limit]
		hasMore = true
	}

	return tasks, hasMore, rows.Err()
}

func (r *Repository) DeleteOldInstances(ctx context.Context, olderThan time.Time, limit int) (int64, error) {
	query := `
      DELETE FROM tasks 
      WHERE id IN (
          SELECT id FROM tasks 
          WHERE status = $1
            AND due_date < $2
            AND (recurrence_parent_id IS NOT NULL OR is_recurrence = false)
          ORDER BY due_date 
          LIMIT $3
        )
    `
	tag, err := r.pool.Exec(ctx, query,
		taskdomain.StatusNew,
		olderThan,
		limit,
	)
	if err != nil {
		return 0, fmt.Errorf("delete old: %w", err)
	}
	return tag.RowsAffected(), nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
		rt     string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.DueDate,
		&task.IsRecurrence,
		&task.RecurrenceParentID,
		&rt,
		&task.RecurrenceConfig, // RecurrenceConfig.Scan()
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	task.RecurrenceType = taskdomain.RecurrenceType(rt)

	return &task, nil
}
