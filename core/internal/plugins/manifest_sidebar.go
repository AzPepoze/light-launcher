package plugins

import (
	"fmt"
	"slices"

	"light-launcher/core/internal/contract"
)

// capabilityBrowseFilters lets the host call source.browse and source.listFilters for a source.
const capabilityBrowseFilters = "browse:filters"

// Group types a sidebar can show.
const (
	groupSearch  = "search"
	groupToggles = "toggles"
	groupSelect  = "select"
	groupMulti   = "multi"
	groupSort    = "sort"
)

var sidebarGroupTypes = map[string]bool{
	groupSearch: true, groupToggles: true, groupSelect: true, groupMulti: true, groupSort: true,
}

// Sort directions a sort group can default to.
const (
	sortAsc  = "asc"
	sortDesc = "desc"
)

// validateSidebar checks the filter groups of one source. A sidebar needs the browse:filters capability.
func validateSidebar(field string, groups []contract.SidebarGroup, capabilities []string) error {
	if len(groups) == 0 {
		return nil
	}
	if !slices.Contains(capabilities, capabilityBrowseFilters) {
		return fieldError(field+".sidebar", "needs the %s capability", capabilityBrowseFilters)
	}
	seen := map[string]bool{}
	for i, group := range groups {
		groupField := fmt.Sprintf("%s.sidebar[%d]", field, i)
		if err := validateSidebarGroup(groupField, group); err != nil {
			return err
		}
		if seen[group.ID] {
			return fieldError(groupField+".id", "duplicate group id %q", group.ID)
		}
		seen[group.ID] = true
	}
	return nil
}

func validateSidebarGroup(field string, group contract.SidebarGroup) error {
	if !sidebarGroupTypes[group.Type] {
		return fieldError(field+".type", "unknown group type %q, use search, toggles, select, multi or sort", group.Type)
	}
	if !itemIDPattern.MatchString(group.ID) {
		return fieldError(field+".id", "must be lowercase letters, digits or dashes, starting with a letter")
	}
	if group.Default != nil && group.Type != groupSort {
		return fieldError(field+".default", "only a sort group can have a default")
	}
	switch group.Type {
	case groupSearch:
		return nil
	case groupToggles:
		return validateChoices(field, group.Options)
	case groupSort:
		return validateSortGroup(field, group)
	default:
		return validateChoiceGroup(field, group)
	}
}

// validateSortGroup checks the options and the optional default, which must name one of the options.
func validateSortGroup(field string, group contract.SidebarGroup) error {
	if err := validateChoices(field, group.Options); err != nil {
		return err
	}
	if group.Default == nil {
		return nil
	}
	if !hasOptionValue(group.Options, group.Default.Value) {
		return fieldError(field+".default.value", "%q is not one of the sort options", group.Default.Value)
	}
	if group.Default.Order != sortAsc && group.Default.Order != sortDesc {
		return fieldError(field+".default.order", "must be asc or desc, not %q", group.Default.Order)
	}
	return nil
}

// validateChoiceGroup checks a select or multi group, which has a label and either fixed options or dynamic ones.
func validateChoiceGroup(field string, group contract.SidebarGroup) error {
	if group.Label == "" {
		return fieldError(field+".label", "must not be empty")
	}
	if group.Dynamic && len(group.Options) > 0 {
		return fieldError(field+".options", "a dynamic group gets its options from the plugin, remove them")
	}
	if !group.Dynamic {
		return validateChoices(field, group.Options)
	}
	return nil
}

func hasOptionValue(options []contract.SidebarOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

// validateChoices checks that a fixed list has options, each with a unique value and a label.
func validateChoices(field string, options []contract.SidebarOption) error {
	if len(options) == 0 {
		return fieldError(field+".options", "must have at least one option")
	}
	seen := map[string]bool{}
	for i, option := range options {
		optionField := fmt.Sprintf("%s.options[%d]", field, i)
		if option.Value == "" {
			return fieldError(optionField+".value", "must not be empty")
		}
		if option.Label == "" {
			return fieldError(optionField+".label", "must not be empty")
		}
		if seen[option.Value] {
			return fieldError(optionField+".value", "duplicate option value %q", option.Value)
		}
		seen[option.Value] = true
	}
	return nil
}
