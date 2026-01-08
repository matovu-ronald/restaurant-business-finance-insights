package imports

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ParsedRow represents a parsed CSV row with mapped fields
type ParsedRow struct {
	LineNumber int
	Raw        map[string]string
	Mapped     map[string]interface{}
	Errors     []string
}

// ParseResult contains the results of parsing a CSV file
type ParseResult struct {
	Headers    []string
	Rows       []ParsedRow
	ValidRows  int
	ErrorRows  int
	TotalRows  int
	SourceType string
}

// Parser handles CSV parsing and validation
type Parser struct {
	sourceType string
	mapping    *MappingProfile
}

// NewParser creates a new CSV parser
func NewParser(sourceType string, mapping *MappingProfile) *Parser {
	return &Parser{
		sourceType: sourceType,
		mapping:    mapping,
	}
}

// Parse parses a CSV file using the configured mapping
func (p *Parser) Parse(reader io.Reader) (*ParseResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	// Read headers
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Clean headers
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	result := &ParseResult{
		Headers:    headers,
		SourceType: p.sourceType,
	}

	// Read all rows
	lineNum := 1
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			lineNum++
			continue
		}
		lineNum++

		row := p.parseRow(headers, record, lineNum)
		result.Rows = append(result.Rows, row)
		result.TotalRows++

		if len(row.Errors) == 0 {
			result.ValidRows++
		} else {
			result.ErrorRows++
		}
	}

	return result, nil
}

func (p *Parser) parseRow(headers []string, record []string, lineNum int) ParsedRow {
	row := ParsedRow{
		LineNumber: lineNum,
		Raw:        make(map[string]string),
		Mapped:     make(map[string]interface{}),
	}

	// Build raw map
	for i, header := range headers {
		if i < len(record) {
			row.Raw[header] = strings.TrimSpace(record[i])
		} else {
			row.Raw[header] = ""
		}
	}

	// Apply mapping
	if p.mapping != nil {
		for sourceCol, targetField := range p.mapping.ColumnMaps {
			if val, ok := row.Raw[sourceCol]; ok && val != "" {
				row.Mapped[targetField] = val
			}
		}

		// Apply defaults for missing fields
		for field, defaultVal := range p.mapping.Defaults {
			if _, exists := row.Mapped[field]; !exists {
				row.Mapped[field] = defaultVal
			}
		}
	}

	// Validate based on source type
	switch p.sourceType {
	case "pos":
		row.Errors = p.validatePOSRow(row)
	case "payroll":
		row.Errors = p.validatePayrollRow(row)
	case "inventory":
		row.Errors = p.validateInventoryRow(row)
	}

	return row
}

func (p *Parser) validatePOSRow(row ParsedRow) []string {
	var errs []string

	// Required fields for POS data
	requiredFields := []string{"date", "total"}
	for _, field := range requiredFields {
		if val, ok := row.Mapped[field]; !ok || val == "" {
			errs = append(errs, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Validate date format
	if dateStr, ok := row.Mapped["date"].(string); ok && dateStr != "" {
		if _, err := parseDate(dateStr); err != nil {
			errs = append(errs, fmt.Sprintf("invalid date format: %s", dateStr))
		}
	}

	// Validate numeric fields
	numericFields := []string{"total", "discounts", "comps", "tax"}
	for _, field := range numericFields {
		if val, ok := row.Mapped[field].(string); ok && val != "" {
			if _, err := parseAmount(val); err != nil {
				errs = append(errs, fmt.Sprintf("invalid numeric value for %s: %s", field, val))
			}
		}
	}

	return errs
}

func (p *Parser) validatePayrollRow(row ParsedRow) []string {
	var errs []string

	// Required fields for payroll data
	requiredFields := []string{"period_start", "period_end", "total_wages"}
	for _, field := range requiredFields {
		if val, ok := row.Mapped[field]; !ok || val == "" {
			errs = append(errs, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Validate date fields
	dateFields := []string{"period_start", "period_end"}
	for _, field := range dateFields {
		if dateStr, ok := row.Mapped[field].(string); ok && dateStr != "" {
			if _, err := parseDate(dateStr); err != nil {
				errs = append(errs, fmt.Sprintf("invalid date format for %s: %s", field, dateStr))
			}
		}
	}

	return errs
}

func (p *Parser) validateInventoryRow(row ParsedRow) []string {
	var errs []string

	// Required fields for inventory data
	requiredFields := []string{"snapshot_date", "item_name", "quantity", "unit_cost"}
	for _, field := range requiredFields {
		if val, ok := row.Mapped[field]; !ok || val == "" {
			errs = append(errs, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Validate numeric fields
	numericFields := []string{"quantity", "unit_cost"}
	for _, field := range numericFields {
		if val, ok := row.Mapped[field].(string); ok && val != "" {
			if _, err := parseAmount(val); err != nil {
				errs = append(errs, fmt.Sprintf("invalid numeric value for %s: %s", field, val))
			}
		}
	}

	return errs
}

// Helper functions for parsing

func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	s = strings.TrimSpace(s)
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unable to parse date")
}

func parseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}

func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	return strconv.Atoi(s)
}

func parseUUID(s string) (uuid.UUID, error) {
	s = strings.TrimSpace(s)
	return uuid.Parse(s)
}

// DefaultMappings returns default column mappings for each source type
func DefaultMappings() map[string]map[string]string {
	return map[string]map[string]string{
		"pos": {
			"Date":          "date",
			"Time":          "time",
			"Total":         "total",
			"Subtotal":      "subtotal",
			"Tax":           "tax",
			"Discounts":     "discounts",
			"Comps":         "comps",
			"Payment Method": "payment_method",
			"Channel":       "channel",
			"Server":        "server",
		},
		"payroll": {
			"Period Start":   "period_start",
			"Period End":     "period_end",
			"Employee":       "employee_name",
			"Hours Worked":   "hours_worked",
			"Hourly Rate":    "hourly_rate",
			"Total Wages":    "total_wages",
			"Super":          "superannuation",
			"Tax Withheld":   "tax_withheld",
		},
		"inventory": {
			"Snapshot Date": "snapshot_date",
			"Item Name":     "item_name",
			"Category":      "category",
			"Quantity":      "quantity",
			"Unit":          "unit",
			"Unit Cost":     "unit_cost",
			"Total Value":   "total_value",
		},
	}
}
