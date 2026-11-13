/** A searchable dropdown gets a search box once it has more options than this. */
export const DROPDOWN_SEARCH_THRESHOLD = 8;

export interface SearchableOption {
	value: string;
	label: string;
}

/** Whether a dropdown shows its search box. */
export function showsDropdownSearch(optionCount: number, searchable: boolean): boolean {
	return searchable && optionCount > DROPDOWN_SEARCH_THRESHOLD;
}

/** Filters dropdown options by label or value, ignoring case and spaces at the ends of the query. */
export function filterDropdownOptions<T extends SearchableOption>(
	options: T[],
	query: string
): T[] {
	const needle = query.trim().toLowerCase();
	if (needle === "") return options;
	return options.filter(
		(option) =>
			option.label.toLowerCase().includes(needle) || option.value.toLowerCase().includes(needle)
	);
}
