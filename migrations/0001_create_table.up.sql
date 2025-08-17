-- Create table
CREATE TABLE IF NOT EXISTS "requests" (
    "domain"              TEXT NOT NULL,                                    -- Domain that was requested
    "created_at"          TIMESTAMPTZ NOT NULL,                             -- Time this record was created
    "updated_at"          TIMESTAMPTZ NOT NULL,                             -- Time this record was created

    PRIMARY KEY ("domain")                                                  -- "domain" is primary key
);
