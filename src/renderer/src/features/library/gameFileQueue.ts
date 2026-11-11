import { errorMessage } from "../../lib/features/rpc";
import type { GameFile } from "../../lib/features/types";

/** The button text while the host waits for the person to finish a browser step. */
export const PREPARING_TEXT = "Finish the step in the browser window";

/**
 * Queues one saved file. setPreparing gets the file id while the call runs, then null.
 * Returns the message to show under the file list, for success or failure.
 */
export async function queueGameFile(
	gameId: number,
	file: GameFile,
	enqueue: (gameId: number, fileId: number) => Promise<number>,
	setPreparing: (fileId: number | null) => void
): Promise<string> {
	setPreparing(file.id);
	try {
		await enqueue(gameId, file.id);
		return `Added "${file.label}" to downloads.`;
	} catch (err) {
		return errorMessage(err);
	} finally {
		setPreparing(null);
	}
}
