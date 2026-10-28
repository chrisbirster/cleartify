import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const component = readFileSync(new URL('./SocialPreview.tsx', import.meta.url), 'utf8');
const router = readFileSync(new URL('../router.tsx', import.meta.url), 'utf8');
const styles = readFileSync(new URL('./SocialPreview.css', import.meta.url), 'utf8');

test('social preview is routed and styled as a standalone demo', () => {
  assert.match(router, /path:\s*['"]\/social['"],\s*component:\s*SocialPreview/);
  assert.match(styles, /social-theme-(?:mint|pink|blue)/);
  assert.match(component, /INTERACTIVE FRONTEND CONCEPT/);
  assert.match(component, /Demo content/);
});

test('sample profiles, playbooks, and installs are explicitly presented as imaginary', () => {
  assert.match(component, /imaginary installs/);
  assert.match(component, /TEXT-ONLY GUIDE · DEMO/);
  assert.match(component, /No account data changed/);
  assert.match(component, /permission approval are required for real installs/);
});

test('demo contains no backend network calls or real authentication', () => {
  assert.doesNotMatch(component, /\b(?:fetch|XMLHttpRequest|WebSocket|EventSource)\s*\(/);
  assert.doesNotMatch(component, /\b(?:axios|graphqlClient|apolloClient)\b/);
  assert.match(component, /createSignal/);
});
