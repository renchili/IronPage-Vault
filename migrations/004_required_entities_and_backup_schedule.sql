CREATE TABLE IF NOT EXISTS document_files (
  id TEXT PRIMARY KEY,
  document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  version_id TEXT NOT NULL UNIQUE REFERENCES document_versions(id) ON DELETE CASCADE,
  file_path TEXT NOT NULL,
  file_sha256 TEXT NOT NULL,
  size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
  page_count INTEGER NOT NULL CHECK (page_count > 0),
  created_by TEXT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS document_files_document_id_idx
  ON document_files(document_id, created_at DESC);

INSERT INTO document_files(
  id, document_id, version_id, file_path, file_sha256,
  size_bytes, page_count, created_by, created_at
)
SELECT
  'fil_' || version.id,
  version.document_id,
  version.id,
  version.file_path,
  version.file_sha256,
  version.size_bytes,
  version.page_count,
  version.created_by,
  version.created_at
FROM document_versions AS version
ON CONFLICT(version_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS redaction_confirmations (
  id TEXT PRIMARY KEY,
  proposal_id TEXT NOT NULL UNIQUE REFERENCES redaction_proposals(id) ON DELETE CASCADE,
  document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  source_version_id TEXT REFERENCES document_versions(id),
  result_version_id TEXT REFERENCES document_versions(id),
  confirmed_by TEXT NOT NULL REFERENCES users(id),
  confirmed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS redaction_confirmations_document_id_idx
  ON redaction_confirmations(document_id, confirmed_at DESC);

INSERT INTO redaction_confirmations(
  id, proposal_id, document_id, confirmed_by, confirmed_at
)
SELECT
  'rcf_' || proposal.id,
  proposal.id,
  proposal.document_id,
  proposal.confirmed_by,
  proposal.confirmed_at
FROM redaction_proposals AS proposal
WHERE proposal.status = 'Confirmed'
  AND proposal.confirmed_by IS NOT NULL
  AND proposal.confirmed_at IS NOT NULL
ON CONFLICT(proposal_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS document_diffs (
  id TEXT PRIMARY KEY,
  left_document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  right_document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  left_version_id TEXT NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
  right_version_id TEXT NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
  result_ciphertext TEXT NOT NULL,
  created_by TEXT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS document_diffs_version_pair_idx
  ON document_diffs(left_version_id, right_version_id, created_at DESC);

INSERT INTO config_entries(key, value) VALUES
  ('backup.schedule_enabled', 'false'),
  ('backup.interval', '24h')
ON CONFLICT(key) DO NOTHING;
