import { createSignal, For, Show, onMount } from 'solid-js';
import './ClearApp.css';

// REAL API ONLY: no fixture users, publications, messages, or fake installations.
type Json = Record<string, unknown>;
type Followup = { id: string; text: string; dueAt: string; status: string };
type Nudge = { followupId: string; text: string; dueAt: string; channel: string; deliveryStatus: string };
type Recipe = { id: string; title: string; summary: string; version?: number };
type Manifest = { manifestSha256: string; permissions?: string[]; specSha256?: string; version?: number };
const origin = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');

class ApiError extends Error {
  constructor(public status: number, public key: string) {
    super(status === 404 || status === 503 ? 'This feature is not enabled or is unavailable.' :
      status === 401 ? 'Please sign in to your private Clear account.' :
      status === 409 ? 'This publication changed or is already installed. Refresh before trying again.' :
      status === 429 ? 'You have reached a temporary rate limit. Try again later.' :
      status === 403 ? 'You do not have permission to perform that action.' :
      key || 'Request failed. Please try again.');
  }
}
async function request(path: string, init: RequestInit = {}): Promise<Json> {
  const response = await fetch(origin + path, {
    ...init, credentials: 'include',
    headers: { ...init.headers, ...(init.body ? { 'content-type': 'application/json' } : {}) },
    cache: 'no-store',
  });
  const result = await response.json().catch(() => ({})) as Json;
  if (!response.ok) throw new ApiError(response.status, String(result.error ?? ''));
  return result;
}
function post(path: string, body: Json) { return request(path, { method: 'POST', body: JSON.stringify(body) }); }
function errText(cause: unknown) { return cause instanceof Error ? cause.message : 'Connection failed. Confirm the local server is running.'; }
function asRows<T>(data: Json, key: string): T[] { const v = data[key]; return Array.isArray(v) ? v as T[] : []; }
function toRecipe(value: unknown): Recipe | null {
  if (!value || typeof value !== 'object') return null;
  const item = value as Json;
  if (typeof item.id !== 'string' || typeof item.title !== 'string') return null;
  return { id: item.id, title: item.title, summary: typeof item.summary === 'string' ? item.summary : '', version: typeof item.version === 'number' ? item.version : undefined };
}
function toManifest(value: Json): Manifest | null {
  const obj = value.manifest && typeof value.manifest === 'object' ? value.manifest as Json : value;
  const digest = obj.manifestSha256 ?? obj.manifest_sha256;
  if (typeof digest !== 'string' || !/^[a-f0-9]{64}$/i.test(digest)) return null;
  return { manifestSha256: digest, permissions: Array.isArray(obj.permissions) ? obj.permissions as string[] : [], version: typeof obj.version === 'number' ? obj.version : undefined };
}
export function ClearApp() {
  const [phone, setPhone] = createSignal('');
  const [code, setCode] = createSignal('');
  const [stage, setStage] = createSignal<'phone' | 'verify' | 'signed-in'>('phone');
  const [busy, setBusy] = createSignal(false);
  const [status, setStatus] = createSignal('');
  const [error, setError] = createSignal('');
  const [text, setText] = createSignal('');
  const [dueAt, setDueAt] = createSignal('');
  const [reply, setReply] = createSignal('');
  const [followups, setFollowups] = createSignal<Followup[]>([]);
  const [nudges, setNudges] = createSignal<Nudge[]>([]);
  const [recipes, setRecipes] = createSignal<Recipe[]>([]);
  const [recipeStatus, setRecipeStatus] = createSignal('');
  const [selected, setSelected] = createSignal<Recipe | null>(null);
  const [manifest, setManifest] = createSignal<Manifest | null>(null);
  const [withSteps, setWithSteps] = createSignal(false);

  async function refreshPrivate() {
    const [f, n] = await Promise.all([
      request('/api/me/clear/followups'),
      request('/api/me/clear/nudges'),
    ]);
    setFollowups(asRows<Followup>(f, 'followups'));
    setNudges(asRows<Nudge>(n, 'nudges'));
  }
  async function refreshRecipes() {
    setRecipeStatus('');
    try {
      const result = await request('/api/protocols');
      const candidates = Array.isArray(result.protocols) ? result.protocols : Array.isArray(result.items) ? result.items : Array.isArray(result.results) ? result.results : [];
      setRecipes(candidates.map(toRecipe).filter((v): v is Recipe => v !== null));
      if (candidates.length && !recipes().length) setRecipeStatus('The catalog returned an unsupported shape. No entries were fabricated.');
      if (!candidates.length) setRecipeStatus('No published recipes returned by the API.');
    } catch (cause) {
      setRecipes([]);
      setRecipeStatus(errText(cause));
    }
  }
  async function checkSession() {
    try {
      await request('/api/me/clear-session');
      setStage('signed-in');
      await refreshPrivate();
    } catch (cause) {
      if (cause instanceof ApiError && (cause.status === 401 || cause.status === 403)) setStage('phone');
      else setStatus(errText(cause));
    }
  }
  onMount(() => { void checkSession(); void refreshRecipes(); });

  async function startLogin(event: SubmitEvent) {
    event.preventDefault(); setBusy(true); setError('');
    try {
      await post('/api/login/start', { phone: phone().trim() });
      setStage('verify');
      setStatus('If the phone belongs to an active account, a code was sent. Locally, check the Debug Sender terminal. The response never confirms account existence.');
    } catch (cause) { setError(errText(cause)); } finally { setBusy(false); }
  }
  async function verifyLogin(event: SubmitEvent) {
    event.preventDefault(); setBusy(true); setError('');
    try {
      await post('/api/login/verify', { phone: phone().trim(), code: code().trim() });
      setStage('signed-in'); setStatus('Signed in with a private Clear session.');
      await refreshPrivate();
    } catch (cause) { setError(errText(cause)); } finally { setBusy(false); }
  }
  async function sendMessage(event: SubmitEvent) {
    event.preventDefault(); setBusy(true); setError('');
    const command = text().trim();
    if (!command) { setBusy(false); return; }
    try {
      const result = await post('/api/me/clear/messages', { text: command, ...(dueAt() ? { dueAt: new Date(dueAt()).toISOString() } : {}) });
      setReply(String(result.reply ?? 'Request accepted.'));
      if (result.deliveryStatus) setStatus('Delivery: ' + String(result.deliveryChannel ?? 'unknown') + ' · ' + String(result.deliveryStatus) + '. Creating a follow-up is not proof of delivery.');
      setText(''); setDueAt('');
      await refreshPrivate();
    } catch (cause) { setError(errText(cause)); } finally { setBusy(false); }
  }
  async function openManifest(item: Recipe) {
    setSelected(item); setManifest(null); setRecipeStatus('');
    try {
      const result = await request('/api/protocols/' + encodeURIComponent(item.id) + '/install-manifest');
      const parsed = toManifest(result);
      if (!parsed) throw new Error('The server did not return a valid install manifest digest. Installation is disabled.');
      setManifest(parsed);
    } catch (cause) {
      setRecipeStatus(errText(cause) + ' The manifest endpoint may require additional flags; installation is disabled without a pinned digest.');
    }
  }
  async function install() {
    const item = selected(), m = manifest();
    if (!item || !m) return;
    setBusy(true); setRecipeStatus('');
    try {
      const result = await post('/api/me/self-serve-guide-installs', {
        publicationId: item.id, acknowledgedManifestSha256: m.manifestSha256, createFollowups: withSteps(),
      });
      setRecipeStatus(result.existing ? 'Already installed for this account.' :
        'Installed. Follow-ups created: ' + String(result.followupsCreated ?? 0) + '. Channel: ' + String(result.deliveryChannel ?? 'none') + '; state: ' + String(result.deliveryStatus ?? 'not requested') + '.');
      setSelected(null); setManifest(null);
      await refreshPrivate();
    } catch (cause) { setRecipeStatus(errText(cause)); }
    finally { setBusy(false); }
  }
  return <main class="clear-app">
    <header class="clear-header"><a href="/" class="clear-brand">✳ cleartify</a><nav><a href="/clear">Clear</a><a href="/recipes">Recipes</a><a href="/social">Social demo</a></nav></header>
    <div class="clear-layout">
      <section class="clear-primary" aria-label="Private Clear conversation">
        <p class="clear-eyebrow">PRIVATE · TEXT FIRST · LOCAL PREVIEW</p>
        <h1>Text Clear<span>✳</span></h1>
        <p class="clear-description">A working, deterministic command interface—not a generative chat assistant. Try <code>remind me to call Alex</code> or <code>done [follow-up ID]</code>. Unspecified due time defaults to 24 hours.</p>
        <Show when={stage() !== 'signed-in'}>
          <section class="clear-panel" aria-label="Phone sign in">
            <h2>Sign in to Clear</h2>
            <Show when={stage() === 'phone'} fallback={<form onSubmit={verifyLogin}><label>Five-digit code<input required inputmode="numeric" autocomplete="one-time-code" pattern="[0-9]{5}" maxlength="5" value={code()} onInput={e => setCode(e.currentTarget.value)} placeholder="Code from Debug Sender"/></label><button disabled={busy()} type="submit">Verify code</button><button type="button" class="clear-link" onClick={() => setStage('phone')}>Change phone</button></form>}>
              <form onSubmit={startLogin}><label>Phone (E.164, e.g. +15551234567)<input type="tel" required pattern="\\+[1-9][0-9]{7,14}" value={phone()} onInput={e => setPhone(e.currentTarget.value)} placeholder="+15551234567"/></label><button disabled={busy()} type="submit">Request sign-in code</button></form>
            </Show>
            <p>For existing accounts only. No account or SMS provider is created by this screen.</p>
          </section>
        </Show>
        <Show when={stage() === 'signed-in'}>
          <section class="clear-panel">
            <h2>Your Clear conversation</h2>
            <Show when={reply()}><p role="status" class="clear-reply">{reply()}</p></Show>
            <form onSubmit={sendMessage}>
              <label>Message to Clear<textarea rows={4} maxlength={4096} value={text()} onInput={e => setText(e.currentTarget.value)} placeholder="remind me to check the recipe"/></label>
              <label>Optional due date/time<input type="datetime-local" value={dueAt()} onInput={e => setDueAt(e.currentTarget.value)} /></label>
              <button type="submit" disabled={busy()}>Send command →</button>
            </form>
            <p>Only <code>remind me to …</code> and <code>done ID</code> change follow-ups. No automatic external actions.</p>
          </section>
          <section class="clear-panel">
            <div class="clear-panel-heading"><h2>My follow-ups</h2><button class="clear-secondary" onClick={() => void refreshPrivate().catch(c => setError(errText(c)))}>Refresh</button></div>
            <Show when={followups().length} fallback={<p>No follow-ups returned for this account.</p>}>
              <ul class="clear-list"><For each={followups()}>{item => <li><strong>{item.text}</strong><span>{item.status} · due {item.dueAt}</span><small>ID: {item.id}</small><Show when={item.status === 'open'}><button class="clear-secondary" disabled={busy()} onClick={() => {setText('done ' + item.id);}}>Complete via command ↗</button></Show></li>}</For></ul>
            </Show>
            <h3>My nudges · delivery truth</h3>
            <Show when={nudges().length} fallback={<p>No active delivered or pending nudges in your inbox.</p>}>
              <ul class="clear-list"><For each={nudges()}>{n => <li><strong>{n.text}</strong><span>{n.channel} · {n.deliveryStatus}</span><small>Due {n.dueAt}</small></li>}</For></ul>
            </Show>
            <p>In-app delivered means persisted for display, not read. Debug SMS is local-only; never assume real SMS delivery.</p>
          </section>
        </Show>
        <Show when={error()}><p role="alert" class="clear-error">{error()}</p></Show>
        <Show when={status()}><p role="status" class="clear-note">{status()}</p></Show>
      </section>
      <aside class="clear-aside">
        <p class="clear-eyebrow">REAL PUBLISHED CATALOG · NO SAMPLE DATA</p><h2>Browse recipes</h2>
        <p>Only records returned by the server appear here. Install requires a pinned manifest and an authenticated session.</p>
        <button class="clear-secondary" onClick={() => void refreshRecipes()}>Refresh published recipes</button>
        <Show when={recipeStatus()}><p role="status" class="clear-note">{recipeStatus()}</p></Show>
        <ul class="clear-list"><For each={recipes()}>{item => <li><strong>{item.title}</strong><p>{item.summary}</p><small>{item.id} · version {item.version ?? 'not provided'}</small><button class="clear-secondary" onClick={() => void openManifest(item)}>Preview install</button></li>}</For></ul>
        <Show when={selected()}>{item => <div class="clear-panel"><h3>Install {item().title}?</h3><Show when={manifest()} fallback={<p>Awaiting a valid install manifest. No install without it.</p>}>{m => <><p>Manifest SHA-256: <code class="clear-digest">{m().manifestSha256}</code></p><p>Permissions: {m().permissions?.length ? m().permissions?.join(', ') : 'none declared'}</p><label class="clear-check"><input type="checkbox" checked={withSteps()} onChange={e => setWithSteps(e.currentTarget.checked)} /> Also create follow-ups for guide steps (default due in 24 hours if unspecified)</label><button disabled={busy() || !!m().permissions?.length || stage() !== 'signed-in'} onClick={() => void install()}>Acknowledge manifest and install</button></>}</Show><button class="clear-link" onClick={() => setSelected(null)}>Cancel</button></div>}</Show>
        <p class="clear-demo">Want the nostalgic MySpace direction? <a href="/social">See the clearly labeled social demo</a>.</p>
      </aside>
    </div>
  </main>;
}
