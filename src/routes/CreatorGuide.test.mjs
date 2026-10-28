import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
const ui=readFileSync(new URL('./CreatorGuide.tsx',import.meta.url),'utf8');
const router=readFileSync(new URL('../router.tsx',import.meta.url),'utf8');

test('creator route uses real OAuth session; no client-provided actor DID',()=>{
 assert.match(router,/path:\s*'\/creator',\s*component:\s*CreatorGuide/);
 assert.match(ui,/href="\/oauth\/login"/);
 assert.match(ui,/credentials:\s*'include'/);
 assert.doesNotMatch(ui,/actorDid|creatorDid\s*:/);
});
test('API draft and publication are separate; public requires private transition and consent',()=>{
 assert.match(ui,/\/api\/me\/protocol-drafts'/);
 assert.match(ui,/\/visibility'/);
 assert.match(ui,/setVisibility\('private'\)/);
 assert.match(ui,/!privateSaved\(\)\s*\|\|\s*!publicConsent\(\)/);
 assert.match(ui,/changeVisibility\('private'\)/);
 assert.match(ui,/No agent code executes/);
});
test('creator errors are truthful about flag-off, OAuth required and conflicts',()=>{
 for(const status of ['401','404','503','409']) assert.ok(ui.includes(status));
 assert.match(ui,/no real Bluesky posting/i);
});

test('large total-size client warning is advisory, never a new submit gate',()=>{
  assert.match(ui,/new TextEncoder\(\)\.encode\(JSON\.stringify/);
  assert.match(ui,/requestSize > 64 \* 1024/);
  assert.match(ui,/preview and submission remain available/);
  assert.match(ui,/setSizeWarning\(/);
  assert.doesNotMatch(ui,/if\s*\(requestSize\s*>\s*64\s*\*\s*1024\)\s*return/);
});
