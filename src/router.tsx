import { createRouter } from '@solidjs/router';
import { BetaActivate } from './routes/BetaActivate';
import { Home } from './routes/Home';
import { SocialPreview } from './routes/SocialPreview';
import { PrivacyNotice, TermsNotice } from './routes/Policy';
import { ProfileSettings } from './routes/ProfileSettings';
import { NotFound } from './routes/NotFound';
import { ProfilePreview } from './routes/ProfilePreview';

export const Router = createRouter({
  routes: [
    { path: '/', component: Home },
    { path: '/social', component: SocialPreview },
    { path: '/privacy', component: PrivacyNotice },
    { path: '/terms', component: TermsNotice },
    { path: '/beta/activate', component: BetaActivate },
    { path: '/settings/profile', component: ProfileSettings },
    { path: '/:profileHandle', matchFilters: { profileHandle: /^@.+$/ }, component: ProfilePreview },
    { path: '/*missing', component: NotFound },
  ],
});
