<script lang="ts">
	import Dropdown from "@components/shared/Dropdown.svelte";
	import BrowseInput from "@components/shared/BrowseInput.svelte";

	export let currentPrefixName: string;
	export let protonDisplayNames: string[];
	export let chosenProtonName: string;
	export let prefixPath: string;
</script>

<section class="panel hero-panel">
	<div class="hero-top">
		<div class="hero-title">
			<span class="material-icons hero-icon">folder</span>
			<div>
				<p class="hero-label">Active prefix</p>
				<h2 class="hero-name">{currentPrefixName}</h2>
			</div>
		</div>
		<div class="runtime-field">
			<label class="field-label" for="prefix-runtime">
				<span class="material-icons">science</span>
				Runtime
			</label>
			<div class="runtime-dropdown" id="prefix-runtime">
				<Dropdown
					options={protonDisplayNames}
					bind:value={chosenProtonName}
				/>
			</div>
		</div>
	</div>

	<div style="margin-top: 16px;">
		<BrowseInput
			bind:value={prefixPath}
			placeholder="Path to WINE prefix..."
			type="folder"
		/>
	</div>
</section>

<style lang="scss">
	.panel {
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-lg);
		padding: 22px 24px;
	}

	.hero-panel {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}

	.hero-top {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		justify-content: space-between;
		gap: 20px;
	}

	.hero-title {
		display: flex;
		align-items: center;
		gap: 14px;
		min-width: 0;

		.hero-icon {
			font-size: 36px;
			color: var(--accent-primary);
			flex-shrink: 0;
		}
	}

	.hero-label {
		margin: 0 0 2px;
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 1px;
		color: var(--text-muted);
	}

	.hero-name {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 800;
		color: var(--text-main);
		text-transform: uppercase;
		letter-spacing: 1px;
		line-height: 1.1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 320px;
	}

	.runtime-field {
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 220px;
		flex: 1;
		max-width: 360px;
	}

	.field-label {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 0.75rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--accent-primary);

		.material-icons {
			font-size: 16px;
		}
	}

	.runtime-dropdown {
		width: 100%;
	}

	@media (max-width: 900px) {
		.hero-name {
			max-width: 100%;
		}

		.runtime-field {
			max-width: 100%;
		}
	}
</style>
