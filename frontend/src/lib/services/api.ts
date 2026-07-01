import type { FeatureCollection, NominatimResult, RouteOption } from '$lib/types/speed';
import { MAP_PROVIDER } from './mapConfig';
import { loadGoogleMaps } from './googleMapsLoader';

const API_BASE = typeof window !== 'undefined' && window.location.hostname !== 'localhost'
	? '/api/v1'
	: 'http://localhost:8080/api/v1';
const NOMINATIM = 'https://nominatim.openstreetmap.org';
const OSRM = 'https://router.project-osrm.org';

export async function fetchSpeedsByBBox(
	minLat: number,
	minLng: number,
	maxLat: number,
	maxLng: number
): Promise<FeatureCollection> {
	const res = await fetch(`${API_BASE}/speeds?bbox=${minLat},${minLng},${maxLat},${maxLng}`);
	if (!res.ok) throw new Error(`API error: ${res.status}`);
	return res.json();
}

export async function fetchSpeedsByRoute(points: [number, number][]): Promise<FeatureCollection> {
	const flat = points.map((p) => `${p[0]},${p[1]}`).join(',');
	const res = await fetch(`${API_BASE}/speeds/route?points=${flat}`);
	if (!res.ok) throw new Error(`API error: ${res.status}`);
	return res.json();
}

export async function geocode(query: string): Promise<NominatimResult[]> {
	const params = new URLSearchParams({
		q: query,
		format: 'json',
		countrycodes: 'ke',
		limit: '5',
		addressdetails: '1'
	});
	const res = await fetch(`${NOMINATIM}/search?${params}`, {
		headers: { 'User-Agent': 'KenyaSpeedLimits/0.1' }
	});
	if (!res.ok) return [];
	return res.json();
}

// Returns every alternative route found (at least one), sorted as the provider ranks them
// (fastest first). Falls back to OSRM automatically if Google fails at request time.
export async function getRoutes(
	startLat: number,
	startLng: number,
	endLat: number,
	endLng: number
): Promise<RouteOption[]> {
	if (MAP_PROVIDER === 'google') {
		try {
			return await getRoutesGoogle(startLat, startLng, endLat, endLng);
		} catch (e) {
			console.error('Google Directions failed, falling back to OSRM', e);
		}
	}
	return getRoutesOSRM(startLat, startLng, endLat, endLng);
}

// Kept for callers that only want the primary route.
export async function getRoute(
	startLat: number,
	startLng: number,
	endLat: number,
	endLng: number
): Promise<RouteOption | null> {
	const routes = await getRoutes(startLat, startLng, endLat, endLng);
	return routes[0] ?? null;
}

async function getRoutesOSRM(
	startLat: number,
	startLng: number,
	endLat: number,
	endLng: number
): Promise<RouteOption[]> {
	const url = `${OSRM}/route/v1/driving/${startLng},${startLat};${endLng},${endLat}?overview=full&geometries=geojson&steps=true&alternatives=true`;
	const res = await fetch(url);
	if (!res.ok) return [];
	const data = await res.json();
	if (!data.routes || data.routes.length === 0) return [];
	return data.routes.map((route: any, i: number) => ({
		coordinates: route.geometry.coordinates.map((c: number[]) => [c[1], c[0]] as [number, number]),
		distance: route.distance,
		duration: route.duration,
		summary: i === 0 ? 'Fastest route' : `Alternative ${i}`
	}));
}

async function getRoutesGoogle(
	startLat: number,
	startLng: number,
	endLat: number,
	endLng: number
): Promise<RouteOption[]> {
	const g = await loadGoogleMaps();
	const directionsService = new g.maps.DirectionsService();
	const result = await directionsService.route({
		origin: { lat: startLat, lng: startLng },
		destination: { lat: endLat, lng: endLng },
		travelMode: g.maps.TravelMode.DRIVING,
		provideRouteAlternatives: true
	});

	return result.routes.map((route) => {
		const coordinates: [number, number][] = route.legs.flatMap((leg) =>
			leg.steps.flatMap((step) => step.path.map((p) => [p.lat(), p.lng()] as [number, number]))
		);
		const distance = route.legs.reduce((sum, leg) => sum + (leg.distance?.value ?? 0), 0);
		const duration = route.legs.reduce((sum, leg) => sum + (leg.duration?.value ?? 0), 0);
		return { coordinates, distance, duration, summary: route.summary || 'Route' };
	});
}
