// Remembers the Home page scroll position while the user navigates away,
// so returning to Home restores the list where they left off.
let homeScrollTop = 0;

export function saveHomeScroll(scrollTop: number): void {
	homeScrollTop = Math.max(0, Math.round(scrollTop));
}

export function getHomeScroll(): number {
	return homeScrollTop;
}
