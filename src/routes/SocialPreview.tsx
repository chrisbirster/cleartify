import { createSignal, For, Show } from 'solid-js';
import './SocialPreview.css';

const playbooks = [
  { id: 'weekly', emoji: '🧹', title: 'Sunday reset club', tag: 'LIFE ADMIN', body: 'A 20-minute reset for people with 87 open browser tabs and 0 clean mugs.', installs: '2.4k', tint: 'mint', steps: ['Empty your brain into a list', 'Pick the three things that matter', 'Schedule a small first step'] },
  { id: 'ship', emoji: '🚀', title: 'Ship it, coward', tag: 'BUILDER ENERGY', body: 'Less perfect planning. More actually pressing publish.', installs: '1.8k', tint: 'pink', steps: ['Name the smallest release', 'Cut everything nonessential', 'Ship and check the result'] },
  { id: 'touch', emoji: '🌱', title: 'Touch grass protocol', tag: 'WELLNESS-ISH', body: 'Gentle reminders to leave your desk, drink water, and remember the sun exists.', installs: '963', tint: 'yellow', steps: ['Stand up and stretch', 'Go outside for 10 minutes', 'Tell Clear how you feel'] },
  { id: 'brain', emoji: '🧠', title: 'Big brain study sesh', tag: 'LEARN SOMETHING', body: 'A low-drama study plan for turning "I should learn that" into progress.', installs: '705', tint: 'lavender', steps: ['Choose one concept', 'Study for 25 minutes', 'Explain it in your own words'] },
] as const;

type Tab = 'playbooks' | 'about' | 'friends';
type Playbook = (typeof playbooks)[number];

export function SocialPreview() {
  const [tab, setTab] = createSignal<Tab>('playbooks');
  const [following, setFollowing] = createSignal(true);
  const [selected, setSelected] = createSignal<Playbook | null>(null);
  const [theme, setTheme] = createSignal<'mint' | 'pink' | 'blue'>('mint');
  const [notice, setNotice] = createSignal('');

  function demoInstall(item: Playbook) {
    setNotice(`“${item.title}” is a preview. Sign-in and permission approval are required for real installs.`);
    setSelected(null);
  }

  return (
    <main class={`social-page social-theme-${theme()}`}>
      <div class="social-marquee"><span>✦ welcome to your corner of the internet ✦ no algorithm, just excellent taste ✦ protocols &gt; productivity guilt ✦</span></div>
      <header class="social-nav">
        <a class="social-logo" href="/">✳ clear<span>tify</span><sup>beta</sup></a>
        <nav aria-label="Primary"><a class="social-nav-active" href="/social">People</a><a href="#playbooks">Playbooks</a><a href="/@chris">Real profiles ↗</a></nav>
        <a class="social-nav-button" href="/">Text Clear ↗</a>
      </header>

      <div class="social-shell">
        <div class="social-notice"><span>✦ INTERACTIVE FRONTEND CONCEPT</span> Demo content. Not a real logged-in account, verified activity, or live install catalog.</div>
        <div class="social-columns">
          <aside class="social-sidebar">
            <div class="social-profile-mini">
              <div class="social-mini-avatar">✳</div>
              <div><strong>YOUR SPACE, YOUR RULES</strong><p>Your methods. Your people. Your own little weird corner of the internet.</p></div>
            </div>
            <div class="social-side-section"><span class="social-side-caption">THE CLUBHOUSE</span><a href="#profile">✧ My profile <b>↗</b></a><a href="#playbooks">⚡ Playbooks <b>04</b></a><button onClick={() => { setTab('friends'); }}>♡ My people <b>03</b></button><a href="#about">✦ About Cleartify <b>↗</b></a></div>
            <div class="social-side-section"><span class="social-side-caption">PICK YOUR VIBE (DEMO)</span><div class="social-swatches"><button aria-label="Mint theme" aria-pressed={theme()==='mint'} class="swatch mint" onClick={() => { setTheme('mint'); }}/><button aria-label="Pink theme" aria-pressed={theme()==='pink'} class="swatch pink" onClick={() => { setTheme('pink'); }}/><button aria-label="Blue theme" aria-pressed={theme()==='blue'} class="swatch blue" onClick={() => { setTheme('blue'); }}/></div><small>Profile customization is just a local preview here.</small></div>
            <div class="social-sticker">you are the<br/>main character <span>✳</span></div>
            <div class="social-side-foot">made for humans with<br />too many side quests ☻</div>
          </aside>

          <div id="profile" class="social-content">
            <section class="social-hero">
              <div class="social-hero-pattern" aria-hidden="true">✳　✴　✳　✷　✳　✴　✳　✷</div>
              <div class="social-hero-stamp">CERTIFIED<br/>INTERNET<br/>PERSON ✦</div>
              <div class="social-hero-profile">
                <div class="social-avatar"><span>CB</span><i>✦</i></div>
                <div class="social-identity"><div class="social-eyebrow">WELCOME TO MY LITTLE INTERNET CORNER</div><h1>chris<span>.exe</span> <em>✳</em></h1><p class="social-handle">@chris <span>·</span> creator / side-quest enthusiast</p><p class="social-bio">building useful things, collecting projects like pokémon, and occasionally remembering to go outside.</p><div class="social-badges"><span>🛠 MAKER MODE</span><span>☻ VERY ONLINE</span><span>⚡ CLEARLY.</span></div></div>
              </div>
            </section>

            <div class="social-profile-bar">
              <div class="social-stats"><div><strong>04</strong><small>public playbooks*</small></div><div><strong>03</strong><small>featured friends*</small></div><div><strong>∞</strong><small>side quests</small></div></div>
              <div class="social-actions"><button class="social-follow" onClick={() => { const next = !following(); setFollowing(next); setNotice(next ? 'Demo: following turned on. No account data changed.' : 'Demo: following turned off. No account data changed.'); }}>{following() ? '♥ Following' : '+ Follow'}</button><button class="social-share" onClick={() => {if (navigator.clipboard) { void navigator.clipboard.writeText(window.location.origin + '/social').then(() => setNotice('Demo link copied!')).catch(() => setNotice('Share this demo page: /social')); } else { setNotice('Share this demo page: /social'); }}}>↗ Share space</button></div>
            </div>

            <div class="social-tabs" role="tablist" aria-label="Profile sections">
              <button role="tab" aria-selected={tab()==='playbooks'} class={tab()==='playbooks'?'active':''} onClick={() => { setTab('playbooks'); }}>⚡ The playbooks</button>
              <button role="tab" aria-selected={tab()==='about'} class={tab()==='about'?'active':''} onClick={() => { setTab('about'); }}>✳ The lore</button>
              <button role="tab" aria-selected={tab()==='friends'} class={tab()==='friends'?'active':''} onClick={() => setTab('friends')}>♡ Top friends</button>
            </div>

            <Show when={notice()}><div class="social-toast" role="status"><span>{notice()}</span><button onClick={() => setNotice('')} aria-label="Dismiss notification">×</button></div></Show>

            <Show when={tab()==='playbooks'}>
              <section class="social-section" id="playbooks"><div class="social-section-heading"><div><span class="social-kicker">STOLEN FROM MY BRAIN, WITH LOVE</span><h2>My public playbooks <span>✦</span></h2><p>Little systems worth sharing. Keep the personality. Skip the spreadsheets.</p></div><div class="social-doodle">too many tabs<br/>not enough time ↙</div></div>
                <div class="social-grid"><For each={playbooks}>{item => <article class={`social-playbook ${item.tint}`}><div class="social-book-top"><span class="social-book-tag">{item.tag}</span><span class="social-book-emoji">{item.emoji}</span></div><h3>{item.title}</h3><p>{item.body}</p><div class="social-book-bottom"><small>☆ {item.installs} imaginary installs</small><button onClick={() => setSelected(item)}>Peek inside ↗</button></div></article>}</For></div>
              </section>
              <div class="social-quote"><span>“</span><div><strong>Your personality is not a productivity problem.</strong><p>But your unread emails might be. Text Clear.</p></div><span>✳</span></div>
            </Show>
            <Show when={tab()==='about'}><section class="social-text-panel" id="about"><span class="social-kicker">THE LORE</span><h2>What is this place?</h2><p>Imagine the personality of old-school MySpace, without the pressure to build another infinite feed. You get a customizable public profile, a few featured friends, and playbooks people can choose to install.</p><p>Your own private messages, installed playbooks and progress stay private unless you explicitly choose to publish something.</p><div class="social-handwritten">yes, you can be a little weird here ✷</div></section></Show>
            <Show when={tab()==='friends'}><section class="social-text-panel"><span class="social-kicker">TOP FRIENDS, 2006 ENERGY</span><h2>The internet is better with people ♡</h2><div class="social-friend-grid"><For each={[['✺','a very online builder'],['☻','the chaos coordinator'],['✦','someone who touches grass']]}>{friend=><div><div>{friend[0]}</div><strong>{friend[1]}</strong><small>Concept-only placeholder</small></div>}</For></div><p>Featured friends are a nostalgic design concept, not a real follower ranking or people you follow.</p></section></Show>
            <footer class="social-footer"><span>✳ cleartify social concept · built to be useful, not addictive</span><a href="/">Back to Cleartify ↗</a></footer>
          </div>
        </div>
      </div>

      <Show when={selected()} keyed>{item => <div class="social-modal-overlay" onClick={() => setSelected(null)}><section class="social-modal" role="dialog" aria-modal="true" aria-label={item.title} onClick={e => e.stopPropagation()}><button class="social-modal-close" onClick={() => setSelected(null)} aria-label="Close preview">×</button><div class="social-modal-emoji">{item.emoji}</div><span class="social-kicker">TEXT-ONLY GUIDE · DEMO</span><h2>{item.title}</h2><p>{item.body}</p><ol><For each={item.steps}>{step => <li>{step}</li>}</For></ol><p class="social-modal-note">Preview only. Real installations require an authenticated user, publication visibility and permission checks.</p><button class="social-modal-action" onClick={() => demoInstall(item)}>Got it, clearly. →</button></section></div>}</Show>
    </main>
  );
}
