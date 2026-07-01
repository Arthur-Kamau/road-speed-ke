/// <reference types="vite/client" />
/// <reference types="google.maps" />

interface ImportMetaEnv {
	readonly VITE_GOOGLE_MAPS_API_KEY?: string;
	readonly VITE_MAP_PROVIDER?: 'google' | 'free';
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
