import type { GameInfo } from "@shared";

export type CardView = "grid" | "list-grid";
export type HomeView = CardView | "sidebar-grid";

export interface CardActions {
	onLaunch: (game: GameInfo) => void;
	onConfigure: (game: GameInfo) => void;
	onSelect: (game: GameInfo, shiftKey: boolean, ctrlKey?: boolean) => void;
}
