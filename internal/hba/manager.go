package hba

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConnectionType represents the type of connection in pg_hba.conf
type ConnectionType string

const (
	Local ConnectionType = "local"
	Host  ConnectionType = "host"
	HostSSL ConnectionType = "hostssl"
	HostNoSSL ConnectionType = "hostnossl"
)

// AuthMethod represents the authentication method
type AuthMethod string

const (
	Trust      AuthMethod = "trust"
	Reject     AuthMethod = "reject"
	MD5        AuthMethod = "md5"
	Password   AuthMethod = "password"
	ScramSHA256 AuthMethod = "scram-sha-256"
	Peer       AuthMethod = "peer"
	Ident      AuthMethod = "ident"
)

// Rule represents a single line in pg_hba.conf
type Rule struct {
	Type       ConnectionType
	Database   string
	User       string
	Address    string // For host/hostssl/hostnossl
	Method     AuthMethod
	Options    string // Additional options
	Comment    string
	IsComment  bool
	OriginalLine string
}

// Manager handles pg_hba.conf file operations
type Manager struct {
	filePath string
}

// NewManager creates a new HBA manager
func NewManager(filePath string) *Manager {
	return &Manager{
		filePath: filePath,
	}
}

// ReadRules reads and parses pg_hba.conf
func (m *Manager) ReadRules() ([]Rule, error) {
	file, err := os.Open(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pg_hba.conf: %w", err)
	}
	defer file.Close()

	var rules []Rule
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		rule := m.parseLine(line)
		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read pg_hba.conf: %w", err)
	}

	return rules, nil
}

// parseLine parses a single line from pg_hba.conf
func (m *Manager) parseLine(line string) Rule {
	originalLine := line
	line = strings.TrimSpace(line)

	// Handle empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return Rule{
			IsComment:    true,
			Comment:      line,
			OriginalLine: originalLine,
		}
	}

	// Remove inline comments
	commentIdx := strings.Index(line, "#")
	comment := ""
	if commentIdx > 0 {
		comment = strings.TrimSpace(line[commentIdx:])
		line = strings.TrimSpace(line[:commentIdx])
	}

	// Split into fields
	fields := strings.Fields(line)
	if len(fields) < 4 {
		// Invalid line, treat as comment
		return Rule{
			IsComment:    true,
			Comment:      originalLine,
			OriginalLine: originalLine,
		}
	}

	rule := Rule{
		Type:         ConnectionType(fields[0]),
		Database:     fields[1],
		User:         fields[2],
		Comment:      comment,
		OriginalLine: originalLine,
	}

	// Parse based on connection type
	if rule.Type == Local {
		// local DATABASE USER METHOD [OPTIONS]
		rule.Method = AuthMethod(fields[3])
		if len(fields) > 4 {
			rule.Options = strings.Join(fields[4:], " ")
		}
	} else {
		// host DATABASE USER ADDRESS METHOD [OPTIONS]
		if len(fields) >= 5 {
			rule.Address = fields[3]
			rule.Method = AuthMethod(fields[4])
			if len(fields) > 5 {
				rule.Options = strings.Join(fields[5:], " ")
			}
		}
	}

	return rule
}

// AddRule adds a new rule to pg_hba.conf
func (m *Manager) AddRule(rule Rule) error {
	// Read existing rules
	rules, err := m.ReadRules()
	if err != nil {
		return err
	}

	// Check if rule already exists
	for _, existingRule := range rules {
		if m.rulesMatch(existingRule, rule) {
			return fmt.Errorf("rule already exists")
		}
	}

	// Add new rule
	rules = append(rules, rule)

	// Write back to file
	return m.WriteRules(rules)
}

// RemoveRule removes a rule from pg_hba.conf
func (m *Manager) RemoveRule(rule Rule) error {
	rules, err := m.ReadRules()
	if err != nil {
		return err
	}

	// Filter out matching rules
	var newRules []Rule
	found := false
	for _, existingRule := range rules {
		if !m.rulesMatch(existingRule, rule) {
			newRules = append(newRules, existingRule)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("rule not found")
	}

	return m.WriteRules(newRules)
}

// WriteRules writes rules back to pg_hba.conf
func (m *Manager) WriteRules(rules []Rule) error {
	// Create backup
	backupPath := m.filePath + ".backup"
	if err := m.copyFile(m.filePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write new file
	file, err := os.Create(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to create pg_hba.conf: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, rule := range rules {
		line := m.formatRule(rule)
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write rule: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// formatRule formats a rule back to a line
func (m *Manager) formatRule(rule Rule) string {
	if rule.IsComment {
		return rule.OriginalLine
	}

	var parts []string
	parts = append(parts, string(rule.Type))
	parts = append(parts, rule.Database)
	parts = append(parts, rule.User)

	if rule.Type != Local {
		parts = append(parts, rule.Address)
	}

	parts = append(parts, string(rule.Method))

	if rule.Options != "" {
		parts = append(parts, rule.Options)
	}

	line := strings.Join(parts, "\t")

	if rule.Comment != "" {
		line += "\t" + rule.Comment
	}

	return line
}

// rulesMatch checks if two rules are equivalent
func (m *Manager) rulesMatch(r1, r2 Rule) bool {
	if r1.IsComment || r2.IsComment {
		return false
	}

	return r1.Type == r2.Type &&
		r1.Database == r2.Database &&
		r1.User == r2.User &&
		r1.Address == r2.Address &&
		r1.Method == r2.Method
}

// copyFile creates a copy of a file
func (m *Manager) copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// ReloadPostgreSQL reloads PostgreSQL configuration
func (m *Manager) ReloadPostgreSQL() error {
	// Try using SQL function first
	// This requires a database connection, which should be passed from caller
	return fmt.Errorf("reload must be called with database connection")
}
