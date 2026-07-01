const apiKey = import.meta.env.VITE_GOOGLE_MAPS_API_KEY ?? '';
const requestedProvider = import.meta.env.VITE_MAP_PROVIDER ?? 'google';

export const GOOGLE_MAPS_API_KEY = apiKey;
export const MAP_PROVIDER: 'google' | 'free' = apiKey && requestedProvider !== 'free' ? 'google' : 'free';
export const GOOGLE_AVAILABLE = !!apiKey;
