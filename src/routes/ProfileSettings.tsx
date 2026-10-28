import { createSignal, onSettled, Show } from 'solid-js';
import * as stylex from '@stylexjs/stylex';

export function ProfileSettings() {
  const base = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');
  const [name, setName] = createSignal('');
  const [bio, setBio] = createSignal('');
  const [visible, setVisible] = createSignal(false);
  const [handle, setHandle] = createSignal('');
  const [message, setMessage] = createSignal('');
  const [loginRequired, setLoginRequired] = createSignal(false);
  const [busy, setBusy] = createSignal(false);
  async function load(signal?: AbortSignal) {
    const res = await fetch(`${base}/api/me/social-profile`, { credentials: 'include', signal });
    if (res.status === 401) { setLoginRequired(true); return; }
    if (res.status === 404) return;
    if (!res.ok) throw new Error('Profile settings are unavailable.');
    const data = await res.json();
    setName(data.displayName); setBio(data.bio); setVisible(data.public); setHandle(data.handle);
  }
  onSettled(() => {
    const controller = new AbortController();
    void load(controller.signal).catch(() => { if (!controller.signal.aborted) setMessage('Could not load your profile.'); });
    return () => controller.abort();
  });
  async function save(event: SubmitEvent) {
    event.preventDefault(); setBusy(true); setMessage('');
    try {
      const res = await fetch(`${base}/api/me/social-profile`, {
        method: 'PUT', credentials: 'include', headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ displayName: name(), bio: bio(), public: visible() }),
      });
      if (res.status === 401) { setLoginRequired(true); return; }
      if (!res.ok) throw new Error('Could not save your profile.');
      await load(); setMessage('Profile saved.');
    } catch (error) { setMessage(error instanceof Error ? error.message : 'Could not save your profile.'); }
    finally { setBusy(false); }
  }
  return <main {...stylex.props(styles.page)}>
    <a href="/">← Cleartify</a><h1>Your profile</h1>
    <Show when={loginRequired()} fallback={
      <form onSubmit={save} {...stylex.props(styles.form)}>
        <label>Display name<input required maxlength="160" value={name()} onInput={e => setName(e.currentTarget.value)} /></label>
        <label>About you<textarea maxlength="2000" value={bio()} onInput={e => setBio(e.currentTarget.value)} /></label>
        <label><input type="checkbox" checked={visible()} onChange={e => setVisible(e.currentTarget.checked)} /> Make my profile public</label>
        <p>Your profile stays private until you choose to share it.</p>
        <button disabled={busy()}>{busy() ? 'Saving…' : 'Save profile'}</button>
      </form>
    }><p>Sign in with your AT Protocol account to set up your profile.</p><a href={`${base}/oauth/login`}>Sign in</a></Show>
    <p role="status">{message()}</p>
    <Show when={handle() && visible()}><a href={`/@${handle()}`}>View your public profile →</a></Show>
  </main>;
}
const styles = stylex.create({
  page: { minHeight: '100vh', padding: '40px 26px', backgroundColor: '#fffaf5', color: '#111936' },
  form: { maxWidth: 560, display: 'grid', gap: 20 },
});
