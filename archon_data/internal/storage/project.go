package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Helper functions
func (p *Project) validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 255 {
		return fmt.Errorf("name cannot be longer than 255 characters")
	}
	return nil
}

func (p *Project) validateDescription(description string) error {
	if len(description) > 1000 {
		return fmt.Errorf("description cannot be longer than 1000 characters")
	}
	return nil
}

func (p *Project) validateParentID(parentID *int64) error {
	if parentID != nil {
		// Check if parent exists
		var exists bool
		err := p.db.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", *parentID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check parent existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("parent project with ID %d does not exist", *parentID)
		}

		// Check for circular references
		if err := p.checkCircularReference(*parentID, p.ID); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) checkCircularReference(parentID, currentID int64) error {
	if parentID == currentID {
		return fmt.Errorf("circular reference detected: project cannot be its own parent")
	}

	// Get all ancestors of the parent
	ancestors := make(map[int64]bool)
	current := parentID
	for current != 0 {
		var nextParentID sql.NullInt64
		err := p.db.QueryRow("SELECT parent_id FROM projects WHERE id = $1", current).Scan(&nextParentID)
		if err != nil {
			return fmt.Errorf("failed to check ancestors: %w", err)
		}
		if !nextParentID.Valid {
			break
		}
		current = nextParentID.Int64
		if current == currentID {
			return fmt.Errorf("circular reference detected: would create a cycle in the project hierarchy")
		}
		ancestors[current] = true
	}

	return nil
}

func (p *Project) validateStatus(status string) error {
	validStatuses := map[string]bool{
		"active":   true,
		"archived": true,
		"deleted":  true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}
	return nil
}

func (p *Project) validatePriority(priority int) error {
	if priority < 1 || priority > 5 {
		return fmt.Errorf("priority must be between 1 and 5")
	}
	return nil
}

func (p *Project) validateStartDate(startDate *time.Time) error {
	if startDate != nil && startDate.After(time.Now()) {
		return fmt.Errorf("start date cannot be in the future")
	}
	return nil
}

func (p *Project) validateEndDate(endDate *time.Time, startDate *time.Time) error {
	if endDate != nil {
		if endDate.Before(time.Now()) {
			return fmt.Errorf("end date cannot be in the past")
		}
		if startDate != nil && endDate.Before(*startDate) {
			return fmt.Errorf("end date cannot be before start date")
		}
	}
	return nil
}

func (p *Project) validateProgress(progress float64) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100")
	}
	return nil
}

func (p *Project) validateBudget(budget *float64) error {
	if budget != nil && *budget < 0 {
		return fmt.Errorf("budget cannot be negative")
	}
	return nil
}

func (p *Project) validateActualCost(actualCost *float64) error {
	if actualCost != nil && *actualCost < 0 {
		return fmt.Errorf("actual cost cannot be negative")
	}
	return nil
}

func (p *Project) validateTags(tags []string) error {
	if len(tags) > 10 {
		return fmt.Errorf("cannot have more than 10 tags")
	}
	for _, tag := range tags {
		if len(tag) > 50 {
			return fmt.Errorf("tag cannot be longer than 50 characters")
		}
	}
	return nil
}

func (p *Project) validateCustomFields(customFields map[string]interface{}) error {
	if len(customFields) > 20 {
		return fmt.Errorf("cannot have more than 20 custom fields")
	}
	for key, value := range customFields {
		if len(key) > 50 {
			return fmt.Errorf("custom field key cannot be longer than 50 characters")
		}
		// Convert value to string to check length
		strValue := fmt.Sprintf("%v", value)
		if len(strValue) > 500 {
			return fmt.Errorf("custom field value cannot be longer than 500 characters")
		}
	}
	return nil
}

// Create creates a new project
func (p *Project) Create(ctx context.Context) error {
	// Validate all fields
	if err := p.validateName(p.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}
	if err := p.validateDescription(p.Description); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}
	if err := p.validateParentID(p.ParentID); err != nil {
		return fmt.Errorf("invalid parent ID: %w", err)
	}
	if err := p.validateStatus(p.Status); err != nil {
		return fmt.Errorf("invalid status: %w", err)
	}
	if err := p.validatePriority(p.Priority); err != nil {
		return fmt.Errorf("invalid priority: %w", err)
	}
	if err := p.validateStartDate(p.StartDate); err != nil {
		return fmt.Errorf("invalid start date: %w", err)
	}
	if err := p.validateEndDate(p.EndDate, p.StartDate); err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}
	if err := p.validateProgress(p.Progress); err != nil {
		return fmt.Errorf("invalid progress: %w", err)
	}
	if err := p.validateBudget(p.Budget); err != nil {
		return fmt.Errorf("invalid budget: %w", err)
	}
	if err := p.validateActualCost(p.ActualCost); err != nil {
		return fmt.Errorf("invalid actual cost: %w", err)
	}
	if err := p.validateTags(p.Tags); err != nil {
		return fmt.Errorf("invalid tags: %w", err)
	}
	if err := p.validateCustomFields(p.CustomFields); err != nil {
		return fmt.Errorf("invalid custom fields: %w", err)
	}

	// Convert tags and custom fields to JSON
	tagsJSON, err := json.Marshal(p.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	customFieldsJSON, err := json.Marshal(p.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	// Insert project
	query := `
		INSERT INTO projects (
			name, description, parent_id, status, priority,
			start_date, end_date, progress, budget, actual_cost,
			tags, custom_fields, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id
	`

	now := time.Now()
	err = p.db.QueryRowContext(ctx, query,
		p.Name, p.Description, p.ParentID, p.Status, p.Priority,
		p.StartDate, p.EndDate, p.Progress, p.Budget, p.ActualCost,
		tagsJSON, customFieldsJSON, now, now,
	).Scan(&p.ID)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// Update updates an existing project
func (p *Project) Update(ctx context.Context) error {
	// Validate all fields
	if err := p.validateName(p.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}
	if err := p.validateDescription(p.Description); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}
	if err := p.validateParentID(p.ParentID); err != nil {
		return fmt.Errorf("invalid parent ID: %w", err)
	}
	if err := p.validateStatus(p.Status); err != nil {
		return fmt.Errorf("invalid status: %w", err)
	}
	if err := p.validatePriority(p.Priority); err != nil {
		return fmt.Errorf("invalid priority: %w", err)
	}
	if err := p.validateStartDate(p.StartDate); err != nil {
		return fmt.Errorf("invalid start date: %w", err)
	}
	if err := p.validateEndDate(p.EndDate, p.StartDate); err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}
	if err := p.validateProgress(p.Progress); err != nil {
		return fmt.Errorf("invalid progress: %w", err)
	}
	if err := p.validateBudget(p.Budget); err != nil {
		return fmt.Errorf("invalid budget: %w", err)
	}
	if err := p.validateActualCost(p.ActualCost); err != nil {
		return fmt.Errorf("invalid actual cost: %w", err)
	}
	if err := p.validateTags(p.Tags); err != nil {
		return fmt.Errorf("invalid tags: %w", err)
	}
	if err := p.validateCustomFields(p.CustomFields); err != nil {
		return fmt.Errorf("invalid custom fields: %w", err)
	}

	// Convert tags and custom fields to JSON
	tagsJSON, err := json.Marshal(p.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	customFieldsJSON, err := json.Marshal(p.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	// Update project
	query := `
		UPDATE projects SET
			name = $1,
			description = $2,
			parent_id = $3,
			status = $4,
			priority = $5,
			start_date = $6,
			end_date = $7,
			progress = $8,
			budget = $9,
			actual_cost = $10,
			tags = $11,
			custom_fields = $12,
			updated_at = $13
		WHERE id = $14
	`

	now := time.Now()
	result, err := p.db.ExecContext(ctx, query,
		p.Name, p.Description, p.ParentID, p.Status, p.Priority,
		p.StartDate, p.EndDate, p.Progress, p.Budget, p.ActualCost,
		tagsJSON, customFieldsJSON, now, p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("project with ID %d not found", p.ID)
	}

	return nil
}

// Delete deletes a project and its children
func (p *Project) Delete(ctx context.Context) error {
	// Start a transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, get all child projects
	var childIDs []int64
	rows, err := tx.QueryContext(ctx, "SELECT id FROM projects WHERE parent_id = $1", p.ID)
	if err != nil {
		return fmt.Errorf("failed to query child projects: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var childID int64
		if err := rows.Scan(&childID); err != nil {
			return fmt.Errorf("failed to scan child project ID: %w", err)
		}
		childIDs = append(childIDs, childID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating child project rows: %w", err)
	}

	// Delete all child projects first
	for _, childID := range childIDs {
		if _, err := tx.ExecContext(ctx, "DELETE FROM projects WHERE id = $1", childID); err != nil {
			return fmt.Errorf("failed to delete child project %d: %w", childID, err)
		}
	}

	// Delete the project itself
	result, err := tx.ExecContext(ctx, "DELETE FROM projects WHERE id = $1", p.ID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("project with ID %d not found", p.ID)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a project by its ID
func (p *Project) GetByID(ctx context.Context, id int64) error {
	query := `
		SELECT id, name, description, parent_id, status, priority,
			start_date, end_date, progress, budget, actual_cost,
			tags, custom_fields, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	var tagsJSON, customFieldsJSON []byte
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.ParentID, &p.Status, &p.Priority,
		&p.StartDate, &p.EndDate, &p.Progress, &p.Budget, &p.ActualCost,
		&tagsJSON, &customFieldsJSON, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("project with ID %d not found", id)
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(tagsJSON, &p.Tags); err != nil {
		return fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	if err := json.Unmarshal(customFieldsJSON, &p.CustomFields); err != nil {
		return fmt.Errorf("failed to unmarshal custom fields: %w", err)
	}

	return nil
}

// UpdateComponent updates a single component in the project while maintaining hierarchy integrity
func (p *Project) UpdateComponent(component *model.Component) error {
	// Load all components
	components, err := p.LoadComponents()
	if err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	// Create a component tree for easier hierarchy management
	tree, err := model.NewComponentTree(components)
	if err != nil {
		return fmt.Errorf("failed to create component tree: %w", err)
	}

	// Get the existing component to check for changes
	existingComp, err := tree.GetComponent(component.ID)
	if err != nil {
		return fmt.Errorf("component not found: %w", err)
	}

	// If parent is changing, validate the new parent-child relationship
	if existingComp.ParentID != component.ParentID {
		// Check if new parent exists
		if component.ParentID != "" {
			if _, err := tree.GetComponent(component.ParentID); err != nil {
				return fmt.Errorf("new parent component not found: %w", err)
			}
		}

		// Check for circular reference
		if err := tree.UpdateParent(component.ID, component.ParentID); err != nil {
			return fmt.Errorf("failed to update parent: %w", err)
		}
	}

	// Validate the updated component
	if err := component.Validate(); err != nil {
		return fmt.Errorf("invalid component: %w", err)
	}

	// Update the component in the tree
	tree.Components[component.ID] = component

	// Convert tree back to slice for saving
	updatedComponents := make([]*model.Component, 0, len(tree.Components))
	for _, comp := range tree.Components {
		updatedComponents = append(updatedComponents, comp)
	}

	// Save the updated components
	if err := p.SaveComponents(updatedComponents); err != nil {
		return fmt.Errorf("failed to save components: %w", err)
	}

	return nil
}

// ChangeType represents the type of change being made to a component
type ChangeType struct {
	Category    string            `json:"category"`     // e.g., "hardware", "software", "configuration", "maintenance"
	SubCategory string            `json:"subCategory"`  // e.g., "rack_equipment", "power_supply", "database"
	Action      string            `json:"action"`       // e.g., "replace", "upgrade", "reconfigure", "repair"
	Properties  map[string]string `json:"properties"`   // Additional metadata about the change type
}

// ComponentChange represents a change to a component
type ComponentChange struct {
	ID              string                 `json:"id"`               // Unique identifier for this change
	Timestamp       time.Time             `json:"timestamp"`        // When the change occurred
	ComponentID     string                `json:"componentId"`      // ID of the component being changed
	ChangeType      ChangeType            `json:"changeType"`       // Type of change
	Reason          string                `json:"reason"`           // Why the change was made
	Notes           string                `json:"notes"`            // Additional notes
	ChangedBy       string                `json:"changedBy"`        // Who made the change
	PreviousState   map[string]interface{} `json:"previousState"`   // Previous state of changed properties
	NewState        map[string]interface{} `json:"newState"`        // New state of changed properties
	RelatedChanges  []string              `json:"relatedChanges"`   // IDs of related changes
	ValidationRules []string              `json:"validationRules"`  // Rules that were applied
}

// ComponentChangeManager handles component changes and maintains change history
type ComponentChangeManager struct {
	project *Project
}

// NewComponentChangeManager creates a new component change manager
func NewComponentChangeManager(project *Project) *ComponentChangeManager {
	return &ComponentChangeManager{project: project}
}

// ApplyChange applies a change to a component
func (m *ComponentChangeManager) ApplyChange(change ComponentChange) error {
	// Load all components
	components, err := m.project.LoadComponents()
	if err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	// Create a component tree for hierarchy management
	tree, err := model.NewComponentTree(components)
	if err != nil {
		return fmt.Errorf("failed to create component tree: %w", err)
	}

	// Get the existing component
	component, err := tree.GetComponent(change.ComponentID)
	if err != nil {
		return fmt.Errorf("component not found: %w", err)
	}

	// Store previous state
	change.PreviousState = make(map[string]interface{})
	for k, v := range component.Properties {
		change.PreviousState[k] = v
	}

	// Apply validation rules based on change type
	if err := m.validateChange(change, component); err != nil {
		return fmt.Errorf("change validation failed: %w", err)
	}

	// Apply the change
	if err := m.applyChangeToComponent(change, component); err != nil {
		return fmt.Errorf("failed to apply change: %w", err)
	}

	// Store new state
	change.NewState = make(map[string]interface{})
	for k, v := range component.Properties {
		change.NewState[k] = v
	}

	// Update the component in the tree
	tree.Components[component.ID] = component

	// Convert tree back to slice for saving
	updatedComponents := make([]*model.Component, 0, len(tree.Components))
	for _, comp := range tree.Components {
		updatedComponents = append(updatedComponents, comp)
	}

	// Save the updated components
	if err := m.project.SaveComponents(updatedComponents); err != nil {
		return fmt.Errorf("failed to save components: %w", err)
	}

	// Save the change record
	if err := m.saveChangeRecord(change); err != nil {
		return fmt.Errorf("failed to save change record: %w", err)
	}

	return nil
}

// validateChange validates a change based on its type and component
func (m *ComponentChangeManager) validateChange(change ComponentChange, component *model.Component) error {
	// Load validation rules for this component type
	rules, err := m.loadValidationRules(component.Type, change.ChangeType)
	if err != nil {
		return fmt.Errorf("failed to load validation rules: %w", err)
	}

	// Apply each validation rule
	for _, rule := range rules {
		if err := m.applyValidationRule(rule, change, component); err != nil {
			return fmt.Errorf("validation rule failed: %w", err)
		}
	}

	return nil
}

// loadValidationRules loads validation rules for a component type and change type
func (m *ComponentChangeManager) loadValidationRules(componentType string, changeType ChangeType) ([]ValidationRule, error) {
	// First try to load specific rules
	rules, err := m.loadSpecificRules(componentType, changeType)
	if err == nil && len(rules) > 0 {
		return rules, nil
	}

	// If no specific rules found, try to load generic rules
	return m.loadGenericRules(componentType, changeType)
}

// loadSpecificRules loads specific validation rules for a component type
func (m *ComponentChangeManager) loadSpecificRules(componentType string, changeType ChangeType) ([]ValidationRule, error) {
	rulesPath := filepath.Join(m.project.Path, "rules", "components", componentType+".json")
	return m.loadRulesFromFile(rulesPath)
}

// loadGenericRules loads generic validation rules for a component type
func (m *ComponentChangeManager) loadGenericRules(componentType string, changeType ChangeType) ([]ValidationRule, error) {
	// Try to find a generic category for this component type
	genericType := m.getGenericType(componentType)
	rulesPath := filepath.Join(m.project.Path, "rules", "generic", genericType+".json")
	return m.loadRulesFromFile(rulesPath)
}

// getGenericType maps specific component types to generic categories
func (m *ComponentChangeManager) getGenericType(componentType string) string {
	// This mapping can be customized or loaded from a configuration file
	genericTypes := map[string]string{
		"power_supply":     "rack_equipment",
		"server":          "rack_equipment",
		"network_switch":  "rack_equipment",
		"database":        "software",
		"web_server":      "software",
		"load_balancer":   "software",
		"firewall":        "security_equipment",
		"ups":            "power_equipment",
	}

	if genericType, exists := genericTypes[componentType]; exists {
		return genericType
	}

	// If no mapping exists, return the original type
	return componentType
}

// ValidationRule represents a validation rule for component changes
type ValidationRule struct {
	Property    string      `json:"property"`
	Required    bool        `json:"required"`
	Type        string      `json:"type,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	Min         interface{} `json:"min,omitempty"`
	Max         interface{} `json:"max,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Dependencies []string   `json:"dependencies,omitempty"`
}

// applyValidationRule applies a single validation rule
func (m *ComponentChangeManager) applyValidationRule(rule ValidationRule, change ComponentChange, component *model.Component) error {
	// Check if property exists in new state
	value, exists := change.NewState[rule.Property]
	if !exists {
		if rule.Required {
			return fmt.Errorf("required property %s is missing", rule.Property)
		}
		return nil
	}

	// Validate based on type
	switch rule.Type {
	case "string":
		if strValue, ok := value.(string); ok {
			if rule.Pattern != "" {
				matched, err := regexp.MatchString(rule.Pattern, strValue)
				if err != nil {
					return fmt.Errorf("invalid pattern for property %s: %w", rule.Property, err)
				}
				if !matched {
					return fmt.Errorf("property %s does not match required pattern", rule.Property)
				}
			}
			if len(rule.Enum) > 0 {
				valid := false
				for _, enumValue := range rule.Enum {
					if strValue == enumValue {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("property %s must be one of: %v", rule.Property, rule.Enum)
				}
			}
		} else {
			return fmt.Errorf("property %s must be a string", rule.Property)
		}

	case "number":
		if numValue, ok := value.(float64); ok {
			if rule.Min != nil {
				if minValue, ok := rule.Min.(float64); ok && numValue < minValue {
					return fmt.Errorf("property %s must be greater than or equal to %v", rule.Property, minValue)
				}
			}
			if rule.Max != nil {
				if maxValue, ok := rule.Max.(float64); ok && numValue > maxValue {
					return fmt.Errorf("property %s must be less than or equal to %v", rule.Property, maxValue)
				}
			}
		} else {
			return fmt.Errorf("property %s must be a number", rule.Property)
		}
	}

	// Check dependencies
	for _, dep := range rule.Dependencies {
		if _, exists := change.NewState[dep]; !exists {
			return fmt.Errorf("property %s depends on %s which is missing", rule.Property, dep)
		}
	}

	return nil
}

// applyChangeToComponent applies the change to the component
func (m *ComponentChangeManager) applyChangeToComponent(change ComponentChange, component *model.Component) error {
	// Handle different change types
	switch change.ChangeType.Category {
	case "hardware":
		return m.applyHardwareChange(change, component)
	case "software":
		return m.applySoftwareChange(change, component)
	case "configuration":
		return m.applyConfigurationChange(change, component)
	case "maintenance":
		return m.applyMaintenanceChange(change, component)
	default:
		// For unknown categories, just apply the new state
		for k, v := range change.NewState {
			if component.Properties == nil {
				component.Properties = make(map[string]interface{})
			}
			component.Properties[k] = v
		}
	}
	return nil
}

// saveChangeRecord saves a change record to the history
func (m *ComponentChangeManager) saveChangeRecord(change ComponentChange) error {
	// Load existing history
	var history []ComponentChange
	historyPath := m.project.GetChangeHistoryPath()
	if data, err := os.ReadFile(historyPath); err == nil {
		if err := json.Unmarshal(data, &history); err != nil {
			return fmt.Errorf("failed to parse change history: %w", err)
		}
	}

	// Add new record
	history = append(history, change)

	// Save history
	if data, err := json.MarshalIndent(history, "", "  "); err == nil {
		if err := os.WriteFile(historyPath, data, 0644); err != nil {
			return fmt.Errorf("failed to save change history: %w", err)
		}
	}

	return nil
}

// GetChangeHistoryPath returns the path to the change history file
func (p *Project) GetChangeHistoryPath() string {
	return filepath.Join(p.Path, "change_history.json")
} 