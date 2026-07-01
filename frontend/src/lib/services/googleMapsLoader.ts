import { GOOGLE_MAPS_API_KEY } from './mapConfig';

let loadPromise: Promise<typeof google> | null = null;

// Injects the Google Maps JS SDK once and resolves with the `google` namespace.
// Safe to call from multiple components — subsequent calls reuse the same script/promise.
export function loadGoogleMaps(): Promise<typeof google> {
	if (loadPromise) return loadPromise;

	loadPromise = new Promise((resolve, reject) => {
		if (typeof window === 'undefined') {
			reject(new Error('Google Maps can only load in the browser'));
			return;
		}
		if (window.google?.maps?.places) {
			resolve(window.google);
			return;
		}

		const script = document.createElement('script');
		script.src = `https://maps.googleapis.com/maps/api/js?key=${GOOGLE_MAPS_API_KEY}&libraries=places&loading=async&v=weekly`;
		script.async = true;
		script.onerror = () => reject(new Error('Failed to load Google Maps SDK'));
		script.onload = () => resolve(window.google);
		document.head.appendChild(script);
	});

	return loadPromise;
}
