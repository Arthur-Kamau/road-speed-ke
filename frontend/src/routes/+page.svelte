<script lang="ts">
	import Map from '$lib/components/Map.svelte';
	import RouteInput from '$lib/components/RouteInput.svelte';
	import SpeedLegend from '$lib/components/SpeedLegend.svelte';
	import RouteInfo from '$lib/components/RouteInfo.svelte';
	import { getRoute, fetchSpeedsByBBox } from '$lib/services/api';
	import { matchRouteToSpeeds, type MatchedSegment } from '$lib/services/matcher';
	import type { SpeedFeature } from '$lib/types/speed';

	let segments = $state<MatchedSegment[]>([]);
	let startMarker = $state<[number, number] | null>(null);
	let endMarker = $state<[number, number] | null>(null);
	let loading = $state(false);
	let routeDistance = $state(0);
	let routeDuration = $state(0);
	let error = $state('');
	let apiAvailable = $state(true);
	let clickMode = $state<'start' | 'end' | null>(null);

	let allFeatures = $state<SpeedFeature[]>([]);

	async function loadAllSpeeds() {
		try {
			const data = await fetchSpeedsByBBox(-5, 33, 2, 42);
			allFeatures = data.features as SpeedFeature[];
			apiAvailable = true;
		} catch {
			apiAvailable = false;
			const res = await fetch('/speeds.json');
			if (res.ok) {
				const data = await res.json();
				allFeatures = data.features as SpeedFeature[];
			}
		}
	}

	loadAllSpeeds();

	async function handleRoute(start: [number, number], end: [number, number]) {
		loading = true;
		error = '';
		startMarker = start;
		endMarker = end;

		try {
			const route = await getRoute(start[0], start[1], end[0], end[1]);
			if (!route) {
				error = 'Could not find a route between those points.';
				segments = [];
				return;
			}

			routeDistance = route.distance;
			routeDuration = route.duration;

			if (allFeatures.length > 0) {
				segments = matchRouteToSpeeds(route.coordinates, allFeatures);
			} else {
				segments = [
					{
						coordinates: route.coordinates,
						speedLimit: null,
						roadName: 'Route (no speed data loaded)'
					}
				];
			}
		} catch (e) {
			error = 'Failed to calculate route. Please try again.';
			segments = [];
		} finally {
			loading = false;
		}
	}

	function handleMapClick(lat: number, lng: number) {
		if (clickMode === 'start') {
			startMarker = [lat, lng];
			clickMode = 'end';
			if (endMarker) {
				handleRoute(startMarker, endMarker);
			}
		} else if (clickMode === 'end') {
			endMarker = [lat, lng];
			clickMode = null;
			if (startMarker) {
				handleRoute(startMarker, [lat, lng]);
			}
		}
	}
</script>

<svelte:head>
	<title>Kenya Speed Limits</title>
	<meta name="description" content="See Kenyan road speed limits on a map. Plan routes and know the speed limit before you drive." />
</svelte:head>

<div class="app">
	<div class="sidebar">
		<div class="header">
			<h1>Kenya Speed Limits</h1>
			<p class="subtitle">Know your speed limits. Avoid NTSA fines.</p>
		</div>

		<RouteInput onRoute={handleRoute} {loading} />

		<div class="click-route">
			<button
				class="click-btn"
				class:active={clickMode !== null}
				onclick={() => (clickMode = clickMode ? null : 'start')}
			>
				{#if clickMode === 'start'}
					Click map to set start point
				{:else if clickMode === 'end'}
					Click map to set destination
				{:else}
					Or click on map to set points
				{/if}
			</button>
		</div>

		{#if error}
			<div class="error">{error}</div>
		{/if}

		{#if !apiAvailable}
			<div class="warning">
				API offline — using bundled speed data. Start the Go server for live data.
			</div>
		{/if}

		{#if segments.length > 0}
			<RouteInfo {segments} distance={routeDistance} duration={routeDuration} />
		{/if}

		<SpeedLegend segmentCount={allFeatures.length} />

		<div class="footer">
			<p>Data sourced from Traffic Act Cap 403 & Kenya Gazette notices.</p>
			<p>
				<a href="https://github.com/Arthur-Kamau/speed" target="_blank" rel="noopener">
					Contribute on GitHub
				</a>
			</p>
		</div>
	</div>

	<div class="map-area">
		<Map {segments} {startMarker} {endMarker} onMapClick={handleMapClick} />
	</div>
</div>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		overflow: hidden;
	}

	.app {
		display: flex;
		height: 100vh;
		width: 100vw;
	}

	.sidebar {
		width: 380px;
		min-width: 380px;
		height: 100vh;
		overflow-y: auto;
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		background: #fafafa;
		border-right: 1px solid #e5e5e5;
		box-sizing: border-box;
	}

	.map-area {
		flex: 1;
		height: 100vh;
	}

	.header h1 {
		margin: 0;
		font-size: 22px;
		font-weight: 800;
		color: #111;
	}

	.subtitle {
		margin: 4px 0 0;
		font-size: 13px;
		color: #666;
	}

	.click-route {
		text-align: center;
	}

	.click-btn {
		background: none;
		border: 1px dashed #ccc;
		border-radius: 8px;
		padding: 8px 16px;
		font-size: 12px;
		color: #888;
		cursor: pointer;
		width: 100%;
	}

	.click-btn:hover,
	.click-btn.active {
		border-color: #2563eb;
		color: #2563eb;
		background: #eff6ff;
	}

	.error {
		background: #fef2f2;
		color: #dc2626;
		padding: 10px 12px;
		border-radius: 8px;
		font-size: 13px;
	}

	.warning {
		background: #fffbeb;
		color: #b45309;
		padding: 10px 12px;
		border-radius: 8px;
		font-size: 12px;
	}

	.footer {
		margin-top: auto;
		padding-top: 12px;
		border-top: 1px solid #eee;
		font-size: 11px;
		color: #999;
	}

	.footer p {
		margin: 2px 0;
	}

	.footer a {
		color: #2563eb;
		text-decoration: none;
	}

	@media (max-width: 768px) {
		.app {
			flex-direction: column;
		}

		.sidebar {
			width: 100%;
			min-width: auto;
			height: auto;
			max-height: 45vh;
			border-right: none;
			border-bottom: 1px solid #e5e5e5;
		}

		.map-area {
			height: 55vh;
		}
	}
</style>
