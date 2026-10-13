import { createEffect, createSignal, onCleanup, Show } from 'solid-js';
import { useParams } from '@solidjs/router';
import * as stylex from '@stylexjs/stylex';

type Profile = {
 did: string; handle: string; displayName: string; bio: string;
 avatarUrl: string; kind: 'person' | 'agent'; claimed: boolean;
};

export function ProfilePreview() {
 const params = useParams();
 const [profile,setProfile]=createSignal<Profile | null>(null);
 const [loading,setLoading]=createSignal(true);
 const [message,setMessage]=createSignal('');
 const apiBase=(import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/,'');

 createEffect(() => {
  const handle=params.handle;
  const controller=new AbortController();
  setLoading(true);setProfile(null);setMessage('');
  void (async ()=>{
   try {
    const res=await fetch(`${apiBase}/api/profiles/${encodeURIComponent(handle ?? '')}`,{signal:controller.signal});
    if(res.status===404){setMessage('This profile is not public or does not exist.');return}
    if(!res.ok){setMessage('Could not load this profile.');return}
    const data=await res.json() as Profile;
    setProfile(data);
   } catch(error) {
    if(!controller.signal.aborted){setMessage('Could not load this profile.')}
   } finally {
    if(!controller.signal.aborted){setLoading(false)}
   }
  })();
  onCleanup(()=>controller.abort());
 }, undefined);

 return <main {...stylex.props(styles.page)}>
  <nav {...stylex.props(styles.nav)}>
   <a href="/" {...stylex.props(styles.brand)}>✦ Cleartify</a>
   <a href="/" {...stylex.props(styles.back)}>← Back home</a>
  </nav>
  <section {...stylex.props(styles.profile)}>
   <div {...stylex.props(styles.cover)} />
   <Show when={loading()}><p role="status">Loading profile…</p></Show>
   <Show when={message()}><p role="status">{message()}</p></Show>
   <Show when={profile()} keyed>{p=>
    <article {...stylex.props(styles.card)}>
     <div {...stylex.props(styles.avatar)} aria-hidden="true">{p.displayName.slice(0,1).toUpperCase()}</div>
     <div>
      <h1>{p.displayName}</h1>
      <p>@{p.handle}</p>
      <p>{p.bio}</p>
      <p><strong>{p.kind==='agent'?'AI assistant':'Person'}</strong> · {p.claimed?'Claimed profile':'Unclaimed profile'}</p>
      <small>AT Protocol DID: {p.did}</small>
      <p {...stylex.props(styles.notice)}>Public information only. Private protocols and run history are never shown.</p>
     </div>
    </article>
   }</Show>
  </section>
 </main>;
}
const styles=stylex.create({
 page:{minHeight:'100vh',backgroundColor:'#fffaf5',color:'#111936'},
 nav:{maxWidth:1080,margin:'0 auto',padding:'22px 26px',display:'flex',justifyContent:'space-between'},
 brand:{fontSize:24,fontWeight:900},back:{color:'#566079'},
 profile:{maxWidth:1080,margin:'0 auto',padding:'0 26px 80px'},
 cover:{height:220,borderRadius:28,background:'linear-gradient(120deg,#bcecff,#ffe57d 48%,#ffb8d6)'},
 card:{display:'grid',gridTemplateColumns:{default:'100px 1fr','@media (max-width: 540px)':'1fr'},gap:24,marginTop:-32,padding:26,backgroundColor:'#fff',borderRadius:22,position:'relative',boxShadow:'0 8px 20px rgba(0,0,0,.06)'},
 avatar:{width:90,height:90,display:'grid',placeItems:'center',borderRadius:22,backgroundColor:'#0ca678',color:'#fff',fontSize:38,fontWeight:900},
 notice:{color:'#596277',fontSize:13}
});
