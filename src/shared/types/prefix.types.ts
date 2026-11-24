import type {
	WirePartition,
	WirePrefixConfigWithProton,
	WirePrefixDrive,
	WireProtonTool,
	WireSetupEntry
} from "../contract/wire.generated";

export type ProtonTool = WireProtonTool;

export type PrefixConfigWithProton = WirePrefixConfigWithProton;

export type Partition = WirePartition;

export type PrefixDrive = WirePrefixDrive;

export type SetupEntry = WireSetupEntry;

export interface PrefixStats {
	name: string;
	createdAt: number | null;
	sizeBytes: number | null;
}
