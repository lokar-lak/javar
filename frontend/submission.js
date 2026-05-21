(() => {
  const platforms = [
    'macOS', 'iOS', 'Linux', 'PlayStation5', 'Android', 'Windows',
    'Xbox360', 'XboxSeries', 'XboxOne', 'Switch', 'WiiU', '3DS',
    'PlayStation4', 'Іншае'
  ];

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
        <button class="submission-close" type="button" aria-label="Закрыць" data-close-submission>×</button>
        <h2 id="submission-title">Дадаць беларусізатар</h2>
        <form class="submission-form" id="submission-form">
          <div class="submission-field">
            <label for="submission-game-title">Назва гульні</label>
            <input id="submission-game-title" name="game_title" type="text" required />
          </div>

          <fieldset class="submission-field">
            <legend>Платформы</legend>
            <p>Абярыце 1 ці болей платформ, дзе ёсць беларусізатар.</p>
            <div class="submission-pills submission-pills--platforms">
              ${platforms.map(p => optionPill('checkbox', 'platforms', p).replace(' required', '')).join('')}
            </div>
          </fieldset>

          <div class="submission-grid">
            <fieldset class="submission-field">
              <legend>Катэгорыя лакалізацыі</legend>
              <div class="submission-checks">
                ${optionPill('radio', 'category', 'Афіцыйная')}
                ${optionPill('radio', 'category', 'Неафіцыйная')}
              </div>
            </fieldset>

            <fieldset class="submission-field">
              <legend>Тып лакалізацыі</legend>
              <div class="submission-checks">
                ${optionPill('checkbox', 'localization_type', 'Тэкст').replace(' required', '')}
                ${optionPill('checkbox', 'localization_type', 'Агучванне').replace(' required', '')}
              </div>
            </fieldset>
          </div>

          <div class="submission-field">
            <label for="submission-authors">Аўтары лакалізацыі</label>
            <input id="submission-authors" name="authors" type="text" required />
          </div>

          <div class="submission-field">
            <label for="submission-game-url">Спасылка на гульню</label>
            <input id="submission-game-url" name="game_url" type="url" required />
          </div>

          <div class="submission-field">
            <label for="submission-translation-url">Спасылка на беларусізатар</label>
            <input id="submission-translation-url" name="translation_url" type="url" required />
          </div>

          <div class="submission-field">
            <label for="submission-description">Апісанне лакалізацыі</label>
            <textarea id="submission-description" name="description" required></textarea>
          </div>

          <div class="submission-error" id="submission-error" role="alert"></div>
          <div class="submission-actions">
            <button class="submission-submit" type="submit">Адправіць</button>
          </div>
        </form>
      </div>`;
    document.body.appendChild(modal);
  }

  function selectedCount(name) {
    return document.querySelectorAll(`#submission-form input[name="${name}"]:checked`).length;
  }

  function showError(message) {
    const error = document.getElementById('submission-error');
    if (!error) return;
    error.textContent = message;
    error.classList.toggle('visible', !!message);
  }

  function openModal() {
    createModal();
    document.getElementById('submission-modal').classList.add('open');
    document.body.classList.add('submission-lock');
    setTimeout(() => document.getElementById('submission-game-title')?.focus(), 0);
  }

  function closeModal() {
    document.getElementById('submission-modal')?.classList.remove('open');
    document.body.classList.remove('submission-lock');
    showError('');
  }

  document.addEventListener('click', (event) => {
    if (event.target.closest('[data-open-submission]')) {
      event.preventDefault();
      openModal();
      return;
    }
    if (event.target.closest('[data-close-submission]') || event.target.id === 'submission-modal') {
      closeModal();
    }
  });

  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') closeModal();
  });

  document.addEventListener('submit', (event) => {
    if (event.target.id !== 'submission-form') return;
    event.preventDefault();
    if (selectedCount('platforms') === 0) {
      showError('Абярыце хаця б адну платформу.');
      return;
    }
    if (selectedCount('localization_type') === 0) {
      showError('Абярыце хаця б адзін тып лакалізацыі.');
      return;
    }
    showError('Захаванне прапановы будзе падключана на наступным этапе.');
  });
})();
