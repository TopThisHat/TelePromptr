<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	// Root page redirects to the traces dashboard
	onMount(() => {
		goto('/traces', { replaceState: true });
	});
</script>

<svelte:head>
	<title>TelePromptr</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center">
	<div class="text-center">
		<div
			class="bg-primary mx-auto flex h-12 w-12 items-center justify-center rounded-xl"
			aria-hidden="true"
		>
			<span class="text-primary-foreground text-lg font-bold">T</span>
		</div>
		<h1 class="text-surface-foreground mt-4 text-display font-bold tracking-tight">
			TelePromptr
		</h1>
		<p class="text-surface-muted mt-2 text-sm">LLM Telemetry Platform</p>
	</div>
</div>
