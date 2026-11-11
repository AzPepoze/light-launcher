import type { Game } from "../../lib/features/types";
import { tileShapeFor, type TileShape } from "./sourceTileShape";

/** Source of games added by hand or found in folders. The server stores them as "manual". */
export const LOCAL_SOURCE = "manual";
const LOCAL_LABEL = "Local games";

export interface SourceLabel {
	source: string;
	label: string;
}

export interface GameGroup {
	source: string;
	label: string;
	shape: TileShape;
	games: Game[];
}

function groupBySource(games: Game[]): Map<string, Game[]> {
	const bySource = new Map<string, Game[]>();
	for (const game of games) {
		const list = bySource.get(game.source);
		if (list) {
			list.push(game);
		} else {
			bySource.set(game.source, [game]);
		}
	}
	return bySource;
}

function makeGroup(source: string, label: string, games: Game[]): GameGroup {
	return { source, label, shape: tileShapeFor(source), games };
}

/** Plugin sources in the order the server lists them, then any others found in the games. */
function pluginSourceIds(sources: SourceLabel[], games: Game[]): string[] {
	const ids = new Set([...sources.map((item) => item.source), ...games.map((game) => game.source)]);
	ids.delete(LOCAL_SOURCE);
	return [...ids];
}

/**
 * Splits games into "Local games" first, then one group per plugin source.
 * Keeps the order of the games inside each group. Empty groups are left out.
 */
export function groupGames(games: Game[], sources: SourceLabel[]): GameGroup[] {
	const bySource = groupBySource(games);
	const labels = new Map(sources.map((item) => [item.source, item.label]));
	const groups: GameGroup[] = [];

	const local = bySource.get(LOCAL_SOURCE);
	if (local) groups.push(makeGroup(LOCAL_SOURCE, LOCAL_LABEL, local));

	for (const source of pluginSourceIds(sources, games)) {
		const sourceGames = bySource.get(source);
		if (sourceGames) groups.push(makeGroup(source, labels.get(source) ?? source, sourceGames));
	}

	return groups;
}
