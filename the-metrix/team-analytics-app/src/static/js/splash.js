/**
 * THE METRIX · Splash Screen
 * X Fly-In animation - shows on first visit per day
 */

(function() {
  'use strict';

  const SPLASH_KEY = 'metrix_splash_shown';
  const SPLASH_DURATION = 3500; // ms total animation time

  // Check if should show splash (once per day)
  function shouldShowSplash() {
    const lastShown = localStorage.getItem(SPLASH_KEY);
    if (!lastShown) return true;

    const lastDate = new Date(parseInt(lastShown));
    const today = new Date();

    // Different day = show splash
    return lastDate.toDateString() !== today.toDateString();
  }

  function markSplashShown() {
    localStorage.setItem(SPLASH_KEY, Date.now().toString());
  }

  // Create splash HTML
  function createSplashHTML() {
    return `
      <div class="metrix-splash-overlay" id="metrix-splash">
        <canvas class="metrix-splash-canvas" id="splash-canvas"></canvas>
        <div class="metrix-splash-scanlines"></div>
        <div class="metrix-splash-vignette"></div>

        <div class="metrix-splash-word-stage">
          <div style="position: relative">
            <div class="metrix-splash-the-prefix" id="splash-the">THE</div>
            <div class="metrix-splash-word-letters">
              <span class="metrix-splash-letter dim" id="splash-m">M</span>
              <span class="metrix-splash-letter dim" id="splash-e">E</span>
              <span class="metrix-splash-letter dim" id="splash-t">T</span>
              <span class="metrix-splash-letter dim" id="splash-r">R</span>
              <span class="metrix-splash-letter dim" id="splash-i">I</span>
              <span class="metrix-splash-letter" id="splash-c">C</span>
              <span class="metrix-splash-letter" id="splash-x" style="position:absolute;opacity:0">X</span>
            </div>
            <div class="metrix-splash-tagline" id="splash-tagline">Engineering Intelligence Platform</div>
          </div>
        </div>

        <div class="metrix-splash-progress">
          <div class="metrix-splash-progress-track">
            <div class="metrix-splash-progress-fill" id="splash-progress"></div>
          </div>
          <div class="metrix-splash-progress-label" id="splash-label">Initializing...</div>
        </div>
      </div>
    `;
  }

  // Inject CSS
  function injectStyles() {
    const style = document.createElement('style');
    style.textContent = `
      .metrix-splash-overlay {
        position: fixed;
        inset: 0;
        z-index: 99999;
        background: #0a0a0f;
        display: flex;
        align-items: center;
        justify-content: center;
        overflow: hidden;
      }

      .metrix-splash-canvas {
        position: absolute;
        inset: 0;
        width: 100%;
        height: 100%;
      }

      .metrix-splash-scanlines {
        position: absolute;
        inset: 0;
        background: repeating-linear-gradient(
          0deg,
          rgba(0,0,0,0.15),
          rgba(0,0,0,0.15) 1px,
          transparent 1px,
          transparent 2px
        );
        opacity: 0.4;
        pointer-events: none;
        z-index: 2;
      }

      .metrix-splash-vignette {
        position: absolute;
        inset: 0;
        background: radial-gradient(
          circle at center,
          transparent 40%,
          rgba(10,10,15,0.6) 100%
        );
        pointer-events: none;
        z-index: 1;
      }

      .metrix-splash-word-stage {
        position: relative;
        z-index: 10;
        display: flex;
        align-items: center;
        justify-content: center;
        text-align: center;
      }

      .metrix-splash-the-prefix {
        font-family: 'Syne', sans-serif;
        font-weight: 700;
        font-size: 16px;
        letter-spacing: 0.4em;
        color: #6b6b8a;
        margin-bottom: 8px;
        opacity: 0;
      }

      .metrix-splash-word-letters {
        display: flex;
        gap: 4px;
        justify-content: center;
        position: relative;
      }

      .metrix-splash-letter {
        font-family: 'Syne', sans-serif;
        font-weight: 900;
        font-size: 72px;
        letter-spacing: -0.02em;
        color: #e8e8f0;
        opacity: 0;
        position: relative;
      }

      .metrix-splash-letter.dim {
        color: #6b6b8a;
      }

      .metrix-splash-tagline {
        font-family: 'DM Mono', monospace;
        font-size: 11px;
        letter-spacing: 0.1em;
        color: #4fffb0;
        margin-top: 20px;
        opacity: 0;
      }

      .metrix-splash-progress {
        position: absolute;
        bottom: 80px;
        left: 50%;
        transform: translateX(-50%);
        width: 320px;
        z-index: 20;
      }

      .metrix-splash-progress-track {
        width: 100%;
        height: 2px;
        background: rgba(79,255,176,0.1);
        border-radius: 2px;
        overflow: hidden;
        margin-bottom: 10px;
      }

      .metrix-splash-progress-fill {
        height: 100%;
        background: linear-gradient(90deg, #4fffb0, #7c6dfa);
        width: 0%;
        transition: width 0.3s ease;
      }

      .metrix-splash-progress-label {
        font-family: 'DM Mono', monospace;
        font-size: 10px;
        color: #6b6b8a;
        text-align: center;
        letter-spacing: 0.06em;
      }

      @media (max-width: 600px) {
        .metrix-splash-letter {
          font-size: 48px;
        }
        .metrix-splash-the-prefix {
          font-size: 12px;
        }
        .metrix-splash-tagline {
          font-size: 9px;
        }
      }
    `;
    document.head.appendChild(style);
  }

  // Setup canvas
  function setupCanvas() {
    const canvas = document.getElementById('splash-canvas');
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    return canvas;
  }

  // Progress bar animation
  function runProgress(duration) {
    const fill = document.getElementById('splash-progress');
    const label = document.getElementById('splash-label');
    const msgs = [
      'Initializing...',
      'Loading metrics...',
      'Syncing DORA...',
      'Building insights...',
      'Ready.'
    ];

    let elapsed = 0;
    const interval = 50;
    const timer = setInterval(() => {
      elapsed += interval;
      const pct = Math.min(100, (elapsed / duration) * 100);
      fill.style.width = pct + '%';
      const msgIdx = Math.floor((pct / 100) * (msgs.length - 1));
      label.textContent = msgs[msgIdx];

      if (pct >= 100) {
        clearInterval(timer);
        label.textContent = 'Ready.';
      }
    }, interval);
  }

  // X Fly-in animation
  function startFlyIn() {
    const ids = ['splash-m', 'splash-e', 'splash-t', 'splash-r', 'splash-i', 'splash-c'];
    const stagger = 80; // ms between letter appearances

    // Fade in METRIC letters with stagger
    ids.forEach((id, i) => {
      const el = document.getElementById(id);
      setTimeout(() => {
        el.style.transition = 'opacity 300ms ease, transform 300ms ease';
        el.style.transform = 'translateY(0)';
        el.style.opacity = '1';
      }, 200 + i * stagger);
    });

    // "THE" fades in
    setTimeout(() => {
      const prefix = document.getElementById('splash-the');
      prefix.style.transition = 'opacity 400ms ease';
      prefix.style.opacity = '1';
    }, 100);

    // X flies in from right, replaces C
    const cEl = document.getElementById('splash-c');
    const xEl = document.getElementById('splash-x');
    const flyDelay = 200 + ids.length * stagger + 200;

    setTimeout(() => {
      // Position X to replace C
      xEl.style.left = cEl.offsetLeft + 'px';
      xEl.style.top = '0px';

      setTimeout(() => {
        // C flickers
        let flickers = 0;
        const flickerInterval = setInterval(() => {
          cEl.style.opacity = cEl.style.opacity === '0' ? '1' : '0';
          flickers++;
          if (flickers >= 6) {
            clearInterval(flickerInterval);
            cEl.style.opacity = '0';
            launchX(xEl, cEl);
          }
        }, 60);
      }, 300);
    }, flyDelay);

    // Tagline
    setTimeout(() => {
      const tl = document.getElementById('splash-tagline');
      tl.style.transition = 'opacity 600ms ease';
      tl.style.opacity = '1';
    }, flyDelay + 1200);

    runProgress(flyDelay + 1400);
  }

  // Launch X with motion blur trails
  function launchX(xEl, cEl) {
    const canvas = setupCanvas();
    const ctx = canvas.getContext('2d');
    const w = canvas.width, h = canvas.height;

    // Get X final position (using C's position as reference)
    const cRect = cEl.getBoundingClientRect();
    const finalX = cRect.left + cRect.width / 2;
    const finalY = cRect.top + cRect.height / 2;

    const travel = 600; // distance X travels from right
    let progress = 0;
    const duration = 700; // ms
    const fontSize = parseInt(window.getComputedStyle(cEl).fontSize);
    const trailCount = 8;

    function easeOutExpo(t) {
      return t === 1 ? 1 : 1 - Math.pow(2, -15 * t);
    }

    function drawTrail(x, alpha, blur) {
      ctx.save();
      ctx.globalAlpha = alpha;
      ctx.filter = `blur(${blur}px)`;
      ctx.font = `900 ${fontSize}px Syne, sans-serif`;
      ctx.fillStyle = `hsl(156, 100%, ${50 + blur * 2}%)`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText('X', x, finalY);
      ctx.restore();
    }

    let lastTime = null;

    function frame(now) {
      if (!lastTime) lastTime = now;
      const dt = (now - lastTime) / duration;
      lastTime = now;
      progress = Math.min(1, progress + dt);
      const eased = easeOutExpo(progress);

      // Clear canvas
      ctx.fillStyle = 'rgba(10, 10, 15, 1)';
      ctx.fillRect(0, 0, w, h);

      // Motion blur trails
      for (let i = trailCount; i >= 0; i--) {
        const trailX = finalX + travel * (1 - easeOutExpo(Math.max(0, progress - i * 0.02)));
        const alpha = (1 - i / trailCount) * 0.15;
        const blur = (i / trailCount) * 8;
        if (alpha > 0.01) drawTrail(trailX, alpha, blur);
      }

      // Main X with glow
      const ix = finalX + travel * (1 - eased);
      const glowIntensity = Math.min(1, eased * 2);

      // Outer glow
      ctx.save();
      ctx.globalAlpha = glowIntensity * 0.6;
      ctx.filter = 'blur(20px)';
      ctx.font = `900 ${fontSize}px Syne, sans-serif`;
      ctx.fillStyle = '#4fffb0';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText('X', ix, finalY);
      ctx.restore();

      // Main letter
      ctx.save();
      ctx.font = `900 ${fontSize}px Syne, sans-serif`;
      ctx.fillStyle = '#e8e8f0';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.shadowBlur = 30 * glowIntensity;
      ctx.shadowColor = '#4fffb0';
      ctx.fillText('X', ix, finalY);
      ctx.restore();

      if (progress < 1) {
        requestAnimationFrame(frame);
      } else {
        // Animation complete - show X in DOM, hide canvas
        setTimeout(() => {
          xEl.style.opacity = '1';
          xEl.style.transition = 'opacity 200ms ease';
          canvas.style.opacity = '0';
          canvas.style.transition = 'opacity 200ms ease';
        }, 100);
      }
    }

    requestAnimationFrame(frame);
  }

  // Remove splash screen
  function removeSplash() {
    const splash = document.getElementById('metrix-splash');
    if (splash) {
      splash.style.opacity = '0';
      splash.style.transition = 'opacity 400ms ease';
      setTimeout(() => splash.remove(), 400);
    }
  }

  // Initialize splash
  function init() {
    if (!shouldShowSplash()) return;

    injectStyles();

    const splashDiv = document.createElement('div');
    splashDiv.innerHTML = createSplashHTML();
    document.body.appendChild(splashDiv.firstElementChild);

    markSplashShown();

    // Start animation after a brief delay
    setTimeout(() => {
      startFlyIn();
    }, 100);

    // Auto-remove after duration
    setTimeout(removeSplash, SPLASH_DURATION);
  }

  // Run on DOM ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
