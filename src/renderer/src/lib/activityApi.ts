import type { GameActivity } from "@shared";

export async function GetGameActivity(): Promise<GameActivity[]> {
	if (typeof window !== "undefined" && window.electronAPI) {
		return window.electronAPI.invoke<GameActivity[]>("GetGameActivity", {});
	}
	throw new Error("electronAPI is not available on window");
}
