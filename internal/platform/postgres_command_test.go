package platform

import (
	"os"
	"strings"
	"testing"
)

func testPostgresCommandConfig(t *testing.T) PostgresCommandConfig {
	t.Helper()
	return PostgresCommandConfig{
		Port:          "5432",
		User:          "ironpage",
		Password:      `secret:with\slashes`,
		Database:      "vault",
		CredentialDir: t.TempDir(),
	}
}

func TestPostgresCommandArgumentsExcludePassword(t *testing.T) {
	config := testPostgresCommandConfig(t)
	for name, args := range map[string][]string{
		"pg_dump":    pgDumpCommandArgs(config, "/backup/vault.dump"),
		"pg_restore": pgRestoreCommandArgs(config, "/backup/vault.dump"),
	} {
		joined := strings.Join(args, " ")
		if strings.Contains(joined, config.Password) || strings.Contains(joined, "password=") {
			t.Fatalf("%s arguments exposed the database password: %q", name, joined)
		}
	}
}

func TestPGPassFileUsesRestrictedModeAndEscaping(t *testing.T) {
	config := testPostgresCommandConfig(t)
	path, err := writePGPassFile(config)
	if err != nil {
		t.Fatalf("write pgpass file: %v", err)
	}
	defer os.Remove(path)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat pgpass file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("pgpass mode = %o, want 600", info.Mode().Perm())
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pgpass file: %v", err)
	}
	expected := `*:5432:vault:ironpage:secret\:with\\slashes` + "\n"
	if string(raw) != expected {
		t.Fatalf("pgpass content = %q, want %q", string(raw), expected)
	}
}

func TestPostgresCommandEnvironmentRemovesAmbientPassword(t *testing.T) {
	t.Setenv("PGPASSWORD", "ambient-secret")
	t.Setenv("PGPASSFILE", "/tmp/ambient-pgpass")
	environment := postgresCommandEnv("/tmp/scoped-pgpass")
	joined := strings.Join(environment, "\n")
	if strings.Contains(joined, "PGPASSWORD=") || strings.Contains(joined, "PGPASSFILE=/tmp/ambient-pgpass") {
		t.Fatalf("ambient postgres credentials remained in command environment")
	}
	if !strings.Contains(joined, "PGPASSFILE=/tmp/scoped-pgpass") {
		t.Fatalf("scoped PGPASSFILE is missing")
	}
}

func TestPostgresCommandConfigRejectsLineBreaks(t *testing.T) {
	config := testPostgresCommandConfig(t)
	config.Password = "secret\ninjected"
	if err := config.validate(); err == nil {
		t.Fatal("expected line break in postgres credentials to be rejected")
	}
}
