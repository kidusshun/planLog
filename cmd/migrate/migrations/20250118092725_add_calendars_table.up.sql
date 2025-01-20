CREATE TABLE IF NOT EXISTS users_calendars (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID FOREIGN KEY REFERENCES users(id),
    plans_id TEXT NOT NULL,
    logs_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
)