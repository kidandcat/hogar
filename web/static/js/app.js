/* ============================================
   Hogar - App JavaScript
   Vanilla ES6+, no frameworks
   ============================================ */

(function () {
  'use strict';

  // --- Fetch wrapper with auth error handling ---
  async function apiFetch(url, options = {}) {
    const defaults = {
      headers: {
        'Content-Type': 'application/json',
      },
    };
    const config = { ...defaults, ...options };
    if (options.headers) {
      config.headers = { ...defaults.headers, ...options.headers };
    }

    try {
      const response = await fetch(url, config);

      if (response.status === 401) {
        window.location.href = '/login';
        return null;
      }

      if (!response.ok) {
        throw new Error(`Error ${response.status}: ${response.statusText}`);
      }

      return response;
    } catch (error) {
      showToast(error.message || 'Error de conexión');
      throw error;
    }
  }

  // --- Toast notifications ---
  function showToast(message, duration = 3000) {
    const toast = document.getElementById('toast');
    if (!toast) return;

    toast.textContent = message;
    toast.classList.add('visible');

    clearTimeout(toast._timeout);
    toast._timeout = setTimeout(() => {
      toast.classList.remove('visible');
    }, duration);
  }

  // --- Navigation: hamburger menu toggle ---
  function initNav() {
    const toggle = document.getElementById('nav-toggle');
    const links = document.getElementById('nav-links');
    if (!toggle || !links) return;

    toggle.addEventListener('click', () => {
      toggle.classList.toggle('active');
      links.classList.toggle('open');
    });

    // Close menu when clicking a link
    links.querySelectorAll('.nav-link').forEach((link) => {
      link.addEventListener('click', () => {
        toggle.classList.remove('active');
        links.classList.remove('open');
      });
    });

    // Close menu when clicking outside
    document.addEventListener('click', (e) => {
      if (!toggle.contains(e.target) && !links.contains(e.target)) {
        toggle.classList.remove('active');
        links.classList.remove('open');
      }
    });
  }

  // --- Shopping list ---
  function initShopping() {
    const shoppingList = document.getElementById('shopping-list');
    if (!shoppingList) return;

    const listId = shoppingList.dataset.listId;

    // Event delegation for item toggling
    shoppingList.addEventListener('click', (e) => {
      const item = e.target.closest('.shopping-item');
      if (!item) return;
      toggleItem(listId, item);
    });

    // Keyboard support
    shoppingList.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        const item = e.target.closest('.shopping-item');
        if (!item) return;
        e.preventDefault();
        toggleItem(listId, item);
      }
    });

    // List selector change
    const listSelect = document.getElementById('list-select');
    if (listSelect) {
      listSelect.addEventListener('change', (e) => {
        window.location.href = `/shopping?list=${e.target.value}`;
      });
    }

    // Initial progress update
    updateProgress();
  }

  function toggleItem(listId, itemEl) {
    const itemId = itemEl.dataset.itemId;
    const wasChecked = itemEl.dataset.checked === 'true';
    const newChecked = !wasChecked;

    // Optimistic UI update
    applyItemState(itemEl, newChecked);
    updateProgress();

    // Sync with server
    apiFetch(`/api/shopping-lists/${listId}/items/${itemId}`, {
      method: 'PATCH',
      body: JSON.stringify({ checked: newChecked }),
    }).catch(() => {
      // Revert on failure
      applyItemState(itemEl, wasChecked);
      updateProgress();
      showToast('No se pudo actualizar. Inténtalo de nuevo.');
    });
  }

  function applyItemState(itemEl, checked) {
    itemEl.dataset.checked = String(checked);
    itemEl.setAttribute('aria-checked', String(checked));

    if (checked) {
      itemEl.classList.add('shopping-item--checked');
      itemEl.querySelector('.shopping-item-check').innerHTML =
        '<svg viewBox="0 0 24 24" class="icon-check"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"/></svg>';
    } else {
      itemEl.classList.remove('shopping-item--checked');
      itemEl.querySelector('.shopping-item-check').innerHTML =
        '<svg viewBox="0 0 24 24" class="icon-circle"><circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" stroke-width="2"/></svg>';
    }
  }

  function updateProgress() {
    const items = document.querySelectorAll('.shopping-item');
    const progressFill = document.getElementById('progress-fill');
    const progressText = document.getElementById('progress-text');

    if (!items.length || !progressFill || !progressText) return;

    const total = items.length;
    const checked = document.querySelectorAll(
      '.shopping-item[data-checked="true"]'
    ).length;
    const percent = total > 0 ? (checked / total) * 100 : 0;

    progressFill.style.width = `${percent}%`;
    progressText.textContent = `${checked} / ${total}`;
  }

  // --- Menu expansion ---
  function initMenuExpand() {
    document.querySelectorAll('[data-menu-expand]').forEach((card) => {
      const header = card.querySelector('.menu-card-header');
      const body = card.querySelector('.menu-card-body');
      if (!header || !body) return;

      header.addEventListener('click', () => {
        const isExpanded = !body.hidden;
        body.hidden = isExpanded;
        card.classList.toggle('expanded', !isExpanded);
      });
    });
  }

  // --- Recipe expansion ---
  function initRecipeExpand() {
    document.querySelectorAll('[data-recipe-expand]').forEach((card) => {
      const header = card.querySelector('.recipe-card-header');
      const body = card.querySelector('.recipe-card-body');
      if (!header || !body) return;

      header.addEventListener('click', () => {
        const isExpanded = !body.hidden;
        body.hidden = isExpanded;
        card.classList.toggle('expanded', !isExpanded);
      });
    });

    // Auto-expand if URL has hash (e.g., #recipe-123)
    if (window.location.hash) {
      const target = document.querySelector(window.location.hash);
      if (target) {
        const body = target.querySelector('.recipe-card-body');
        if (body) {
          body.hidden = false;
          target.classList.add('expanded');
          requestAnimationFrame(() => {
            target.scrollIntoView({ behavior: 'smooth', block: 'start' });
          });
        }
      }
    }
  }

  // --- Initialize everything on DOM ready ---
  document.addEventListener('DOMContentLoaded', () => {
    initNav();
    initShopping();
    initMenuExpand();
    initRecipeExpand();
  });
})();
