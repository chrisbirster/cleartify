import * as stylex from '@stylexjs/stylex';

export function NotFound() {
  return <main {...stylex.props(styles.page)}>
    <h1>Page not found</h1>
    <p>We couldn’t find that page.</p>
    <a href="/">Back to Cleartify</a>
  </main>;
}

const styles = stylex.create({
  page: { minHeight: '100vh', padding: '64px 26px', backgroundColor: '#fffaf5', color: '#111936' },
});
