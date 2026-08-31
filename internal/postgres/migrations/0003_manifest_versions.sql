CREATE TABLE manifest_versions (
    version bigserial PRIMARY KEY,
    document text NOT NULL,
    checksum text NOT NULL,
    activated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    activated_by text NOT NULL DEFAULT 'admin'
);

CREATE TABLE active_manifest (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    version bigint NOT NULL UNIQUE REFERENCES manifest_versions(version)
);
