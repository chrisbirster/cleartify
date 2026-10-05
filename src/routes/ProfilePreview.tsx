import { For } from 'solid-js';
import { useParams } from '@solidjs/router';
import * as stylex from '@stylexjs/stylex';

const protocols = [
  { emoji: '🚀', name: 'Ship It', description: 'Feature → dev → main → tag → verify.', access: 'Free' },
  { emoji: '🧹', name: 'Weekly Reset', description: 'Clear loose ends and decide what actually matters.', access: 'Free' },
  { emoji: '💡', name: 'Weekend MVP', description: 'Keep a new idea tiny enough to ship this weekend.', access: '$4' },
  { emoji: '📚', name: 'Exam Sprint', description: 'Turn a deadline into a daily study and review loop.', access: '$3/mo' },
  { emoji: '🔁', name: 'Follow Through', description: 'Keep chasing a delegated outcome until it is done.', access: 'Free' },
  { emoji: '🧪', name: 'Try It Before We Build It', description: 'Research → prototype → evidence → go/no-go.', access: '$7' },
] as const;

export function ProfilePreview() {
  const params = useParams();
  const isChris = () => params.handle?.toLowerCase() === 'chris';

  return (
    <main {...stylex.props(styles.page)}>
      <nav {...stylex.props(styles.nav)}>
        <a href="/" {...stylex.props(styles.brand)}>✦ Cleartify</a>
        <a href="/" {...stylex.props(styles.back)}>← Back home</a>
      </nav>

      <section {...stylex.props(styles.profile)}>
        <div {...stylex.props(styles.cover)}>
          <span {...stylex.props(styles.coverSticker)}>your first friend ✨</span>
        </div>
        <div {...stylex.props(styles.identity)}>
          <div {...stylex.props(styles.avatar)}>C</div>
          <div>
            <div {...stylex.props(styles.nameRow)}>
              <h1 {...stylex.props(styles.name)}>{isChris() ? 'Chris' : `@${params.handle}`}</h1>
              <span {...stylex.props(styles.claimed)}>✓ claimed</span>
            </div>
            <p {...stylex.props(styles.bio)}>Building weird things, shipping software, and trying to have fewer tabs open in my brain.</p>
            <div {...stylex.props(styles.stats)}>
              <strong>10 protocols running</strong>
              <span>6 public</span>
              <span>2 premium</span>
              <span>your first Cleartify friend</span>
            </div>
          </div>
        </div>

        <div {...stylex.props(styles.columns)}>
          <aside {...stylex.props(styles.about)}>
            <h2>Chris's corner of the internet</h2>
            <p>Think Myspace energy, except instead of music autoplay you can steal the systems that keep someone's life moving.</p>
            <div {...stylex.props(styles.note)}>“Can Clear just handle this?”<br /><strong>Clearly.</strong></div>
            <button {...stylex.props(styles.follow)}>Following ✓</button>
          </aside>

          <section>
            <div {...stylex.props(styles.sectionHeader)}>
              <div>
                <div {...stylex.props(styles.eyebrow)}>WHAT I ACTUALLY USE</div>
                <h2>Chris's protocols</h2>
              </div>
              <span {...stylex.props(styles.liveBadge)}>● 10 running now</span>
            </div>
            <div {...stylex.props(styles.protocolGrid)}>
              <For each={protocols}>{(protocol) => (
                <article {...stylex.props(styles.protocol)}>
                  <div {...stylex.props(styles.icon)}>{protocol.emoji}</div>
                  <div {...stylex.props(styles.protocolMain)}>
                    <h3>{protocol.name}</h3>
                    <p>{protocol.description}</p>
                  </div>
                  <div {...stylex.props(styles.protocolFooter)}>
                    <span>{protocol.access}</span>
                    <button>{protocol.access === 'Free' ? 'Install' : 'Unlock'}</button>
                  </div>
                </article>
              )}</For>
            </div>
          </section>
        </div>
      </section>
    </main>
  );
}

const styles = stylex.create({
  page: { minHeight: '100vh', backgroundColor: '#fffaf5', color: '#111936' },
  nav: { maxWidth: 1080, margin: '0 auto', padding: '22px 26px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' },
  brand: { fontSize: 24, fontWeight: 950, letterSpacing: '-.04em' },
  back: { fontWeight: 800, color: '#566079' },
  profile: { maxWidth: 1080, margin: '0 auto', padding: '0 26px 80px' },
  cover: { height: 220, borderRadius: 28, background: 'linear-gradient(120deg,#bcecff,#ffe57d 48%,#ffb8d6)', position: 'relative', overflow: 'hidden' },
  coverSticker: { position: 'absolute', right: 30, bottom: 28, backgroundColor: '#fff', padding: '11px 15px', borderRadius: 12, fontFamily: 'Comic Sans MS, cursive', fontWeight: 900, transform: 'rotate(4deg)' },
  identity: { display: 'grid', gridTemplateColumns: '120px 1fr', gap: 25, alignItems: 'end', marginTop: -50, padding: '0 28px', position: 'relative' },
  avatar: { width: 120, height: 120, borderRadius: 32, border: '7px solid #fffaf5', backgroundColor: '#0ca678', color: '#fff', display: 'grid', placeItems: 'center', fontWeight: 950, fontSize: 48 },
  nameRow: { display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' },
  name: { margin: 0, fontSize: 43, lineHeight: 1, letterSpacing: '-.05em' },
  claimed: { borderRadius: 999, padding: '6px 10px', backgroundColor: '#d9f8ec', color: '#087f5b', fontWeight: 850, fontSize: 12 },
  bio: { color: '#5a6479', maxWidth: 650, lineHeight: 1.5 },
  stats: { display: 'flex', flexWrap: 'wrap', gap: 15, fontSize: 13, color: '#687187' },
  columns: { display: 'grid', gridTemplateColumns: { default: '290px 1fr', '@media (max-width: 780px)': '1fr' }, gap: 30, marginTop: 54 },
  about: { backgroundColor: '#f5eaff', borderRadius: 23, padding: 23, alignSelf: 'start' },
  note: { margin: '20px 0', padding: 16, borderRadius: 15, backgroundColor: '#fff', fontFamily: 'Comic Sans MS, cursive', transform: 'rotate(-2deg)' },
  follow: { width: '100%', border: 0, borderRadius: 12, padding: 12, backgroundColor: '#111936', color: '#fff', fontWeight: 900 },
  sectionHeader: { display: 'flex', justifyContent: 'space-between', alignItems: 'end', gap: 20, marginBottom: 18 },
  eyebrow: { color: '#087f5b', fontSize: 11, letterSpacing: '.16em', fontWeight: 900 },
  liveBadge: { backgroundColor: '#fff1c2', borderRadius: 999, padding: '8px 11px', fontSize: 12, fontWeight: 900 },
  protocolGrid: { display: 'grid', gridTemplateColumns: { default: 'repeat(2,1fr)', '@media (max-width: 620px)': '1fr' }, gap: 16 },
  protocol: { backgroundColor: '#fff', border: '1px solid #ece7e1', borderRadius: 19, padding: 18, display: 'grid', gridTemplateColumns: '48px 1fr', gap: 12, boxShadow: '0 8px 22px rgba(35,29,24,.04)' },
  icon: { width: 48, height: 48, borderRadius: 14, backgroundColor: '#edf8ff', display: 'grid', placeItems: 'center', fontSize: 23 },
  protocolMain: { minWidth: 0 },
  protocolFooter: { gridColumn: '1 / -1', borderTop: '1px solid #f0ece8', paddingTop: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontWeight: 900, color: '#4e5870' },
});
