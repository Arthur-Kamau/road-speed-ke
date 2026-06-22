<script lang="ts">
	import { geocode } from '$lib/services/api';
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

	function debounceSearch(query: string, setter: (r: NominatimResult[]) => void) {
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
		{#if startFocused && startResults.length > 0}
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
		{#if endFocused && endResults.length > 0}
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
</style>
