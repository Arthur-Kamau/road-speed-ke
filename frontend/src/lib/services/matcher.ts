import type { SpeedFeature } from '$lib/types/speed';

export interface MatchedSegment {
	coordinates: [number, number][];
	speedLimit: number | null;
	roadName: string;
}

function distanceMeters(lat1: number, lng1: number, lat2: number, lng2: number): number {
	const R = 6371000;
	const dLat = ((lat2 - lat1) * Math.PI) / 180;
	const dLng = ((lng2 - lng1) * Math.PI) / 180;
	const a =
		Math.sin(dLat / 2) ** 2 +
		Math.cos((lat1 * Math.PI) / 180) *
			Math.cos((lat2 * Math.PI) / 180) *
			Math.sin(dLng / 2) ** 2;
	return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
}

function pointToSegmentDistance(
	px: number,
	py: number,
	ax: number,
	ay: number,
	bx: number,
	by: number
): number {
	const dx = bx - ax;
	const dy = by - ay;
	if (dx === 0 && dy === 0) return distanceMeters(px, py, ax, ay);

	let t = ((px - ax) * dx + (py - ay) * dy) / (dx * dx + dy * dy);
	t = Math.max(0, Math.min(1, t));

	return distanceMeters(px, py, ax + t * dx, ay + t * dy);
}

function nearestSpeedFeature(
	lat: number,
	lng: number,
	features: SpeedFeature[],
	maxDistMeters: number = 200
): SpeedFeature | null {
	let best: SpeedFeature | null = null;
	let bestDist = maxDistMeters;

	for (const feature of features) {
		const coords = feature.geometry.coordinates;
		for (let i = 0; i < coords.length - 1; i++) {
			const [aLng, aLat] = coords[i];
			const [bLng, bLat] = coords[i + 1];
			const d = pointToSegmentDistance(lat, lng, aLat, aLng, bLat, bLng);
			if (d < bestDist) {
				bestDist = d;
				best = feature;
			}
		}
	}

	return best;
}

export function matchRouteToSpeeds(
	routeCoords: [number, number][],
	features: SpeedFeature[]
): MatchedSegment[] {
	if (routeCoords.length === 0) return [];

	const segments: MatchedSegment[] = [];
	let currentFeature: SpeedFeature | null = null;
	let currentCoords: [number, number][] = [];

	for (let i = 0; i < routeCoords.length; i++) {
		const [lat, lng] = routeCoords[i];
		const matched = nearestSpeedFeature(lat, lng, features);

		const matchedId = matched?.properties?.road_name ?? null;
		const currentId = currentFeature?.properties?.road_name ?? null;

		if (matchedId !== currentId) {
			if (currentCoords.length > 1) {
				segments.push({
					coordinates: [...currentCoords],
					speedLimit: currentFeature?.properties?.speed_limit_kmh ?? null,
					roadName: currentFeature?.properties?.road_name ?? 'Unknown road'
				});
			}
			currentFeature = matched;
			currentCoords = i > 0 ? [routeCoords[i - 1], routeCoords[i]] : [routeCoords[i]];
		} else {
			currentCoords.push(routeCoords[i]);
		}
	}

	if (currentCoords.length > 1) {
		segments.push({
			coordinates: currentCoords,
			speedLimit: currentFeature?.properties?.speed_limit_kmh ?? null,
			roadName: currentFeature?.properties?.road_name ?? 'Unknown road'
		});
	}

	return segments;
}
