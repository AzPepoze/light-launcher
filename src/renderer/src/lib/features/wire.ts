// Payloads and replies of the Go API, as they arrive on the wire.
// The generated file is built from the Go structs by core/cmd/gentypes; field names
// match their json tags. Mappers turn these into the UI types in types.ts.

// Every generated interface is re-exported so callers keep importing wire types from
// this module. Do not redeclare a generated type here: change the Go struct and run
// go run ./core/cmd/gentypes instead.
export type * from "@shared/contract/wire.generated";

// The types below have no exported Go struct: the handlers use unexported payload
// structs or send events as maps. They stay handwritten and documented.

/** The payload of LibraryUpdate. Mirrors librarysvc.updatePayload. */
export interface WireGameUpdate {
	id: number;
	title?: string;
	notes?: string;
	tags?: string[];
	hidden?: boolean;
	favorite?: boolean;
	exePath?: string;
	exeArgs?: string;
	installDir?: string;
}

/** The payload of LibraryCheckUpdates. Mirrors librarysvc.checkPayload. */
export interface WireUpdateCheckRequest {
	ids?: number[];
	force?: boolean;
}

/** The reply of LibraryDownloadFile. Mirrors librarysvc.downloadFileResult. */
export interface WireQueuedDownload {
	id: number;
}

/** A removal event on plugin-state. The Go host emits it as a map, not a struct. */
export interface WireRemovedPlugin {
	id: string;
	removed: true;
}

/** A plugin-notify event. The Go host emits it as a map, not a struct. */
export interface WirePluginNotify {
	plugin: string;
	level: string;
	message: string;
}
