import type { Game, SourceBadge, SourceCard } from "../../lib/features/types";
import type { TransferKind } from "./transferState";

/** A state shown on a cover corner. "Installed" is never shown: it is the normal state. */
export type TileState = "updateAvailable" | "notInstalled" | TransferKind | "favorite";

export interface TileCard {
	title: string;
	subtitle: string;
	/** Developer and engine, shown on hover when the game has them. */
	details: string[];
	/** The states to show, in corner order. */
	states: TileState[];
}

/** Without a card, a source gets the favorite and update states. */
const DEFAULT_BADGES: SourceBadge[] = ["favorite", "updateAvailable"];

/** The text of one game field. Only strings and numbers show; anything else is blank. */
function fieldText(game: Game, field: string): string {
	const value: unknown = (game as unknown as Record<string, unknown>)[field];
	return typeof value === "string" || typeof value === "number" ? String(value) : "";
}

function detailsOf(game: Game): string[] {
	return [game.developer, game.engine].filter((value): value is string => Boolean(value));
}

/**
 * The states for one game. A card can turn favorite and update off. Downloading, installing and
 * not installed come from the host, so every source shows them. Browse results are not in the
 * library yet, so they get no not-installed state.
 */
export function tileStates(game: Game, card: SourceCard | null, transfer: TransferKind | null): TileState[] {
	const badges = card ? (card.badges ?? []) : DEFAULT_BADGES;
	const states: TileState[] = [];
	if (badges.includes("updateAvailable") && game.updateAvailable) states.push("updateAvailable");
	if (transfer) states.push(transfer);
	else if (game.id > 0 && !game.installed) states.push("notInstalled");
	if (badges.includes("favorite") && game.favorite) states.push("favorite");
	return states;
}

/** What a tile shows. Title and version are the defaults for missing card fields. */
export function cardView(game: Game, card: SourceCard | null, transfer: TransferKind | null): TileCard {
	return {
		title: fieldText(game, card?.title ?? "title"),
		subtitle: fieldText(game, card?.subtitle ?? "version"),
		details: detailsOf(game),
		states: tileStates(game, card, transfer)
	};
}
