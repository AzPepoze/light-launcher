export class ResourceHistoryTracker {
	private readonly maxPoints: number;
	cpu = $state<number[]>([]);
	ram = $state<number[]>([]);
	gpuMap = $state<Record<number, number[]>>({});

	constructor(maxPoints = 30) {
		this.maxPoints = maxPoints;
	}

	private clamp(val: number): number {
		return Number.isFinite(val) ? Math.min(100, Math.max(0, val)) : 0;
	}

	private updateBuffer(buffer: number[], val: number): number[] {
		if (buffer.length === 0) {
			return Array(this.maxPoints).fill(val);
		}
		return [...buffer.slice(-(this.maxPoints - 1)), val];
	}

	push(usage: { cpu: number; ram: number; gpus: number[] }) {
		this.cpu = this.updateBuffer(this.cpu, this.clamp(usage.cpu));
		this.ram = this.updateBuffer(this.ram, this.clamp(usage.ram));

		const newGpuMap: Record<number, number[]> = {};
		for (let i = 0; i < usage.gpus.length; i++) {
			const gpuVal = this.clamp(usage.gpus[i]);
			const existing = this.gpuMap[i] ?? [];
			newGpuMap[i] = this.updateBuffer(existing, gpuVal);
		}

		this.gpuMap = newGpuMap;
	}

	getGpuHistory(idx: number): number[] {
		return this.gpuMap[idx] ?? [];
	}

	clear() {
		this.cpu = [];
		this.ram = [];
		this.gpuMap = {};
	}
}
