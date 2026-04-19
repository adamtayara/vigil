'use strict';

// Tab navigation
document.querySelectorAll('.nav-tab').forEach(tab => {
  tab.addEventListener('click', () => {
    const target = tab.dataset.target;
    document.querySelectorAll('.nav-tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.module-section').forEach(s => s.classList.remove('active'));
    tab.classList.add('active');
    const section = document.getElementById(target);
    if (section) section.classList.add('active');
  });
});

// Severity filter cards in hero
document.querySelectorAll('.sev-card[data-filter]').forEach(card => {
  card.addEventListener('click', () => {
    const filter = card.dataset.filter;
    card.classList.toggle('active');
    applyFilters();
  });
});

// Finding card expand/collapse
document.querySelectorAll('.finding-header').forEach(header => {
  header.addEventListener('click', () => {
    header.closest('.finding-card').classList.toggle('open');
  });
});

// Filter buttons within sections
document.querySelectorAll('.filter-btn[data-sev]').forEach(btn => {
  btn.addEventListener('click', () => {
    btn.classList.toggle('active');
    filterSection(btn.closest('.module-section'));
  });
});

// Search inputs
document.querySelectorAll('.filter-input').forEach(input => {
  input.addEventListener('input', () => {
    filterSection(input.closest('.module-section'));
  });
});

function filterSection(section) {
  if (!section) return;
  const activeFilters = Array.from(section.querySelectorAll('.filter-btn.active[data-sev]'))
                             .map(b => parseInt(b.dataset.sev));
  const searchText = (section.querySelector('.filter-input')?.value || '').toLowerCase();

  section.querySelectorAll('.finding-card').forEach(card => {
    const sev = parseInt(card.dataset.sev);
    const text = card.textContent.toLowerCase();
    const sevMatch = activeFilters.length === 0 || activeFilters.includes(sev);
    const textMatch = !searchText || text.includes(searchText);
    card.style.display = sevMatch && textMatch ? '' : 'none';
  });
}

function applyFilters() {
  const activeFilters = Array.from(document.querySelectorAll('.sev-card.active[data-filter]'))
                             .map(c => parseInt(c.dataset.filter));
  document.querySelectorAll('.finding-card').forEach(card => {
    const sev = parseInt(card.dataset.sev);
    if (activeFilters.length === 0) {
      card.style.display = '';
    } else {
      card.style.display = activeFilters.includes(sev) ? '' : 'none';
    }
  });
}

// Sortable table columns
document.querySelectorAll('th[data-sort]').forEach(th => {
  th.style.cursor = 'pointer';
  th.addEventListener('click', () => {
    const table = th.closest('table');
    const col = parseInt(th.dataset.sort);
    const rows = Array.from(table.querySelectorAll('tbody tr'));
    const asc = th.dataset.order !== 'asc';
    th.dataset.order = asc ? 'asc' : 'desc';

    rows.sort((a, b) => {
      const aVal = a.cells[col]?.textContent.trim() || '';
      const bVal = b.cells[col]?.textContent.trim() || '';
      const aNum = parseFloat(aVal.replace(/[^0-9.-]/g, ''));
      const bNum = parseFloat(bVal.replace(/[^0-9.-]/g, ''));
      if (!isNaN(aNum) && !isNaN(bNum)) return asc ? aNum - bNum : bNum - aNum;
      return asc ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
    });

    const tbody = table.querySelector('tbody');
    rows.forEach(r => tbody.appendChild(r));
  });
});

// Score ring animation
(function () {
  const ring = document.querySelector('.score-track');
  const fill = document.querySelector('.score-fill');
  if (!ring || !fill) return;

  const r = parseFloat(fill.getAttribute('r'));
  const circ = 2 * Math.PI * r;
  const score = parseInt(document.querySelector('.score-num')?.textContent) || 0;

  fill.style.strokeDasharray = circ;
  fill.style.strokeDashoffset = circ;

  let color = '#22c55e';
  if (score < 80) color = '#f97316';
  if (score < 60) color = '#ef4444';
  fill.style.stroke = color;

  requestAnimationFrame(() => {
    fill.style.transition = 'stroke-dashoffset 1s ease-out';
    fill.style.strokeDashoffset = circ * (1 - score / 100);
  });

  // Color the score number too
  const numEl = document.querySelector('.score-num');
  if (numEl) numEl.style.color = color;
})();

// Auto-open first tab
(function () {
  const firstTab = document.querySelector('.nav-tab');
  if (firstTab) firstTab.click();
})();
