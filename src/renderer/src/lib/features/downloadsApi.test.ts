import { beforeEach, describe, expect, it, vi } from "vitest";
import { downloadsApi } from "./downloadsApi";
import { rpc } from "./rpc";

vi.mock("./rpc", () => ({ rpc: vi.fn() }));

const rpcMock = vi.mocked(rpc);

describe("downloadsApi.enqueueFile", () => {
	beforeEach(() => {
		rpcMock.mockReset();
	});

	it("sends the game id and the file id, never a link", async () => {
		rpcMock.mockResolvedValue({ id: 12 });

		await downloadsApi.enqueueFile(3, 9);

		expect(rpcMock).toHaveBeenCalledWith("LibraryDownloadFile", { gameId: 3, fileId: 9 });
	});

	it("returns the id of the queued download", async () => {
		rpcMock.mockResolvedValue({ id: 12 });

		await expect(downloadsApi.enqueueFile(3, 9)).resolves.toBe(12);
	});

	it("passes the host error on to the caller", async () => {
		rpcMock.mockRejectedValue(new Error("file not found"));

		await expect(downloadsApi.enqueueFile(3, 9)).rejects.toThrow("file not found");
	});
});
