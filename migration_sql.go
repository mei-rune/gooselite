package goose

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const sqlCmdPrefix = "-- +goose "

func endsWithSemicolon(line string) bool {

	prev := ""
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "--") {
			break
		}
		prev = word
	}

	return strings.HasSuffix(prev, ";")
}

// Split the given sql script into individual statements.
//
// The base case is to simply split on semicolons, as these
// naturally terminate a statement.
//
// However, more complex cases like pl/pgsql can have semicolons
// within a statement. For these cases, we provide the explicit annotations
// 'StatementBegin' and 'StatementEnd' to allow the script to
// tell us to ignore semicolons.
func splitSQLStatements(r io.Reader, direction bool) (stmts []string, err error) {

	var buf bytes.Buffer
	scanner := bufio.NewScanner(r)

	// track the count of each section
	// so we can diagnose scripts with no annotations
	upSections := 0
	downSections := 0

	statementEnded := false
	ignoreSemicolons := false
	directionIsActive := false

	for scanner.Scan() {

		line := scanner.Text()

		// handle any goose-specific commands
		if strings.HasPrefix(line, sqlCmdPrefix) {
			cmd := strings.TrimSpace(line[len(sqlCmdPrefix):])
			switch cmd {
			case "Up":
				directionIsActive = (direction == true)
				upSections++
				break

			case "Down":
				directionIsActive = (direction == false)
				downSections++
				break

			case "StatementBegin":
				if directionIsActive {
					ignoreSemicolons = true
				}
				break

			case "StatementEnd":
				if directionIsActive {
					statementEnded = (ignoreSemicolons == true)
					ignoreSemicolons = false
				}
				break
			}
		}

		if !directionIsActive {
			continue
		}

		if _, err := buf.WriteString(line + "\n"); err != nil {
			return nil, err
		}

		if !ignoreSemicolons && (statementEnded || endsWithSemicolon(line)) {
			statementEnded = false
			stmts = append(stmts, buf.String())
			buf.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan migration: %w", err)
	}

	// diagnose likely migration script errors
	if ignoreSemicolons {
		log.Println("WARNING: saw '-- +goose StatementBegin' with no matching '-- +goose StatementEnd'")
	}

	if upSections == 0 && downSections == 0 {
		return nil, fmt.Errorf("no Up/Down annotations found, so no statements were executed")
	}

	return
}

// Run a migration specified in raw SQL.
//
// Sections of the script can be annotated with a special comment,
// starting with "-- +goose" to specify whether the section should
// be applied during an Up or Down migration
//
// All statements following an Up or Down directive are grouped together
// until another direction directive is found.
func runSQLMigration(ctx context.Context, conf *DBConfig, dialect SqlDialect, db *sql.DB, script string, v int64, direction bool) error {
	f, err := os.Open(script)
	if err != nil {
		return fmt.Errorf("open migration file: %w", err)
	}
	defer f.Close()

	// find each statement, checking annotations for up/down direction
	// and execute each of them in the current transaction
	stmts, err := splitSQLStatements(f, direction)
	if err != nil {
		return fmt.Errorf("split SQL statements: %w", err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("split SQL statements: %w", err)
	}

	txn, err := db.Begin()
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}

	var committed bool
	defer func() {
		if !committed {
			txn.Rollback()
		}
	}()

	for _, query := range stmts {
		query = strings.TrimSpace(query)
		if strings.HasSuffix(query, ";") &&
			!strings.HasSuffix(query, "END;") &&
			!strings.HasSuffix(query, "End;") &&
			!strings.HasSuffix(query, "end;") {
			query = strings.TrimSuffix(query, ";")
		}
		if _, err = txn.Exec(query); err != nil {
			fmt.Println("Executing:", query)
			return fmt.Errorf("FAIL %s (%v), quitting migration.", filepath.Base(script), err)
		}
	}

	if err := dialect.InsertVersionSql(ctx, txn, v, direction); err != nil {
		return fmt.Errorf("insert version row: %w", err)
	}

	committed = true
	if err := txn.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
