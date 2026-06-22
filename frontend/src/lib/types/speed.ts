export interface SpeedFeature {
	type: 'Feature';
	properties: {
		id: number;
		road_name: string;
		road_class: string;
		speed_limit_kmh: number;
		direction: string;
		source: string;
		verified: boolean;
		county: string;
		last_updated: string;
	};
	geometry: {
		type: 'LineString';
		coordinates: [number, number][];
	};
}

export interface FeatureCollection {
	type: 'FeatureCollection';
	features: SpeedFeature[];
}

export interface RouteStep {
	geometry: [number, number][];
	speedLimit: number | null;
	roadName: string;
}

export interface NominatimResult {
	display_name: string;
	lat: string;
	lon: string;
}

export const SPEED_COLORS: Record<string, string> = {
	none: '#9ca3af',
	'30': '#3b82f6',
	'50': '#22c55e',
	'80': '#eab308',
	'100': '#f97316',
	'110': '#ef4444',
};

export function getSpeedColor(speedKmh: number | null): string {
	if (speedKmh === null) return SPEED_COLORS.none;
	if (speedKmh <= 30) return SPEED_COLORS['30'];
	if (speedKmh <= 50) return SPEED_COLORS['50'];
	if (speedKmh <= 80) return SPEED_COLORS['80'];
	if (speedKmh <= 100) return SPEED_COLORS['100'];
	return SPEED_COLORS['110'];
}

export function getSpeedLabel(speedKmh: number | null): string {
	if (speedKmh === null) return 'Unknown';
	return `${speedKmh} km/h`;
}
