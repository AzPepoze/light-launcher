import type { GameInfo } from "@shared";
import type { CardActions } from "../types";

export interface CardClickModifiers {
	shiftKey?: boolean;
	ctrlKey?: boolean;
	metaKey?: boolean;
}

/** In selection mode the click toggles selection; otherwise it launches the game. */
export function activateCard(
	game: GameInfo,
	isSelectionMode: boolean,
	event: CardClickModifiers | undefined,
	actions: CardActions
): void {
	if (isSelectionMode) {
		actions.onSelect(game, Boolean(event?.shiftKey), Boolean(event?.ctrlKey || event?.metaKey));
		return;
	}
	actions.onLaunch(game);
}

export function configureCard(game: GameInfo, actions: CardActions): void {
	actions.onConfigure(game);
}
