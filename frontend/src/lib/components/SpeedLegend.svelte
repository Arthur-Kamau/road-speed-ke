<script lang="ts">
	import { SPEED_COLORS } from '$lib/types/speed';

	interface Props {
		segmentCount?: number;
		totalDistance?: number;
	}

	let { segmentCount = 0, totalDistance = 0 }: Props = $props();

	const legendItems = [
		{ label: 'No data', color: SPEED_COLORS.none },
		{ label: '≤ 30 km/h (School zone)', color: SPEED_COLORS['30'] },
		{ label: '≤ 50 km/h (Urban)', color: SPEED_COLORS['50'] },
		{ label: '51–80 km/h (Peri-urban)', color: SPEED_COLORS['80'] },
		{ label: '81–100 km/h (Highway)', color: SPEED_COLORS['100'] },
		{ label: '101–110 km/h (Expressway)', color: SPEED_COLORS['110'] }
	];

	function formatDistance(meters: number): string {
		if (meters >= 1000) return `${(meters / 1000).toFixed(1)} km`;
		return `${Math.round(meters)} m`;
	}
</script>

<div class="legend">
	<h3>Speed Limits</h3>
	{#each legendItems as item}
		<div class="legend-item">
			<span class="color-bar" style="background:{item.color}"></span>
			<span class="label">{item.label}</span>
		</div>
	{/each}

	{#if segmentCount > 0}
		<div class="stats">
			<span>{segmentCount} segments</span>
			{#if totalDistance > 0}
				<span>{formatDistance(totalDistance)}</span>
			{/if}
		</div>
	{/if}
</div>

<style>
	.legend {
		background: rgba(255, 255, 255, 0.95);
		border-radius: 10px;
		padding: 14px 16px;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
		min-width: 180px;
	}

	h3 {
		margin: 0 0 10px;
		font-size: 14px;
		font-weight: 700;
		color: #1a1a1a;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 3px 0;
	}

	.color-bar {
		display: block;
		width: 24px;
		height: 6px;
		border-radius: 3px;
		flex-shrink: 0;
	}

	.label {
		font-size: 12px;
		color: #444;
	}

	.stats {
		margin-top: 10px;
		padding-top: 8px;
		border-top: 1px solid #e5e5e5;
		font-size: 11px;
		color: #888;
		display: flex;
		justify-content: space-between;
	}
</style>
