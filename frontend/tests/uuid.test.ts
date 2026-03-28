import { describe, it, expect } from 'vitest';
import { generateId } from '../src/lib/uuid';

describe('generateId', () => {
	it('returns a valid UUID v4 string', () => {
		const id = generateId();
		const uuidV4Regex = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
		expect(id).toMatch(uuidV4Regex);
	});

	it('returns unique values on successive calls', () => {
		const ids = new Set(Array.from({ length: 100 }, () => generateId()));
		expect(ids.size).toBe(100);
	});
});
