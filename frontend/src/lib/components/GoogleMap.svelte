<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getSpeedColor } from '$lib/types/speed';
	import { loadGoogleMaps } from '$lib/services/googleMapsLoader';
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
	let map: google.maps.Map | null = $state(null);
	let overlays: (google.maps.Polyline | google.maps.Marker | google.maps.marker.AdvancedMarkerElement)[] = [];
	let clickListener: google.maps.MapsEventListener | null = null;

	onMount(async () => {
		const g = await loadGoogleMaps();
		const m = new g.maps.Map(mapContainer, {
			center: { lat: -1.2864, lng: 36.8172 },
			zoom: 7,
			mapId: 'speed-ke-map',
			zoomControl: true,
			zoomControlOptions: { position: g.maps.ControlPosition.RIGHT_BOTTOM },
			streetViewControl: false,
			mapTypeControl: true,
			mapTypeControlOptions: {
				position: g.maps.ControlPosition.TOP_RIGHT,
				style: g.maps.MapTypeControlStyle.DROPDOWN_MENU
			}
		});

		clickListener = m.addListener('click', (e: google.maps.MapMouseEvent) => {
			if (e.latLng) onMapClick?.(e.latLng.lat(), e.latLng.lng());
		});

		map = m;
	});

	onDestroy(() => {
		clearOverlays();
		clickListener?.remove();
	});

	function clearOverlays() {
		for (const o of overlays) {
			if (o instanceof google.maps.Polyline) o.setMap(null);
			else if (o instanceof google.maps.Marker) o.setMap(null);
			else if ('map' in o) o.map = null;
		}
		overlays = [];
	}

	function createPinMarker(
		m: google.maps.Map,
		pos: [number, number],
		color: string,
		label: string
	): google.maps.Marker {
		return new google.maps.Marker({
			map: m,
			position: { lat: pos[0], lng: pos[1] },
			label: { text: label, color: '#fff', fontWeight: '700', fontSize: '12px' },
			icon: {
				path: google.maps.SymbolPath.CIRCLE,
				scale: 12,
				fillColor: color,
				fillOpacity: 1,
				strokeColor: '#fff',
				strokeWeight: 3
			}
		});
	}

	$effect(() => {
		const m = map;
		const segs = segments;
		const alts = alternativeRoutes;
		const sm = startMarker;
		const em = endMarker;

		if (!m) return;

		clearOverlays();

		for (const alt of alts) {
			const poly = new google.maps.Polyline({
				map: m,
				path: alt.map((c) => ({ lat: c[0], lng: c[1] })),
				strokeColor: '#6b7280',
				strokeWeight: 4,
				strokeOpacity: 0.5,
				icons: [{ icon: { path: 'M 0,-1 0,1', strokeOpacity: 0.6, scale: 3 }, offset: '0', repeat: '16px' }],
				zIndex: 1
			});
			poly.setOptions({ strokeOpacity: 0 });
			overlays.push(poly);
		}

		for (const seg of segs) {
			const color = getSpeedColor(seg.speedLimit);
			const path = seg.coordinates.map((c) => ({ lat: c[0], lng: c[1] }));

			const shadow = new google.maps.Polyline({
				map: m,
				path,
				strokeColor: '#000',
				strokeWeight: 8,
				strokeOpacity: 0.2,
				zIndex: 2
			});
			overlays.push(shadow);

			const line = new google.maps.Polyline({
				map: m,
				path,
				strokeColor: color,
				strokeWeight: 5,
				strokeOpacity: 0.9,
				zIndex: 3
			});

			const infoWindow = new google.maps.InfoWindow({
				content: `<div style="font-family:sans-serif;font-size:13px;">
					<strong>${seg.roadName}</strong><br>
					<span style="color:${color};font-size:18px;font-weight:700;">
						${seg.speedLimit !== null ? seg.speedLimit + ' km/h' : 'No data'}
					</span>
				</div>`
			});
			line.addListener('click', (e: google.maps.PolyMouseEvent) => {
				infoWindow.setPosition(e.latLng!);
				infoWindow.open(m);
			});
			overlays.push(line);
		}

		if (sm) overlays.push(createPinMarker(m, sm, '#22c55e', 'A'));
		if (em) overlays.push(createPinMarker(m, em, '#ef4444', 'B'));

		if (segs.length > 0) {
			const bounds = new google.maps.LatLngBounds();
			for (const seg of segs) {
				for (const c of seg.coordinates) {
					bounds.extend({ lat: c[0], lng: c[1] });
				}
			}
			m.fitBounds(bounds, 50);
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
</style>
