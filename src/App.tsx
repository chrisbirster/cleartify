import { Route } from '@solidjs/router';
import { Home } from './routes/Home';
import { ProfilePreview } from './routes/ProfilePreview';

export function App() {
  return (
    <>
      <Route path="/" component={Home} />
      <Route path="/@:handle" component={ProfilePreview} />
    </>
  );
}
