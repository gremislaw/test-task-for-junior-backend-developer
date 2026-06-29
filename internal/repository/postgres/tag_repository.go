package postgres

import (
    "context"
    "fmt"

    "example.com/taskservice/internal/domain/task"
    "github.com/jackc/pgx/v5"
)

func (r *Repository) CreateTag(ctx context.Context, name string) (*taskdomain.Tag, error) {
    var tag taskdomain.Tag
    query := `INSERT INTO tags (name) VALUES ($1) RETURNING id, name, is_system`
    err := r.pool.QueryRow(ctx, query, name).Scan(&tag.ID, &tag.Name, &tag.IsSystem)
    if err != nil {
        return nil, fmt.Errorf("create tag: %w", err)
    }
    return &tag, nil
}

func (r *Repository) GetTagByID(ctx context.Context, id int64) (*taskdomain.Tag, error) {
    var tag taskdomain.Tag
    query := `SELECT id, name, is_system FROM tags WHERE id = $1`
    err := r.pool.QueryRow(ctx, query, id).Scan(&tag.ID, &tag.Name, &tag.IsSystem)
    if err != nil {
        return nil, fmt.Errorf("get tag: %w", err)
    }
    return &tag, nil
}

func (r *Repository) DeleteTag(ctx context.Context, id int64) error {
    tag, err := r.GetTagByID(ctx, id)
    if err != nil { return err }
    if tag.IsSystem { return ErrSystemTagProtected }

    res, err := r.pool.Exec(ctx, `DELETE FROM tags WHERE id = $1`, id)
    if err != nil { return fmt.Errorf("delete tag: %w", err) }
    if res.RowsAffected() == 0 { return ErrNotFound }
    return nil
}

func (r *Repository) AssignTags(ctx context.Context, taskID int64, tagIDs []int64) error {
    if len(tagIDs) == 0 { return nil }

    tx, err := r.pool.Begin(ctx)
    if err != nil { return fmt.Errorf("begin tx: %w", err) }
    defer tx.Rollback(ctx)

    query := `INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
    for _, tid := range tagIDs {
        if _, err := tx.Exec(ctx, query, taskID, tid); err != nil {
            return fmt.Errorf("assign tag %d: %w", tid, err)
        }
    }
    return tx.Commit(ctx)
}

func (r *Repository) RemoveTags(ctx context.Context, taskID int64, tagIDs []int64) error {
    if len(tagIDs) == 0 { return nil }

    tx, err := r.pool.Begin(ctx)
    if err != nil { return fmt.Errorf("begin tx: %w", err) }
    defer tx.Rollback(ctx)

    query := `DELETE FROM task_tags WHERE task_id = $1 AND tag_id = ANY($2)`
    _, err = tx.Exec(ctx, query, taskID, tagIDs)
    if err != nil { return fmt.Errorf("remove tags: %w", err) }

    return tx.Commit(ctx)
}