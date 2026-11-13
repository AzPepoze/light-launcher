/** A group with more choices than this gets a search field, even when the group is not searchable. */
export const CHOICE_SEARCH_THRESHOLD = 8;

export interface FilterableChoice {
	value: string;
	label: string;
}

export interface ChoiceFilter<T> {
	/** Selected choices first, then the choices that match the query. */
	choices: T[];
	/** True when a query is typed and no unselected choice matches it. */
	noMatches: boolean;
}

/** Whether a choice group shows its search field. */
export function showsChoiceSearch(optionCount: number, searchable: boolean): boolean {
	return searchable || optionCount > CHOICE_SEARCH_THRESHOLD;
}

/**
 * Filters loaded choices by label, ignoring case and spaces at the ends of the query.
 * Selected choices stay at the top even when they do not match.
 */
export function filterChoices<T extends FilterableChoice>(
	choices: T[],
	query: string,
	isSelected: (value: string) => boolean
): ChoiceFilter<T> {
	const needle = query.trim().toLowerCase();
	const selected = choices.filter((choice) => isSelected(choice.value));
	const matching = choices.filter(
		(choice) =>
			!isSelected(choice.value) && (needle === "" || choice.label.toLowerCase().includes(needle))
	);
	return {
		choices: [...selected, ...matching],
		noMatches: needle !== "" && matching.length === 0
	};
}
