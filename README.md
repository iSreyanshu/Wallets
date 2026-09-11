<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>EVM Wallet Generator</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Amaranth:ital,wght@0,400;0,700;1,400;1,700&display=swap" rel="stylesheet">
  <style>
    :root {
      --ink: #18221e;
      --muted: #66736d;
      --line: #dfe7e2;
      --paper: #f7faf7;
      --card: rgba(255, 255, 255, 0.84);
      --mint: #c9f27c;
      --mint-dark: #618b21;
      --night: #18221e;
      --orange: #f1a36b;
      --radius: 18px;
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      color: var(--ink);
      background:
        radial-gradient(circle at 8% 0%, rgba(201, 242, 124, .32), transparent 30%),
        radial-gradient(circle at 96% 16%, rgba(241, 163, 107, .2), transparent 25%),
        var(--paper);
      font-family: 'Amaranth', sans-serif;
      line-height: 1.55;
    }

    .page { width: min(900px, calc(100% - 36px)); margin: 0 auto; padding: 42px 0 64px; }
    header { position: relative; padding: 42px 0 34px; border-bottom: 1px solid var(--line); }
    .eyebrow { display: inline-flex; align-items: center; gap: 8px; margin-bottom: 20px; color: var(--mint-dark); font-size: .78rem; font-weight: 700; letter-spacing: .1em; text-transform: uppercase; }
    .eyebrow::before { content: ''; width: 9px; height: 9px; border-radius: 50%; background: var(--orange); box-shadow: 0 0 0 5px rgba(241, 163, 107, .18); }
    h1 { max-width: 680px; margin: 0; font-size: clamp(3.1rem, 9vw, 6.7rem); font-weight: 700; letter-spacing: -.045em; line-height: .92; }
    .intro { max-width: 620px; margin: 24px 0 0; color: var(--muted); font-size: clamp(1.05rem, 2vw, 1.3rem); }
    .badges { display: flex; flex-wrap: wrap; gap: 9px; margin-top: 25px; }
    .badge { display: inline-flex; align-items: center; min-height: 29px; padding: 4px 11px; border: 1px solid var(--line); border-radius: 999px; background: var(--card); color: var(--muted); font-size: .75rem; font-weight: 700; }
    .badge strong { color: var(--ink); margin-right: 5px; }
    main { display: grid; gap: 18px; margin-top: 28px; }
    section { padding: 27px 28px 29px; border: 1px solid var(--line); border-radius: var(--radius); background: var(--card); box-shadow: 0 14px 34px rgba(44, 67, 53, .06); }
    h2 { display: flex; align-items: center; gap: 11px; margin: 0 0 13px; font-size: 1.65rem; line-height: 1.1; }
    h2::before { content: ''; width: 7px; height: 24px; border-radius: 10px; background: var(--mint); }
    p, li { color: var(--muted); font-size: 1rem; }
    p { margin: 0 0 16px; }
    strong { color: var(--ink); }
    ol, ul { margin: 0 0 18px; padding-left: 22px; }
    li { padding: 3px 0; }
    code { font-family: inherit; }
    .code-wrap { position: relative; margin: 17px 0 18px; }
    pre { margin: 0; padding: 19px 54px 19px 19px; overflow-x: auto; border: 1px solid #2b3a33; border-radius: 13px; background: var(--night); color: #e8f6e9; font-family: 'Amaranth', sans-serif; font-size: .94rem; line-height: 1.7; white-space: pre-wrap; word-break: break-word; }
    .copy-btn { position: absolute; top: 10px; right: 10px; display: grid; width: 32px; height: 32px; place-items: center; border: 1px solid rgba(255,255,255,.16); border-radius: 9px; background: rgba(255,255,255,.08); color: #e8f6e9; cursor: pointer; transition: background .2s, transform .2s; }
    .copy-btn:hover { background: rgba(201, 242, 124, .22); transform: translateY(-1px); }
    .copy-btn svg { width: 15px; height: 15px; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
    .copy-btn.copied { color: var(--mint); }
    .callout { margin-top: 20px; padding: 15px 17px; border-left: 3px solid var(--orange); border-radius: 0 11px 11px 0; background: #fff4eb; color: #805534; font-size: .95rem; }
    .footer { padding-top: 32px; color: var(--muted); text-align: center; font-size: .9rem; }
    .footer span { color: var(--orange); }

    @media (max-width: 600px) {
      .page { width: min(100% - 24px, 900px); padding-top: 18px; }
      header { padding: 30px 4px 28px; }
      h1 { font-size: clamp(3rem, 17vw, 4.8rem); }
      section { padding: 23px 18px 24px; }
      h2 { font-size: 1.45rem; }
      pre { font-size: .86rem; }
    }
  </style>
</head>
<body>
  <div class="page">
    <header>
      <div class="eyebrow">Local key generation toolkit</div>
      <h1>EVM Wallet<br>Generator</h1>
      <p class="intro">A fast, local wallet generator with normal, vanity and exclusion modes. Your keys stay on your machine.</p>
      <div class="badges">
        <span class="badge"><strong>Zig</strong> 0.13.0</span>
        <span class="badge"><strong>Crypto</strong> secp256k1 + Keccak</span>
        <span class="badge">OS-backed randomness</span>
      </div>
    </header>

    <main>
      <section>
        <h2>Quick Start</h2>
        <div class="code-wrap">
          <pre><code>git clone https://github.com/iSreyanshu/Wallets.git
cd Wallets
chmod +x install.sh run.sh
./install.sh
./run.sh</code></pre>
          <button class="copy-btn" type="button" aria-label="Copy Quick Start commands" title="Copy" data-copy="git clone https://github.com/iSreyanshu/Wallets.git
cd Wallets
chmod +x install.sh run.sh
./install.sh
./run.sh">
            <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="9" y="9" width="11" height="11" rx="2"></rect><path d="M15 9V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h3"></path></svg>
          </button>
        </div>
        <p>The installer handles compiler, build and OpenSSL prerequisites. It installs Zig in <strong>~/.local/zig</strong> and does not require a system-wide Zig package.</p>
      </section>

      <section>
        <h2>Modes</h2>
        <p>Each run opens an interactive menu and asks for the wallet count:</p>
        <ol>
          <li>Generate normal wallets.</li>
          <li>Generate vanity wallets by matching up to 10 hexadecimal characters after <strong>0x</strong> or at the end of the address.</li>
          <li>Generate wallets excluding a comma-separated list of hexadecimal characters, such as <strong>b,6,f</strong>.</li>
        </ol>
        <p>The output is written to <strong>wallets.csv</strong> (address has no <strong>0x</strong>; private key keeps its <strong>0x</strong> prefix):</p>
        <div class="code-wrap">
          <pre><code>address,private_key
023bcd...,0x...</code></pre>
          <button class="copy-btn" type="button" aria-label="Copy CSV example" title="Copy" data-copy="address,private_key
023bcd...,0x...">
            <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="9" y="9" width="11" height="11" rx="2"></rect><path d="M15 9V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h3"></path></svg>
          </button>
        </div>
      </section>

      <section>
        <h2>Performance</h2>
        <p>The default pool size is <strong>100 workers per detected CPU core</strong>, capped by the requested wallet count. Override it with <strong>--worker</strong>:</p>
        <div class="code-wrap">
          <pre><code>./run.sh --count 100000 --worker 1000</code></pre>
          <button class="copy-btn" type="button" aria-label="Copy performance command" title="Copy" data-copy="./run.sh --count 100000 --worker 1000">
            <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="9" y="9" width="11" height="11" rx="2"></rect><path d="M15 9V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h3"></path></svg>
          </button>
        </div>
      </section>

      <section>
        <h2>Manual Build</h2>
        <p>The build links OpenSSL's secp256k1 implementation and includes the Ethereum Keccak-256 implementation in <strong>deps/keccak</strong>.</p>
        <div class="code-wrap">
          <pre><code>zig build -Doptimize=ReleaseFast
zig build run -Doptimize=ReleaseFast -- --count 1000 --worker 400</code></pre>
          <button class="copy-btn" type="button" aria-label="Copy build commands" title="Copy" data-copy="zig build -Doptimize=ReleaseFast
zig build run -Doptimize=ReleaseFast -- --count 1000 --worker 400">
            <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="9" y="9" width="11" height="11" rx="2"></rect><path d="M15 9V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h3"></path></svg>
          </button>
        </div>
      </section>

      <section>
        <h2>Security</h2>
        <ul>
          <li>Private keys come from <strong>std.crypto.random</strong>, backed by the operating system.</li>
          <li><strong>wallets.csv</strong> contains spendable secrets. Treat it like a password vault.</li>
          <li>Never commit, upload or paste generated private keys into chat or issue trackers.</li>
          <li>This tool does not connect to a blockchain or check balances.</li>
        </ul>
        <div class="callout">Keep generated wallet files local, encrypted and out of version control.</div>
      </section>
    </main>

    <div class="footer">Local <span>•</span> Fast <span>•</span> Private</div>
  </div>

  <script>
    document.querySelectorAll('.copy-btn').forEach((button) => {
      button.addEventListener('click', async () => {
        await navigator.clipboard.writeText(button.dataset.copy);
        button.classList.add('copied');
        button.setAttribute('aria-label', 'Copied');
        setTimeout(() => {
          button.classList.remove('copied');
          button.setAttribute('aria-label', 'Copy');
        }, 1400);
      });
    });
  </script>
</body>
</html>
