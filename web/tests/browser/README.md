Run the scroll-lock regression tests from `web/`:

```sh
npm ci
npx playwright install chromium
npm run test:browser
```

The tests start a local Vite server and mount the real review, attachment, confirmation,
and Markdown editor components. API responses use isolated fixtures; no backend or
account is needed. Scrolling is checked with actual browser wheel input.
