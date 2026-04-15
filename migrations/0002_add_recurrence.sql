ALTER TABLE tasks 
  ADD COLUMN IF NOT EXISTS due_date TIMESTAMPTZ NOT NULL DEFAULT (created_at);

ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS recurrence_parent_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS recurrence_type VARCHAR(20),
  ADD COLUMN IF NOT EXISTS recurrence_config JSONB;

CREATE UNIQUE INDEX IF NOT EXISTS idx_task_parent_due
  ON tasks(recurrence_parent_id, due_date) 
  WHERE recurrence_parent_id IS NOT NULL;