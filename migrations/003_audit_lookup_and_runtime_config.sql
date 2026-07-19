ALTER TABLE audit_logs
  ADD COLUMN IF NOT EXISTS source_ip_lookup TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS audit_logs_source_ip_lookup_idx
  ON audit_logs(source_ip_lookup);

DELETE FROM config_entries
WHERE key = 'backup.local_volume'
  AND value = '/var/lib/ironpage/backups';
