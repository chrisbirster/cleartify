import { render } from '@solidjs/web';
import { App } from './App';
import { Router } from './router';
import './styles/global.css';

const root = document.getElementById('root');

if (!root) throw new Error('Missing #root');

render(
  () => (
    <Router>
      {props => <App>{props.children}</App>}
    </Router>
  ),
  root,
);
