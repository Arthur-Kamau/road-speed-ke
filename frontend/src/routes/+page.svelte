<script lang="ts">
	import LeafletMap from '$lib/components/Map.svelte';
	import GoogleMap from '$lib/components/GoogleMap.svelte';
	import RouteInput from '$lib/components/RouteInput.svelte';
	import SpeedLegend from '$lib/components/SpeedLegend.svelte';
	import RouteInfo from '$lib/components/RouteInfo.svelte';
	import { getRoutes, fetchSpeedsByBBox } from '$lib/services/api';
	import { matchRouteToSpeeds, type MatchedSegment } from '$lib/services/matcher';
	import type { SpeedFeature, RouteOption } from '$lib/types/speed';
	import { MAP_PROVIDER, GOOGLE_AVAILABLE } from '$lib/services/mapConfig';

	let segments = $state<MatchedSegment[]>([]);
	let startMarker = $state<[number, number] | null>(null);
	let endMarker = $state<[number, number] | null>(null);
	let loading = $state(false);
	let routeDistance = $state(0);
	let routeDuration = $state(0);
	let error = $state('');
	let apiAvailable = $state(true);
	let clickMode = $state<'start' | 'end' | null>(null);
	let mapTiles = $state<'google' | 'osm'>(MAP_PROVIDER === 'google' ? 'google' : 'osm');

	let routeOptions = $state<RouteOption[]>([]);
	let selectedRouteIndex = $state(0);
	let alternativeRoutes = $derived(
		routeOptions.filter((_, i) => i !== selectedRouteIndex).map((r) => r.coordinates)
	);

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
			const routes = await getRoutes(start[0], start[1], end[0], end[1]);
			if (routes.length === 0) {
				error = 'Could not find a route between those points.';
				segments = [];
				routeOptions = [];
				return;
			}

			routeOptions = routes;
			selectRoute(0);
		} catch (e) {
			error = 'Failed to calculate route. Please try again.';
			segments = [];
			routeOptions = [];
		} finally {
			loading = false;
		}
	}

	function selectRoute(index: number) {
		const route = routeOptions[index];
		if (!route) return;
		selectedRouteIndex = index;
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

		{#if routeOptions.length > 1}
			<div class="route-options">
				{#each routeOptions as route, i}
					<button
						class="route-option"
						class:active={i === selectedRouteIndex}
						onclick={() => selectRoute(i)}
					>
						<span class="route-option-summary">{route.summary}</span>
						<span class="route-option-meta">
							{(route.distance / 1000).toFixed(1)} km · {Math.round(route.duration / 60)} min
						</span>
					</button>
				{/each}
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
		{#if GOOGLE_AVAILABLE}
			<div class="map-toggle">
				<button
					class="toggle-btn"
					class:active={mapTiles === 'google'}
					onclick={() => (mapTiles = 'google')}
				>Google Maps</button>
				<button
					class="toggle-btn"
					class:active={mapTiles === 'osm'}
					onclick={() => (mapTiles = 'osm')}
				>OpenStreetMap</button>
			</div>
		{/if}

		{#if mapTiles === 'google'}
			<GoogleMap {segments} {alternativeRoutes} {startMarker} {endMarker} onMapClick={handleMapClick} />
		{:else}
			<LeafletMap {segments} {alternativeRoutes} {startMarker} {endMarker} onMapClick={handleMapClick} />
		{/if}
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
		position: relative;
	}

	.map-toggle {
		position: absolute;
		top: 10px;
		left: 10px;
		z-index: 1000;
		display: flex;
		background: #fff;
		border-radius: 8px;
		box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
		overflow: hidden;
	}

	.toggle-btn {
		padding: 6px 14px;
		border: none;
		background: #fff;
		font-size: 12px;
		font-weight: 600;
		color: #666;
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}

	.toggle-btn:not(:last-child) {
		border-right: 1px solid #e5e5e5;
	}

	.toggle-btn.active {
		background: #2563eb;
		color: #fff;
	}

	.toggle-btn:hover:not(.active) {
		background: #f0f0f0;
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

	.route-options {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.route-option {
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: #f5f5f5;
		border: 2px solid transparent;
		border-radius: 8px;
		padding: 8px 12px;
		font-size: 12px;
		cursor: pointer;
		text-align: left;
	}

	.route-option-summary {
		font-weight: 600;
		color: #333;
	}

	.route-option-meta {
		color: #888;
	}

	.route-option.active {
		border-color: #2563eb;
		background: #eff6ff;
	}

	.route-option.active .route-option-meta {
		color: #2563eb;
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
