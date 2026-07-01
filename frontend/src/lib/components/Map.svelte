<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getSpeedColor } from '$lib/types/speed';
	import type { MatchedSegment } from '$lib/services/matcher';

	interface Props {
		segments?: MatchedSegment[];
		alternativeRoutes?: [number, number][][];
		startMarker?: [number, number] | null;
		endMarker?: [number, number] | null;
		onMapClick?: (lat: number, lng: number) => void;
	}

	let {
		segments = [],
		alternativeRoutes = [],
		startMarker = null,
		endMarker = null,
		onMapClick
	}: Props = $props();

	let mapContainer: HTMLDivElement;
	let map: L.Map | null = $state(null);
	let layerGroup: L.LayerGroup | null = $state(null);
	let startMarkerLayer: L.Marker | null = null;
	let endMarkerLayer: L.Marker | null = null;
	let leaflet: typeof import('leaflet') | null = $state(null);

	onMount(async () => {
		const L = await import('leaflet');
		leaflet = L;

		const m = L.map(mapContainer, {
			center: [-1.2864, 36.8172],
			zoom: 7,
			zoomControl: false
		});

		L.control.zoom({ position: 'bottomright' }).addTo(m);

		L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
			attribution:
				'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
			maxZoom: 19
		}).addTo(m);

		const lg = L.layerGroup().addTo(m);

		m.on('click', (e: L.LeafletMouseEvent) => {
			onMapClick?.(e.latlng.lat, e.latlng.lng);
		});

		map = m;
		layerGroup = lg;
	});

	onDestroy(() => {
		map?.remove();
	});

	function createIcon(L: typeof import('leaflet'), color: string, label: string): L.DivIcon {
		return L.divIcon({
			className: 'custom-marker',
			html: `<div style="background:${color};width:24px;height:24px;border-radius:50%;border:3px solid white;box-shadow:0 2px 6px rgba(0,0,0,0.3);display:flex;align-items:center;justify-content:center;color:white;font-size:12px;font-weight:700;">${label}</div>`,
			iconSize: [24, 24],
			iconAnchor: [12, 12]
		});
	}

	$effect(() => {
		const L = leaflet;
		const m = map;
		const lg = layerGroup;
		const segs = segments;
		const alts = alternativeRoutes;
		const sm = startMarker;
		const em = endMarker;

		if (!L || !m || !lg) return;

		lg.clearLayers();

		for (const alt of alts) {
			const latlngs = alt.map((c) => L.latLng(c[0], c[1]));
			lg.addLayer(
				L.polyline(latlngs, { color: '#6b7280', weight: 4, opacity: 0.6, dashArray: '1 8' })
			);
		}

		for (const seg of segs) {
			const color = getSpeedColor(seg.speedLimit);
			const latlngs = seg.coordinates.map((c) => L.latLng(c[0], c[1]));

			lg.addLayer(
				L.polyline(latlngs, { color: '#000', weight: 8, opacity: 0.2 })
			);

			const line = L.polyline(latlngs, { color, weight: 5, opacity: 0.9 });
			line.bindPopup(
				`<div style="font-family:sans-serif;font-size:13px;">
					<strong>${seg.roadName}</strong><br>
					<span style="color:${color};font-size:18px;font-weight:700;">
						${seg.speedLimit !== null ? seg.speedLimit + ' km/h' : 'No data'}
					</span>
				</div>`
			);
			lg.addLayer(line);
		}

		if (startMarkerLayer) {
			m.removeLayer(startMarkerLayer);
			startMarkerLayer = null;
		}
		if (endMarkerLayer) {
			m.removeLayer(endMarkerLayer);
			endMarkerLayer = null;
		}

		if (sm) {
			startMarkerLayer = L.marker(sm, { icon: createIcon(L, '#22c55e', 'A') }).addTo(m);
		}
		if (em) {
			endMarkerLayer = L.marker(em, { icon: createIcon(L, '#ef4444', 'B') }).addTo(m);
		}

		if (segs.length > 0) {
			const allCoords = segs.flatMap((s) => s.coordinates);
			const bounds = L.latLngBounds(allCoords.map((c) => L.latLng(c[0], c[1])));
			m.fitBounds(bounds, { padding: [50, 50] });
		}
	});
</script>

<div class="map-wrapper">
	<div bind:this={mapContainer} class="map"></div>
</div>

<style>
	.map-wrapper {
		width: 100%;
		height: 100%;
		position: relative;
	}

	.map {
		width: 100%;
		height: 100%;
	}

	:global(.custom-marker) {
		background: none !important;
		border: none !important;
	}
</style>
