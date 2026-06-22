<script lang="ts">
	import { getSpeedColor } from '$lib/types/speed';
	import type { MatchedSegment } from '$lib/services/matcher';

	interface Props {
		segments: MatchedSegment[];
		distance: number;
		duration: number;
	}

	let { segments, distance, duration }: Props = $props();

	function formatDuration(seconds: number): string {
		const h = Math.floor(seconds / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		if (h > 0) return `${h}h ${m}m`;
		return `${m} min`;
	}

	function formatDistance(meters: number): string {
		return `${(meters / 1000).toFixed(1)} km`;
	}

	let speedBreakdown = $derived.by(() => {
		const buckets: Record<string, number> = {};
		for (const seg of segments) {
			const key = seg.speedLimit !== null ? `${seg.speedLimit}` : 'unknown';
			let dist = 0;
			for (let i = 0; i < seg.coordinates.length - 1; i++) {
				const [lat1, lng1] = seg.coordinates[i];
				const [lat2, lng2] = seg.coordinates[i + 1];
				const R = 6371000;
				const dLat = ((lat2 - lat1) * Math.PI) / 180;
				const dLng = ((lng2 - lng1) * Math.PI) / 180;
				const a =
					Math.sin(dLat / 2) ** 2 +
					Math.cos((lat1 * Math.PI) / 180) *
						Math.cos((lat2 * Math.PI) / 180) *
						Math.sin(dLng / 2) ** 2;
				dist += R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
			}
			buckets[key] = (buckets[key] || 0) + dist;
		}
		return Object.entries(buckets)
			.map(([key, dist]) => ({
				speed: key === 'unknown' ? null : parseInt(key),
				distance: dist,
				color: getSpeedColor(key === 'unknown' ? null : parseInt(key))
			}))
			.sort((a, b) => (a.speed ?? 999) - (b.speed ?? 999));
	});

	let totalMatched = $derived(speedBreakdown.reduce((s, b) => s + b.distance, 0));
</script>

<div class="route-info">
	<div class="summary">
		<div class="stat">
			<span class="stat-value">{formatDistance(distance)}</span>
			<span class="stat-label">Distance</span>
		</div>
		<div class="stat">
			<span class="stat-value">{formatDuration(duration)}</span>
			<span class="stat-label">Est. Time</span>
		</div>
		<div class="stat">
			<span class="stat-value">{segments.length}</span>
			<span class="stat-label">Zones</span>
		</div>
	</div>

	<h3>Speed Distribution</h3>

	<div class="bar">
		{#each speedBreakdown as bucket}
			<div
				class="bar-segment"
				style="width:{(bucket.distance / totalMatched) * 100}%;background:{bucket.color}"
				title="{bucket.speed !== null ? bucket.speed + ' km/h' : 'No data'}: {(bucket.distance / 1000).toFixed(1)} km"
			></div>
		{/each}
	</div>

	<div class="breakdown">
		{#each speedBreakdown as bucket}
			<div class="breakdown-row">
				<span class="color-dot" style="background:{bucket.color}"></span>
				<span class="speed-label">
					{bucket.speed !== null ? `${bucket.speed} km/h` : 'No data'}
				</span>
				<span class="distance">{(bucket.distance / 1000).toFixed(1)} km</span>
				<span class="percent">{((bucket.distance / totalMatched) * 100).toFixed(0)}%</span>
			</div>
		{/each}
	</div>

	<div class="segments-list">
		<h3>Route Segments</h3>
		{#each segments as seg, i}
			<div class="segment-item">
				<span class="seg-color" style="background:{getSpeedColor(seg.speedLimit)}"></span>
				<div class="seg-info">
					<span class="seg-name">{seg.roadName}</span>
					<span class="seg-speed">
						{seg.speedLimit !== null ? `${seg.speedLimit} km/h` : 'No data'}
					</span>
				</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.route-info {
		background: rgba(255, 255, 255, 0.95);
		border-radius: 12px;
		padding: 16px;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
		width: 340px;
		max-height: calc(100vh - 280px);
		overflow-y: auto;
	}

	.summary {
		display: flex;
		gap: 12px;
		margin-bottom: 16px;
	}

	.stat {
		flex: 1;
		text-align: center;
		padding: 8px;
		background: #f8f8f8;
		border-radius: 8px;
	}

	.stat-value {
		display: block;
		font-size: 16px;
		font-weight: 700;
		color: #1a1a1a;
	}

	.stat-label {
		font-size: 10px;
		color: #888;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	h3 {
		margin: 0 0 8px;
		font-size: 13px;
		font-weight: 700;
		color: #1a1a1a;
	}

	.bar {
		display: flex;
		height: 12px;
		border-radius: 6px;
		overflow: hidden;
		margin-bottom: 12px;
	}

	.bar-segment {
		height: 100%;
		min-width: 2px;
		transition: width 0.3s;
	}

	.breakdown {
		margin-bottom: 16px;
	}

	.breakdown-row {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 4px 0;
		font-size: 12px;
	}

	.color-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.speed-label {
		flex: 1;
		color: #333;
	}

	.distance {
		color: #666;
		font-variant-numeric: tabular-nums;
	}

	.percent {
		color: #999;
		width: 32px;
		text-align: right;
		font-variant-numeric: tabular-nums;
	}

	.segments-list {
		border-top: 1px solid #eee;
		padding-top: 12px;
	}

	.segment-item {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 6px 0;
		border-bottom: 1px solid #f5f5f5;
	}

	.segment-item:last-child {
		border-bottom: none;
	}

	.seg-color {
		width: 6px;
		height: 28px;
		border-radius: 3px;
		flex-shrink: 0;
	}

	.seg-info {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.seg-name {
		font-size: 12px;
		color: #333;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 240px;
	}

	.seg-speed {
		font-size: 11px;
		color: #888;
	}
</style>
