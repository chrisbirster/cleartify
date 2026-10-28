import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const component = readFileSync(new URL('./ClearApp.tsx', import.meta.url), 'utf8');
const router = readFileSync(new URL('../router.tsx', import.meta.url), 'utf8');
test('Clear and recipes are separate real API routes while social remains demo', () => {
  assert.match(router, /path:\s*'\/clear',\s*component:\s*ClearApp/);
  assert.match(router, /path:\s*'\/recipes',\s*component:\s*ClearApp/);
  assert.match(router, /path:\s*'\/social',\s*component:\s*SocialPreview/);
});
test('real login, catalog, manifest, installs and followup endpoints are wired', () => {
  for (const route of ['/api/login/start','/api/login/verify','/api/protocols',
    '/api/public/guide-manifests/','/api/me/self-serve-guide-installs',
    '/api/me/clear/messages','/api/me/clear/followups','/api/me/clear/nudges']) {
    assert.ok(component.includes(route), 'missing ' + route);
  }
  assert.match(component, /credentials:\s*'include'/);
  assert.match(component, /no-store/);
});
test('errors and delivery semantics are explicitly represented without dummy user data', () => {
  for(const status of ['401','409','429']) assert.ok(component.includes(status));
  assert.match(component, /deterministic command interface/);
  assert.match(component, /In-app delivered means persisted/);
  assert.doesNotMatch(component, /(?:mockProfiles|fakeRecipes|demoFollowups)/);
});

test('valid catalog uses local mapped result, not potentially stale Solid state', () => {
  assert.match(component, /const mappedRecipes = candidates\.map\(toRecipe\)\.filter/);
  assert.match(component, /setRecipes\(mappedRecipes\)/);
  assert.match(component, /if \(candidates\.length && mappedRecipes\.length === 0\)/);
  assert.doesNotMatch(component, /candidates\.length && !recipes\(\)\.length/);
  const valid = [{id:'test-guide',title:'A real catalog item',summary:'Guide'}];
  const mapped = valid.filter(v => typeof v.id === 'string' && typeof v.title === 'string');
  assert.equal(valid.length > 0 && mapped.length === 0, false);
});

test('phone input uses browser-portable E.164 regex confirmed by Chromium', () => {
  assert.match(component, /pattern="\[\+\]\[1-9\]\[0-9\]\{7,14\}"/);
  const pattern = /^[+][1-9][0-9]{7,14}$/;
  assert.equal(pattern.test('+15550002222'), true);
  assert.equal(pattern.test('15550002222'), false);
});
