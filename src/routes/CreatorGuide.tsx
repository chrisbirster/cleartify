import { createSignal, For, Show } from 'solid-js';
import './CreatorGuide.css';

type Step = { key: number; text: string; dueAt: string };
type Draft = { id: string; slug: string; title: string; summary: string; license: string; version: number; steps: Step[] };
type Json = Record<string, unknown>;
class CreatorError extends Error {
  constructor(public code: number, detail: string) {
    super(code === 401 ? 'A verified Bluesky / AT Protocol creator session is required. Sign in using the real OAuth flow.' :
      code === 404 || code === 503 ? 'Creator drafts or publishing are disabled or unavailable. Check local feature flags.' :
      code === 409 ? 'This slug/version or visibility transition conflicts with an existing publication.' :
      code === 403 ? 'Origin or creator permission check failed.' :
      code === 400 ? 'Draft fields, guide steps, license or provenance were rejected by the server.' :
      'Request failed: ' + detail);
  }
}
const apiBase = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');
async function request(path: string, body: Json): Promise<Json> {
  const response = await fetch(apiBase + path, { method: 'POST', credentials: 'include',
    headers: { 'content-type': 'application/json' }, body: JSON.stringify(body), cache: 'no-store' });
  const result = await response.json().catch(() => ({})) as Json;
  if (!response.ok) throw new CreatorError(response.status, String(result.error ?? 'unavailable'));
  return result;
}
const message = (cause: unknown) => cause instanceof Error ? cause.message : 'Unable to reach the creator API.';
function validSlug(value: string) { return /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value) && value.length <= 100; }

export function CreatorGuide() {
  const [title, setTitle] = createSignal('');
  const [slug, setSlug] = createSignal('');
  const [summary, setSummary] = createSignal('');
  const [license, setLicense] = createSignal('CC0-1.0');
  const [version, setVersion] = createSignal(1);
  const [steps, setSteps] = createSignal<Step[]>([{ key: 1, text: '', dueAt: '' }]);
  const [nextKey, setNextKey] = createSignal(2);
  const [preview, setPreview] = createSignal<Draft | null>(null);
  const [draftId, setDraftId] = createSignal('');
  const [visibility, setVisibility] = createSignal<'private' | 'unlisted' | 'public'>('private');
  const [privateSaved, setPrivateSaved] = createSignal(false);
  const [publicConsent, setPublicConsent] = createSignal(false);
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal('');
  const [notice, setNotice] = createSignal('');
  const [sizeWarning, setSizeWarning] = createSignal('');
  const [published, setPublished] = createSignal(false);

  const change = () => {
    setPreview(null);setPublicConsent(false);
    if (draftId()) setNotice('Draft already created: edits require a new slug/version. The saved draft is immutable.');
  };
  const editStep = (key: number, field: 'text' | 'dueAt', value: string) => {
    setSteps(all => all.map(s => s.key === key ? { ...s, [field]: value } : s));change();
  };
  function build(): Draft | null {
    setError('');
    const clean = steps().map(s => ({ ...s, text: s.text.trim() }));
    if (!title().trim() || title().length > 160 || !validSlug(slug()) || !summary().trim() || summary().length > 1000 ||
      clean.length < 1 || clean.length > 32 || clean.some(s => !s.text || s.text.length > 4096)) {
      setError('Enter a title, a lowercase hyphenated slug, a summary, and 1–32 nonempty text steps.');return null;
    }
    if (clean.some(s => s.dueAt && !Number.isFinite(Date.parse(s.dueAt)))) {
      setError('Invalid due date; choose a date/time or leave the step unscheduled.');return null;
    }
    // Advisory only. The server remains authoritative about request-size limits.
    const requestSize = new TextEncoder().encode(JSON.stringify({
      slug: slug().trim(), version: version(), title: title().trim(),
      summary: summary().trim(), license: license(),
      spec: { steps: clean.map(s => ({ operation: 'guide', text: s.text,
        ...(s.dueAt ? { dueAt: new Date(s.dueAt).toISOString() } : {}) })), permissions: [] },
    })).byteLength;
    setSizeWarning(requestSize > 64 * 1024 ?
      'Large guide request (' + Math.ceil(requestSize / 1024) + ' KiB). The server may reject it; preview and submission remain available.' : '');
    return { id: '', title: title().trim(), slug: slug().trim(), summary: summary().trim(),
      license: license(), version: version(), steps: clean };
  }
  function makePreview(event: SubmitEvent) {
    event.preventDefault();
    const candidate = build();
    if (candidate) {setPreview(candidate);setNotice('Preview only. Nothing has been submitted or published.');}
  }
  async function saveDraft() {
    const p = preview(); if (!p || busy()) return;
    setBusy(true);setError('');setNotice('');
    try {
      const result = await request('/api/me/protocol-drafts', {
        slug: p.slug,version: p.version,title: p.title,summary: p.summary,license: p.license,
        spec: {steps: p.steps.map(s => ({operation:'guide',text:s.text,
          ...(s.dueAt ? {dueAt:new Date(s.dueAt).toISOString()} : {})})),permissions:[]},
      });
      const id=result.id;
      if(typeof id !== 'string' || !id) throw new Error('The creator API did not return a draft ID.');
      setDraftId(id);
      setVisibility('private');setNotice('Draft saved. It is not published or publicly discoverable.');
    } catch(cause) {setError(message(cause));}
    finally {setBusy(false);}
  }
  async function changeVisibility(next: 'private' | 'unlisted' | 'public') {
    if (!draftId() || busy()) return;
    if(next === 'public' && (!privateSaved() || !publicConsent())) {
      setError('First save privately, then explicitly confirm that the guide may be shared publicly.');return;
    }
    setBusy(true);setError('');setNotice('');
    try {
      const result=await request('/api/me/protocol-drafts/'+encodeURIComponent(draftId())+'/visibility', {
        visibility:next,
        provenance:'Original text guide authored and consented to by this verified creator.',
      });
      setVisibility(next);
      if(next==='private') {setPrivateSaved(true);setNotice('Private visibility recorded. Only you can move it to public or unlisted.');}
      else {setPublished(true);setNotice('Visibility accepted by server: '+String(result.visibility ?? next)+'. This is an actual publication transition, not a demo action.');}
    } catch(cause) {setError(message(cause));}
    finally {setBusy(false);}
  }
  return <main class="creator-page">
    <header class="creator-header"><a href="/">✳ cleartify</a><nav><a href="/clear">Clear</a><a href="/recipes">Recipes</a><a href="/social">Social demo</a></nav></header>
    <div class="creator-shell">
      <section class="creator-main">
        <p class="creator-eyebrow">CREATOR STUDIO · LIVE API · VERIFIED OAUTH REQUIRED</p>
        <h1>Make a guide worth sharing.</h1>
        <p>Write simple, inert text steps. Preview locally, save as a draft, record <strong>private</strong> visibility first, then choose whether to share. No agent code executes.</p>
        <div class="creator-auth"><strong>Creator identity</strong><p>This page never creates a fake account or accepts a DID from a form. The server resolves your identity only from the existing verified AT Protocol OAuth session.</p><a href="/oauth/login">Sign in with Bluesky / AT Protocol →</a><small>If OAuth is disabled or configured for a public HTTPS origin, local login may not be available. A phone Clear session does not count as creator OAuth.</small></div>
        <form onSubmit={makePreview} class="creator-form">
          <label>Title<input required maxlength="160" value={title()} disabled={!!draftId()} onInput={e=>{setTitle(e.currentTarget.value);change();}} placeholder="My tiny weekly reset"/></label>
          <label>Slug<input required pattern="[a-z0-9]+(-[a-z0-9]+)*" maxlength="100" value={slug()} disabled={!!draftId()} onInput={e=>{setSlug(e.currentTarget.value);change();}} placeholder="tiny-weekly-reset"/></label>
          <label>Summary<textarea required maxlength="1000" value={summary()} disabled={!!draftId()} onInput={e=>{setSummary(e.currentTarget.value);change();}} placeholder="What the reader gets from this guide"/></label>
          <div class="creator-two"><label>Version<input type="number" min="1" max="1000000" value={version()} disabled={!!draftId()} onInput={e=>{setVersion(Number(e.currentTarget.value));change();}}/></label><label>Open license<select value={license()} disabled={!!draftId()} onChange={e=>{setLicense(e.currentTarget.value);change();}}><option value="CC0-1.0">CC0 1.0</option><option value="CC-BY-4.0">CC BY 4.0</option><option value="MIT">MIT</option><option value="Apache-2.0">Apache 2.0</option></select></label></div>
          <h2>Text steps <span>(1–32)</span></h2>
          <For each={steps()}>{(step,index)=><fieldset class="creator-step"><legend>Step {index()+1}</legend><label>Instruction<textarea required maxlength="4096" disabled={!!draftId()} value={step.text} onInput={e=>editStep(step.key,'text',e.currentTarget.value)} placeholder="Write one helpful action…"/></label><label>Optional due time (local)<input type="datetime-local" disabled={!!draftId()} value={step.dueAt} onInput={e=>editStep(step.key,'dueAt',e.currentTarget.value)} /></label><Show when={steps().length>1&&!draftId()}><button type="button" class="creator-soft" onClick={()=>{setSteps(all=>all.filter(s=>s.key!==step.key));change();}}>Remove step</button></Show></fieldset>}</For>
          <Show when={!draftId()}><button type="button" class="creator-soft" disabled={steps().length>=32} onClick={()=>{setSteps(all=>[...all,{key:nextKey(),text:'',dueAt:''}]);setNextKey(n=>n+1);change();}}>+ Add step</button></Show>
          <button disabled={!!draftId()} type="submit">Preview guide (no API write)</button>
        </form>
      </section>
      <aside class="creator-side">
        <h2>Preview & publishing</h2><Show when={preview()} fallback={<p>Preview your guide before creating an immutable draft.</p>}>{p=><article class="creator-preview"><p class="creator-pill">NOT YET PUBLISHED</p><h3>{p().title}</h3><p>{p().summary}</p><ol><For each={p().steps}>{s=><li>{s.text}<Show when={s.dueAt}><small>Due: {s.dueAt}</small></Show></li>}</For></ol><p>License: {p().license} · version {p().version}</p></article>}</Show>
        <Show when={preview()&&!draftId()}><button onClick={()=>void saveDraft()} disabled={busy()}>Save immutable private draft</button></Show>
        <Show when={draftId()}>
          <p class="creator-pill">Draft ID: <code>{draftId()}</code></p>
          <p>Visibility: <strong>{visibility()}</strong></p>
          <Show when={!privateSaved()}><button disabled={busy()} onClick={()=>void changeVisibility('private')}>1. Record private visibility</button><p>Required before public/unlisted; this cannot be skipped.</p></Show>
          <Show when={privateSaved()&&!published()}>
            <p>2. Make a deliberate visibility choice.</p>
            <label class="creator-check"><input type="checkbox" checked={publicConsent()} onChange={e=>setPublicConsent(e.currentTarget.checked)} /> I own or have rights to share these steps. I understand public means discoverable by anyone and unlisted means accessible to people with the link.</label>
            <div class="creator-actions"><button class="creator-soft" disabled={!publicConsent()||busy()} onClick={()=>void changeVisibility('unlisted')}>Publish unlisted</button><button disabled={!publicConsent()||busy()} onClick={()=>void changeVisibility('public')}>Publish publicly</button></div>
          </Show>
        </Show>
        <Show when={sizeWarning()}><p class="creator-notice" role="status">{sizeWarning()}</p></Show>
        <Show when={error()}><p class="creator-error" role="alert">{error()}</p></Show><Show when={notice()}><p class="creator-notice" role="status">{notice()}</p></Show>
        <hr/><p class="creator-fine">The server validates owner DID, license, provenance, visibility transition, and inert step schema. Public publishing requires explicit consent here and valid OAuth on the server. There is no real Bluesky posting in this page.</p>
      </aside>
    </div>
  </main>;
}
