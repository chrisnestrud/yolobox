<template>
  <div class="home-content">
    <section class="home-section masthead-section">
      <div class="section-inner masthead-inner">
        <AsciiLogo />
      </div>
    </section>

    <!-- ── Problem / Solution ── -->
    <section class="home-section comparison-section">
      <div class="section-inner">
        <div class="section-label">The Problem</div>
        <h2 class="section-title">
          You're one bad prompt away from
          <code class="danger-code">rm -rf ~</code>
        </h2>
        <div class="terminals-grid">
          <div class="terminal-col">
            <div class="terminal-label label-danger">Without a sandbox</div>
            <div class="term term-danger">
              <div class="term-header">
                <span class="dot dot-red"></span>
                <span class="dot dot-yellow"></span>
                <span class="dot dot-green"></span>
                <span class="term-title">zsh</span>
              </div>
              <div class="term-body">
                <div class="term-line"><span class="prompt">$</span> claude --dangerously-skip-permissions</div>
                <div class="term-line"><span class="output">Claude&gt; I'll reorganize your files...</span></div>
                <div class="term-line"><span class="prompt">$</span> <span class="danger-text">rm -rf ~/*</span></div>
                <div class="term-line"><span class="prompt">$</span> ls ~/.ssh</div>
                <div class="term-line"><span class="danger-text">No such file or directory</span></div>
                <div class="term-line"><span class="prompt">$</span> ls ~/Documents</div>
                <div class="term-line"><span class="danger-text">No such file or directory</span></div>
                <div class="term-line"><span class="prompt">You&gt;</span> Wtf Claude you deleted everything!</div>
                <div class="term-line"><span class="danger-text">Claude&gt;</span> You're absolutely right!</div>
              </div>
            </div>
          </div>
          <div class="terminal-col">
            <div class="terminal-label label-safe">With yolobox</div>
            <div class="term term-safe">
              <div class="term-header">
                <span class="dot dot-red"></span>
                <span class="dot dot-yellow"></span>
                <span class="dot dot-green"></span>
                <span class="term-title">yolobox</span>
              </div>
              <div class="term-body">
                <div class="term-line"><span class="prompt">$</span> yolobox claude</div>
                <div class="term-line"><span class="output">Claude&gt; I'll reorganize your files...</span></div>
                <div class="term-line"><span class="prompt">$</span> rm -rf ~/*</div>
                <div class="term-line"><span class="comment"># nuked the container's home. oh well.</span></div>
                <div class="term-line"><span class="prompt">$</span> exit</div>
                <div class="term-line"><span class="comment"># back on the host:</span></div>
                <div class="term-line"><span class="prompt">$</span> ls ~/.ssh</div>
                <div class="term-line"><span class="safe-text">id_ed25519  config</span>  <span class="comment"># safe</span></div>
                <div class="term-line"><span class="prompt">You&gt;</span> 😎</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ── How It Works ── -->
    <section class="home-section steps-section">
      <div class="section-inner">
        <div class="section-label">The Solution</div>
        <h2 class="section-title">Let your AI go full send. Your home directory stays home.</h2>
        <p class="section-desc">
          yolobox gives the agent full permissions inside a container while keeping the important host boundaries intact.
        </p>
        <div class="solution-grid">
          <div class="solution-item">Project mounted at its real path</div>
          <div class="solution-item">Full permissions and sudo inside the box</div>
          <div class="solution-item">Home directory not mounted by default</div>
          <div class="solution-item">Persistent volumes keep tools and state across runs</div>
        </div>
        <div class="steps-grid">
          <div class="step-card">
            <div class="step-number">01</div>
            <h3>Mount your project</h3>
            <p>
              Your project directory is mounted at its real host path for seamless
              session continuity. <code>/Users/you/project</code> inside equals
              <code>/Users/you/project</code> outside.
            </p>
          </div>
          <div class="step-card">
            <div class="step-number">02</div>
            <h3>Isolate everything else</h3>
            <p>
              Your home directory is never mounted. SSH keys, credentials, cloud
              configs, dotfiles, other projects &mdash; all invisible to the agent.
            </p>
          </div>
          <div class="step-card">
            <div class="step-number">03</div>
            <h3>Let the AI cook</h3>
            <p>
              Full sudo access, persistent named volumes, pre-installed tools. The
              agent installs anything it needs. Sessions survive restarts.
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- ── Quick Start ── -->
    <section class="home-section quickstart-section">
      <div class="section-inner">
        <div class="section-label">Quick Start</div>
        <h2 class="section-title">Two commands. That's it.</h2>
        <div class="quickstart-terminal">
          <div class="term">
            <div class="term-header">
              <span class="dot dot-red"></span>
              <span class="dot dot-yellow"></span>
              <span class="dot dot-green"></span>
              <span class="term-title">zsh</span>
            </div>
            <div class="term-body">
              <div class="term-line"><span class="comment"># Install</span></div>
              <div class="term-line"><span class="prompt">$</span> brew install finbarr/tap/yolobox</div>
              <div class="term-line">&nbsp;</div>
              <div class="term-line"><span class="comment"># Run from any project</span></div>
              <div class="term-line"><span class="prompt">$</span> cd /path/to/your/project</div>
              <div class="term-line"><span class="prompt">$</span> yolobox claude  <span class="comment"># or codex, gemini, copilot</span></div>
            </div>
          </div>
        </div>
        <div class="install-alt">
          Or use the install script: <code>curl -fsSL https://raw.githubusercontent.com/finbarr/yolobox/master/install.sh | bash</code>
        </div>
        <div class="quickstart-cta">
          <a href="/getting-started" class="cta-docs">Read the docs</a>
        </div>
      </div>
    </section>

    <section class="home-section guide-section">
      <div class="section-inner">
        <div class="section-label">Find the Right Page</div>
        <h2 class="section-title">Start with the task you are trying to solve</h2>
        <div class="guide-grid">
          <a v-for="guide in guides" :key="guide.title" :href="guide.href" class="guide-card">
            <div class="guide-kicker">{{ guide.kicker }}</div>
            <h3>{{ guide.title }}</h3>
            <p>{{ guide.copy }}</p>
          </a>
        </div>
      </div>
    </section>

    <!-- ── Supported Agents ── -->
    <section class="home-section agents-section">
      <div class="section-inner">
        <div class="section-label">Supported Agents</div>
        <h2 class="section-title">Every AI, full auto</h2>
        <div class="agents-grid">
          <div class="agent-card" v-for="agent in agents" :key="agent.name">
            <div class="agent-name">{{ agent.name }}</div>
            <div class="agent-cmd">{{ agent.cmd }}</div>
            <div class="agent-expands">{{ agent.expands }}</div>
          </div>
        </div>
        <div class="agent-note">
          <strong>It's the AI's box, not yours.</strong> You launch the agent and let
          it work. It has sudo &mdash; if it needs a compiler, database, or framework,
          it just installs it.
        </div>
      </div>
    </section>

  </div>
</template>

<script setup>
import AsciiLogo from './AsciiLogo.vue'

const agents = [
  { name: 'Claude Code', cmd: 'yolobox claude', expands: '--dangerously-skip-permissions' },
  { name: 'OpenAI Codex', cmd: 'yolobox codex', expands: '--dangerously-bypass-approvals-and-sandbox' },
  { name: 'Gemini CLI', cmd: 'yolobox gemini', expands: '--yolo' },
  { name: 'GitHub Copilot', cmd: 'yolobox copilot', expands: '--yolo' },
  { name: 'OpenCode', cmd: 'yolobox opencode', expands: 'runs natively' },
]

const guides = [
  {
    kicker: 'Install',
    title: 'Get running fast',
    href: '/getting-started',
    copy: 'Install yolobox, pick a runtime, and launch your first agent session.',
  },
  {
    kicker: 'Customize',
    title: 'Add packages per project',
    href: '/customizing',
    copy: 'Use packages or a Dockerfile fragment without taking on a permanent base-image fork.',
  },
  {
    kicker: 'Configure',
    title: 'Set your defaults',
    href: '/configuration',
    copy: 'Understand global config, project config, copied instructions, and forwarded credentials.',
  },
  {
    kicker: 'Reference',
    title: 'Find the exact flag',
    href: '/flags',
    copy: 'Look up runtime, network, Docker, resource, and customization flags by category.',
  },
  {
    kicker: 'Security',
    title: 'Understand the boundary',
    href: '/security',
    copy: 'See what yolobox protects, what it does not, and how to harden your setup further.',
  },
  {
    kicker: 'Workflow',
    title: 'Use the right command',
    href: '/commands',
    copy: 'Keep the AI shortcut workflow front and center, and use the shell and maintenance commands intentionally.',
  },
]
</script>

<style scoped>
/* ── Layout ── */
.home-content {
  width: 100%;
  overflow: hidden;
}

.home-section {
  padding: 80px 24px;
  position: relative;
}

.home-section + .home-section {
  border-top: 1px solid var(--vp-c-divider);
}

.masthead-section {
  padding-top: 52px;
  padding-bottom: 18px;
}

.masthead-section + .home-section {
  border-top: none;
}

.masthead-inner {
  display: flex;
  justify-content: flex-start;
}

.comparison-section {
  padding-top: 0;
}

.steps-section .section-desc {
  max-width: none;
}

.section-inner {
  max-width: 960px;
  margin: 0 auto;
}

/* ── Section Headings ── */
.section-label {
  font-family: var(--vp-font-family-mono);
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--yolo-orange);
  text-transform: uppercase;
  letter-spacing: 0.15em;
  margin-bottom: 16px;
}

.section-title {
  font-family: var(--vp-font-family-base);
  font-weight: 800;
  font-size: clamp(1.6rem, 4vw, 2.25rem);
  line-height: 1.2;
  letter-spacing: -0.03em;
  margin-bottom: 16px;
  color: var(--vp-c-text-1);
}

.section-desc {
  color: var(--vp-c-text-2);
  font-size: 1.05rem;
  max-width: 580px;
  line-height: 1.7;
}

.guide-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
  margin-top: 32px;
}

.guide-card {
  display: block;
  padding: 22px;
  border-radius: 16px;
  text-decoration: none;
  color: inherit;
  background: linear-gradient(180deg, var(--yolo-surface-2), var(--yolo-surface-1));
  border: 1px solid var(--yolo-card-border);
  transition: transform 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.guide-card:hover {
  transform: translateY(-2px);
  border-color: var(--yolo-accent-border-strong);
  box-shadow: var(--yolo-card-shadow-hover);
}

.guide-kicker {
  font-family: var(--vp-font-family-mono);
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--yolo-orange);
  margin-bottom: 10px;
}

.guide-card h3 {
  margin: 0 0 10px;
  font-size: 1.05rem;
  color: var(--vp-c-text-1);
}

.guide-card p {
  margin: 0;
  color: var(--vp-c-text-2);
  line-height: 1.6;
}

.danger-code {
  color: #ef4444;
  background: var(--yolo-danger-soft);
  padding: 2px 10px;
  border-radius: 6px;
  font-family: var(--vp-font-family-mono);
  font-size: 0.85em;
}

/* ── Terminal Blocks ── */
.term {
  background: var(--yolo-terminal-bg);
  border: 1px solid var(--yolo-terminal-border);
  border-radius: 12px;
  overflow: hidden;
  transition: border-color 0.3s, box-shadow 0.3s;
}

.term:hover {
  border-color: var(--yolo-terminal-border-strong);
  box-shadow: var(--yolo-card-shadow);
}

.term-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--yolo-terminal-header-bg);
  border-bottom: 1px solid var(--yolo-card-border);
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}
.dot-red { background: #ff5f57; }
.dot-yellow { background: #febc2e; }
.dot-green { background: #28c840; }

.term-title {
  margin-left: 8px;
  font-family: var(--vp-font-family-mono);
  font-size: 0.75rem;
  color: var(--yolo-term-faint);
}

.term-body {
  padding: 20px 24px;
  font-family: var(--vp-font-family-mono);
  font-size: 0.85rem;
  line-height: 1.75;
  color: var(--yolo-term-text);
}

.term-body .prompt { color: var(--yolo-orange); }
.term-body .comment { color: var(--yolo-term-faint); }
.term-body .output { color: var(--yolo-term-muted); }
.term-body .danger-text { color: #ef4444; }
.term-body .safe-text { color: #34d399; }

.term-danger { border-color: var(--yolo-danger-border); }
.term-danger:hover { border-color: var(--yolo-danger-border-strong); box-shadow: var(--yolo-danger-shadow); }

.term-safe { border-color: var(--yolo-safe-border); }
.term-safe:hover { border-color: var(--yolo-safe-border-strong); box-shadow: var(--yolo-safe-shadow); }

/* ── Quick Start ── */
.quickstart-section {
  padding-bottom: 80px;
}

.quickstart-terminal {
  max-width: 600px;
  margin-top: 32px;
}

.install-alt {
  margin-top: 20px;
  font-size: 0.85rem;
  color: var(--vp-c-text-3);
}

.install-alt code {
  font-family: var(--vp-font-family-mono);
  font-size: 0.8rem;
  color: var(--vp-c-text-2);
  background: var(--yolo-inline-code-bg);
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--yolo-inline-code-border);
}

.quickstart-cta {
  margin-top: 28px;
}

.cta-docs {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: var(--vp-font-family-mono);
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--yolo-orange) !important;
  border-bottom: 1px solid rgba(255, 107, 0, 0.3) !important;
  padding-bottom: 2px;
  transition: all 0.2s;
  text-decoration: none !important;
}

.cta-docs:hover {
  color: var(--yolo-gold) !important;
  border-bottom-color: var(--yolo-gold) !important;
}

.cta-docs::after {
  content: '\2192';
}

/* ── Comparison Grid ── */
.terminals-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  margin-top: 48px;
}

.terminal-label {
  font-family: var(--vp-font-family-mono);
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  margin-bottom: 12px;
}
.label-danger { color: #ef4444; }
.label-safe { color: #34d399; }

/* ── Steps ── */
.steps-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-top: 48px;
}

.step-card {
  padding: 32px 28px;
  background: var(--yolo-surface-1);
  border: 1px solid var(--yolo-card-border);
  border-radius: 14px;
  transition: border-color 0.3s, transform 0.3s;
}

.step-card:hover {
  border-color: var(--yolo-card-border-strong);
  transform: translateY(-3px);
  box-shadow: var(--yolo-card-shadow);
}

.step-number {
  font-weight: 800;
  font-size: 2.2rem;
  color: var(--yolo-orange);
  line-height: 1;
  margin-bottom: 16px;
  opacity: 0.6;
}

.step-card h3 {
  font-weight: 700;
  font-size: 1.15rem;
  margin-bottom: 10px;
  letter-spacing: -0.02em;
  color: var(--vp-c-text-1);
}

.step-card p {
  color: var(--vp-c-text-2);
  font-size: 0.9rem;
  line-height: 1.65;
}

.step-card code {
  font-family: var(--vp-font-family-mono);
  font-size: 0.8rem;
  color: var(--yolo-amber);
  background: var(--yolo-accent-soft);
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--yolo-accent-border);
}

.solution-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
  margin-top: 28px;
  margin-bottom: 40px;
}

.solution-item {
  padding: 16px 18px;
  border-radius: 12px;
  background: var(--yolo-accent-soft);
  border: 1px solid var(--yolo-accent-border);
  color: var(--vp-c-text-1);
  font-weight: 500;
  line-height: 1.5;
}

/* ── Agents ── */
.agents-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
  margin-top: 32px;
}

.agent-card {
  padding: 24px;
  background: var(--yolo-surface-1);
  border: 1px solid var(--yolo-card-border);
  border-radius: 12px;
  transition: border-color 0.3s, transform 0.3s;
}

.agent-card:hover {
  border-color: var(--yolo-accent-border-strong);
  transform: translateY(-3px);
  box-shadow: var(--yolo-card-shadow);
}

.agent-name {
  font-weight: 700;
  font-size: 1rem;
  margin-bottom: 8px;
  letter-spacing: -0.01em;
  color: var(--vp-c-text-1);
}

.agent-cmd {
  font-family: var(--vp-font-family-mono);
  font-size: 0.8rem;
  color: var(--yolo-orange);
  background: var(--yolo-accent-soft);
  border: 1px solid var(--yolo-accent-border);
  padding: 4px 10px;
  border-radius: 6px;
  display: inline-block;
  margin-bottom: 10px;
}

.agent-expands {
  font-family: var(--vp-font-family-mono);
  font-size: 0.72rem;
  color: var(--vp-c-text-3);
  word-break: break-all;
}

.agent-note {
  margin-top: 32px;
  padding: 20px 24px;
  background: var(--yolo-accent-soft);
  border: 1px solid var(--yolo-accent-border);
  border-radius: 10px;
  color: var(--vp-c-text-2);
  font-size: 0.95rem;
  line-height: 1.6;
}

.agent-note strong {
  color: var(--vp-c-text-1);
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .guide-grid,
  .solution-grid,
  .terminals-grid {
    grid-template-columns: 1fr;
  }

  .steps-grid {
    grid-template-columns: 1fr;
  }

  .term-body {
    font-size: 0.75rem;
    padding: 16px;
  }

  .home-section {
    padding: 60px 24px;
  }
}
</style>
