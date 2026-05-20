package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Apply applies desired state to all instances (idempotent).
func Apply(ctx context.Context, cfg *Config, dryRun bool) error {
	password := os.Getenv("POSTGRES_ADMIN_PASSWORD")
	if password == "" {
		password = os.Getenv("PGPASSWORD")
	}
	if password == "" && !dryRun {
		return fmt.Errorf("set POSTGRES_ADMIN_PASSWORD or PGPASSWORD")
	}
	for _, inst := range cfg.Instances {
		if dryRun {
			fmt.Printf("Instance: %s:%d\n", inst.Host, inst.Port)
			for _, d := range inst.Databases {
				fmt.Printf("  database: %s owner %s\n", d.Name, d.Owner)
			}
			for _, u := range inst.Users {
				fmt.Printf("  user: %s\n", u.Name)
			}
			continue
		}
		if err := applyInstance(ctx, inst, password); err != nil {
			return fmt.Errorf("%s:%d: %w", inst.Host, inst.Port, err)
		}
	}
	return nil
}

func applyInstance(ctx context.Context, inst Instance, adminPassword string) error {
	connStr := fmt.Sprintf("host=%s port=%d user=postgres password=%s dbname=postgres sslmode=disable",
		inst.Host, inst.Port, adminPassword)
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	for _, u := range inst.Users {
		if err := ensureUser(ctx, conn, u); err != nil {
			return err
		}
	}
	for _, d := range inst.Databases {
		if err := ensureDatabase(ctx, inst, adminPassword, d); err != nil {
			return err
		}
	}
	for _, u := range inst.Users {
		for _, dbName := range u.Databases {
			if err := grantConnect(ctx, inst, adminPassword, dbName, u.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureUser(ctx context.Context, conn *pgx.Conn, u User) error {
	var exists bool
	err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname=$1)`, u.Name).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		sql := fmt.Sprintf(`CREATE ROLE %s WITH LOGIN PASSWORD $1`, quoteIdent(u.Name))
		_, err = conn.Exec(ctx, sql, u.Password)
		return err
	}
	sql := fmt.Sprintf(`ALTER ROLE %s WITH PASSWORD $1`, quoteIdent(u.Name))
	_, err = conn.Exec(ctx, sql, u.Password)
	return err
}

func ensureDatabase(ctx context.Context, inst Instance, adminPassword string, d Database) error {
	admin, err := pgx.Connect(ctx, fmt.Sprintf("host=%s port=%d user=postgres password=%s dbname=postgres sslmode=disable",
		inst.Host, inst.Port, adminPassword))
	if err != nil {
		return err
	}
	defer admin.Close(ctx)
	var exists bool
	if err := admin.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, d.Name).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	// CREATE DATABASE cannot run in a transaction block in some drivers; use simple Exec outside tx.
	sql := fmt.Sprintf(`CREATE DATABASE %s OWNER %s`, quoteIdent(d.Name), quoteIdent(d.Owner))
	_, err = admin.Exec(ctx, sql)
	return err
}

func grantConnect(ctx context.Context, inst Instance, adminPassword, dbName, user string) error {
	connStr := fmt.Sprintf("host=%s port=%d user=postgres password=%s dbname=%s sslmode=disable",
		inst.Host, inst.Port, adminPassword, dbName)
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	sql := fmt.Sprintf(`GRANT CONNECT ON DATABASE %s TO %s`, quoteIdent(dbName), quoteIdent(user))
	_, err = conn.Exec(ctx, sql)
	return err
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
