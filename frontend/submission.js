(() => {
  const fallbackTexts = {
    common: { actions: { submit: 'Адправіць', submitting: 'Адпраўляю...' } },
    submission: {
      platforms: ['macOS', 'iOS', 'Linux', 'PlayStation', 'Android', 'Windows', 'Xbox', 'Nintendo Switch', 'Іншая'],
      close: 'Закрыць',
      title: 'Дадаць беларусізатар',
      gameTitle: 'Назва гульні',
      platformsLegend: 'Платформы',
      platformsHint: 'Выберыце 1 ці болей платформ, дзе ёсць беларусізатар.',
      category: 'Катэгорыя лакалізацыі',
      official: 'Афіцыйная',
      unofficial: 'Неафіцыйная',
      localizationType: 'Тып лакалізацыі',
      text: 'Тэкст',
      voice: 'Агучванне',
      authors: 'Аўтары лакалізацыі',
      gameUrl: 'Спасылка на гульню',
      translationUrl: 'Спасылка на беларусізатар',
      description: 'Апісанне лакалізацыі',
      platformRequired: 'Выберыце хаця б адну платформу.',
      typeRequired: 'Выберыце хаця б адзін тып лакалізацыі.',
      duplicate: 'Падобная гульня ўжо ёсць у каталогу{names}',
      genericError: 'Не ўдалося адправіць прапанову.',
      success: 'Дзякуй! Прапанова адпраўлена на мадэрацыю.',
      networkError: 'Адбылася памылка інтэрнэт-злучэння. Паспрабуйце яшчэ раз.'
    }
  };
  const tr = window.t || ((path, vars = {}) => {
    const value = path.split('.').reduce((acc, key) => acc?.[key], fallbackTexts);
    if (typeof value !== 'string') return value;
    return value.replace(/\{(\w+)\}/g, (_, key) => vars[key] ?? '');
  });
  const copy = tr('submission');
  const platforms = copy.platforms;

  function optionPill(type, name, label) {
    return `
      <label class="submission-pill">
        <input type="${type}" name="${name}" value="${label}" required />
        <span>${label}</span>
      </label>`;
  }

  function createModal() {
    if (document.getElementById('submission-modal')) return;

    const modal = document.createElement('div');
    modal.className = 'submission-modal';
    modal.id = 'submission-modal';
    modal.innerHTML = `
      <div class="submission-dialog" role="dialog" aria-modal="true" aria-labelledby="submission-title">
        <button class="submission-close" type="button" aria-label="${copy.close}" data-close-submission>×</button>
        <h2 id="submission-title">${copy.title}</h2>
        <form class="submission-form" id="submission-form">
          <div class="submission-field">
            <label for="submission-game-title">${copy.gameTitle}</label>
            <input id="submission-game-title" name="game_title" type="text" required />
          </div>

          <fieldset class="submission-field">
            <legend>${copy.platformsLegend}</legend>
            <p>${copy.platformsHint}</p>
            <div class="submission-pills submission-pills--platforms">
              ${platforms.map(p => optionPill('checkbox', 'platforms', p).replace(' required', '')).join('')}
            </div>
          </fieldset>

          <div class="submission-grid">
            <fieldset class="submission-field">
              <legend>${copy.category}</legend>
              <div class="submission-checks">
                <label class="submission-pill">
                  <input type="radio" name="category" value="official" required />
                  <span>${copy.official}</span>
                </label>
                <label class="submission-pill">
                  <input type="radio" name="category" value="unofficial" required />
                  <span>${copy.unofficial}</span>
                </label>
              </div>
            </fieldset>

            <fieldset class="submission-field">
              <legend>${copy.localizationType}</legend>
              <div class="submission-checks">
                ${optionPill('checkbox', 'localization_type', copy.text).replace(' required', '')}
                ${optionPill('checkbox', 'localization_type', copy.voice).replace(' required', '')}
              </div>
            </fieldset>
          </div>

          <div class="submission-field submission-unofficial-field" hidden>
            <label for="submission-authors">${copy.authors}</label>
            <input id="submission-authors" name="authors" type="text" />
          </div>

          <div class="submission-field">
            <label for="submission-game-url">${copy.gameUrl}</label>
            <input id="submission-game-url" name="game_url" type="url" required />
          </div>

          <div class="submission-field submission-unofficial-field" hidden>
            <label for="submission-translation-url">${copy.translationUrl}</label>
            <input id="submission-translation-url" name="translation_url" type="url" />
          </div>

          <div class="submission-field">
            <label for="submission-description">${copy.description}</label>
            <textarea id="submission-description" name="description"></textarea>
          </div>

          <div class="submission-error" id="submission-error" role="alert"></div>
          <div class="submission-actions">
            <button class="submission-submit" type="submit">${tr('common.actions.submit')}</button>
          </div>
        </form>
      </div>`;
    document.body.appendChild(modal);
  }

  function selectedCount(name) {
    return document.querySelectorAll(`#submission-form input[name="${name}"]:checked`).length;
  }

  function selectedValues(name) {
    return [...document.querySelectorAll(`#submission-form input[name="${name}"]:checked`)].map(input => input.value);
  }

  function showError(message) {
    const error = document.getElementById('submission-error');
    if (!error) return;
    error.textContent = message;
    error.classList.toggle('visible', !!message);
  }

  function showSuccess(message) {
    const toast = document.createElement('div');
    toast.className = 'submission-toast';
    toast.textContent = message;
    document.body.appendChild(toast);
    requestAnimationFrame(() => toast.classList.add('show'));
    setTimeout(() => {
      toast.classList.remove('show');
      setTimeout(() => toast.remove(), 250);
    }, 3200);
  }

  function openModal() {
    createModal();
    updateUnofficialFields();
    document.getElementById('submission-modal').classList.add('open');
    document.body.classList.add('submission-lock');
    setTimeout(() => document.getElementById('submission-game-title')?.focus(), 0);
  }

  function closeModal() {
    document.getElementById('submission-modal')?.classList.remove('open');
    document.body.classList.remove('submission-lock');
    showError('');
  }

  function setSubmitting(isSubmitting) {
    const button = document.querySelector('#submission-form .submission-submit');
    if (!button) return;
    button.disabled = isSubmitting;
    button.textContent = isSubmitting ? tr('common.actions.submitting') : tr('common.actions.submit');
  }

  function updateUnofficialFields() {
    const isUnofficial = document.querySelector('#submission-form input[name="category"]:checked')?.value === 'unofficial';
    document.querySelectorAll('.submission-unofficial-field').forEach(field => {
      field.hidden = !isUnofficial;
      field.querySelectorAll('input, textarea, select').forEach(input => {
        if (input.id === 'submission-authors') input.required = isUnofficial;
        if (!isUnofficial) input.value = '';
      });
    });
  }

  document.addEventListener('click', (event) => {
    if (event.target.closest('[data-open-submission]')) {
      event.preventDefault();
      openModal();
      return;
    }
    if (event.target.closest('[data-close-submission]')) {
      closeModal();
    }
  });

  document.addEventListener('change', (event) => {
    if (event.target.closest('#submission-form input[name="category"]')) {
      updateUnofficialFields();
    }
  });

  document.addEventListener('submit', async (event) => {
    if (event.target.id !== 'submission-form') return;
    event.preventDefault();
    showError('');
    if (selectedCount('platforms') === 0) {
      showError(copy.platformRequired);
      return;
    }
    if (selectedCount('localization_type') === 0) {
      showError(copy.typeRequired);
      return;
    }
    const form = event.target;
    const data = new FormData(form);
    const category = data.get('category') || '';
    const body = {
      game_title: data.get('game_title')?.trim() || '',
      platforms: selectedValues('platforms'),
      category: category,
      localization_type: selectedValues('localization_type'),
      authors: category === 'unofficial' ? (data.get('authors')?.trim() || '') : '',
      game_url: data.get('game_url')?.trim() || '',
      translation_url: category === 'unofficial' ? (data.get('translation_url')?.trim() || '') : '',
      description: data.get('description')?.trim() || ''
    };

    setSubmitting(true);
    try {
      const res = await fetch('/api/translation-submissions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      const payload = await res.json().catch(() => ({}));
      if (res.status === 409) {
        const names = (payload.similar_games || []).map(g => `«${g.title}»`).join(', ');
        showError(tr('submission.duplicate', { names: names ? ': ' + names : '.' }));
        return;
      }
      if (!res.ok) {
        showError(payload.error || copy.genericError);
        return;
      }
      form.reset();
      updateUnofficialFields();
      closeModal();
      showSuccess(copy.success);
    } catch {
      showError(copy.networkError);
    } finally {
      setSubmitting(false);
    }
  });
})();
