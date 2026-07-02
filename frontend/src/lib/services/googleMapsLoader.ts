import { GOOGLE_MAPS_API_KEY } from './mapConfig';

let loadPromise: Promise<typeof google> | null = null;

export function loadGoogleMaps(): Promise<typeof google> {
	if (loadPromise) return loadPromise;

	loadPromise = new Promise((resolve, reject) => {
		if (typeof window === 'undefined') {
			reject(new Error('Google Maps can only load in the browser'));
			return;
		}
		if (window.google?.maps?.ControlPosition) {
			resolve(window.google);
			return;
		}

		const script = document.createElement('script');
		script.src = `https://maps.googleapis.com/maps/api/js?key=${GOOGLE_MAPS_API_KEY}&libraries=places&v=weekly&loading=async`;
		script.async = true;
		script.onerror = () => reject(new Error('Failed to load Google Maps SDK'));
		script.onload = () => {
			const check = () => {
				if (window.google?.maps?.ControlPosition) resolve(window.google);
				else setTimeout(check, 50);
			};
			check();
		};
		document.head.appendChild(script);
	});

	return loadPromise;
}
