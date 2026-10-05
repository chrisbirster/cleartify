import { render } from '@solidjs/web';
import { Router } from '@solidjs/router';
import { App } from './App';
import './styles/global.css';

const root = document.getElementById('root');

if (!root) throw new Error('Missing #root');

render(
  () => (
    <Router>
      <App />
    </Router>
  ),
  root,
);
