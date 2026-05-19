/**
 * THE METRIX · Feedback Widget
 * Drop-in script — adds floating feedback button + inline reaction API to any page.
 *
 * Usage:
 *   <script src="feedback-widget.js"></script>
 *
 * Inline reactions (attach to any card/component):
 *   <div data-feedback-component="ai-recommendations" data-feedback-label="AI Recommendations"></div>
 *   // Widget auto-discovers these and injects thumbs up/down
 *
 *   Or manually:
 *   MetrixFeedback.react('recommendation-card', 'up', { extra: 'context' });
 *
 * Config (before script tag):
 *   window.METRIX_FEEDBACK_CONFIG = {
 *     endpoint: 'http://localhost:8765/feedback',  // POST endpoint
 *     page: 'dora-dashboard',                      // auto-detected from URL if omitted
 *     userName: 'alice',                            // optional
 *   };
 */

(function() {
  'use strict';

  // ===================== CONFIG =====================
  const config = window.METRIX_FEEDBACK_CONFIG || {};
  const ENDPOINT = config.endpoint || 'http://localhost:8765/feedback';
  const PAGE = config.page || (location.pathname.split('/').pop().replace('.html','') || 'unknown');
  const USER = config.userName || null;
  const STORE_KEY = 'metrix_feedback';

  // ===================== STORAGE =====================
  function loadFeedback() {
    try { return JSON.parse(localStorage.getItem(STORE_KEY) || '[]'); } catch(e) { return []; }
  }

  function storeFeedback(entry) {
    const all = loadFeedback();
    all.push(entry);
    localStorage.setItem(STORE_KEY, JSON.stringify(all));
  }

  async function submit(entry) {
    entry.id = Date.now().toString(36) + Math.random().toString(36).slice(2,6);
    entry.timestamp = new Date().toISOString();
    entry.page = PAGE;
    entry.user = USER;
    entry.userAgent = navigator.userAgent.slice(0, 80);

    storeFeedback(entry);

    try {
      await fetch(ENDPOINT, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(entry),
      });
    } catch(e) {
      // Silently fail — already saved locally
    }

    return entry;
  }

  // ===================== STYLES =====================
  const styles = `
    @import url('https://fonts.googleapis.com/css2?family=DM+Mono:wght@400;500&family=Syne:wght@700;800&display=swap');

    .mfb-root * { box-sizing: border-box; margin: 0; padding: 0; font-family: 'DM Mono', monospace; }

    /* ---- FLOATING TRIGGER ---- */
    .mfb-trigger {
      position: fixed;
      bottom: 28px;
      right: 28px;
      z-index: 9998;
      display: flex;
      align-items: center;
      gap: 0;
      cursor: pointer;
      filter: drop-shadow(0 4px 20px rgba(79,255,176,0.25));
      transition: filter 0.2s;
    }
    .mfb-trigger:hover { filter: drop-shadow(0 6px 28px rgba(79,255,176,0.45)); }

    .mfb-trigger-btn {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 10px 16px 10px 12px;
      background: #12121a;
      border: 1px solid rgba(79,255,176,0.35);
      border-radius: 24px;
      color: #4fffb0;
      font-size: 11px;
      letter-spacing: 0.06em;
      text-transform: uppercase;
      transition: all 0.2s;
      white-space: nowrap;
    }
    .mfb-trigger-btn:hover { background: #1a1a26; border-color: #4fffb0; }

    .mfb-trigger-icon {
      width: 26px; height: 26px;
      background: linear-gradient(135deg, #4fffb0, #7c6dfa);
      border-radius: 50%;
      display: flex; align-items: center; justify-content: center;
      font-size: 13px;
      flex-shrink: 0;
      animation: mfb-pulse 3s ease-in-out infinite;
    }
    @keyframes mfb-pulse {
      0%,100% { box-shadow: 0 0 0 0 rgba(79,255,176,0.4); }
      50% { box-shadow: 0 0 0 6px rgba(79,255,176,0); }
    }

    .mfb-unread-dot {
      position: absolute;
      top: -3px; right: -3px;
      width: 9px; height: 9px;
      background: #ff6b6b;
      border-radius: 50%;
      border: 2px solid #0a0a0f;
      display: none;
    }
    .mfb-unread-dot.visible { display: block; }

    /* ---- PANEL ---- */
    .mfb-panel {
      position: fixed;
      bottom: 84px;
      right: 28px;
      width: 360px;
      z-index: 9999;
      background: #12121a;
      border: 1px solid #2a2a3a;
      border-radius: 14px;
      overflow: hidden;
      box-shadow: 0 24px 60px rgba(0,0,0,0.6), 0 0 0 1px rgba(79,255,176,0.08);
      transform: translateY(12px) scale(0.97);
      opacity: 0;
      pointer-events: none;
      transition: all 0.22s cubic-bezier(0.34,1.56,0.64,1);
    }
    .mfb-panel.open {
      transform: translateY(0) scale(1);
      opacity: 1;
      pointer-events: all;
    }

    .mfb-panel-header {
      padding: 14px 16px 12px;
      border-bottom: 1px solid #2a2a3a;
      display: flex;
      align-items: center;
      justify-content: space-between;
      background: #1a1a26;
    }
    .mfb-panel-title {
      font-family: 'Syne', sans-serif;
      font-weight: 800;
      font-size: 12px;
      color: #e8e8f0;
      letter-spacing: 0.06em;
      text-transform: uppercase;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .mfb-panel-title::before {
      content: '✦';
      color: #4fffb0;
      font-size: 10px;
    }
    .mfb-close {
      background: none;
      border: none;
      color: #6b6b8a;
      cursor: pointer;
      font-size: 16px;
      padding: 2px 6px;
      border-radius: 4px;
      transition: color 0.15s;
      line-height: 1;
    }
    .mfb-close:hover { color: #e8e8f0; }

    .mfb-panel-body { padding: 16px; }

    /* ---- TYPE SELECTOR ---- */
    .mfb-type-label {
      font-size: 9px;
      text-transform: uppercase;
      letter-spacing: 0.1em;
      color: #6b6b8a;
      margin-bottom: 8px;
    }

    .mfb-type-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 6px;
      margin-bottom: 14px;
    }

    .mfb-type-btn {
      display: flex;
      align-items: center;
      gap: 7px;
      padding: 9px 12px;
      background: #1a1a26;
      border: 1px solid #2a2a3a;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.15s;
      font-size: 11px;
      color: #6b6b8a;
      text-align: left;
    }
    .mfb-type-btn:hover { border-color: #4a4a6a; color: #e8e8f0; }
    .mfb-type-btn.selected { border-color: #4fffb0; color: #4fffb0; background: rgba(79,255,176,0.06); }
    .mfb-type-btn.selected.bug { border-color: #ff6b6b; color: #ff6b6b; background: rgba(255,107,107,0.06); }
    .mfb-type-btn.selected.feature { border-color: #7c6dfa; color: #7c6dfa; background: rgba(124,109,250,0.06); }
    .mfb-type-btn.selected.idea { border-color: #ffd93d; color: #ffd93d; background: rgba(255,217,61,0.06); }
    .mfb-type-btn.selected.broken { border-color: #ff6b6b; color: #ff6b6b; background: rgba(255,107,107,0.1); }
    .mfb-type-icon { font-size: 15px; flex-shrink: 0; }

    /* ---- TEXTAREA ---- */
    .mfb-textarea {
      width: 100%;
      background: #0a0a0f;
      border: 1px solid #2a2a3a;
      border-radius: 8px;
      padding: 10px 12px;
      color: #e8e8f0;
      font-family: 'DM Mono', monospace;
      font-size: 11px;
      resize: none;
      height: 88px;
      outline: none;
      transition: border-color 0.2s;
      line-height: 1.6;
      margin-bottom: 10px;
    }
    .mfb-textarea:focus { border-color: #4fffb0; }
    .mfb-textarea::placeholder { color: #3a3a5a; }

    /* ---- NAME FIELD ---- */
    .mfb-name-row {
      display: flex;
      gap: 8px;
      margin-bottom: 12px;
    }
    .mfb-name-input {
      flex: 1;
      background: #0a0a0f;
      border: 1px solid #2a2a3a;
      border-radius: 6px;
      padding: 8px 10px;
      color: #e8e8f0;
      font-family: 'DM Mono', monospace;
      font-size: 11px;
      outline: none;
      transition: border-color 0.2s;
    }
    .mfb-name-input:focus { border-color: #4a4a6a; }
    .mfb-name-input::placeholder { color: #3a3a5a; }

    /* ---- SUBMIT ---- */
    .mfb-submit {
      width: 100%;
      padding: 11px;
      background: #4fffb0;
      color: #000;
      border: none;
      border-radius: 8px;
      font-family: 'DM Mono', monospace;
      font-size: 11px;
      font-weight: 600;
      cursor: pointer;
      letter-spacing: 0.05em;
      text-transform: uppercase;
      transition: all 0.2s;
    }
    .mfb-submit:hover { background: #3de8a0; transform: translateY(-1px); box-shadow: 0 4px 16px rgba(79,255,176,0.3); }
    .mfb-submit:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }

    /* ---- SUCCESS ---- */
    .mfb-success {
      padding: 24px 16px;
      text-align: center;
      display: none;
    }
    .mfb-success.visible { display: block; }
    .mfb-success-icon { font-size: 32px; margin-bottom: 10px; }
    .mfb-success-title { font-family: 'Syne', sans-serif; font-weight: 800; font-size: 14px; color: #4fffb0; margin-bottom: 6px; }
    .mfb-success-sub { font-size: 10px; color: #6b6b8a; line-height: 1.6; }
    .mfb-success-another { margin-top: 14px; background: none; border: 1px solid #2a2a3a; border-radius: 6px; padding: 7px 16px; color: #6b6b8a; font-family: 'DM Mono', monospace; font-size: 10px; cursor: pointer; transition: all 0.15s; letter-spacing: 0.04em; }
    .mfb-success-another:hover { color: #e8e8f0; border-color: #4a4a6a; }

    /* ---- PAGE BADGE ---- */
    .mfb-page-badge {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      font-size: 9px;
      padding: 2px 7px;
      border-radius: 3px;
      background: #1a1a26;
      border: 1px solid #2a2a3a;
      color: #6b6b8a;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      margin-bottom: 12px;
    }
    .mfb-page-badge::before { content: '◎'; color: #7c6dfa; font-size: 8px; }

    /* ---- INLINE REACTION ---- */
    .mfb-inline-reaction {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 6px;
      border-radius: 6px;
      background: transparent;
      transition: background 0.15s;
    }
    .mfb-inline-reaction:hover { background: rgba(255,255,255,0.03); }

    .mfb-react-label {
      font-size: 9px;
      color: #6b6b8a;
      letter-spacing: 0.06em;
      text-transform: uppercase;
      white-space: nowrap;
    }

    .mfb-react-btn {
      display: flex;
      align-items: center;
      gap: 3px;
      padding: 4px 8px;
      background: #1a1a26;
      border: 1px solid #2a2a3a;
      border-radius: 20px;
      cursor: pointer;
      font-size: 12px;
      color: #6b6b8a;
      transition: all 0.15s;
      font-family: 'DM Mono', monospace;
    }
    .mfb-react-btn:hover { border-color: #4a4a6a; color: #e8e8f0; transform: scale(1.05); }
    .mfb-react-btn.voted-up { background: rgba(79,255,176,0.1); border-color: rgba(79,255,176,0.4); color: #4fffb0; }
    .mfb-react-btn.voted-down { background: rgba(255,107,107,0.1); border-color: rgba(255,107,107,0.4); color: #ff6b6b; }
    .mfb-react-btn.voted-up:hover, .mfb-react-btn.voted-down:hover { transform: none; cursor: default; }

    .mfb-react-count { font-size: 9px; min-width: 8px; }

    /* ---- TOAST ---- */
    .mfb-toast {
      position: fixed;
      bottom: 90px;
      right: 28px;
      background: #12121a;
      border: 1px solid rgba(79,255,176,0.3);
      border-radius: 8px;
      padding: 10px 14px;
      font-size: 11px;
      color: #4fffb0;
      z-index: 10000;
      transform: translateY(8px);
      opacity: 0;
      transition: all 0.25s;
      pointer-events: none;
      display: flex;
      align-items: center;
      gap: 8px;
      box-shadow: 0 8px 24px rgba(0,0,0,0.4);
    }
    .mfb-toast.show { transform: translateY(0); opacity: 1; }
    .mfb-toast-icon { font-size: 14px; }
  `;

  // ===================== DOM INJECT =====================
  function inject() {
    // Styles
    const styleEl = document.createElement('style');
    styleEl.textContent = styles;
    document.head.appendChild(styleEl);

    // Root wrapper
    const root = document.createElement('div');
    root.className = 'mfb-root';
    root.innerHTML = `
      <!-- FLOATING TRIGGER -->
      <div class="mfb-trigger" id="mfb-trigger" onclick="MetrixFeedback._toggle()">
        <div style="position:relative">
          <div class="mfb-trigger-btn">
            <div class="mfb-trigger-icon">✦</div>
            Feedback
          </div>
          <div class="mfb-unread-dot" id="mfb-dot"></div>
        </div>
      </div>

      <!-- PANEL -->
      <div class="mfb-panel" id="mfb-panel">
        <div class="mfb-panel-header">
          <div class="mfb-panel-title">The Metrix · Feedback</div>
          <button class="mfb-close" onclick="MetrixFeedback._close()">✕</button>
        </div>

        <div id="mfb-form-wrap">
          <div class="mfb-panel-body">
            <div class="mfb-page-badge" id="mfb-page-badge">${PAGE}</div>

            <div class="mfb-type-label">Type</div>
            <div class="mfb-type-grid">
              <button class="mfb-type-btn feature" onclick="MetrixFeedback._selectType('feature', this)">
                <span class="mfb-type-icon">✨</span> Feature Request
              </button>
              <button class="mfb-type-btn bug" onclick="MetrixFeedback._selectType('bug', this)">
                <span class="mfb-type-icon">🐛</span> Bug Report
              </button>
              <button class="mfb-type-btn idea" onclick="MetrixFeedback._selectType('idea', this)">
                <span class="mfb-type-icon">💡</span> Idea / Suggestion
              </button>
              <button class="mfb-type-btn broken" onclick="MetrixFeedback._selectType('broken', this)">
                <span class="mfb-type-icon">🔥</span> This Is Broken
              </button>
            </div>

            <textarea class="mfb-textarea" id="mfb-text"
              placeholder="What's on your mind? The more specific, the better..."></textarea>

            <div class="mfb-name-row">
              <input class="mfb-name-input" id="mfb-name" placeholder="Your name (optional)"
                value="${USER || ''}" />
            </div>

            <button class="mfb-submit" id="mfb-submit" onclick="MetrixFeedback._submitWidget()">
              Send Feedback →
            </button>
          </div>
        </div>

        <div class="mfb-success" id="mfb-success">
          <div class="mfb-success-icon">🛸</div>
          <div class="mfb-success-title">Feedback received</div>
          <div class="mfb-success-sub">
            Logged locally and sent to the team.<br>
            This helps The Metrix get better.
          </div>
          <button class="mfb-success-another" onclick="MetrixFeedback._resetForm()">
            ↩ Submit another
          </button>
        </div>
      </div>

      <!-- TOAST -->
      <div class="mfb-toast" id="mfb-toast">
        <span class="mfb-toast-icon">✦</span>
        <span id="mfb-toast-text">Feedback saved</span>
      </div>
    `;
    document.body.appendChild(root);
  }

  // ===================== PANEL LOGIC =====================
  let panelOpen = false;
  let selectedType = null;

  function toggle() {
    panelOpen ? close() : open();
  }

  function open() {
    panelOpen = true;
    document.getElementById('mfb-panel').classList.add('open');
    document.getElementById('mfb-dot').classList.remove('visible');
    setTimeout(() => document.getElementById('mfb-text')?.focus(), 250);
  }

  function close() {
    panelOpen = false;
    document.getElementById('mfb-panel').classList.remove('open');
  }

  function selectType(type, btn) {
    selectedType = type;
    document.querySelectorAll('.mfb-type-btn').forEach(b => b.classList.remove('selected'));
    btn.classList.add('selected');
  }

  async function submitWidget() {
    const text = document.getElementById('mfb-text').value.trim();
    const name = document.getElementById('mfb-name').value.trim();

    if (!text && !selectedType) {
      document.getElementById('mfb-text').focus();
      document.getElementById('mfb-text').style.borderColor = '#ff6b6b';
      setTimeout(() => { document.getElementById('mfb-text').style.borderColor = ''; }, 1500);
      return;
    }

    const btn = document.getElementById('mfb-submit');
    btn.disabled = true;
    btn.textContent = 'Sending...';

    await submit({
      kind: 'widget',
      type: selectedType || 'general',
      text,
      user: name || USER || null,
    });

    document.getElementById('mfb-form-wrap').style.display = 'none';
    document.getElementById('mfb-success').classList.add('visible');
  }

  function resetForm() {
    selectedType = null;
    document.getElementById('mfb-text').value = '';
    document.getElementById('mfb-submit').disabled = false;
    document.getElementById('mfb-submit').textContent = 'Send Feedback →';
    document.querySelectorAll('.mfb-type-btn').forEach(b => b.classList.remove('selected'));
    document.getElementById('mfb-form-wrap').style.display = 'block';
    document.getElementById('mfb-success').classList.remove('visible');
  }

  // ===================== INLINE REACTIONS =====================
  // Track votes in localStorage keyed by component
  function getVotes() {
    try { return JSON.parse(localStorage.getItem('metrix_votes') || '{}'); } catch(e) { return {}; }
  }

  function setVote(component, sentiment) {
    const v = getVotes();
    v[component] = sentiment;
    localStorage.setItem('metrix_votes', JSON.stringify(v));
  }

  // count store: { component: { up: N, down: N } }
  function getCounts() {
    try { return JSON.parse(localStorage.getItem('metrix_reaction_counts') || '{}'); } catch(e) { return {}; }
  }

  function incrementCount(component, sentiment) {
    const c = getCounts();
    if (!c[component]) c[component] = { up: 0, down: 0 };
    c[component][sentiment]++;
    localStorage.setItem('metrix_reaction_counts', JSON.stringify(c));
    return c[component];
  }

  function buildReactionHTML(component, label) {
    const votes = getVotes();
    const counts = getCounts();
    const voted = votes[component];
    const up = counts[component]?.up || 0;
    const down = counts[component]?.down || 0;

    return `
      <div class="mfb-inline-reaction" data-mfb-component="${component}">
        <span class="mfb-react-label">${label || ''}</span>
        <button class="mfb-react-btn ${voted === 'up' ? 'voted-up' : ''}"
          onclick="MetrixFeedback.react('${component}', 'up')"
          ${voted ? 'disabled' : ''}>
          👍 <span class="mfb-react-count" id="mfb-count-up-${component}">${up || ''}</span>
        </button>
        <button class="mfb-react-btn ${voted === 'down' ? 'voted-down' : ''}"
          onclick="MetrixFeedback.react('${component}', 'down')"
          ${voted ? 'disabled' : ''}>
          👎 <span class="mfb-react-count" id="mfb-count-down-${component}">${down || ''}</span>
        </button>
      </div>`;
  }

  // Auto-discover [data-feedback-component] elements and inject reactions
  function autoInject() {
    document.querySelectorAll('[data-feedback-component]').forEach(el => {
      if (el.querySelector('.mfb-inline-reaction')) return; // already injected
      const component = el.getAttribute('data-feedback-component');
      const label = el.getAttribute('data-feedback-label') || '';
      const wrap = document.createElement('div');
      wrap.innerHTML = buildReactionHTML(component, label);
      el.appendChild(wrap.firstElementChild);
    });
  }

  async function react(component, sentiment, extra = {}) {
    const votes = getVotes();
    if (votes[component]) return; // already voted

    setVote(component, sentiment);
    const counts = incrementCount(component, sentiment);

    // Update UI
    const upEl = document.getElementById(`mfb-count-up-${component}`);
    const downEl = document.getElementById(`mfb-count-down-${component}`);
    if (upEl) upEl.textContent = counts.up || '';
    if (downEl) downEl.textContent = counts.down || '';

    // Disable both buttons
    const wrap = document.querySelector(`[data-mfb-component="${component}"]`);
    if (wrap) {
      wrap.querySelectorAll('.mfb-react-btn').forEach(btn => {
        btn.disabled = true;
        if (btn.onclick?.toString().includes(`'${sentiment}'`)) {
          btn.classList.add(sentiment === 'up' ? 'voted-up' : 'voted-down');
        }
      });
    }

    await submit({
      kind: 'reaction',
      component,
      sentiment,
      ...extra,
    });

    showToast(sentiment === 'up' ? '👍 Thanks for the signal!' : '👎 Noted — we\'ll look into it');
  }

  // ===================== TOAST =====================
  let toastTimer;
  function showToast(msg) {
    const t = document.getElementById('mfb-toast');
    document.getElementById('mfb-toast-text').textContent = msg;
    t.classList.add('show');
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => t.classList.remove('show'), 2800);
  }

  // ===================== CLOSE ON OUTSIDE CLICK =====================
  document.addEventListener('click', (e) => {
    if (!panelOpen) return;
    const panel = document.getElementById('mfb-panel');
    const trigger = document.getElementById('mfb-trigger');
    if (!panel.contains(e.target) && !trigger.contains(e.target)) close();
  }, true);

  // ===================== PUBLIC API =====================
  window.MetrixFeedback = {
    react,
    showToast,
    open,
    close,
    // Internal (used by inline onclick attrs)
    _toggle: toggle,
    _close: close,
    _selectType: selectType,
    _submitWidget: submitWidget,
    _resetForm: resetForm,
  };

  // ===================== INIT =====================
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => { inject(); autoInject(); });
  } else {
    inject();
    autoInject();
  }

  // Re-run autoInject on DOM mutations (for dynamically rendered cards)
  const observer = new MutationObserver(() => autoInject());
  observer.observe(document.body, { childList: true, subtree: true });

})();
