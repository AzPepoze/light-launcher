/** Pulls the dropped file paths out of either a raw array or a `{ data }` event envelope. */
export function droppedFiles(event: unknown): string[] {
	const payload = eventPayload(event);
	if (!Array.isArray(payload)) return [];
	return payload.filter((entry): entry is string => typeof entry === "string");
}

function eventPayload(event: unknown): unknown {
	if (Array.isArray(event)) return event;
	if (event === null || typeof event !== "object" || !("data" in event)) return undefined;
	return (event as { data?: unknown }).data;
}
