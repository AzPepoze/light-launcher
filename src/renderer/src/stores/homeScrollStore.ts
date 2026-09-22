let homeScrollTop = 0;

export function saveHomeScroll(scrollTop: number): void {
	homeScrollTop = Math.max(0, Math.round(scrollTop));
}

export function getHomeScroll(): number {
	return homeScrollTop;
}
