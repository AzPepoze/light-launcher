export class ResourceHistoryTracker {
	private maxPoints: number;
	cpu = $state<number[]>([]);
	ram = $state<number[]>([]);
	gpuMap = $state<Record<number, number[]>>({});

	constructor(maxPoints = 30) {
		this.maxPoints = maxPoints;
	}

	private parsePercent(val: string | undefined): number {
		if (!val) return 0;
		// Handle formats like "34%", "34.5%", "12.4 GB / 32 GB (39%)"
		if (val.includes("(") && val.includes("%)")) {
			const match = val.match(/\((\d+(?:\.\d+)?)%\)/);
			if (match) return parseFloat(match[1]);
		}
		const cleaned = val.replace(/[^0-9.]/g, "");
		const num = parseFloat(cleaned);
		return isNaN(num) ? 0 : Math.min(100, Math.max(0, num));
	}

	push(usage: { cpu: string; ram: string; gpu: string; gpus?: string[] }) {
		const cpuVal = this.parsePercent(usage.cpu);
		const ramVal = this.parsePercent(usage.ram);

		this.cpu = [...this.cpu.slice(-(this.maxPoints - 1)), cpuVal];
		this.ram = [...this.ram.slice(-(this.maxPoints - 1)), ramVal];

		// Per-GPU history
		const gpuList = usage.gpus && usage.gpus.length > 0 ? usage.gpus : [usage.gpu];
		const newMap: Record<number, number[]> = { ...this.gpuMap };

		for (let i = 0; i < gpuList.length; i++) {
			const gpuVal = this.parsePercent(gpuList[i]);
			const existing = newMap[i] || [];
			newMap[i] = [...existing.slice(-(this.maxPoints - 1)), gpuVal];
		}

		this.gpuMap = newMap;
	}

	getGpuHistory(idx: number): number[] {
		return this.gpuMap[idx] || (this.gpuMap[0] ? this.gpuMap[0] : []);
	}
}
