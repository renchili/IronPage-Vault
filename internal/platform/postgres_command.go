package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PostgresCommandConfig contains the non-networked local PostgreSQL identity
// required by pg_dump and pg_restore. Passwords are supplied through a
// short-lived 0600 PGPASSFILE and never placed in subprocess arguments.
type PostgresCommandConfig struct {
	Port          string
	User          string
	Password      string
	Database      string
	CredentialDir string
}

func (c PostgresCommandConfig) validate() error {
	for name, value := range map[string]string{
		"port":           c.Port,
		"user":           c.User,
		"password":       c.Password,
		"database":       c.Database,
		"credential_dir": c.CredentialDir,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("postgres command %s is required", name)
		}
		if strings.ContainsAny(value, "\r\n") {
			return fmt.Errorf("postgres command %s contains a line break", name)
		}
	}
	if !filepath.IsAbs(c.CredentialDir) {
		return fmt.Errorf("postgres command credential_dir must be absolute")
	}
	return nil
}

func pgpassEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, ":", `\:`)
}

func writePGPassFile(config PostgresCommandConfig) (string, error) {
	if err := config.validate(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(config.CredentialDir, 0700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(config.CredentialDir, ".ironpage-pgpass-")
	if err != nil {
		return "", err
	}
	path := file.Name()
	remove := true
	defer func() {
		if remove {
			_ = file.Close()
			_ = os.Remove(path)
		}
	}()
	if err := file.Chmod(0600); err != nil {
		return "", err
	}
	line := strings.Join([]string{"*", pgpassEscape(config.Port), pgpassEscape(config.Database), pgpassEscape(config.User), pgpassEscape(config.Password)}, ":") + "\n"
	if _, err := file.WriteString(line); err != nil {
		return "", err
	}
	if err := file.Sync(); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	remove = false
	return path, nil
}

func postgresCommandEnv(pgpassPath string) []string {
	environment := make([]string, 0, len(os.Environ())+1)
	for _, value := range os.Environ() {
		key := value
		if index := strings.IndexByte(value, '='); index >= 0 {
			key = value[:index]
		}
		if key == "PGPASSWORD" || key == "PGPASSFILE" {
			continue
		}
		environment = append(environment, value)
	}
	return append(environment, "PGPASSFILE="+pgpassPath)
}

func runPostgresCommand(name string, args []string, config PostgresCommandConfig) error {
	pgpassPath, err := writePGPassFile(config)
	if err != nil {
		return err
	}
	defer os.Remove(pgpassPath)
	command := exec.Command(name, args...)
	command.Env = postgresCommandEnv(pgpassPath)
	return command.Run()
}
