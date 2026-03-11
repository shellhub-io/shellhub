package migrate

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

// FieldMismatch records a single field-level difference between Mongo and PG.
type FieldMismatch struct {
	Table    string
	RecordID string
	Field    string
	Expected string
	Actual   string
}

// TableReport tracks validation results for a single table.
type TableReport struct {
	Compared   int64
	Mismatches []FieldMismatch
	Missing    int64
}

// ValidationReport accumulates results across all validated tables.
type ValidationReport struct {
	TotalCompared int64
	TotalMismatch int64
	TotalMissing  int64
	PerTable      map[string]*TableReport
}

// NewValidationReport creates a new empty validation report.
func NewValidationReport() *ValidationReport {
	return &ValidationReport{
		PerTable: make(map[string]*TableReport),
	}
}

func (r *ValidationReport) getTable(name string) *TableReport {
	t, ok := r.PerTable[name]
	if !ok {
		t = &TableReport{}
		r.PerTable[name] = t
	}

	return t
}

// AddCompared records that n records were compared for the given table.
func (r *ValidationReport) AddCompared(table string, n int64) {
	r.TotalCompared += n
	r.getTable(table).Compared += n
}

// AddMissing records a record present in Mongo but not found in PG.
func (r *ValidationReport) AddMissing(table, recordID string) {
	r.TotalMissing++
	r.getTable(table).Missing++
	log.WithFields(log.Fields{"table": table, "id": recordID}).Error("Missing record in PG")
}

// AddMismatch records a field-level mismatch.
func (r *ValidationReport) AddMismatch(m FieldMismatch) {
	r.TotalMismatch++
	t := r.getTable(m.Table)
	t.Mismatches = append(t.Mismatches, m)
	log.WithFields(log.Fields{
		"table":    m.Table,
		"id":       m.RecordID,
		"field":    m.Field,
		"expected": m.Expected,
		"actual":   m.Actual,
	}).Error("Field mismatch")
}

// CheckField compares two values using fmt.Sprintf for stringification.
func (r *ValidationReport) CheckField(table, recordID, field string, expected, actual any) {
	e := fmt.Sprintf("%v", expected)
	a := fmt.Sprintf("%v", actual)
	if e != a {
		r.AddMismatch(FieldMismatch{Table: table, RecordID: recordID, Field: field, Expected: e, Actual: a})
	}
}

// CheckTime compares two time.Time values truncated to millisecond precision.
func (r *ValidationReport) CheckTime(table, recordID, field string, expected, actual time.Time) {
	e := expected.Truncate(time.Millisecond).UTC()
	a := actual.Truncate(time.Millisecond).UTC()
	if !e.Equal(a) {
		r.AddMismatch(FieldMismatch{Table: table, RecordID: recordID, Field: field, Expected: e.String(), Actual: a.String()})
	}
}

// CheckTimePtr compares two *time.Time values, treating nil and zero as equivalent.
func (r *ValidationReport) CheckTimePtr(table, recordID, field string, expected, actual *time.Time) {
	var e, a time.Time
	if expected != nil {
		e = expected.Truncate(time.Millisecond).UTC()
	}
	if actual != nil {
		a = actual.Truncate(time.Millisecond).UTC()
	}
	if !e.Equal(a) {
		r.AddMismatch(FieldMismatch{
			Table: table, RecordID: recordID, Field: field,
			Expected: e.String(), Actual: a.String(),
		})
	}
}

// CheckFloat compares two float64 values with epsilon tolerance.
func (r *ValidationReport) CheckFloat(table, recordID, field string, expected, actual float64) {
	if math.Abs(expected-actual) > 1e-9 {
		r.AddMismatch(FieldMismatch{
			Table: table, RecordID: recordID, Field: field,
			Expected: fmt.Sprintf("%v", expected), Actual: fmt.Sprintf("%v", actual),
		})
	}
}

// CheckStrings compares two string slices treating nil as empty and ignoring order.
func (r *ValidationReport) CheckStrings(table, recordID, field string, expected, actual []string) {
	e := normalizeStrings(expected)
	a := normalizeStrings(actual)
	if fmt.Sprintf("%v", e) != fmt.Sprintf("%v", a) {
		r.AddMismatch(FieldMismatch{
			Table: table, RecordID: recordID, Field: field,
			Expected: fmt.Sprintf("%v", e), Actual: fmt.Sprintf("%v", a),
		})
	}
}

// CheckStringMap compares two map[string]string values (order-independent).
func (r *ValidationReport) CheckStringMap(table, recordID, field string, expected, actual map[string]string) {
	if len(expected) != len(actual) {
		r.AddMismatch(FieldMismatch{
			Table: table, RecordID: recordID, Field: field,
			Expected: formatMap(expected), Actual: formatMap(actual),
		})

		return
	}

	for k, v := range expected {
		if actual[k] != v {
			r.AddMismatch(FieldMismatch{
				Table: table, RecordID: recordID, Field: field,
				Expected: formatMap(expected), Actual: formatMap(actual),
			})

			return
		}
	}
}

// HasErrors returns true if there are any mismatches or missing records.
func (r *ValidationReport) HasErrors() bool {
	return r.TotalMismatch > 0 || r.TotalMissing > 0
}

// Log prints a summary of the validation report.
func (r *ValidationReport) Log() {
	log.WithFields(log.Fields{
		"total_compared": r.TotalCompared,
		"total_mismatch": r.TotalMismatch,
		"total_missing":  r.TotalMissing,
	}).Info("Deep validation summary")

	for name, t := range r.PerTable {
		log.WithFields(log.Fields{
			"table":      name,
			"compared":   t.Compared,
			"mismatches": len(t.Mismatches),
			"missing":    t.Missing,
		}).Info("Table validation result")
	}
}

func formatMap(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, len(keys))
	for i, k := range keys {
		pairs[i] = fmt.Sprintf("%s:%s", k, m[k])
	}

	return fmt.Sprintf("map%v", pairs)
}

func normalizeStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	out := make([]string, len(s))
	copy(out, s)
	sort.Strings(out)

	return out
}

func (m *Migrator) deepValidate(ctx context.Context) error {
	report := NewValidationReport()

	validators := []struct {
		name string
		fn   func(ctx context.Context, r *ValidationReport) error
	}{
		{"systems", m.deepValidateSystems},
		{"users", m.deepValidateUsers},
		{"namespaces", m.deepValidateNamespaces},
		{"memberships", m.deepValidateMemberships},
		{"tags", m.deepValidateTags},
		{"devices", m.deepValidateDevices},
		{"device_tags", m.deepValidateDeviceTags},
		{"sessions", m.deepValidateSessions},
		{"public_keys", m.deepValidatePublicKeys},
		{"public_key_tags", m.deepValidatePublicKeyTags},
		{"api_keys", m.deepValidateAPIKeys},
	}

	for _, v := range validators {
		log.WithField("table", v.name).Info("Deep validating")
		if err := v.fn(ctx, report); err != nil {
			return fmt.Errorf("deep validation of %s failed: %w", v.name, err)
		}
	}

	report.Log()

	if report.HasErrors() {
		return fmt.Errorf("deep validation found %d mismatches and %d missing records",
			report.TotalMismatch, report.TotalMissing)
	}

	return nil
}
