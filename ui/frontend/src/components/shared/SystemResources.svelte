<script lang="ts">
	export let sysInfo: {
		os: string;
		kernel: string;
		cpu: string;
		gpu: string;
		ram: string;
		driver: string;
	};
	export let sysUsage: {
		cpu: string;
		ram: string;
		gpu: string;
	};
</script>

<div class="status-grid">
	<!-- OS -->
	<div class="status-box">
		<div class="box-header">
			<div class="icon-label">
				<span class="material-icons mini-icon">laptop_windows</span>
				<span class="label">SYSTEM</span>
			</div>
		</div>
		<div class="system-info">
			<span class="os-text" title={sysInfo.os}>{sysInfo.os}</span>
			<span class="kernel-text">Kernel: {sysInfo.kernel}</span>
		</div>
	</div>

	<!-- CPU -->
	<div class="status-box">
		<div class="box-header">
			<div class="icon-label">
				<span class="material-icons mini-icon">memory</span>
				<span class="label">CPU</span>
			</div>
			<span class="usage">{sysUsage.cpu}</span>
		</div>
		<div class="progress-bg">
			<div class="progress-fill" style="width: {sysUsage.cpu}"></div>
		</div>
		<span class="info-text" title={sysInfo.cpu}>{sysInfo.cpu}</span>
	</div>

	<!-- RAM -->
	<div class="status-box">
		<div class="box-header">
			<div class="icon-label">
				<span class="material-icons mini-icon">storage</span>
				<span class="label">RAM</span>
			</div>
			<span class="usage">
				{sysUsage.ram.includes("(")
					? sysUsage.ram.split("(").pop().replace(")", "")
					: "0%"}
			</span>
		</div>
		<div class="progress-bg">
			<div
				class="progress-fill"
				style="width: {sysUsage.ram.includes('(')
					? sysUsage.ram.split('(').pop().replace(')', '')
					: '0%'}"
			></div>
		</div>
		<span class="info-text">{sysUsage.ram.split(" / ")[0]} used</span>
	</div>

	<!-- GPU -->
	<div class="status-box">
		<div class="box-header">
			<div class="icon-label">
				<span class="material-icons mini-icon">videogame_asset</span>
				<span class="label">GPU</span>
			</div>
			<span class="usage">{sysUsage.gpu}</span>
		</div>
		<div class="progress-bg">
			<div
				class="progress-fill"
				style="width: {sysUsage.gpu}; background: var(--accent-secondary, #b197fc)"
			></div>
		</div>
		<span class="info-text" title="{sysInfo.gpu} ({sysInfo.driver})">
			{sysInfo.gpu}
		</span>
	</div>
</div>

<style lang="scss">
	.status-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 20px;
	}

	.status-box {
		background: var(--bg-base);
		padding: 16px;
		border-radius: var(--radius-md);
		border: 2px solid var(--glass-border);
		display: flex;
		flex-direction: column;
		min-width: 0;
		transition: border-color var(--transition-fast);

		&:hover {
			border-color: var(--glass-border-bright);
		}

		.box-header {
			display: flex;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 12px;

			.icon-label {
				display: flex;
				align-items: center;
				gap: 6px;
				color: var(--text-muted);

				.mini-icon {
					font-size: 16px;
				}

				.label {
					font-size: 0.75rem;
					font-weight: 800;
					text-transform: uppercase;
					letter-spacing: 0.5px;
				}
			}

			.usage {
				font-size: 0.95rem;
				font-weight: 800;
				color: var(--accent-primary);
			}
		}

		.info-text {
			display: block;
			font-size: 0.75rem;
			font-weight: 700;
			color: var(--text-muted);
			margin-top: 10px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.system-info {
			display: flex;
			flex-direction: column;
			gap: 4px;
			overflow: hidden;

			.os-text {
				font-size: 0.85rem;
				font-weight: 800;
				color: var(--text-main);
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}

			.kernel-text {
				font-size: 0.7rem;
				font-weight: 600;
				color: var(--text-muted);
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}
	}

	.progress-bg {
		height: 6px;
		background: var(--bg-input);
		border-radius: var(--radius-pill);
		overflow: hidden;

		.progress-fill {
			height: 100%;
			background: var(--accent-primary);
			box-shadow: 0 0 8px var(--accent-primary);
			transition: width 0.3s ease;
		}
	}
</style>
