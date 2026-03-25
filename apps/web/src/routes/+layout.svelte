<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<div class="bg-surface text-surface-foreground min-h-screen antialiased">
	{@render children()}
</div>
