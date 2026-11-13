import { describe, expect, it } from "vitest";
import type { Game, GameChange, GameFile } from "../../lib/features/types";
import {
	backendResult,
	checkedSummary,
	collapseUpToDate,
	emptyCounts,
	gamesWithUpdates,
	groupBySource,
	historyCover,
	historyStep,
	issueLevel,
	issueMessage,
	lastCheckedLabel,
	latestCheckByGame,
	liveHistoryItem,
	matchesFilter,
	newestChangelog,
	nextCheckLabel,
	pagingCursor,
	pickUpdateFile,
	plainMessage,
	progressLabel,
	progressValue,
	runMessage,
	skippedNote,
	tally,
	versionPair,
	visibleHistory,
	type UpdateCheckResult,
	type UpdateHistoryItem,
	type UpdateIssue
} from "./updatesView";

function sampleGame(overrides: Partial<Game> = {}): Game {
	return {
		id: 1,
		source: "sample",
		url: "https://example.com/game",
		title: "Sample Game",
		version: "1.2",
		developer: null,
		engine: null,
		tags: [],
		cover: null,
		coverLocal: null,
		screenshots: [],
		screenshotLocals: [],
		description: null,
		running: false,
		favorite: false,
		notes: null,
		installed: true,
		installedVersion: "1.0",
		installDir: null,
		installSize: null,
		exePath: null,
		exeArgs: null,
		playtimeS: 0,
		lastPlayedAt: null,
		addedAt: "2026-01-01T00:00:00Z",
		updateAvailable: true,
		origin: "plugin",
		sourceKind: "plugin",
		folderPath: null,
		scanFolder: null,
		localConfig: null,
		pinned: false,
		...overrides
	};
}

function sampleFile(id: number): GameFile {
	return {
		id,
		label: "example.zip",
		url: "https://example.com/example.zip",
		filename: "example.zip",
		size: null,
		host: null,
		kind: "archive",
		platform: "any",
		requiresAuth: false
	};
}

function sampleChange(overrides: Partial<GameChange>): GameChange {
	return {
		id: 1,
		checkedAt: "2026-01-01T00:00:00Z",
		from: null,
		to: null,
		changelog: null,
		seen: false,
		applied: false,
		...overrides
	};
}

function sampleItem(overrides: Partial<UpdateHistoryItem>): UpdateHistoryItem {
	return {
		id: 1,
		gameId: 1,
		title: "Sample Game",
		checkedAt: "2026-01-01T00:00:00Z",
		from: null,
		to: null,
		result: "current",
		error: null,
		coverPath: null,
		coverUrl: null,
		source: "sample",
		...overrides
	};
}

function sampleResult(overrides: Partial<UpdateCheckResult>): UpdateCheckResult {
	return {
		gameId: 1,
		title: "Sample Game",
		from: null,
		to: null,
		result: "current",
		error: null,
		coverPath: null,
		coverUrl: null,
		...overrides
	};
}

function sampleIssue(overrides: Partial<UpdateIssue>): UpdateIssue {
	return {
		source: "shop",
		label: "Sample Shop",
		kind: "error",
		message: "boom",
		games: 1,
		...overrides
	};
}

describe("gamesWithUpdates", () => {
	it("uses the singular for one game", () => {
		expect(gamesWithUpdates(1)).toBe("1 game has a newer release");
		expect(gamesWithUpdates(3)).toBe("3 games have a newer release");
	});
});

describe("lastCheckedLabel", () => {
	it("says when no check was ever saved", () => {
		expect(lastCheckedLabel(null)).toBe("Never checked");
	});

	it("shows how long ago the last saved check ran", () => {
		const now = Date.parse("2026-01-01T01:00:00Z");
		expect(lastCheckedLabel("2026-01-01T00:55:00Z", now)).toBe("Last checked 5m ago");
		expect(lastCheckedLabel("2026-01-01T00:59:40Z", now)).toBe("Last checked just now");
	});
});

describe("nextCheckLabel", () => {
	it("is empty when no check is scheduled", () => {
		expect(nextCheckLabel(null)).toBe("");
		expect(nextCheckLabel("")).toBe("");
	});

	it("is empty when the scheduled time cannot be read", () => {
		expect(nextCheckLabel("not a date")).toBe("");
	});

	it("says soon when the time already passed or is exactly now", () => {
		const now = Date.parse("2026-01-01T01:00:00Z");
		expect(nextCheckLabel("2026-01-01T00:59:00Z", now)).toBe("Next check soon");
		expect(nextCheckLabel("2026-01-01T01:00:00Z", now)).toBe("Next check soon");
	});

	it("counts the wait in minutes, hours and days", () => {
		const now = Date.parse("2026-01-01T01:00:00Z");
		expect(nextCheckLabel("2026-01-01T01:00:30Z", now)).toBe("Next check in under a minute");
		expect(nextCheckLabel("2026-01-01T01:05:00Z", now)).toBe("Next check in 5m");
		expect(nextCheckLabel("2026-01-01T04:00:00Z", now)).toBe("Next check in 3h");
		expect(nextCheckLabel("2026-01-03T01:00:00Z", now)).toBe("Next check in 2d");
	});

	it("rounds up so a wait never shows zero", () => {
		const now = Date.parse("2026-01-01T01:00:00Z");
		expect(nextCheckLabel("2026-01-01T01:00:01Z", now)).toBe("Next check in under a minute");
		expect(nextCheckLabel("2026-01-01T01:05:30Z", now)).toBe("Next check in 6m");
		expect(nextCheckLabel("2026-01-01T01:59:59Z", now)).toBe("Next check in 1h");
		expect(nextCheckLabel("2026-01-02T01:00:01Z", now)).toBe("Next check in 2d");
	});

	it("keeps an exact unit on that unit", () => {
		const now = Date.parse("2026-01-01T01:00:00Z");
		expect(nextCheckLabel("2026-01-01T01:01:00Z", now)).toBe("Next check in 1m");
		expect(nextCheckLabel("2026-01-01T02:00:00Z", now)).toBe("Next check in 1h");
		expect(nextCheckLabel("2026-01-02T01:00:00Z", now)).toBe("Next check in 1d");
	});
});

describe("pagingCursor", () => {
	it("is the oldest saved id, so the next page starts below it", () => {
		expect(
			pagingCursor([sampleItem({ id: 9 }), sampleItem({ id: 7 }), sampleItem({ id: 8 })])
		).toBe(7);
	});

	it("ignores live rows and starts from the top when nothing is saved", () => {
		expect(pagingCursor([sampleItem({ id: -1 })])).toBe(0);
		expect(pagingCursor([])).toBe(0);
	});
});

describe("progress", () => {
	it("is indeterminate until the total is known", () => {
		expect(progressValue(0, 0)).toBeNull();
		expect(progressLabel(0, 0)).toBe("Checking for updates");
	});

	it("is done over total with a count label", () => {
		expect(progressValue(12, 159)).toBeCloseTo(12 / 159);
		expect(progressLabel(12, 159)).toBe("Checking 12 of 159");
	});
});

describe("skippedNote", () => {
	it("is empty when nothing was skipped", () => {
		expect(skippedNote(0, [])).toBe("");
	});

	it("names the source that needs a key", () => {
		const issues = [sampleIssue({ kind: "auth", label: "Sample Shop" })];
		expect(skippedNote(12, issues)).toBe("Skipped 12 games (Sample Shop needs an API key)");
	});

	it("counts skipped games without a reason when no setup issue is known", () => {
		expect(skippedNote(1, [])).toBe("Skipped 1 game");
	});
});

describe("runMessage", () => {
	it("gives the run result in one line", () => {
		const counts = { ...emptyCounts(), updates: 3 };
		expect(runMessage(147, counts)).toBe("Checked 147 games, 3 updates");
	});

	it("adds failures and singular words", () => {
		expect(runMessage(1, { ...emptyCounts(), updates: 1, failed: 2 })).toBe(
			"Checked 1 game, 1 update, 2 failed"
		);
	});
});

describe("checkedSummary", () => {
	it("uses the singular for one game", () => {
		expect(checkedSummary(1)).toBe("Checked 1 game");
		expect(checkedSummary(5)).toBe("Checked 5 games");
	});
});

describe("issueMessage", () => {
	it("tells the person a source needs an API key and how many games were skipped", () => {
		const issue = sampleIssue({ kind: "auth", label: "Sample Shop", games: 12 });
		expect(issueMessage(issue)).toBe("Sample Shop needs an API key. 12 games were skipped.");
		expect(issueLevel(issue)).toBe("warning");
	});

	it("uses the plain message for other failures", () => {
		const issue = sampleIssue({ message: "Host unreachable (code -32603)", games: 1 });
		expect(issueMessage(issue)).toBe("Sample Shop: Host unreachable. 1 game was skipped.");
		expect(issueLevel(issue)).toBe("error");
	});
});

describe("plainMessage", () => {
	it("removes error codes", () => {
		expect(plainMessage("API key is not configured (code -32603)")).toBe(
			"API key is not configured"
		);
	});

	it("returns null for empty text", () => {
		expect(plainMessage(null)).toBeNull();
		expect(plainMessage("  (code -1)  ")).toBeNull();
	});
});

describe("tally", () => {
	it("counts each outcome", () => {
		let counts = emptyCounts();
		for (const outcome of ["update", "current", "error", "skipped", "skipped"] as const) {
			counts = tally(counts, sampleResult({ result: outcome }));
		}
		expect(counts).toEqual({ updates: 1, failed: 1, upToDate: 1, skipped: 2 });
	});
});

describe("backendResult", () => {
	it("maps each filter to the backend result, with an empty value for all checks", () => {
		expect(backendResult("attention")).toBe("");
		expect(backendResult("all")).toBe("");
		expect(backendResult("update")).toBe("update");
		expect(backendResult("error")).toBe("error");
		expect(backendResult("current")).toBe("current");
	});
});

describe("matchesFilter", () => {
	it("keeps skipped games out of the attention filter", () => {
		expect(matchesFilter("attention", "update")).toBe(true);
		expect(matchesFilter("attention", "error")).toBe(true);
		expect(matchesFilter("attention", "skipped")).toBe(false);
		expect(matchesFilter("all", "skipped")).toBe(true);
		expect(matchesFilter("current", "current")).toBe(true);
	});
});

describe("collapseUpToDate", () => {
	it("keeps the newest up-to-date check of each game and every other check", () => {
		const entries = [
			sampleItem({ id: 5, gameId: 1, result: "current" }),
			sampleItem({ id: 4, gameId: 1, result: "current" }),
			sampleItem({ id: 3, gameId: 2, result: "update" }),
			sampleItem({ id: 2, gameId: 2, result: "current" }),
			sampleItem({ id: 1, gameId: 2, result: "current" })
		];
		expect(collapseUpToDate(entries).map((entry) => entry.id)).toEqual([5, 3, 2]);
	});
});

describe("visibleHistory", () => {
	const entries = [
		sampleItem({ id: 4, gameId: 1, result: "current" }),
		sampleItem({ id: 3, gameId: 1, result: "current" }),
		sampleItem({ id: 2, gameId: 2, result: "error" }),
		sampleItem({ id: 1, gameId: 3, result: "skipped" })
	];

	it("shows updates and failures by default", () => {
		expect(visibleHistory(entries, "attention").map((entry) => entry.id)).toEqual([2]);
	});

	it("collapses repeated up-to-date checks under All", () => {
		expect(visibleHistory(entries, "all").map((entry) => entry.id)).toEqual([4, 2, 1]);
	});

	it("does not collapse under the Up to date filter", () => {
		expect(visibleHistory(entries, "current").map((entry) => entry.id)).toEqual([4, 3]);
	});
});

describe("latestCheckByGame", () => {
	it("keeps the newest check time of each game", () => {
		const latest = latestCheckByGame([
			sampleItem({ gameId: 1, checkedAt: "2026-03-01T10:00:00Z" }),
			sampleItem({ gameId: 1, checkedAt: "2026-02-01T10:00:00Z" }),
			sampleItem({ gameId: 2, checkedAt: "2026-01-01T10:00:00Z" })
		]);
		expect(latest.get(1)).toBe("2026-03-01T10:00:00Z");
		expect(latest.get(2)).toBe("2026-01-01T10:00:00Z");
	});
});

describe("groupBySource", () => {
	it("groups by label, with the most recent check first inside a group", () => {
		const games = [
			sampleGame({ id: 1, source: "beta", title: "Zeta" }),
			sampleGame({ id: 2, source: "alpha", title: "Bravo" }),
			sampleGame({ id: 3, source: "beta", title: "Alpha" }),
			sampleGame({ id: 4, source: "beta", title: "Mid" })
		];
		const lastChecked = new Map([
			[1, "2026-01-01T10:00:00Z"],
			[3, "2026-03-01T10:00:00Z"],
			[4, "2026-01-01T10:00:00Z"]
		]);
		const groups = groupBySource(
			games,
			{ alpha: "Alpha source", beta: "Beta source" },
			lastChecked
		);

		expect(groups.map((group) => group.label)).toEqual(["Alpha source", "Beta source"]);
		expect(groups[1].games.map((game) => game.title)).toEqual(["Alpha", "Mid", "Zeta"]);
	});

	it("falls back to the source id when no label is known", () => {
		const groups = groupBySource([sampleGame({ source: "gamma" })], {}, new Map());
		expect(groups[0].label).toBe("gamma");
	});

	it("returns no groups for no games", () => {
		expect(groupBySource([], {}, new Map())).toEqual([]);
	});
});

describe("versionPair", () => {
	it("shows old and new versions", () => {
		expect(versionPair(sampleGame())).toEqual({ from: "1.0", to: "1.2" });
	});

	it("fills unknown versions with words", () => {
		expect(versionPair(sampleGame({ installedVersion: null, version: null }))).toEqual({
			from: "unknown",
			to: "new"
		});
	});
});

describe("newestChangelog", () => {
	it("returns the changelog of the newest check that has one", () => {
		const changes = [
			sampleChange({ id: 1, checkedAt: "2026-01-01T00:00:00Z", changelog: "Old notes" }),
			sampleChange({ id: 2, checkedAt: "2026-02-01T00:00:00Z", changelog: "New notes" }),
			sampleChange({ id: 3, checkedAt: "2026-03-01T00:00:00Z", changelog: "   " })
		];
		expect(newestChangelog(changes)).toBe("New notes");
	});

	it("returns null when no change has notes", () => {
		expect(newestChangelog([sampleChange({ changelog: null })])).toBeNull();
		expect(newestChangelog([])).toBeNull();
	});
});

describe("historyStep", () => {
	it("maps each result to a step", () => {
		expect(historyStep(sampleItem({ result: "current" })).kind).toBe("checked");
		expect(historyStep(sampleItem({ result: "update" })).kind).toBe("updated");
		expect(historyStep(sampleItem({ result: "skipped" }))).toMatchObject({
			kind: "skipped",
			title: "Skipped"
		});
	});

	it("shows the error in plain words, without the code", () => {
		expect(
			historyStep(sampleItem({ result: "error", error: "Host unreachable (code -32603)" }))
		).toMatchObject({
			kind: "failed",
			icon: "error",
			title: "Host unreachable"
		});
	});

	it("uses a plain title when an error has no message", () => {
		expect(historyStep(sampleItem({ result: "error", error: null })).title).toBe("Check failed");
	});
});

describe("historyCover", () => {
	it("uses the remote cover URL of the game", () => {
		const cover = historyCover(
			sampleItem({ gameId: 9, coverUrl: "https://example.test/9.jpg", coverPath: "/a/9.jpg" })
		);
		expect(cover).toEqual({ id: 9, cover: "https://example.test/9.jpg", coverLocal: null });
	});
});

describe("liveHistoryItem", () => {
	it("turns a check result into a history row", () => {
		const row = liveHistoryItem(
			sampleResult({
				gameId: 4,
				title: "Other Game",
				from: "1.0",
				to: "2.0",
				result: "update",
				coverUrl: "u"
			}),
			-1,
			"2026-01-01T00:00:00Z"
		);
		expect(row).toMatchObject({
			id: -1,
			gameId: 4,
			title: "Other Game",
			result: "update",
			coverUrl: "u"
		});
	});
});

describe("pickUpdateFile", () => {
	it("picks the only file", () => {
		const file = sampleFile(7);
		expect(pickUpdateFile([file])).toEqual({ kind: "single", file });
	});

	it("reports no file or several files", () => {
		expect(pickUpdateFile([])).toEqual({ kind: "none" });
		expect(pickUpdateFile([sampleFile(1), sampleFile(2)])).toEqual({ kind: "many" });
	});
});
