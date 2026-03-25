<script lang="ts">
	/**
	 * Icon wrapper component.
	 *
	 * Standardizes icon sizing, color inheritance, and accessibility
	 * for Lucide icons used throughout the TelePromptr dashboard.
	 *
	 * @example
	 * ```svelte
	 * <Icon icon={Activity} size="md" />
	 * <Icon icon={Settings} size="lg" label="Settings" />
	 * ```
	 */
	import type { Component, SvelteComponent } from 'svelte';

	/**
	 * Accepts both Svelte 5 function components and Svelte 4 class components
	 * (lucide-svelte currently exports SvelteComponentTyped classes).
	 */
	type AnyComponent = Component<any> | typeof SvelteComponent<any>;

	/** Predefined icon sizes in pixels. */
	const SIZES = {
		xs: 14,
		sm: 16,
		md: 20,
		lg: 24,
		xl: 32
	} as const;

	type IconSize = keyof typeof SIZES;

	interface IconProps {
		/** The Lucide icon component to render. */
		icon: AnyComponent;
		/** Predefined size name. Defaults to 'md'. */
		size?: IconSize;
		/** Accessible label for meaningful (non-decorative) icons. */
		label?: string;
		/** Additional CSS classes. */
		class?: string;
		/** Stroke width override. */
		strokeWidth?: number;
	}

	let {
		icon: IconComponent,
		size = 'md',
		label,
		class: className = '',
		strokeWidth = 2
	}: IconProps = $props();

	let pixelSize = $derived(SIZES[size]);
	let ariaHidden = $derived(!label);
</script>

<IconComponent
	width={pixelSize}
	height={pixelSize}
	stroke-width={strokeWidth}
	aria-hidden={ariaHidden}
	aria-label={label}
	role={label ? 'img' : undefined}
	class="inline-block shrink-0 {className}"
/>
