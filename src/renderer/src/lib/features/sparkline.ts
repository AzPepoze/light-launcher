export interface SparkPoint {
	x: number;
	y: number;
}

export interface SparkShape {
	/** Smooth line through the samples. */
	line: string;
	/** The line closed down to the baseline, for the fill. */
	area: string;
	/** The newest sample, or null when there are none. */
	last: SparkPoint | null;
}

/** Scales samples to a width by height box. The largest sample touches the top edge. */
export function sparkPoints(samples: number[], width: number, height: number): SparkPoint[] {
	const max = Math.max(0, ...samples);
	const lastIndex = samples.length - 1;
	return samples.map((value, index) => ({
		x: lastIndex === 0 ? width : (index / lastIndex) * width,
		y: max > 0 ? height - 1 - (value / max) * (height - 2) : height - 1
	}));
}

export function sparkShape(samples: number[], width: number, height: number): SparkShape {
	const points = sparkPoints(samples, width, height);
	if (points.length === 0) {
		const flat = `M0 ${height - 1} H${width}`;
		return { line: flat, area: `${flat} V${height} H0 Z`, last: null };
	}
	const [first, ...rest] = points;
	let line = `M${first.x} ${first.y}`;
	let previous = first;
	for (const point of rest) {
		const middleX = (previous.x + point.x) / 2;
		line += ` C${middleX} ${previous.y} ${middleX} ${point.y} ${point.x} ${point.y}`;
		previous = point;
	}
	const last = points[points.length - 1];
	return {
		line,
		area: `${line} V${height} H${first.x} Z`,
		last
	};
}
