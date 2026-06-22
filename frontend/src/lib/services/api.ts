import type { FeatureCollection, NominatimResult } from '$lib/types/speed';

const API_BASE = 'http://localhost:8080/api/v1';
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

export async function getRoute(
	startLat: number,
	startLng: number,
	endLat: number,
	endLng: number
): Promise<{ coordinates: [number, number][]; distance: number; duration: number } | null> {
	const url = `${OSRM}/route/v1/driving/${startLng},${startLat};${endLng},${endLat}?overview=full&geometries=geojson&steps=true`;
	const res = await fetch(url);
	if (!res.ok) return null;
	const data = await res.json();
	if (!data.routes || data.routes.length === 0) return null;
	const route = data.routes[0];
	return {
		coordinates: route.geometry.coordinates.map((c: number[]) => [c[1], c[0]] as [number, number]),
		distance: route.distance,
		duration: route.duration
	};
}
