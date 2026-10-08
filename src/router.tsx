import { createRouter } from '@solidjs/router';
import { Home } from './routes/Home';
import { ProfilePreview } from './routes/ProfilePreview';

export const Router = createRouter({
  routes: [
    { path: '/', component: Home },
    { path: '/@:handle', component: ProfilePreview },
  ],
});
