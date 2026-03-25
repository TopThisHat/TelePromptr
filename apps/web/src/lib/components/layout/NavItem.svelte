<script lang="ts">
	/**
	 * NavItem - Sidebar navigation link.
	 *
	 * Renders a navigation link with an icon, label, and active state
	 * indicator. When the sidebar is collapsed, only the icon is shown
	 * with a tooltip for the label.
	 *
	 * @example
	 * ```svelte
	 * <NavItem href="/traces" icon={Activity} label="Traces" active={true} collapsed={false} />
	 * ```
	 */
	import type { Component, SvelteComponent } from 'svelte';
	import Icon from '$lib/components/ui/Icon.svelte';

	/**
	 * Accepts both Svelte 5 function components and Svelte 4 class components
	 * (lucide-svelte currently exports SvelteComponentTyped classes).
	 */
	type AnyComponent = Component<any> | typeof SvelteComponent<any>;

	interface NavItemProps {
		/** Route path this item links to. */
		href: string;
		/** Lucide icon component. */
		icon: AnyComponent;
		/** Display label for the navigation item. */
		label: string;
		/** Whether this item represents the current route. */
		active?: boolean;
		/** Whether the sidebar is in collapsed (icon-only) mode. */
		collapsed?: boolean;
	}

	let { href, icon, label, active = false, collapsed = false }: NavItemProps = $props();
</script>

<a
	{href}
	aria-current={active ? 'page' : undefined}
	title={collapsed ? label : undefined}
	class="group relative flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium
		transition-colors duration-[var(--transition-base)]
		{collapsed ? 'justify-center' : ''}
		{active
		? 'bg-primary/10 text-primary dark:bg-primary/15 dark:text-primary'
		: 'text-surface-sidebar-foreground/70 hover:bg-surface-border/50 hover:text-surface-sidebar-foreground'}"
>
	<!-- Active indicator bar -->
	{#if active}
		<span
			class="bg-primary absolute top-1/2 -translate-y-1/2 rounded-full
				{collapsed ? '-left-1 h-6 w-1' : '-left-0.5 h-5 w-1'}"
			aria-hidden="true"
		></span>
	{/if}

	<Icon {icon} size="sm" />

	{#if !collapsed}
		<span class="truncate">{label}</span>
	{/if}
</a>
