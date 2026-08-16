import type { LogColor, LogLevel } from "./types";

export const PASTEL_PALETTE: LogColor[] = [
	{
		name: "pastel-pink",
		ansi: "\x1b[38;2;255;179;186m",
		hex: "#ffb3ba",
		cssBg: "rgba(255, 179, 186, 0.25)",
		cssBorder: "rgba(255, 179, 186, 0.5)"
	},
	{
		name: "pastel-peach",
		ansi: "\x1b[38;2;255;223;186m",
		hex: "#ffdfba",
		cssBg: "rgba(255, 223, 186, 0.25)",
		cssBorder: "rgba(255, 223, 186, 0.5)"
	},
	{
		name: "pastel-buttercup",
		ansi: "\x1b[38;2;253;255;182m",
		hex: "#fdffb6",
		cssBg: "rgba(253, 255, 182, 0.25)",
		cssBorder: "rgba(253, 255, 182, 0.5)"
	},
	{
		name: "pastel-mint",
		ansi: "\x1b[38;2;186;255;201m",
		hex: "#baffc9",
		cssBg: "rgba(186, 255, 201, 0.25)",
		cssBorder: "rgba(186, 255, 201, 0.5)"
	},
	{
		name: "pastel-sky",
		ansi: "\x1b[38;2;186;225;255m",
		hex: "#bae1ff",
		cssBg: "rgba(186, 225, 255, 0.25)",
		cssBorder: "rgba(186, 225, 255, 0.5)"
	},
	{
		name: "pastel-lavender",
		ansi: "\x1b[38;2;232;223;245m",
		hex: "#e8dff5",
		cssBg: "rgba(232, 223, 245, 0.25)",
		cssBorder: "rgba(232, 223, 245, 0.5)"
	},
	{
		name: "pastel-seafoam",
		ansi: "\x1b[38;2;181;234;215m",
		hex: "#b5ead7",
		cssBg: "rgba(181, 234, 215, 0.25)",
		cssBorder: "rgba(181, 234, 215, 0.5)"
	},
	{
		name: "pastel-rose",
		ansi: "\x1b[38;2;252;225;228m",
		hex: "#fce1e4",
		cssBg: "rgba(252, 225, 228, 0.25)",
		cssBorder: "rgba(252, 225, 228, 0.5)"
	},
	{
		name: "pastel-violet",
		ansi: "\x1b[38;2;189;178;255m",
		hex: "#bdb2ff",
		cssBg: "rgba(189, 178, 255, 0.25)",
		cssBorder: "rgba(189, 178, 255, 0.5)"
	},
	{
		name: "pastel-magenta",
		ansi: "\x1b[38;2;255;198;255m",
		hex: "#ffc6ff",
		cssBg: "rgba(255, 198, 255, 0.25)",
		cssBorder: "rgba(255, 198, 255, 0.5)"
	},
	{
		name: "pastel-aqua",
		ansi: "\x1b[38;2;155;246;255m",
		hex: "#9bf6ff",
		cssBg: "rgba(155, 246, 255, 0.25)",
		cssBorder: "rgba(155, 246, 255, 0.5)"
	},
	{
		name: "pastel-sage",
		ansi: "\x1b[38;2;202;255;191m",
		hex: "#caffbf",
		cssBg: "rgba(202, 255, 191, 0.25)",
		cssBorder: "rgba(202, 255, 191, 0.5)"
	},
	{
		name: "pastel-periwinkle",
		ansi: "\x1b[38;2;160;196;255m",
		hex: "#a0c4ff",
		cssBg: "rgba(160, 196, 255, 0.25)",
		cssBorder: "rgba(160, 196, 255, 0.5)"
	}
];

export const ANSI_RESET = "\x1b[0m";
export const ANSI_DIM = "\x1b[2m";
export const ANSI_BOLD = "\x1b[1m";

export const LEVEL_COLORS: Record<LogLevel, { ansi: string; hex: string; bg: string }> = {
	debug: {
		ansi: "\x1b[38;2;176;190;197m",
		hex: "#b0bec5",
		bg: "rgba(176, 190, 197, 0.2)"
	},
	info: {
		ansi: "\x1b[38;2;129;212;250m",
		hex: "#81d4fa",
		bg: "rgba(129, 212, 250, 0.2)"
	},
	warn: {
		ansi: "\x1b[38;2;255;224;130m",
		hex: "#ffe082",
		bg: "rgba(255, 224, 130, 0.2)"
	},
	error: {
		ansi: "\x1b[38;2;239;154;154m",
		hex: "#ef9a9a",
		bg: "rgba(239, 154, 154, 0.2)"
	}
};

export function hashString(str: string): number {
	let hash = 5381;
	const clean = str.trim().toLowerCase();
	for (let i = 0; i < clean.length; i++) {
		hash = ((hash << 5) + hash) ^ clean.charCodeAt(i);
	}
	return Math.abs(hash);
}

export function getCategoryColor(category: string): LogColor {
	if (!category || category.trim() === "") {
		return PASTEL_PALETTE[0];
	}
	const hash = hashString(category);
	const index = hash % PASTEL_PALETTE.length;
	return PASTEL_PALETTE[index];
}
