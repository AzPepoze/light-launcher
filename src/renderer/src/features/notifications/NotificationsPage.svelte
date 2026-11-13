<script lang="ts">
	import { onMount } from "svelte";
	import Badge from "@components/shared/Badge.svelte";
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import Tabs from "@components/shared/Tabs.svelte";
	import {
		groupByDay,
		notifications,
		relativeTime,
		type DayGroup,
		type NotificationType
	} from "@stores/notificationStore";
	import {
		countNotifications,
		filterNotifications,
		filterTabs,
		isNotificationFilter,
		type NotificationFilter
	} from "./notificationFilter";

	const READ_DELAY_MS = 1_000;
	const CLOCK_REFRESH_MS = 30_000;

	const TYPE_ICONS: Record<NotificationType, string> = {
		error: "error",
		warning: "warning",
		success: "check_circle",
		info: "info"
	};

	const DAY_LABELS: Record<DayGroup, string> = {
		today: "Today",
		yesterday: "Yesterday",
		earlier: "Earlier"
	};

	interface EmptyCopy {
		icon: string;
		title: string;
		hint: string;
	}

	const EMPTY_COPY: Record<NotificationFilter, EmptyCopy> = {
		all: {
			icon: "notifications_off",
			title: "No notifications yet",
			hint: "Downloads, updates and plugin messages show up here."
		},
		unread: {
			icon: "mark_email_read",
			title: "Nothing unread",
			hint: "Every notice has been read."
		},
		errors: {
			icon: "check_circle",
			title: "No errors",
			hint: "Problems show up here when something fails."
		}
	};

	const history = notifications.history;
	const unreadCount = notifications.unreadCount;

	let filter = $state<NotificationFilter>("all");
	let now = $state(Date.now());

	const counts = $derived(countNotifications($history));
	const tabs = $derived(filterTabs(counts));
	const subtitle = $derived(pageSubtitle($history.length, $unreadCount));
	const entries = $derived(filterNotifications($history, filter));
	const groups = $derived(groupByDay(entries, now));
	const emptyCopy = $derived(EMPTY_COPY[filter]);

	onMount(() => {
		const readTimer = setTimeout(() => notifications.markAllRead(), READ_DELAY_MS);
		const clockTimer = setInterval(() => (now = Date.now()), CLOCK_REFRESH_MS);
		return () => {
			clearTimeout(readTimer);
			clearInterval(clockTimer);
		};
	});

	function noticeWord(total: number): string {
		return total === 1 ? "notice" : "notices";
	}

	function pageSubtitle(total: number, unread: number): string {
		if (total === 0) return "No notifications yet";
		if (unread === 0) return `${total} ${noticeWord(total)}, all read`;
		return `${total} ${noticeWord(total)}, ${unread} unread`;
	}

	function selectFilter(id: string) {
		if (isNotificationFilter(id)) filter = id;
	}

	function fullTime(time: number): string {
		return new Date(time).toLocaleString();
	}
</script>

<div class="notifications">
	<PageHeader title="Notifications" {subtitle} icon="notifications">
		<Button
			size="sm"
			icon="done_all"
			disabled={$unreadCount === 0}
			onclick={() => notifications.markAllRead()}
		>
			Mark all read
		</Button>
		<Button
			size="sm"
			icon="delete_sweep"
			disabled={$history.length === 0}
			onclick={() => notifications.clearHistory()}
		>
			Clear all
		</Button>
	</PageHeader>

	<div class="filters">
		<Tabs
			tabs={tabs}
			value={filter}
			onChange={selectFilter}
			aria-label="Notification filters"
			showZeroBadges
		/>
	</div>

	{#if entries.length === 0}
		<EmptyState icon={emptyCopy.icon} title={emptyCopy.title} hint={emptyCopy.hint} />
	{:else}
		<div class="groups">
			{#each groups as group (group.day)}
				<section class="group" aria-labelledby={`day-${group.day}`}>
					<h2 class="day" id={`day-${group.day}`}>{DAY_LABELS[group.day]}</h2>
					<ul>
						{#each group.entries as entry (entry.id)}
							<li class="row {entry.type}">
								<span class="type-icon">
									<Icon name={TYPE_ICONS[entry.type]} size={20} />
								</span>
								<div class="body">
									<p class="message">
										{entry.message}
										{#if entry.count > 1}<span class="count">x{entry.count}</span>{/if}
									</p>
									<div class="meta">
										<Badge>{entry.source ?? "Launcher"}</Badge>
										<span class="time" title={fullTime(entry.time)}>
											{relativeTime(entry.time, now)}
										</span>
									</div>
									{#if entry.action}
										<div class="item-action">
											<Button size="sm" onclick={() => entry.action?.run()}>
												{entry.action.label}
											</Button>
										</div>
									{/if}
								</div>
								{#if !entry.read}
									<span class="unread-dot" role="img" aria-label="Unread"></span>
								{/if}
								<IconButton
									icon="close"
									label="Remove notification"
									size={32}
									tone="danger"
									onclick={() => notifications.remove(entry.id)}
								/>
							</li>
						{/each}
					</ul>
				</section>
			{/each}
		</div>
	{/if}
</div>

<style lang="scss">
	.notifications {
		max-width: 900px;
	}

	.filters {
		margin-bottom: 20px;
	}

	.groups {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.group {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.day {
		margin: 0;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	ul {
		margin: 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.row {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 12px 14px;
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-md);
		background: var(--bg-surface);
		transition: border-color var(--transition-fast);

		&:hover {
			border-color: var(--glass-border-bright);
		}

		&.error .type-icon {
			color: var(--danger);
		}

		&.warning .type-icon {
			color: var(--warning);
		}

		&.success .type-icon {
			color: var(--success);
		}

		&.info .type-icon {
			color: var(--accent-primary);
		}
	}

	.type-icon {
		display: flex;
		padding-top: 1px;
	}

	.body {
		flex: 1;
		min-width: 0;
	}

	.message {
		margin: 0;
		color: var(--text-main);
		font-size: 0.9rem;
		line-height: 1.45;
		overflow-wrap: anywhere;
	}

	.count {
		margin-left: 6px;
		color: var(--text-muted);
		font-size: 0.8rem;
		font-weight: 700;
	}

	.meta {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: 6px;
	}

	.time {
		color: var(--text-muted);
		font-size: 0.75rem;
	}

	.item-action {
		margin-top: 10px;
	}

	.unread-dot {
		width: 8px;
		height: 8px;
		margin-top: 6px;
		border-radius: 50%;
		background: var(--accent-primary);
		flex-shrink: 0;
	}

	@media (prefers-reduced-motion: reduce) {
		.row {
			transition: none;
		}
	}
</style>
