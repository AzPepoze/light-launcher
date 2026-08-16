<script lang="ts">
	import { tick } from "svelte";

	export let availablePrefixes: string[] = [];
	export let onCreate: (name: string) => void;
	export let onCancel: () => void;

	let newPrefixName = "";
	let createError = "";
	let nameInput: HTMLInputElement;

	const namePattern = /^[a-zA-Z0-9_-]+$/;

	function validateName(raw: string): string | null {
		const name = raw.trim();
		if (!name) return "Enter a prefix name";
		if (name === "Default") return '"Default" is reserved';
		if (!namePattern.test(name)) {
			return "Use only letters, numbers, hyphens, and underscores";
		}
		if (availablePrefixes.includes(name)) {
			return "A prefix with this name already exists";
		}
		return null;
	}

	function submitCreate() {
		const name = newPrefixName.trim();
		const error = validateName(name);
		if (error) {
			createError = error;
			return;
		}
		onCreate(name);
	}

	function handleCreateKeydown(e: KeyboardEvent) {
		if (e.key === "Enter") {
			e.preventDefault();
			submitCreate();
		}
		if (e.key === "Escape") {
			e.preventDefault();
			onCancel();
		}
	}

	$: if (newPrefixName) {
		const err = validateName(newPrefixName);
		createError = err && newPrefixName.trim() ? err : "";
	}

	$: canCreate =
		!!newPrefixName.trim() && validateName(newPrefixName) === null;

	import { onMount } from "svelte";
	onMount(() => {
		tick().then(() => nameInput?.focus());
	});
</script>

<div class="list-row create-row" role="form" aria-label="Create new prefix">
	<span class="material-icons row-icon">create_new_folder</span>
	<div class="create-fields">
		<input
			bind:this={nameInput}
			type="text"
			class="name-input"
			class:invalid={!!createError}
			placeholder="Prefix name…"
			bind:value={newPrefixName}
			on:keydown={handleCreateKeydown}
			autocomplete="off"
			spellcheck="false"
			aria-label="New prefix name"
		/>
		{#if createError}
			<p class="field-error" role="alert">{createError}</p>
		{/if}
	</div>
	<div class="create-btns">
		<button
			type="button"
			class="icon-action"
			title="Cancel"
			aria-label="Cancel"
			on:click={onCancel}
		>
			<span class="material-icons">close</span>
		</button>
		<button
			type="button"
			class="icon-action confirm"
			title="Create"
			aria-label="Create prefix"
			disabled={!canCreate}
			on:click={submitCreate}
		>
			<span class="material-icons">check</span>
		</button>
	</div>
</div>

<style lang="scss">
	.list-row {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		border-radius: var(--radius-md);
		box-sizing: border-box;
	}

	.create-row {
		padding: 10px 10px 10px 12px;
		background: var(--bg-input);
		border: 2px solid var(--accent-primary);
		align-items: flex-start;

		.row-icon {
			font-size: 20px;
			color: var(--accent-primary);
			margin-top: 8px;
			flex-shrink: 0;
		}
	}

	.create-fields {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.name-input {
		width: 100%;
		box-sizing: border-box;
		padding: 8px 10px;
		font-size: 0.85rem;
		font-weight: 700;
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-sm);
		color: var(--text-main);
		outline: none;
		transition: border-color var(--transition-fast);

		&::placeholder {
			color: var(--text-dim);
			font-weight: 600;
		}

		&:focus {
			border-color: var(--accent-primary);
		}

		&.invalid {
			border-color: var(--danger);
		}
	}

	.field-error {
		margin: 0;
		font-size: 0.68rem;
		font-weight: 700;
		color: var(--danger);
		line-height: 1.3;
	}

	.create-btns {
		display: flex;
		flex-direction: column;
		gap: 4px;
		flex-shrink: 0;
	}

	.icon-action {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		padding: 0;
		border: none;
		border-radius: var(--radius-sm);
		background: var(--bg-elevated);
		color: var(--text-muted);
		cursor: pointer;
		transition:
			background var(--transition-fast),
			color var(--transition-fast),
			transform var(--transition-spring);

		.material-icons {
			font-size: 18px;
		}

		&:hover {
			background: var(--bg-surface);
			color: var(--text-main);
		}

		&.confirm {
			background: var(--accent-primary);
			color: var(--bg-base);

			&:hover:not(:disabled) {
				background: var(--accent-hover);
			}
		}

		&:disabled {
			opacity: 0.4;
			cursor: not-allowed;
		}
	}
</style>
