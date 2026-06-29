package task

import (
	"context"
	"fmt"

	"example.com/taskservice/internal/domain/task"
)

type TagService struct {
	repo Repository
}

func NewTagService(repo Repository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) CreateTag(ctx context.Context, name string) (*taskdomain.Tag, error) {
	if name == "" {
		return nil, fmt.Errorf("tag name required")
	}
	return s.repo.CreateTag(ctx, name)
}

func (s *TagService) DeleteTag(ctx context.Context, id int64) error {
	return s.repo.DeleteTag(ctx, id)
}

func (s *TagService) AssignTagsToTask(ctx context.Context, taskID int64, tagIDs []int64) error {
	if taskID <= 0 || len(tagIDs) == 0 {
		return fmt.Errorf("invalid input")
	}
	return s.repo.AssignTags(ctx, taskID, tagIDs)
}

func (s *TagService) RemoveTagsFromTask(ctx context.Context, taskID int64, tagIDs []int64) error {
	if taskID <= 0 || len(tagIDs) == 0 {
		return fmt.Errorf("invalid input")
	}
	return s.repo.RemoveTags(ctx, taskID, tagIDs)
}
