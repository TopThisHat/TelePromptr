<script lang="ts">
	/**
	 * Sidebar - Fixed left navigation panel.
	 *
	 * Contains the TelePromptr logo, main navigation links,
	 * and a collapse toggle. Adapts to desktop (expanded),
	 * tablet (collapsed icon-only), and mobile (overlay) modes.
	 *
	 * @example
	 * ```svelte
	 * <Sidebar collapsed={false} oncollapse={toggleSidebar} currentPath={$page.url.pathname} />
	 * ```
	 */
	import { page } from '$app/state';
	import NavItem from './NavItem.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import {
		TracesIcon,
		EvaluationsIcon,
		PromptsIcon,
		AnalyticsIcon,
		SettingsIcon,
		SidebarCollapseIcon,
		SidebarExpandIcon
	} from '$lib/components/ui/icons';

	interface SidebarProps {
		/** Whether the sidebar is collapsed to icon-only mode. */
		collapsed?: boolean;
		/** Callback fired when the collapse toggle is clicked. */
		oncollapse?: () => void;
	}

	let { collapsed = false, oncollapse }: SidebarProps = $props();

	/** Main navigation items (displayed in order). */
	const navItems = [
		{ href: '/traces', icon: TracesIcon, label: 'Traces' },
		{ href: '/evaluations', icon: EvaluationsIcon, label: 'Evaluations' },
		{ href: '/prompts', icon: PromptsIcon, label: 'Prompts' },
		{ href: '/analytics', icon: AnalyticsIcon, label: 'Analytics' }
	] as const;

	/** Bottom-grouped navigation items. */
	const bottomItems = [{ href: '/settings', icon: SettingsIcon, label: 'Settings' }] as const;

	let currentPath = $derived(page.url.pathname);

	/**
	 * Check if a nav item is active based on the current route.
	 * Matches exact path or any sub-path.
	 */
	function isActive(href: string): boolean {
		return currentPath === href || currentPath.startsWith(href + '/');
	}
</script>

<aside
	class="bg-surface-sidebar border-surface-border z-[var(--z-sidebar)] flex shrink-0 flex-col border-r
		transition-[width] duration-[var(--transition-slow)]
		{collapsed ? 'w-[var(--spacing-sidebar-collapsed)]' : 'w-[var(--spacing-sidebar)]'}"
	aria-label="Application sidebar"
>
	<!-- Logo / Brand -->
	<div
		class="border-surface-border flex h-[var(--spacing-header)] items-center border-b px-4
			{collapsed ? 'justify-center' : 'gap-3'}"
	>
		<div
			class="bg-primary flex h-8 w-8 shrink-0 items-center justify-center rounded-lg"
			aria-hidden="true"
		>
			<span class="text-primary-foreground text-sm font-bold">T</span>
		</div>
		{#if !collapsed}
			<span class="text-surface-sidebar-foreground truncate text-sm font-semibold tracking-tight">
				TelePromptr
			</span>
		{/if}
	</div>

	<!-- Main navigation -->
	<nav class="flex flex-1 flex-col gap-1 overflow-y-auto p-3" aria-label="Main navigation">
		<div class="flex flex-1 flex-col gap-1">
			{#each navItems as item (item.href)}
				<NavItem
					href={item.href}
					icon={item.icon}
					label={item.label}
					active={isActive(item.href)}
					{collapsed}
				/>
			{/each}
		</div>

		<!-- Bottom items (Settings, etc.) -->
		<div class="border-surface-border flex flex-col gap-1 border-t pt-3">
			{#each bottomItems as item (item.href)}
				<NavItem
					href={item.href}
					icon={item.icon}
					label={item.label}
					active={isActive(item.href)}
					{collapsed}
				/>
			{/each}
		</div>
	</nav>

	<!-- Collapse toggle -->
	<div class="border-surface-border border-t p-3">
		<button
			onclick={oncollapse}
			class="text-surface-muted hover:bg-surface-border/50 hover:text-surface-sidebar-foreground flex w-full items-center justify-center rounded-lg p-2 transition-colors"
			aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
		>
			<Icon icon={collapsed ? SidebarExpandIcon : SidebarCollapseIcon} size="sm" />
		</button>
	</div>
</aside>
