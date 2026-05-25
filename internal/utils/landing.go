package utils

const LandingPageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Flipay - Mini Payment Gateway API</title>
  <style>
    body { margin: 0; font-family: Inter, Arial, sans-serif; background: #0f172a; color: #e2e8f0; }
    .container { max-width: 960px; margin: 0 auto; padding: 56px 24px; }
    .badge { display: inline-block; padding: 8px 12px; border-radius: 999px; background: #064e3b; color: #bbf7d0; font-size: 14px; }
    h1 { font-size: 44px; margin: 18px 0 10px; line-height: 1.1; }
    p { color: #cbd5e1; font-size: 18px; line-height: 1.6; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 16px; margin-top: 28px; }
    .card { background: #111827; border: 1px solid #334155; border-radius: 16px; padding: 20px; }
    .card h2 { margin: 0 0 12px; font-size: 20px; }
    ul { padding-left: 20px; color: #cbd5e1; }
    li { margin: 8px 0; }
    a { color: #38bdf8; text-decoration: none; }
    a:hover { text-decoration: underline; }
    code { background: #020617; color: #a7f3d0; padding: 3px 6px; border-radius: 6px; }
    .actions { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 24px; }
    .button { padding: 12px 16px; border-radius: 10px; background: #2563eb; color: white; font-weight: 700; }
    .button.secondary { background: #1e293b; border: 1px solid #475569; }
  </style>
</head>
<body>
  <main class="container">
    <span class="badge">API Running • Portfolio Backend Project</span>
    <h1>Flipay - Mini Payment Gateway API</h1>
    <p>Fintech backend simulation built with Go, Gin, PostgreSQL, Redis queue worker, JWT authentication, idempotency, and webhook callback flow.</p>
    <div class="actions">
      <a class="button" href="/swagger/index.html">Open Swagger Docs</a>
      <a class="button secondary" href="/health">Check Health</a>
    </div>
    <section class="grid">
      <div class="card">
        <h2>Core Features</h2>
        <ul>
          <li>JWT Register & Login</li>
          <li>Create Payment</li>
          <li>Virtual Account & QRIS Simulation</li>
          <li>Async Redis Worker</li>
          <li>Webhook Callback Simulation</li>
        </ul>
      </div>
      <div class="card">
        <h2>Main Endpoints</h2>
        <ul>
          <li><code>GET /health</code></li>
          <li><code>POST /api/v1/auth/register</code></li>
          <li><code>POST /api/v1/auth/login</code></li>
          <li><code>POST /api/v1/payments</code></li>
          <li><code>GET /api/v1/payments/history</code></li>
        </ul>
      </div>
      <div class="card">
        <h2>Tech Stack</h2>
        <ul>
          <li>Golang + Gin</li>
          <li>PostgreSQL</li>
          <li>Redis</li>
          <li>Zap Logger</li>
          <li>Swagger + Docker</li>
        </ul>
      </div>
    </section>
  </main>
</body>
</html>`
