package contract

// SidebarGroup is one filter group a source asks the Library sidebar to show.
type SidebarGroup struct {
	Type        string          `json:"type"`
	ID          string          `json:"id"`
	Label       string          `json:"label,omitempty"`
	Placeholder string          `json:"placeholder,omitempty"`
	Options     []SidebarOption `json:"options,omitempty"`
	Dynamic     bool            `json:"dynamic,omitempty"`
	Searchable  bool            `json:"searchable,omitempty"`
	Orderable   bool            `json:"orderable,omitempty"`
	Default     *SortDefault    `json:"default,omitempty"`
}

// SidebarOption is one choice of a toggles, select, multi or sort group.
type SidebarOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

// SortDefault is the choice and direction a sort group starts with.
type SortDefault struct {
	Value string `json:"value"`
	Order string `json:"order"`
}
