package config

type Severity string

const (
	// SeverityIgnore ignores validation failures
	SeverityIgnore Severity = "ignore"

	// SeverityWarning warns on validation failures
	SeverityWarning Severity = "warning"

	// SeverityDanger errors on validation failures
	SeverityDanger Severity = "danger"
)

func (severity *Severity) IsActionable() bool {
	return *severity == SeverityWarning || *severity == SeverityDanger
}
