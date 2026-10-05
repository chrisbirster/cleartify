import { For, createSignal } from 'solid-js';
import * as stylex from '@stylexjs/stylex';

const features = [
  ['💬', 'Feels like texting a friend', 'Natural conversations, not clunky commands.'],
  ['🧠', 'Remembers your chaos', 'Learns what matters to you and how you like things done.'],
  ['✅', 'Actually gets things done', 'Follows through until the task is complete.'],
  ['✨', 'Asks before the important stuff', 'Clear keeps you in control of consequential actions.'],
] as const;

const starterProtocols = [
  ['📅', 'Weekly Reset', 'Reviews your week, clears loose ends, and sets priorities.'],
  ['✈️', 'Trip Planning', 'Researches the trip, organizes the details, and follows up.'],
  ['🚀', 'Ship It', 'Turns an idea into a plan and keeps the work moving.'],
] as const;

export function Home() {
  const [joined, setJoined] = createSignal(false);
  const [submitting, setSubmitting] = createSignal(false);

  async function submitWaitlist(event: SubmitEvent) {
    event.preventDefault();
    const form = event.currentTarget as HTMLFormElement;
    const data = new FormData(form);

    setSubmitting(true);
    try {
      const response = await fetch('/api/waitlist', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          name: data.get('name'),
          email: data.get('email'),
          phone: data.get('phone'),
          emailUpdates: data.get('emailUpdates') === 'on',
        }),
      });

      if (!response.ok) throw new Error('waitlist request failed');
      setJoined(true);
      form.reset();
    } catch {
      // M0 has no API yet. Keep the visual flow usable while the backend lands.
      setJoined(true);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main {...stylex.props(styles.page)}>
      <nav {...stylex.props(styles.nav)}>
        <a href="#top" {...stylex.props(styles.brand)}>✦ Cleartify</a>
        <div {...stylex.props(styles.navLinks)}>
          <a href="#how">How it works</a>
          <a href="#protocols">Protocols</a>
          <a href="/@chris">Profiles</a>
          <a href="#privacy">Privacy</a>
        </div>
        <a href="#waitlist" {...stylex.props(styles.navCta)}>Join the waitlist →</a>
      </nav>

      <section id="top" {...stylex.props(styles.hero)}>
        <div {...stylex.props(styles.heroCopy)}>
          <div {...stylex.props(styles.eyebrow)}>YOUR LIFE, WITH BACKUP</div>
          <h1 {...stylex.props(styles.heroTitle)}>
            Text Clear.<br />
            <span {...stylex.props(styles.highlight)}>Stop raw-dogging your life.</span>
          </h1>
          <p {...stylex.props(styles.lede)}>
            Cleartify is the personal AI assistant you text. It remembers your life,
            follows through, and learns how you like things done.
          </p>
          <div {...stylex.props(styles.heroActions)}>
            <a href="#waitlist" {...stylex.props(styles.primaryButton)}>Get early access →</a>
            <span {...stylex.props(styles.scribble)}>No app. Just text it. ↗</span>
          </div>
          <div {...stylex.props(styles.quote)}>“Can you handle this?” — <strong>Clearly.</strong></div>
        </div>

        <div {...stylex.props(styles.phoneWrap)}>
          <span {...stylex.props(styles.stickerPink)}>no app lol</span>
          <span {...stylex.props(styles.stickerYellow)}>literally<br />just text it</span>
          <span {...stylex.props(styles.stickerBlue)}>main character<br />support ✨</span>
          <div {...stylex.props(styles.phone)}>
            <div {...stylex.props(styles.phoneTop)}>9:41 <span>●●●</span></div>
            <div {...stylex.props(styles.contact)}><span {...stylex.props(styles.avatar)}>C</span><strong>Clear ✨</strong></div>
            <div {...stylex.props(styles.messages)}>
              <div {...stylex.props(styles.userBubble)}>did i forget anything today?<br />i'm overwhelmed lol</div>
              <div {...stylex.props(styles.clearBubble)}>
                You're not behind. Here's what matters:<br /><br />
                • 10am — team sync<br />
                • proposal due today<br />
                • move gym to 6?<br /><br />
                I can draft the proposal and move the calendar block. Want me to?
              </div>
              <div {...stylex.props(styles.userBubble)}>yes pls</div>
              <div {...stylex.props(styles.clearBubble)}>On it. Clearly. 💪</div>
            </div>
            <div {...stylex.props(styles.messageInput)}>Text Clear… <span>↑</span></div>
          </div>
        </div>
      </section>

      <section {...stylex.props(styles.quickBar)}>
        <div>⚡ <strong>No app required</strong><small>Just text. It's that simple.</small></div>
        <div>💗 <strong>Feels like a real assistant</strong><small>Helpful, proactive, and actually gets you.</small></div>
        <div>🎉 <strong>You're always in control</strong><small>Clear acts within the permissions you choose.</small></div>
      </section>

      <section {...stylex.props(styles.section)}>
        <div {...stylex.props(styles.eyebrow)}>A MORE CAPABLE YOU</div>
        <h2 {...stylex.props(styles.sectionTitle)}>Why people are into this.</h2>
        <div {...stylex.props(styles.cardGrid)}>
          <For each={features}>{([icon, title, body], index) => (
            <article {...stylex.props(styles.featureCard, cardTone[index() % cardTone.length])}>
              <div {...stylex.props(styles.cardIcon)}>{icon}</div>
              <h3>{title}</h3>
              <p>{body}</p>
            </article>
          )}</For>
        </div>
      </section>

      <section id="protocols" {...stylex.props(styles.protocolSection)}>
        <div {...stylex.props(styles.protocolHeader)}>
          <div>
            <div {...stylex.props(styles.eyebrow)}>WAYS OF DOING THINGS</div>
            <h2 {...stylex.props(styles.sectionTitle)}>Protocols for real life.</h2>
            <p {...stylex.props(styles.sectionCopy)}>
              Protocols are reusable playbooks Clear can run for you. Everyone starts with useful basics.
              Power users and creators can publish their own.
            </p>
          </div>
          <div {...stylex.props(styles.bigSticker)}>Like having a<br />Type A bestie<br />in your pocket. 🙂</div>
        </div>
        <div {...stylex.props(styles.protocolGrid)}>
          <For each={starterProtocols}>{([icon, title, body]) => (
            <article {...stylex.props(styles.protocolCard)}>
              <div {...stylex.props(styles.cardIcon)}>{icon}</div>
              <h3>{title}</h3>
              <p>{body}</p>
              <button type="button" {...stylex.props(styles.installButton)}>Included ✓</button>
            </article>
          )}</For>
        </div>
        <div {...stylex.props(styles.socialCallout)}>
          <div>
            <strong>People have profiles, too.</strong>
            <p>See what protocols creators use, follow them, install public ones, or subscribe to premium ones.</p>
          </div>
          <a href="/@chris" {...stylex.props(styles.secondaryButton)}>Meet Chris, your first friend →</a>
        </div>
      </section>

      <section id="how" {...stylex.props(styles.section)}>
        <h2 {...stylex.props(styles.sectionTitle)}>Three steps. Zero agent management.</h2>
        <div {...stylex.props(styles.steps)}>
          <div><span>1</span><strong>Text Clear</strong><p>Save Clear in your contacts and talk normally.</p></div>
          <div><span>2</span><strong>Connect what you trust</strong><p>Email, calendar, services, and protocols are opt-in.</p></div>
          <div><span>3</span><strong>Clear follows through</strong><p>Specialist agents and tools stay behind the scenes.</p></div>
        </div>
      </section>

      <section id="privacy" {...stylex.props(styles.trustSection)}>
        <div>
          <div {...stylex.props(styles.eyebrow)}>YOUR LIFE, YOURS</div>
          <h2 {...stylex.props(styles.sectionTitle)}>Private by design. 🔒</h2>
          <p {...stylex.props(styles.sectionCopy)}>Clear works for you. Permissions are explicit, important actions require approval, and public profile activity is opt-in.</p>
        </div>
        <div {...stylex.props(styles.trustPoints)}>
          <div>🛡️ <strong>You choose the permissions</strong></div>
          <div>🔐 <strong>High-stakes actions require approval</strong></div>
          <div>👁️‍🗨️ <strong>Your private protocols stay private</strong></div>
        </div>
      </section>

      <section id="waitlist" {...stylex.props(styles.waitlist)}>
        <div>
          <div {...stylex.props(styles.eyebrow)}>READY WHEN YOU ARE</div>
          <h2 {...stylex.props(styles.waitlistTitle)}>A calmer, more organized you is a text away.</h2>
          <p>Join the beta list. We'll send product updates by email, then invite selected testers with a one-time beta link and code.</p>
        </div>
        {joined() ? (
          <div {...stylex.props(styles.successBox)}>🎉 You're on the list. Clearly.</div>
        ) : (
          <form onSubmit={submitWaitlist} {...stylex.props(styles.form)}>
            <input name="name" required placeholder="Your name" />
            <input name="email" required type="email" placeholder="you@email.com" />
            <input name="phone" required type="tel" placeholder="Mobile number" />
            <label {...stylex.props(styles.checkbox)}><input name="emailUpdates" type="checkbox" checked /> Send me product/newsletter updates.</label>
            <button disabled={submitting()} {...stylex.props(styles.primaryButton)}>{submitting() ? 'Joining…' : 'Join the waitlist →'}</button>
          </form>
        )}
      </section>

      <footer {...stylex.props(styles.footer)}>
        <strong>✦ Cleartify</strong>
        <span>Less to juggle. More you.</span>
        <span>Clearly.</span>
      </footer>
    </main>
  );
}

const cardTone = [
  stylex.create({ x: { backgroundColor: '#f8eaff' } }).x,
  stylex.create({ x: { backgroundColor: '#e7fbf2' } }).x,
  stylex.create({ x: { backgroundColor: '#e7f5ff' } }).x,
  stylex.create({ x: { backgroundColor: '#fff1dc' } }).x,
];

const styles = stylex.create({
  page: { minHeight: '100vh', overflow: 'hidden' },
  nav: { maxWidth: 1180, margin: '0 auto', padding: '20px 28px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 24 },
  brand: { fontSize: 25, fontWeight: 900, letterSpacing: '-0.04em' },
  navLinks: { display: { default: 'flex', '@media (max-width: 760px)': 'none' }, gap: 28, fontSize: 14, fontWeight: 700 },
  navCta: { backgroundColor: '#12b886', color: '#fff', padding: '12px 18px', borderRadius: 999, fontWeight: 800, fontSize: 14 },
  hero: { maxWidth: 1180, margin: '0 auto', minHeight: 620, padding: '52px 28px 70px', display: 'grid', gridTemplateColumns: { default: '1.05fr .95fr', '@media (max-width: 860px)': '1fr' }, gap: 48, alignItems: 'center' },
  heroCopy: { position: 'relative', zIndex: 2 },
  eyebrow: { fontSize: 12, fontWeight: 900, letterSpacing: '.18em', color: '#087f5b', marginBottom: 12 },
  heroTitle: { margin: 0, fontSize: { default: 72, '@media (max-width: 560px)': 49 }, lineHeight: .98, letterSpacing: '-0.065em', fontWeight: 950 },
  highlight: { backgroundColor: '#ffe873', borderRadius: 18, padding: '0 10px 6px', boxDecorationBreak: 'clone' },
  lede: { maxWidth: 610, fontSize: 20, lineHeight: 1.6, color: '#39445f', marginTop: 30 },
  heroActions: { display: 'flex', flexWrap: 'wrap', gap: 22, alignItems: 'center', marginTop: 30 },
  primaryButton: { border: 0, cursor: 'pointer', display: 'inline-flex', justifyContent: 'center', backgroundColor: '#0ca678', color: '#fff', padding: '15px 23px', borderRadius: 14, fontWeight: 900, boxShadow: '0 10px 26px rgba(12,166,120,.18)' },
  secondaryButton: { display: 'inline-flex', padding: '12px 17px', borderRadius: 12, backgroundColor: '#fff', border: '1px solid #dfe4ea', fontWeight: 800 },
  scribble: { fontFamily: 'Comic Sans MS, Marker Felt, cursive', transform: 'rotate(-3deg)', fontWeight: 700 },
  quote: { marginTop: 30, color: '#596277', fontSize: 14 },
  phoneWrap: { position: 'relative', display: 'grid', placeItems: 'center', minHeight: 520 },
  phone: { width: 330, maxWidth: '82vw', minHeight: 555, backgroundColor: '#f8f9fb', border: '10px solid #15171d', borderRadius: 54, boxShadow: '0 34px 80px rgba(18,25,45,.24)', padding: '17px 15px 18px', transform: 'rotate(2deg)', position: 'relative', zIndex: 2 },
  phoneTop: { display: 'flex', justifyContent: 'space-between', fontSize: 12, fontWeight: 900, padding: '0 7px 10px' },
  contact: { display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 5, borderBottom: '1px solid #e5e7eb', paddingBottom: 12, fontSize: 13 },
  avatar: { width: 34, height: 34, display: 'grid', placeItems: 'center', borderRadius: '50%', backgroundColor: '#72a8bf', color: '#fff', fontWeight: 900 },
  messages: { display: 'flex', flexDirection: 'column', gap: 12, padding: '24px 4px 18px' },
  userBubble: { alignSelf: 'flex-end', maxWidth: '82%', backgroundColor: '#1687ff', color: '#fff', padding: '10px 12px', borderRadius: '17px 17px 5px 17px', lineHeight: 1.35, fontSize: 13 },
  clearBubble: { alignSelf: 'flex-start', maxWidth: '88%', backgroundColor: '#e9eaed', color: '#111827', padding: '11px 12px', borderRadius: '17px 17px 17px 5px', lineHeight: 1.45, fontSize: 13 },
  messageInput: { position: 'absolute', bottom: 18, left: 18, right: 18, border: '1px solid #d8dce3', color: '#9aa1ad', borderRadius: 999, padding: '10px 12px', display: 'flex', justifyContent: 'space-between', backgroundColor: '#fff' },
  stickerPink: { position: 'absolute', zIndex: 3, top: 35, left: -4, backgroundColor: '#ff9dc3', padding: '9px 13px', borderRadius: 12, fontFamily: 'Comic Sans MS, cursive', fontWeight: 900, transform: 'rotate(-7deg)' },
  stickerYellow: { position: 'absolute', zIndex: 3, top: 210, left: -14, backgroundColor: '#ffe873', padding: '10px 13px', borderRadius: 8, fontFamily: 'Comic Sans MS, cursive', fontWeight: 800, transform: 'rotate(-5deg)' },
  stickerBlue: { position: 'absolute', zIndex: 3, top: 110, right: -10, backgroundColor: '#bfe7ff', padding: '12px 14px', borderRadius: '45% 55% 46% 54%', fontFamily: 'Comic Sans MS, cursive', fontWeight: 900, textAlign: 'center', transform: 'rotate(6deg)' },
  quickBar: { borderTop: '1px solid #e8edf2', borderBottom: '1px solid #e8edf2', backgroundColor: '#f4fbff', display: 'grid', gridTemplateColumns: { default: 'repeat(3,1fr)', '@media (max-width: 700px)': '1fr' }, gap: 1, padding: '17px max(24px, calc((100vw - 1120px)/2))', ':is(> div)': { display: 'grid', gridTemplateColumns: '35px 1fr', columnGap: 8, padding: '8px 24px' }, ':is(small)': { gridColumn: '2', color: '#677085', marginTop: 3 } },
  section: { maxWidth: 1180, margin: '0 auto', padding: '84px 28px' },
  sectionTitle: { margin: 0, fontSize: { default: 45, '@media (max-width: 560px)': 36 }, lineHeight: 1.05, letterSpacing: '-0.045em' },
  sectionCopy: { color: '#536079', lineHeight: 1.7, maxWidth: 720 },
  cardGrid: { display: 'grid', gridTemplateColumns: { default: 'repeat(4,1fr)', '@media (max-width: 900px)': 'repeat(2,1fr)', '@media (max-width: 520px)': '1fr' }, gap: 18, marginTop: 34 },
  featureCard: { borderRadius: 24, padding: 25, minHeight: 210, border: '1px solid rgba(10,20,40,.05)', ':is(h3)': { margin: '16px 0 8px', fontSize: 20 }, ':is(p)': { color: '#4f5b72', lineHeight: 1.55 } },
  cardIcon: { width: 46, height: 46, display: 'grid', placeItems: 'center', borderRadius: 15, backgroundColor: 'rgba(255,255,255,.72)', fontSize: 24 },
  protocolSection: { padding: '84px max(28px, calc((100vw - 1124px)/2))', backgroundColor: '#eaf8ff' },
  protocolHeader: { display: 'grid', gridTemplateColumns: { default: '1fr auto', '@media (max-width: 720px)': '1fr' }, alignItems: 'center', gap: 40 },
  bigSticker: { backgroundColor: '#ffe873', padding: '16px 20px', borderRadius: 12, fontFamily: 'Comic Sans MS, cursive', fontWeight: 900, transform: 'rotate(4deg)', boxShadow: '0 8px 22px rgba(20,30,50,.08)' },
  protocolGrid: { display: 'grid', gridTemplateColumns: { default: 'repeat(3,1fr)', '@media (max-width: 760px)': '1fr' }, gap: 18, marginTop: 30 },
  protocolCard: { backgroundColor: '#fff', borderRadius: 20, padding: 22, boxShadow: '0 8px 26px rgba(49,98,122,.07)', ':is(h3)': { margin: '14px 0 7px' }, ':is(p)': { minHeight: 58, color: '#58657b', lineHeight: 1.5 } },
  installButton: { border: 0, borderRadius: 999, padding: '9px 12px', backgroundColor: '#e9fbf4', color: '#087f5b', fontWeight: 800 },
  socialCallout: { marginTop: 28, backgroundColor: '#fff', borderRadius: 22, padding: 24, display: 'flex', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between', gap: 20, ':is(p)': { color: '#59667a', marginBottom: 0 } },
  steps: { display: 'grid', gridTemplateColumns: { default: 'repeat(3,1fr)', '@media (max-width: 720px)': '1fr' }, gap: 24, marginTop: 35, ':is(> div)': { padding: 22, borderTop: '3px solid #0ca678' }, ':is(span)': { display: 'grid', placeItems: 'center', width: 32, height: 32, borderRadius: '50%', backgroundColor: '#d5f5ea', color: '#087f5b', fontWeight: 900, marginBottom: 14 }, ':is(strong)': { display: 'block', fontSize: 18 }, ':is(p)': { color: '#5c667b', lineHeight: 1.5 } },
  trustSection: { padding: '72px max(28px, calc((100vw - 1124px)/2))', backgroundColor: '#ecfbea', display: 'grid', gridTemplateColumns: { default: '1fr 1fr', '@media (max-width: 760px)': '1fr' }, gap: 45, alignItems: 'center' },
  trustPoints: { display: 'grid', gap: 12, ':is(> div)': { backgroundColor: 'rgba(255,255,255,.7)', borderRadius: 14, padding: 16 } },
  waitlist: { maxWidth: 1124, margin: '70px auto', borderRadius: 30, padding: { default: 46, '@media (max-width: 560px)': 27 }, background: 'linear-gradient(135deg,#e8dfff,#ffdcea)', display: 'grid', gridTemplateColumns: { default: '1fr .9fr', '@media (max-width: 800px)': '1fr' }, gap: 42, alignItems: 'center' },
  waitlistTitle: { fontSize: { default: 44, '@media (max-width: 560px)': 34 }, lineHeight: 1.05, margin: '0 0 14px', letterSpacing: '-.04em' },
  form: { display: 'grid', gap: 12, ':is(input:not([type=checkbox]))': { width: '100%', border: '1px solid rgba(30,40,70,.14)', backgroundColor: '#fff', padding: '14px 15px', borderRadius: 12, outline: 'none' } },
  checkbox: { display: 'flex', gap: 8, alignItems: 'flex-start', fontSize: 13, color: '#596075' },
  successBox: { backgroundColor: '#fff', borderRadius: 18, padding: 25, fontWeight: 900, fontSize: 20 },
  footer: { maxWidth: 1180, margin: '0 auto', padding: '30px 28px 44px', display: 'flex', flexWrap: 'wrap', justifyContent: 'space-between', gap: 20, color: '#536078' },
});
