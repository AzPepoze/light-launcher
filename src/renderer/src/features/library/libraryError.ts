import { errorMessage } from "../../lib/features/rpc";

const MISSING_METHOD = /method not found/i;
const MISSING_METHOD_TEXT = "This source does not support this yet.";

/** A short message for a failed library call. A plugin without the method gets plain words. */
export function libraryErrorMessage(err: unknown): string {
	const text = errorMessage(err);
	return MISSING_METHOD.test(text) ? MISSING_METHOD_TEXT : text;
}
