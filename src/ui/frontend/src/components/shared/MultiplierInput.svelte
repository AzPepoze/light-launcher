<script lang="ts">
	import RangeSlider from "@components/shared/RangeSlider.svelte";

	export let value: string = "2";
	export let onChange: ((val: string) => void) | null = null;

	function handleChange(val: number) {
		const newVal = val.toString();
		value = newVal;
		onChange?.(newVal);
	}
</script>

<RangeSlider
	value={parseInt(value) || 2}
	min={2}
	max={4}
	inputMin={1}
	inputMax={100}
	unit="x"
	showWarning={false}
	snapValues={[2, 3, 4]}
	onChange={handleChange}
/>
