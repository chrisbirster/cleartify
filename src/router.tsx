import { createRouter } from '@solidjs/router';
import { BetaActivate } from './routes/BetaActivate';
import { Home } from './routes/Home';
import { PrivacyNotice, TermsNotice } from './routes/Policy';
import { ProfilePreview } from './routes/ProfilePreview';

export const Router = createRouter({
  routes: [
    { path: '/', component: Home },
    { path: '/privacy', component: PrivacyNotice },
    { path: '/terms', component: TermsNotice },
    { path: '/beta/activate', component: BetaActivate },
    { path: '/@:handle', component: ProfilePreview },
  ],
});
