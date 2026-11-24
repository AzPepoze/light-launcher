/** True when the drag carries files from outside the page. */
export function carriesFiles(event: DragEvent): boolean {
	return event.dataTransfer?.types.includes("Files") ?? false;
}

/** Counts nested dragenter and dragleave events, so the highlight does not flicker over children. */
export class DragDepth {
	private depth = 0;

	enter(): void {
		this.depth += 1;
	}

	/** Returns true when the drag has left the element for good. */
	leave(): boolean {
		this.depth -= 1;
		if (this.depth > 0) return false;
		this.reset();
		return true;
	}

	reset(): void {
		this.depth = 0;
	}
}
