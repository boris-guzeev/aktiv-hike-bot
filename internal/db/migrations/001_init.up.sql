CREATE TABLE  IF NOT EXISTS hikes (
    id              SERIAL PRIMARY KEY,
    title_ru        TEXT NOT NULL,
    title_en        TEXT,
    description_ru  TEXT NOT NULL,
    description_en  TEXT,
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ NOT NULL,
    photo_file_id   TEXT,
    is_published    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hikes_published_start ON hikes (is_published, starts_at);

CREATE TABLE tg_users (
  id            SERIAL PRIMARY KEY,
  tg_user_id    BIGINT UNIQUE NOT NULL,
  tg_username   TEXT,
  full_name     TEXT,
  lang          TEXT NOT NULL DEFAULT 'ru', -- 'ru' | 'en'
  is_admin      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tg_users_username ON tg_users (tg_username);

CREATE TABLE bookings (
  id         SERIAL PRIMARY KEY,
  hike_id    INT NOT NULL REFERENCES hikes(id) ON DELETE RESTRICT,
  user_id    INT NOT NULL REFERENCES tg_users(id) ON DELETE RESTRICT,
  status     TEXT NOT NULL DEFAULT 'pending', -- pending|approved|rejected|cancelled
  note       TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(hike_id, user_id)
);

CREATE INDEX idx_bookings_hike_status ON bookings (hike_id, status);

CREATE TABLE payments (
  id            SERIAL PRIMARY KEY,
  booking_id    INT NOT NULL REFERENCES bookings(id) ON DELETE RESTRICT,
  amount        NUMERIC(12,2) NOT NULL DEFAULT 0,
  currency      TEXT NOT NULL DEFAULT 'GEL',
  proof_file_id TEXT,   -- telegram file_id на чек/фото
  status        TEXT NOT NULL DEFAULT 'submitted', -- submitted|verified|rejected
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_booking_status ON payments (booking_id, status);

