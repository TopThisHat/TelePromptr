<script lang="ts">
	import Sidebar from '$lib/components/layout/Sidebar.svelte';
	import Header from '$lib/components/layout/Header.svelte';

	let { children } = $props();

	/** Sidebar collapsed state (desktop/tablet). */
	let sidebarCollapsed = $state(false);

	/** Mobile sidebar overlay open state. */
	let mobileOpen = $state(false);

	function toggleSidebar(): void {
		sidebarCollapsed = !sidebarCollapsed;
	}

	function openMobile(): void {
		mobileOpen = true;
	}

	function closeMobile(): void {
		mobileOpen = false;
	}

	/** Close mobile overlay when Escape is pressed. */
	function handleKeydown(event: KeyboardEvent): void {
		if (event.key === 'Escape' && mobileOpen) {
			closeMobile();
		}
	}
</script>

<svelte:document onkeydown={handleKeydown} />

<div class="flex h-screen overflow-hidden">
	<!-- Desktop / Tablet sidebar (hidden on mobile) -->
	<div class="hidden lg:flex">
		<Sidebar collapsed={sidebarCollapsed} oncollapse={toggleSidebar} />
	</div>

	<!-- Mobile sidebar overlay -->
	{#if mobileOpen}
		<!-- Backdrop -->
		<div class="fixed inset-0 z-[var(--z-modal-backdrop)] lg:hidden">
			<button
				class="absolute inset-0 bg-black/50 backdrop-blur-xs"
				onclick={closeMobile}
				aria-label="Close navigation menu"
				tabindex="-1"
			></button>
			<div
				class="relative z-[var(--z-modal)] h-full w-[var(--spacing-sidebar)] max-w-[80vw]"
				role="dialog"
				aria-modal="true"
				aria-label="Navigation menu"
			>
				<Sidebar collapsed={false} oncollapse={closeMobile} />
			</div>
		</div>
	{/if}

	<!-- Main content area -->
	<div class="flex min-w-0 flex-1 flex-col">
		<Header onmobileopen={openMobile} />

		<main
			id="main-content"
			class="flex-1 overflow-y-auto p-[var(--spacing-content)]"
		>
			{@render children()}
		</main>
	</div>
</div>
