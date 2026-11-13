/** Escape closes the game panel, unless another dialog is open on top of it. */
export function closesDetailOn(key: string, panelOpen: boolean, dialogOpen: boolean): boolean {
	return key === "Escape" && panelOpen && !dialogOpen;
}
