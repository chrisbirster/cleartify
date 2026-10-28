import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
const read = (file) => readFileSync(new URL(file, import.meta.url), 'utf8');

test('Vite development proxy stays on loopback and preserves relative /api paths', () => {
  const config = read('../vite.config.ts');
  assert.match(config, /127\.0\.0\.1:8080/);
  assert.match(config, /'\/api'/);
  assert.match(config, /strictPort: true/);
});
test('manual local journey is explicit about synthetic publication and real phone challenge', () => {
  const guide = read('../docs/local-clear-browser-journey.md');
  assert.match(guide, /synthetic/i);
  assert.match(guide, /not a real ATProto-authenticated creator publication/i);
  assert.match(guide, /no authentication bypass/i);
  assert.match(guide, /deterministic commands only/i);
});
test('fixture refuses execution without explicit local acknowledgment', () => {
  const run = spawnSync(process.execPath, [new URL('./dev-seed-guide.mjs', import.meta.url).pathname], { encoding: 'utf8', cwd: new URL('..', import.meta.url).pathname });
  assert.equal(run.status, 2);
  assert.match(run.stderr, /acknowledgment required/i);
});
