import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';
export default defineConfig({
	build: {
		sourcemap: true
	},
	plugins: [sveltekit()],
	server: {
		proxy: {
			'/api': 'http://localhost:8080'
		}
	},
	test: {
		include: ['tests/**/*.test.ts']
	}
});
