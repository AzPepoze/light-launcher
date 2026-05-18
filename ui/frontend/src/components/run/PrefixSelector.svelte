<script lang="ts">
	import Dropdown from "@components/shared/Dropdown.svelte";
	
	export let availablePrefixes: string[] = [];
	export let selectedPrefixName: string = "Default";
	export let prefixPath: string = "";
	export let baseDir: string = "";
	export let onPrefixChange: (name: string) => Promise<void>;
	export let onBrowsePrefix: () => Promise<void>;
</script>

<div class="form-group">
	<label for="winePrefix">WINEPREFIX</label>
	<div class="input-group">
		<div class="dropdown-wrapper">
			<Dropdown
				options={[...availablePrefixes, "Custom..."]}
				bind:value={selectedPrefixName}
				onChange={onPrefixChange}
			/>
		</div>
		<button on:click={onBrowsePrefix} class="btn">Browse</button>
	</div>
	{#if selectedPrefixName === "Custom..." || !prefixPath.startsWith(baseDir)}
		<div class="path-display">{prefixPath}</div>
	{/if}
</div>

<style lang="scss">
	.form-group label {
		display: block;
		font-size: 0.9rem;
		font-weight: 800;
		color: var(--accent-primary);
		text-transform: uppercase;
		letter-spacing: 1px;
		margin-bottom: 10px;
	}
	.input-group {
		display: flex;
		gap: 12px;
		width: 100%;
		.dropdown-wrapper {
			flex: 1;
		}
	}
	.path-display {
		margin-top: 10px;
		font-size: 0.8rem;
		color: var(--text-muted);
		word-break: break-all;
		padding: 10px 16px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
	}
	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 12px 24px;
		border-radius: var(--radius-pill);
		font-weight: 700;
		font-size: 0.9rem;
		cursor: pointer;
		transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
		border: 2px solid rgba(255, 255, 255, 0.05);
		background: var(--bg-surface);
		color: var(--text-main);

		&:hover {
			background: var(--bg-elevated);
			border-color: var(--accent-secondary);
			transform: scale(1.05);
		}

		&:active {
			transform: scale(0.95);
		}
	}
</style>
