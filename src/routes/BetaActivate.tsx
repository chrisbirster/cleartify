import { createSignal } from 'solid-js';
import * as stylex from '@stylexjs/stylex';

type Stage = 'invite' | 'code' | 'done';

export function BetaActivate() {
  const apiBase = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');
  const inviteToken = new URLSearchParams(window.location.search).get('token') ?? '';

  const [stage, setStage] = createSignal<Stage>('invite');
  const [activationToken, setActivationToken] = createSignal('');
  const [phoneHint, setPhoneHint] = createSignal('');
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal('');

  async function request(path: string, body: unknown) {
    const response = await fetch(`${apiBase}${path}`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(body),
    });

    const payload = await response.json().catch(() => ({})) as Record<string, unknown>;
    if (!response.ok) {
      throw new Error(
        typeof payload.message === 'string'
          ? payload.message
          : 'Something went wrong. Try again.',
      );
    }
    return payload;
  }

  async function sendCode() {
    if (!inviteToken) {
      setError('This invitation link is missing its token.');
      return;
    }

    setBusy(true);
    setError('');
    try {
      const payload = await request('/api/beta/start', { token: inviteToken });
      setActivationToken(String(payload.activationToken ?? ''));
      setPhoneHint(String(payload.phoneHint ?? 'your phone'));
      setStage('code');
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Could not send your code.');
    } finally {
      setBusy(false);
    }
  }

  async function resendCode() {
    setBusy(true);
    setError('');
    try {
      await request('/api/beta/resend', { activationToken: activationToken() });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Could not resend your code.');
    } finally {
      setBusy(false);
    }
  }

  async function verifyCode(event: SubmitEvent) {
    event.preventDefault();
    const form = event.currentTarget as HTMLFormElement;
    const data = new FormData(form);
    const code = String(data.get('code') ?? '').trim();

    setBusy(true);
    setError('');
    try {
      await request('/api/beta/verify', {
        activationToken: activationToken(),
        code,
      });
      setStage('done');
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Could not verify your code.');
    } finally {
      setBusy(false);
    }
  }

  return (
    <main {...stylex.props(styles.page)}>
      <a href="/" {...stylex.props(styles.brand)}>Cleartify</a>

      <section {...stylex.props(styles.card)}>
        {stage() === 'invite' && (
          <>
            <div {...stylex.props(styles.eyebrow)}>YOU'RE IN</div>
            <h1 {...stylex.props(styles.title)}>Meet Clear.</h1>
            <p {...stylex.props(styles.copy)}>
              One last thing: verify the phone that will text Clear. We won't send the
              five-digit code until you press the button below.
            </p>
            {error() && <div role="alert" {...stylex.props(styles.error)}>{error()}</div>}
            <button disabled={busy() || !inviteToken} onClick={sendCode} {...stylex.props(styles.primary)}>
              {busy() ? 'Sending…' : 'Text me my code →'}
            </button>
          </>
        )}

        {stage() === 'code' && (
          <>
            <div {...stylex.props(styles.eyebrow)}>CHECK YOUR PHONE</div>
            <h1 {...stylex.props(styles.title)}>Five digits. Clearly.</h1>
            <p {...stylex.props(styles.copy)}>
              We sent a beta code to <strong>{phoneHint()}</strong>. It expires in ten minutes.
            </p>
            <form onSubmit={verifyCode} {...stylex.props(styles.form)}>
              <input
                name="code"
                inputmode="numeric"
                autocomplete="one-time-code"
                pattern="[0-9]{5}"
                maxlength="5"
                placeholder="48217"
                required
                {...stylex.props(styles.codeInput)}
              />
              {error() && <div role="alert" {...stylex.props(styles.error)}>{error()}</div>}
              <button disabled={busy()} {...stylex.props(styles.primary)}>
                {busy() ? 'Checking…' : 'Activate Clear →'}
              </button>
            </form>
            <button disabled={busy()} onClick={resendCode} {...stylex.props(styles.secondary)}>
              Send a new code
            </button>
          </>
        )}

        {stage() === 'done' && (
          <>
            <div {...stylex.props(styles.eyebrow)}>CLEAR IS ON</div>
            <h1 {...stylex.props(styles.title)}>You're in.</h1>
            <p {...stylex.props(styles.copy)}>
              Your phone is verified and your starter protocols are installed. Clear's welcome
              text is the beginning of the private assistant experience—no social profile required.
            </p>
            <a href="/" {...stylex.props(styles.primaryLink)}>Back to Cleartify →</a>
          </>
        )}
      </section>

      <p {...stylex.props(styles.note)}>
        Your beta account is private by default. ATProto only enters the picture if you later
        choose to create a public Cleartify Social profile.
      </p>
    </main>
  );
}

const styles = stylex.create({
  page: {
    minHeight: '100vh',
    backgroundColor: '#fbfaf7',
    color: '#111827',
    fontFamily: 'Inter, ui-sans-serif, system-ui, sans-serif',
    padding: '32px 20px 72px',
  },
  brand: {
    display: 'block',
    width: 'fit-content',
    margin: '0 auto 64px',
    color: '#111827',
    fontSize: 22,
    fontWeight: 900,
    textDecoration: 'none',
  },
  card: {
    maxWidth: 560,
    margin: '0 auto',
    backgroundColor: '#ffffff',
    border: '1px solid #e8e5de',
    borderRadius: 28,
    padding: 'clamp(28px, 6vw, 52px)',
    boxShadow: '0 24px 80px rgba(17, 24, 39, 0.08)',
  },
  eyebrow: {
    color: '#159a74',
    fontSize: 12,
    fontWeight: 900,
    letterSpacing: 1.8,
    marginBottom: 14,
  },
  title: {
    margin: '0 0 16px',
    fontSize: 'clamp(40px, 8vw, 64px)',
    lineHeight: 0.98,
    letterSpacing: -2.5,
  },
  copy: {
    color: '#596075',
    fontSize: 17,
    lineHeight: 1.65,
    marginBottom: 28,
  },
  form: {
    display: 'grid',
    gap: 14,
  },
  codeInput: {
    width: '100%',
    boxSizing: 'border-box',
    border: '2px solid #d8d4ca',
    borderRadius: 16,
    padding: '16px 18px',
    fontSize: 30,
    fontWeight: 900,
    letterSpacing: 10,
    textAlign: 'center',
    outline: 'none',
  },
  primary: {
    width: '100%',
    border: 0,
    borderRadius: 999,
    backgroundColor: '#19b889',
    color: '#ffffff',
    padding: '15px 22px',
    fontSize: 16,
    fontWeight: 900,
    cursor: 'pointer',
  },
  primaryLink: {
    display: 'block',
    textAlign: 'center',
    borderRadius: 999,
    backgroundColor: '#19b889',
    color: '#ffffff',
    padding: '15px 22px',
    fontSize: 16,
    fontWeight: 900,
    textDecoration: 'none',
  },
  secondary: {
    width: '100%',
    marginTop: 12,
    border: 0,
    backgroundColor: 'transparent',
    color: '#596075',
    padding: 10,
    fontWeight: 800,
    cursor: 'pointer',
  },
  error: {
    backgroundColor: '#fff2f2',
    border: '1px solid #ffc9c9',
    color: '#c92a2a',
    borderRadius: 12,
    padding: '11px 13px',
    marginBottom: 14,
    fontSize: 14,
    fontWeight: 700,
  },
  note: {
    maxWidth: 520,
    margin: '24px auto 0',
    color: '#7a7f8f',
    textAlign: 'center',
    fontSize: 13,
    lineHeight: 1.6,
  },
});
