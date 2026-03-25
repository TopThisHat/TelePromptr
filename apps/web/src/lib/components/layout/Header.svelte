<script lang="ts">
	/**
	 * Header - Top application bar.
	 *
	 * Contains the mobile hamburger menu, breadcrumbs,
	 * theme toggle, and user actions area. Sits fixed at
	 * the top of the content area.
	 *
	 * @example
	 * ```svelte
	 * <Header onmobileopen={openMobileSidebar} />
	 * ```
	 */
	import { page } from '$app/state';
	import { theme } from '$lib/stores/theme.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import {
		MenuIcon,
		LightModeIcon,
		DarkModeIcon,
		ChevronRightIcon
	} from '$lib/components/ui/icons';

	interface HeaderProps {
		/** Callback to open the mobile sidebar overlay. */
		onmobileopen?: () => void;
	}

	let { onmobileopen }: HeaderProps = $props();

	/** Derive breadcrumbs from the current pathname. */
	let breadcrumbs = $derived.by(() => {
		const pathname = page.url.pathname;
		const segments = pathname.split('/').filter(Boolean);
		return segments.map((segment, index) => ({
			label: segment.charAt(0).toUpperCase() + segment.slice(1),
			href: '/' + segments.slice(0, index + 1).join('/'),
			current: index === segments.length - 1
		}));
	});
</script>

<header
	class="bg-surface-header border-surface-border z-[var(--z-header)] flex h-[var(--spacing-header)] shrink-0 items-center border-b px-4 backdrop-blur-sm"
>
	<!-- Mobile hamburger -->
	<button
		onclick={onmobileopen}
		class="text-surface-muted hover:text-surface-foreground mr-3 rounded-md p-1.5 lg:hidden"
		aria-label="Open navigation menu"
	>
		<Icon icon={MenuIcon} size="md" />
	</button>

	<!-- Breadcrumbs -->
	<nav class="flex flex-1 items-center gap-1 text-sm" aria-label="Breadcrumb">
		<ol class="flex items-center gap-1">
			{#each breadcrumbs as crumb, i (crumb.href)}
				<li class="flex items-center gap-1">
					{#if i > 0}
						<Icon
							icon={ChevronRightIcon}
							size="xs"
							class="text-surface-muted"
						/>
					{/if}
					{#if crumb.current}
						<span class="text-surface-foreground font-medium" aria-current="page">
							{crumb.label}
						</span>
					{:else}
						<a
							href={crumb.href}
							class="text-surface-muted hover:text-surface-foreground transition-colors"
						>
							{crumb.label}
						</a>
					{/if}
				</li>
			{/each}
		</ol>
	</nav>

	<!-- Actions -->
	<div class="flex items-center gap-1">
		<!-- Theme toggle -->
		<button
			onclick={() => theme.toggle()}
			class="text-surface-muted hover:bg-surface-border/50 hover:text-surface-foreground rounded-lg p-2 transition-colors"
			aria-label="Toggle {theme.resolved === 'dark' ? 'light' : 'dark'} mode"
		>
			{#if theme.resolved === 'dark'}
				<Icon icon={LightModeIcon} size="sm" />
			{:else}
				<Icon icon={DarkModeIcon} size="sm" />
			{/if}
		</button>
	</div>
</header>
