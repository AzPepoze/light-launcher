import { rpc } from "./rpc";
import type {
	Game,
	GamePreview,
	IncomingListResponse,
	LibraryEditableFields,
	LibraryGetResponse,
	LibraryListResponse,
	SearchResponse,
	UpdatesResponse
} from "./types";

export const libraryApi = {
	list: () => rpc<LibraryListResponse>("LibraryList"),
	get: (id: number) => rpc<LibraryGetResponse>("LibraryGet", { id }),
	update: (id: number, fields: LibraryEditableFields) => rpc<Game>("LibraryUpdate", { id, fields }),
	remove: (id: number, deleteFiles: boolean) => rpc<void>("LibraryRemove", { id, deleteFiles }),
	search: (query: string) => rpc<SearchResponse>("LibrarySearch", { query }),
	preview: (url: string) => rpc<GamePreview>("LibraryPreview", { url }),
	addFromUrl: (url: string) => rpc<Game>("LibraryAddFromUrl", { url }),
	linkLocal: (id: number, path: string) => rpc<Game>("LibraryLinkLocal", { id, path }),
	checkUpdates: (force: boolean) => rpc<UpdatesResponse>("LibraryCheckUpdates", { force }),
	setFavorite: (id: number, favorite: boolean) => rpc<Game>("LibrarySetFavorite", { id, favorite }),
	launch: (id: number) => rpc<void>("LibraryLaunch", { id }),
	incomingList: () => rpc<IncomingListResponse>("IncomingList"),
	incomingLink: (name: string, gameId: number) => rpc<void>("IncomingLink", { name, gameId }),
	incomingIgnore: (name: string) => rpc<void>("IncomingIgnore", { name }),
	incomingCreateLocal: (name: string) => rpc<Game>("IncomingCreateLocal", { name })
};
