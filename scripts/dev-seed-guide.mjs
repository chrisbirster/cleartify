// OFFLINE DEVELOPMENT FIXTURE ONLY. Must never run against a real account DB.
import { DatabaseSync } from 'node:sqlite';
import { resolve, relative } from 'node:path';
import { existsSync, realpathSync } from 'node:fs';

const root = resolve(process.cwd(), '.local');
const file = resolve(root, 'cleartify-dev.db');
if (process.argv[2] !== '--i-understand-local-fixtures-only') {
  console.error('Explicit local-fixture acknowledgment required.');
  process.exit(2);
}
if (!existsSync(file) || !realpathSync(file).startsWith(root + '/')) {
  console.error('Expected an existing migrated SQLite DB inside .local/. Refusing other paths.');
  process.exit(2);
}
if (relative(root, realpathSync(file)).startsWith('..')) process.exit(2);

const db = new DatabaseSync(file);
try {
  const now = new Date().toISOString();
  const did = 'did:plc:aaaaaaaaaaaaaaaaaaaaaaaa';
  db.exec('BEGIN IMMEDIATE');
  for (const [id, phone] of [['local-creator-a', '+15550001111'], ['local-reader-b', '+15550002222']]) {
    db.prepare(`INSERT OR IGNORE INTO users(id,display_name,primary_email,status,created_at,updated_at)
                VALUES(?,?,?,'active',?,?)`).run(id, id, id + '@example.invalid', now, now);
    db.prepare(`INSERT OR IGNORE INTO phone_identities(id,user_id,phone_e164,verified_at,created_at)
                VALUES(?,?,?,?,?)`).run('phone-' + id, id, phone, now, now);
  }
  // Explicit synthetic public guide publication, not a creator/OAuth demonstration.
  const spec = JSON.stringify({ steps: [{ operation: 'guide', text: 'Take one small step', dueAt: new Date(Date.now() - 60_000).toISOString() }], permissions: [] });
  db.prepare(`INSERT OR IGNORE INTO published_protocols
    (id,creator_did,slug,version,title,summary,license,spec_json,status,created_at)
    VALUES(?,?,?,?,?,?,?,?,'published',?)`).run('local-guide-001',did,'local-guide-001',1,'Local demo: one small step','Synthetic local-only guide. Not a real public creator publication.','CC0-1.0',spec,now);
  db.prepare(`INSERT OR IGNORE INTO creator_publications
    (protocol_id,visibility,original_source_id,provenance_note,approved_at,updated_at)
    VALUES(?,'public','','synthetic local development fixture',?,?)`).run('local-guide-001',now,now);
  db.exec('COMMIT');
  console.log('LOCAL FIXTURE ONLY: A creator publication + B verified-phone login test identity seeded.');
  console.log('B phone: +15550002222; login requires a real challenge from the loopback Debug Sender.');
  console.log('A publication is seeded offline; the browser does NOT implement ATProto creator publishing.');
} catch (error) {
  try { db.exec('ROLLBACK'); } catch {}
  console.error('Fixture failed. Confirm migrations ran first; no real data should be used.');
  process.exitCode = 1;
} finally { db.close(); }
