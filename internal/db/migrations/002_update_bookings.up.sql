ALTER TABLE bookings
  ADD COLUMN IF NOT EXISTS assigned_admin_id INT REFERENCES tg_users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS taken_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX IF NOT EXISTS idx_bookings_assigned_admin ON bookings (assigned_admin_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings (status);