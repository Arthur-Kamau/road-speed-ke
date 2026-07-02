<script lang="ts">
	import { onMount } from 'svelte';
	import { geocode } from '$lib/services/api';
	import { MAP_PROVIDER } from '$lib/services/mapConfig';
	import { loadGoogleMaps } from '$lib/services/googleMapsLoader';
	import type { NominatimResult } from '$lib/types/speed';

	interface Props {
		onRoute: (start: [number, number], end: [number, number]) => void;
		loading?: boolean;
	}

	let { onRoute, loading = false }: Props = $props();

	let startQuery = $state('');
	let endQuery = $state('');
	let startResults = $state<NominatimResult[]>([]);
	let endResults = $state<NominatimResult[]>([]);
	let startCoord = $state<[number, number] | null>(null);
	let endCoord = $state<[number, number] | null>(null);
	let startFocused = $state(false);
	let endFocused = $state(false);
	let searchTimeout: ReturnType<typeof setTimeout>;

	let startInputEl: HTMLInputElement;
	let endInputEl: HTMLInputElement;
	let useGooglePlaces = $state(MAP_PROVIDER === 'google');

	onMount(() => {
		if (!useGooglePlaces) return;

		loadGoogleMaps()
			.then((g) => {
				const options: google.maps.places.AutocompleteOptions = {
					componentRestrictions: { country: 'ke' },
					fields: ['geometry', 'name', 'formatted_address']
				};

				const startAc = new g.maps.places.Autocomplete(startInputEl, options);
				startAc.addListener('place_changed', () => {
					const place = startAc.getPlace();
					if (place.geometry?.location) {
						startCoord = [place.geometry.location.lat(), place.geometry.location.lng()];
						startQuery = place.name ?? place.formatted_address ?? startInputEl.value;
					}
				});

				const endAc = new g.maps.places.Autocomplete(endInputEl, options);
				endAc.addListener('place_changed', () => {
					const place = endAc.getPlace();
					if (place.geometry?.location) {
						endCoord = [place.geometry.location.lat(), place.geometry.location.lng()];
						endQuery = place.name ?? place.formatted_address ?? endInputEl.value;
					}
				});

				// Prevent the form from submitting when user presses Enter to select
				// an autocomplete suggestion (Google fires keydown before place_changed)
				const suppressEnter = (e: KeyboardEvent) => {
					if (e.key === 'Enter') {
						const pacVisible = document.querySelector('.pac-container:not([style*="display: none"])');
						if (pacVisible) e.preventDefault();
					}
				};
				startInputEl.addEventListener('keydown', suppressEnter);
				endInputEl.addEventListener('keydown', suppressEnter);
			})
			.catch((e) => {
				console.error('Google Places unavailable, falling back to manual search', e);
				useGooglePlaces = false;
			});
	});

	function debounceSearch(query: string, setter: (r: NominatimResult[]) => void) {
		if (useGooglePlaces) return;
		clearTimeout(searchTimeout);
		if (query.length < 2) {
			setter([]);
			return;
		}
		searchTimeout = setTimeout(async () => {
			const results = await geocode(query);
			setter(results);
		}, 400);
	}

	let canPreview = $derived(startCoord !== null && endCoord !== null);

	function selectStart(r: NominatimResult) {
		startQuery = r.display_name.split(',').slice(0, 2).join(',');
		startCoord = [parseFloat(r.lat), parseFloat(r.lon)];
		startResults = [];
		startFocused = false;
	}

	function selectEnd(r: NominatimResult) {
		endQuery = r.display_name.split(',').slice(0, 2).join(',');
		endCoord = [parseFloat(r.lat), parseFloat(r.lon)];
		endResults = [];
		endFocused = false;
	}

	async function resolveAndGo() {
		// When using Google Places, read coordinates from the input if place_changed
		// already fired — don't fall back to Nominatim geocode.
		if (useGooglePlaces) {
			if (startCoord && endCoord) {
				onRoute(startCoord, endCoord);
			}
			return;
		}

		if (!startCoord && startQuery.length >= 2) {
			const results = await geocode(startQuery);
			if (results.length > 0) {
				startCoord = [parseFloat(results[0].lat), parseFloat(results[0].lon)];
				startQuery = results[0].display_name.split(',').slice(0, 2).join(',');
			}
		}
		if (!endCoord && endQuery.length >= 2) {
			const results = await geocode(endQuery);
			if (results.length > 0) {
				endCoord = [parseFloat(results[0].lat), parseFloat(results[0].lon)];
				endQuery = results[0].display_name.split(',').slice(0, 2).join(',');
			}
		}
		if (startCoord && endCoord) {
			onRoute(startCoord, endCoord);
		}
	}

	function handleSwap() {
		[startQuery, endQuery] = [endQuery, startQuery];
		[startCoord, endCoord] = [endCoord, startCoord];
		if (useGooglePlaces) {
			startInputEl.value = startQuery;
			endInputEl.value = endQuery;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') resolveAndGo();
	}
</script>

<div class="route-input">
	<h2>Plan Your Route</h2>

	<div class="input-group">
		<label for="start">From</label>
		<div class="input-wrapper">
			<span class="dot start-dot"></span>
			<input
				id="start"
				type="text"
				placeholder="e.g. Nairobi CBD"
				bind:this={startInputEl}
				bind:value={startQuery}
				onfocus={() => (startFocused = true)}
				onblur={() => setTimeout(() => (startFocused = false), 200)}
				onkeydown={handleKeydown}
				oninput={() => {
					startCoord = null;
					debounceSearch(startQuery, (r) => {
						startResults = r;
					});
				}}
			/>
		</div>
		{#if !useGooglePlaces && startFocused && startResults.length > 0}
			<ul class="suggestions">
				{#each startResults as r}
					<li>
						<button type="button" onmousedown={() => selectStart(r)}>
							{r.display_name}
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	<button class="swap-btn" onclick={handleSwap} title="Swap start and destination">
		&#8645;
	</button>

	<div class="input-group">
		<label for="end">To</label>
		<div class="input-wrapper">
			<span class="dot end-dot"></span>
			<input
				id="end"
				type="text"
				placeholder="e.g. Mombasa"
				bind:this={endInputEl}
				bind:value={endQuery}
				onfocus={() => (endFocused = true)}
				onblur={() => setTimeout(() => (endFocused = false), 200)}
				onkeydown={handleKeydown}
				oninput={() => {
					endCoord = null;
					debounceSearch(endQuery, (r) => {
						endResults = r;
					});
				}}
			/>
		</div>
		{#if !useGooglePlaces && endFocused && endResults.length > 0}
			<ul class="suggestions">
				{#each endResults as r}
					<li>
						<button type="button" onmousedown={() => selectEnd(r)}>
							{r.display_name}
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	<button
		class="preview-btn"
		disabled={loading || (startQuery.length < 2 || endQuery.length < 2)}
		onclick={resolveAndGo}
	>
		{#if loading}
			Calculating route...
		{:else}
			Preview Route
		{/if}
	</button>
</div>

<style>
	.route-input {
		background: rgba(255, 255, 255, 0.95);
		border-radius: 12px;
		padding: 16px;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
		width: 340px;
	}

	h2 {
		margin: 0 0 12px;
		font-size: 16px;
		font-weight: 700;
		color: #1a1a1a;
	}

	.input-group {
		position: relative;
		margin-bottom: 4px;
	}

	label {
		font-size: 11px;
		font-weight: 600;
		color: #888;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.input-wrapper {
		display: flex;
		align-items: center;
		gap: 8px;
		background: #f5f5f5;
		border-radius: 8px;
		padding: 0 10px;
		margin-top: 4px;
	}

	.dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.start-dot {
		background: #22c55e;
	}
	.end-dot {
		background: #ef4444;
	}

	input {
		flex: 1;
		border: none;
		background: transparent;
		padding: 10px 0;
		font-size: 14px;
		outline: none;
		color: #1a1a1a;
	}

	.suggestions {
		position: absolute;
		top: 100%;
		left: 0;
		right: 0;
		background: white;
		border-radius: 8px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
		list-style: none;
		margin: 4px 0 0;
		padding: 4px;
		z-index: 1000;
		max-height: 200px;
		overflow-y: auto;
	}

	.suggestions li button {
		display: block;
		width: 100%;
		text-align: left;
		padding: 8px 10px;
		border: none;
		background: none;
		font-size: 12px;
		color: #333;
		cursor: pointer;
		border-radius: 6px;
	}

	.suggestions li button:hover {
		background: #f0f0f0;
	}

	.swap-btn {
		display: block;
		margin: 2px auto;
		background: none;
		border: 1px solid #ddd;
		border-radius: 50%;
		width: 28px;
		height: 28px;
		font-size: 14px;
		cursor: pointer;
		color: #666;
		line-height: 1;
	}

	.swap-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	.preview-btn {
		width: 100%;
		margin-top: 12px;
		padding: 12px;
		background: #2563eb;
		color: white;
		border: none;
		border-radius: 8px;
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.15s;
	}

	.preview-btn:hover:not(:disabled) {
		background: #1d4ed8;
	}

	.preview-btn:disabled {
		background: #93c5fd;
		cursor: not-allowed;
	}

	/* Google Places Autocomplete dropdown (rendered at body level) */
	:global(.pac-container) {
		z-index: 10000 !important;
		border-radius: 8px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
		border: none;
		margin-top: 4px;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	}

	:global(.pac-item) {
		padding: 8px 12px;
		font-size: 13px;
		cursor: pointer;
		border-top: 1px solid #f0f0f0;
	}

	:global(.pac-item:first-child) {
		border-top: none;
	}

	:global(.pac-item:hover),
	:global(.pac-item-selected) {
		background: #f0f4ff;
	}

	:global(.pac-icon) {
		display: none;
	}

	:global(.pac-item-query) {
		font-size: 13px;
		color: #333;
	}
</style>
