import type { Partition, SetupEntry } from "@shared";

/** D: through Y:. Z: is never allowed. */
const AUTO_LETTERS = Array.from({ length: 22 }, (_, index) => `${String.fromCharCode(100 + index)}:`);

/** Every letter a setup row can use: C: and D: through Y:. */
export const DRIVE_LETTERS = ["c:", ...AUTO_LETTERS];

export function letterLabel(letter: string): string {
	return letter.toUpperCase();
}

export function letterOptions(letters: string[]): { value: string; label: string }[] {
	return letters.map((letter) => ({ value: letter, label: letterLabel(letter) }));
}

/** Letters a row may pick: its own letter plus every letter no other row uses. */
export function lettersAvailableTo(entries: SetupEntry[], current: string): string[] {
	const taken = entries.filter((entry) => entry.letter !== current).map((entry) => entry.letter);
	return DRIVE_LETTERS.filter((letter) => !taken.includes(letter));
}

/** The first free letter from D:, or an empty string when all are used. */
export function nextFreeLetter(entries: SetupEntry[]): string {
	return AUTO_LETTERS.find((letter) => !entries.some((entry) => entry.letter === letter)) ?? "";
}

/** The letter a disk mount point already has in the setup, or an empty string. */
export function letterForPath(entries: SetupEntry[], path: string): string {
	return entries.find((entry) => entry.path === path)?.letter ?? "";
}

export function usedBytes(partition: Partition): number {
	return Math.max(0, partition.totalBytes - partition.freeBytes);
}

/** Share of the disk in use, from 0 to 1. */
export function usedRatio(partition: Partition): number {
	if (partition.totalBytes <= 0) return 0;
	return usedBytes(partition) / partition.totalBytes;
}

export function cloneEntries(entries: SetupEntry[]): SetupEntry[] {
	return entries.map((entry) => ({ letter: entry.letter, path: entry.path }));
}

export function isSameSetup(a: SetupEntry[], b: SetupEntry[]): boolean {
	return (
		a.length === b.length &&
		a.every((entry, index) => entry.letter === b[index].letter && entry.path === b[index].path)
	);
}
