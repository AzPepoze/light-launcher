import type { WirePrefixConfigWithProton, WireProtonTool } from "../contract/wire.generated";

export type ProtonTool = WireProtonTool;

export type PrefixConfigWithProton = WirePrefixConfigWithProton;

export interface PrefixStats {
	name: string;
	createdAt: number | null;
	sizeBytes: number | null;
}
