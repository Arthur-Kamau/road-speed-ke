<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getSpeedColor } from '$lib/types/speed';
	import type { MatchedSegment } from '$lib/services/matcher';

	interface Props {
		segments?: MatchedSegment[];
		startMarker?: [number, number] | null;
		endMarker?: [number, number] | null;
		onMapClick?: (lat: number, lng: number) => void;
	}

	let { segments = [], startMarker = null, endMarker = null, onMapClick }: Props = $props();

	let mapContainer: HTMLDivElement;
	let map: L.Map;
	let layerGroup: L.LayerGroup;
	let startMarkerLayer: L.Marker | null = null;
	let endMarkerLayer: L.Marker | null = null;
	let L: typeof import('leaflet');

	onMount(async () => {
		L = await import('leaflet');

		map = L.map(mapContainer, {
			center: [-1.2864, 36.8172],
			zoom: 7,
			zoomControl: false
		});

		L.control.zoom({ position: 'bottomright' }).addTo(map);

		L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
			attribution:
				'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
			maxZoom: 19
		}).addTo(map);

		layerGroup = L.layerGroup().addTo(map);

		map.on('click', (e: L.LeafletMouseEvent) => {
			onMapClick?.(e.latlng.lat, e.latlng.lng);
		});
	});

	onDestroy(() => {
		map?.remove();
	});

	function createIcon(color: string, label: string): L.DivIcon {
		return L.divIcon({
			className: 'custom-marker',
			html: `<div style="background:${color};width:24px;height:24px;border-radius:50%;border:3px solid white;box-shadow:0 2px 6px rgba(0,0,0,0.3);display:flex;align-items:center;justify-content:center;color:white;font-size:12px;font-weight:700;">${label}</div>`,
			iconSize: [24, 24],
			iconAnchor: [12, 12]
		});
	}

	$effect(() => {
		if (!map || !L || !layerGroup) return;

		layerGroup.clearLayers();

		for (const seg of segments) {
			const color = getSpeedColor(seg.speedLimit);
			const latlngs = seg.coordinates.map((c) => L.latLng(c[0], c[1]));

			const bgLine = L.polyline(latlngs, {
				color: '#000',
				weight: 8,
				opacity: 0.2
			});
			layerGroup.addLayer(bgLine);

			const line = L.polyline(latlngs, {
				color,
				weight: 5,
				opacity: 0.9
			});

			line.bindPopup(
				`<div style="font-family:sans-serif;font-size:13px;">
					<strong>${seg.roadName}</strong><br>
					<span style="color:${color};font-size:18px;font-weight:700;">
						${seg.speedLimit !== null ? seg.speedLimit + ' km/h' : 'No data'}
					</span>
				</div>`
			);

			layerGroup.addLayer(line);
		}

		if (startMarkerLayer) {
			map.removeLayer(startMarkerLayer);
			startMarkerLayer = null;
		}
		if (endMarkerLayer) {
			map.removeLayer(endMarkerLayer);
			endMarkerLayer = null;
		}

		if (startMarker) {
			startMarkerLayer = L.marker(startMarker, {
				icon: createIcon('#22c55e', 'A')
			}).addTo(map);
		}

		if (endMarker) {
			endMarkerLayer = L.marker(endMarker, {
				icon: createIcon('#ef4444', 'B')
			}).addTo(map);
		}

		if (segments.length > 0) {
			const allCoords = segments.flatMap((s) => s.coordinates);
			const bounds = L.latLngBounds(allCoords.map((c) => L.latLng(c[0], c[1])));
			map.fitBounds(bounds, { padding: [50, 50] });
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
