const ERROR_CODE = /\s*\((?:code|error)\s*-?\d+\)/gi;
const URL = /\bhttps?:\/\/\S+/gi;
const RPC_PREFIX = /^Error invoking remote method '[^']*':\s*/i;
const ERROR_PREFIX = /^(?:uncaught\s+)?error:\s*/i;

/** Turns an error or message into one plain sentence for a notice. */
export function plainMessage(err: unknown, fallback = "Something went wrong"): string {
	const firstLine = rawText(err).split("\n")[0] ?? "";
	const plain = firstLine
		.replace(ERROR_CODE, "")
		.replace(URL, "")
		.replace(RPC_PREFIX, "")
		.replace(ERROR_PREFIX, "")
		.replace(/\s+/g, " ")
		.trim();
	return plain || fallback;
}

function rawText(err: unknown): string {
	if (typeof err === "string") return err;
	if (err instanceof Error) return err.message;
	if (typeof err === "object" && err !== null && "message" in err) {
		const message = (err as { message: unknown }).message;
		if (typeof message === "string") return message;
	}
	return String(err ?? "");
}
