/**
 * Skeleton Loader Component
 *
 * Universal loading state component for consistent skeleton screens.
 *
 * Usage:
 *   <skeleton-loader type="kpi" count="4"></skeleton-loader>
 *   <skeleton-loader type="chart" height="300"></skeleton-loader>
 *   <skeleton-loader type="table" rows="10"></skeleton-loader>
 *
 * Attributes:
 *   - type: 'kpi' | 'chart' | 'table' | 'bar' | 'stat' | 'text' | 'metric'
 *   - count: number of items to render (for kpi type)
 *   - rows: number of rows (for table type)
 *   - height: custom height in pixels (for chart type)
 */
class SkeletonLoader extends HTMLElement {
  static get observedAttributes() {
    return ['type', 'count', 'rows', 'height'];
  }

  connectedCallback() {
    this.render();
  }

  attributeChangedCallback() {
    if (this.isConnected) {
      this.render();
    }
  }

  render() {
    const type = this.getAttribute('type') || 'kpi';
    const count = parseInt(this.getAttribute('count') || '1');
    const rows = parseInt(this.getAttribute('rows') || '5');
    const height = this.getAttribute('height') || '300';

    let html = '';

    if (type === 'kpi') {
      // Grid layout for KPI cards
      html = '<div class="kpi-grid" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 14px;">';
      for (let i = 0; i < count; i++) {
        html += '<div class="skeleton skeleton-kpi"></div>';
      }
      html += '</div>';

    } else if (type === 'chart') {
      // Single chart skeleton with custom height
      html = `<div class="skeleton skeleton-chart" style="height: ${height}px"></div>`;

    } else if (type === 'table') {
      // Table skeleton with multiple rows
      html = '<div class="skeleton-table">';
      for (let i = 0; i < rows; i++) {
        html += '<div class="skeleton skeleton-bar"></div>';
      }
      html += '</div>';

    } else if (type === 'bar') {
      // Single bar skeleton (for lists)
      for (let i = 0; i < count; i++) {
        html += '<div class="skeleton skeleton-bar"></div>';
      }

    } else if (type === 'stat') {
      // Stat card skeletons
      for (let i = 0; i < count; i++) {
        html += '<div class="skeleton skeleton-stat"></div>';
      }

    } else if (type === 'text') {
      // Text line skeletons
      for (let i = 0; i < count; i++) {
        html += '<div class="skeleton skeleton-text"></div>';
      }

    } else if (type === 'metric') {
      // Metric skeletons (taller than text)
      for (let i = 0; i < count; i++) {
        html += '<div class="skeleton skeleton-metric"></div>';
      }
    }

    this.innerHTML = html;
  }
}

// Register the custom element
customElements.define('skeleton-loader', SkeletonLoader);
