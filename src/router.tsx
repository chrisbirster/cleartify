import { createRouter } from '@solidjs/router';
import { Home } from './routes/Home';
import { PrivacyNotice, TermsNotice } from './routes/Policy';
import { ProfilePreview } from './routes/ProfilePreview';

export const Router = createRouter({
  routes: [
    { path: '/', component: Home },
    { path: '/privacy', component: PrivacyNotice },
    { path: '/terms', component: TermsNotice },
    { path: '/@:handle', component: ProfilePreview },
  ],
});
