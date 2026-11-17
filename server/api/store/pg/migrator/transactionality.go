package migrator

import (
	"fmt"
	"strings"
)

var nonTransactionalStatements = []string{
	"VACUUM",
	"ALTER SYSTEM",
	"CREATE INDEX CONCURRENTLY",
	"DROP INDEX CONCURRENTLY",
	"REINDEX INDEX CONCURRENTLY",
	"REINDEX TABLE CONCURRENTLY",
	"CREATE DATABASE",
	"DROP DATABASE",
}

func checkTransactionality(file string, chunks []string) error {
	transactional := strings.Contains(file, ".tx.")
	needsAutocommit := false

	for _, chunk := range chunks {
		statement := stripSQLComments(chunk)

		keyword := findNonTransactionalStatement(statement)
		if keyword == "" {
			continue
		}

		needsAutocommit = true

		if transactional {
			return fmt.Errorf("%s: %q cannot run inside a transaction, so the file must not carry the .tx. suffix", file, keyword)
		}

		if n := countStatements(statement); n > 1 {
			return fmt.Errorf("%s: %q shares its --bun:split chunk with %d other statement(s), which simple-protocol mode batches into an implicit transaction", file, keyword, n-1)
		}
	}

	if !transactional && !needsAutocommit {
		return fmt.Errorf("%s: no statement in it needs to run outside a transaction, so the file must carry the .tx. suffix", file)
	}

	return nil
}

func stripSQLComments(chunk string) string {
	lines := strings.Split(chunk, "\n")
	kept := make([]string, 0, len(lines))

	for _, line := range lines {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}

		kept = append(kept, line)
	}

	return strings.Join(kept, "\n")
}

func findNonTransactionalStatement(chunk string) string {
	for statement := range strings.SplitSeq(chunk, ";") {
		normalized := strings.Join(strings.Fields(strings.ToUpper(statement)), " ")

		for _, keyword := range nonTransactionalStatements {
			if strings.HasPrefix(normalized, keyword) {
				return keyword
			}
		}
	}

	return ""
}

func countStatements(statement string) int {
	count := 0

	for s := range strings.SplitSeq(statement, ";") {
		if strings.TrimSpace(s) != "" {
			count++
		}
	}

	return count
}
