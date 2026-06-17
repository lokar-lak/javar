// ── API ───────────────────────────────────────────────────────────────────
const API = '';

async function apiFetch(path) {
  const res = await fetch(API + path);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

// ── Helpers ───────────────────────────────────────────────────────────────
function stars(n, total = 5) {
  n = Math.round(n);
  return Array.from({length: total}, (_, i) => i < n ? '★' : '☆').join('');
}

function releaseYear(dateStr) {
  if (!dateStr) return '';
  const localized = String(dateStr).match(/^(\d{2})-(\d{2})-(\d{4})$/);
  if (localized) return localized[3];
  return new Date(dateStr).getFullYear();
}

// Placeholder when cover image is missing
const PLACEHOLDER_GRADIENTS = [
  'linear-gradient(135deg,#1F1F2E,#2E2E44)',
  'linear-gradient(135deg,#16213e,#1F1F2E)',
  'linear-gradient(135deg,#0f3460,#1a1a2e)',
  'linear-gradient(135deg,#2d1b69,#1F1F2E)',
  'linear-gradient(135deg,#1a2a1a,#1F1F2E)',
  'linear-gradient(135deg,#3d0000,#1F1F2E)',
];

function coverStyle(game, idx = 0) {
  if (game.cover_url) return `background-image:url('${game.cover_url}')`;
  return `background:${PLACEHOLDER_GRADIENTS[idx % PLACEHOLDER_GRADIENTS.length]}`;
}

function coverImageHTML(game, idx = 0, options = {}) {
  const loading = options.loading || (idx < 3 ? 'eager' : 'lazy');
  const priority = options.priority || (idx === 0 ? 'high' : 'auto');
  if (!game.cover_url) return '';
  return `<img class="cover-img" src="${game.cover_url}" alt="${game.title || ''}" loading="${loading}" decoding="async" fetchpriority="${priority}" style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;object-position:center;display:block"/>`;
}

// Whether there are only AI translations (no manual ones)
// Backend computes has_only_ai in ListGames
function hasOnlyAI(game) {
  return !!game.has_only_ai;
}

function bestRating(game) {
  return game.best_rating || 0;
}

// ── Game Card ─────────────────────────────────────────────────────────────
function renderGameCard(game, idx = 0) {
  const genre = game.genres?.[0]?.name || '';
  const year  = releaseYear(game.release_date);
  const rating = bestRating(game);
  const aiOnly = hasOnlyAI(game);
  const off = game.has_official;
  const verified = game.has_verified;

  return `
    <div class="game-card" onclick="location.href='game.html?slug=${game.slug}'" style="cursor:pointer">
      <div class="game-card__cover game-card__cover--placeholder" style="${!game.cover_url ? coverStyle(game, idx) : ''}">
        ${coverImageHTML(game, idx)}
        ${!game.cover_url ? `<span style="font-size:40px;opacity:.2">🎮</span>` : ''}
        ${aiOnly ? `<div class="game-card__ai-badge">${t('common.translationTypes.ai').toUpperCase()}</div>` : ''}
        ${off ? `<div class="game-card__official-badge">${t('common.officialStatus.official').toUpperCase()}</div>` : ''}
        ${verified ? `<div class="game-card__verified-badge">${t('common.status.verifiedShort')}</div>` : ''}
      </div>
      <div class="game-card__body">
        <div class="game-card__meta">
          ${genre ? `<span class="badge badge--genre">${genre}</span>` : '<span></span>'}
          <span class="stars">${stars(rating)}</span>
        </div>
        <div class="game-card__title">${game.title}</div>
        <div class="game-card__dev">${game.developer}${year ? ` • ${year}` : ''}</div>
        <a href="game.html?slug=${game.slug}" class="btn btn--primary" style="width:fit-content">${t('common.actions.details')}</a>
      </div>
    </div>`;
}

// ── Skeleton grid ─────────────────────────────────────────────────────────
function skeletonGrid(n = 6) {
  return Array.from({length: n}, () => `
    <div class="game-card">
      <div style="aspect-ratio:16/9" class="skeleton"></div>
      <div class="game-card__body" style="gap:10px">
        <div class="skeleton" style="height:16px;width:60%"></div>
        <div class="skeleton" style="height:20px;width:80%"></div>
        <div class="skeleton" style="height:14px;width:50%"></div>
        <div class="skeleton" style="height:32px;width:110px;margin-top:4px"></div>
      </div>
    </div>`).join('');
}
