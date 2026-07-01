// Map/routing/geocoding provider selection.
//
// 'google' uses Google Places Autocomplete + Directions API (faster autocomplete,
// alternative routes). 'free' uses Nominatim + OSRM (no API key, no billing) —
// see CLAUDE.md "Key Design Decisions". Map tiles stay OpenStreetMap/Leaflet either
// way; only geocoding and routing switch.
//
// Falls back to 'free' automatically if no API key is configured, so builds
// without a key (e.g. CI, contributors without a Google Cloud account) still work.
const apiKey = import.meta.env.VITE_GOOGLE_MAPS_API_KEY ?? '';
const requestedProvider = import.meta.env.VITE_MAP_PROVIDER ?? 'google';

export const GOOGLE_MAPS_API_KEY = apiKey;
export const MAP_PROVIDER: 'google' | 'free' = apiKey && requestedProvider !== 'free' ? 'google' : 'free';
