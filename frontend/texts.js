const TEXTS = {
  common: {
    titlePrefix: 'JAVAR',
    nav: {
      home: 'Галоўная',
      catalog: 'Каталог',
      contacts: 'Кантакты',
      faq: 'FAQ',
      submit: 'Дадаць беларусізатар',
      searchPlaceholder: 'Пошук...'
    },
    footer: {
      copyright: '© {year} Javar. Каталог беларускіх лакалізацый. Усе правы абароненыя.'
    },
    actions: {
      details: 'Падрабязней',
      submit: 'Адправіць',
      submitting: 'Адпраўляю...',
      confirm: 'Так',
      cancel: 'Не',
      backToCatalog: 'Вярнуцца ў каталог'
    },
    status: {
      noRating: 'Яшчэ без ацэнак',
      loadingError: 'Адбылася памылка загрузкі.',
      gameNotFound: 'Гульня не знойдзена.',
      verified: 'Пераклад правераны',
      verifiedShort: 'Правераны',
      verifiedWithDate: 'Пераклад правераны {date}',
      incomplete: 'Пераклад няпоўны',
      incompleteShort: 'Няпоўны',
      broken: 'Пераклад не працуе',
      brokenShort: 'Не працуе'
    },
    labels: {
      reviewCount: 'Колькасць водгукаў: {count}',
      search: 'Пошук: {value}',
      genre: 'Жанр: {value}'
    },
    translationTypes: {
      manual: 'Ручны пераклад',
      ai: 'Машынны пераклад',
      manualShort: 'Ручны',
      aiShort: 'Машынны'
    },
    orthography: {
      academic: 'Акадэмічны',
      tarashkevitsa: 'Альтэрнатыўны',
      lacinka: 'Лацінка'
    },
    officialStatus: {
      official: 'Афіцыйны',
      semiOfficial: 'Паўафіцыйны',
      unofficial: 'Неафіцыйны'
    }
  },
  pages: {
    home: {
      title: 'Галоўная',
      statsLabel: 'Колькасць перакладаў',
      translationsTitle: 'Пераклады',
      officialCount: 'Афіцыйных',
      semiOfficialCount: 'Паўафіцыйных',
      unofficialCount: 'Неафіцыйных',
      newReleases: 'Новыя рэлізы',
      gameOfDay: 'Гульня дня',
      aboutTitle: 'Пра нас',
      aboutText: 'Javar — спроба сабраць пад адным дахам увесь размаіты і разнастайны свет беларускіх лакалізацый. Раскіданыя па розных пляцоўках, яны часта застаюцца незаўважанымі і недаацэненымі. Мы хочам змяніць гэта, стварыўшы месца, дзе можна лёгка знайсці і ацаніць працу беларускіх перакладчыка. На Javar вы знойдзеце каталог гульняў з беларускімі лакалізацыямі, рэйтынгі і водгукі супольнасці, а таксама магчымасць падзяліцца сваім досведам і ўражаннямі. Наша мэта — не толькі папулярызаваць беларускія лакалізацыі, але і стварыць актыўную і падтрымліваючую супольнасць вакол іх.',
      noGames: 'Гульняў пакуль няма.'
    },
    catalog: {
      title: 'Каталог',
      heading: 'Каталог',
      searchPlaceholder: 'Назва або распрацоўшчык...',
      genre: 'Жанр',
      allGenres: 'Усе жанры',
      type: 'Тып перакладу',
      allTypes: 'Усе тыпы',
      orthography: 'Правапіс',
      anyOrthography: 'Любы правапіс',
      status: 'Статус',
      anyStatus: 'Любы статус',
      translator: 'Аўтар перакладу',
      allTranslators: 'Усе аўтары',
      sort: 'Сартаванне',
      sortCreatedDesc: 'Па даце стварэння беларусізатара (новыя)',
      sortCreatedAsc: 'Па даце стварэння беларусізатара (старыя)',
      sortReleaseDesc: 'Па даце выхаду гульні (новыя)',
      sortReleaseAsc: 'Па даце выхаду гульні (старыя)',
      sortSteamDesc: 'Па рэйтынгу Steam',
      sortSteamAsc: 'Па рэйтынгу Steam (адваротны)',
      sortTranslationDesc: 'Па рэйтынгу перакладу',
      sortTranslationAsc: 'Па рэйтынгу перакладу (адваротны)',
      emptyTitle: 'Нічога не знойдзена',
      emptyHint: 'Паспрабуйце змяніць фільтры',
      errorTitle: 'Адбылася памылка загрузкі',
      errorHint: 'Праверце, ці запушчаны сервер',
      prevPage: '← Назад',
      nextPage: 'Далей →'
    },
    game: {
      title: 'Гульня',
      steamPositive: 'станоўчых водгукаў',
      translationRating: 'Рэйтынг перакладу',
      avgRating: 'сярэдняя ацэнка',
      translationsCount: 'Перакладаў',
      available: 'даступна',
      totalClicks: 'Усяго пераходаў',
      searchClicks: 'на пошук',
      releaseDate: 'Дата выхаду',
      developer: 'Распрацоўшчык',
      publisher: 'Выдавец',
      genres: 'Жанры',
      platforms: 'Платформы',
      aboutGame: 'Пра гульню',
      translations: 'Пераклады',
      translatorFallback: 'Беларусізатар',
      clicks: 'Пераходаў: {count}',
      authors: 'Аўтары',
      studio: 'Студыя',
      translated: 'Перакладзена',
      findTranslation: 'Знайсці пераклад ↗',
      reviews: 'Водгукі',
      reviewsFor: 'Водгукі — {name}',
      leaveReview: 'Пакінуць водгук',
      reviewUpdateHint: 'Калі вы ўжо ацэньвалі гэты пераклад, ваша адзнака абновіцца.',
      replaceReviewConfirm: 'Вы ўжо ацэньвалі гэты пераклад. Хочаце змяніць сваю ацэнку?',
      nameOptional: 'Імя (неабавязкова)',
      anonymous: 'Выпадковае імя',
      rating: 'Рэйтынг',
      comment: 'Каментарый (неабавязкова)',
      reviewPlaceholder: 'Можна пакінуць кароткі водгук...',
      noReviews: 'Водгукаў пакуль няма. Будзьце першым!',
      updateError: 'Адбылася памылка абнаўлення.',
      linkError: 'Не ўдалося атрымаць спасылку',
      chooseRating: 'Вы не ацанілі гульню. Калі ласка, пастаўце адзнаку.',
      writeComment: 'Напішыце каментарый',
      submitError: 'Адбылася памылка адпраўкі'
    },
    contacts: {
      title: 'Кантакты',
      heading: 'Звяжыцеся з намі',
      text: 'Ёсць пытанні, прапановы або хочаце далучыцца да праекту?<br/>Напішыце нам у Telegram-бот, і мы адкажам як мага хутчэй.',
      botLabel: 'Наш Telegram-бот:'
    },
    faq: {
      title: 'FAQ',
      heading: 'Частыя пытанні',
      intro: 'Адказы на асноўныя пытанні пра беларусізатары, спасылкі і даданне новых лакалізацый у каталог.',
      items: [
        { question: 'Што рабіць, калі спасылка на беларусізатар не адкрываецца ці не вышукваецца?', answer: 'Звяжыцеся з намі праз наш тэлеграм-бот, мы пастараемся як мага хутчэй паправіць праблему.' },
        { question: 'Я ведаю гульню, у якой ёсць беларуская лакалізацыя/беларусізатар. Ці можна яе дадаць на сайт?', answer: 'Нават трэба! Калі ласка, запоўніце нашу форму, націснуўшы кнопку "Дадаць беларусізатар".' },
        { question: 'Дзе знайсці інструкцыю ўсталявання беларусізатара?', answer: 'Інструкцыі ўсталявання пішуцца аўтарамі перакладу даступныя на сайце, дзе апублікаваны беларусізатар. Мы, на жаль, не можам заўсёды гарантаваць, што яны будуць актуальны.' },
        { question: 'Што рабіць, калі я ўсталёўваю беларусіатар, а ён не працуе?', answer: 'У такіх выпадках найлепей звяртацца непасрэдна да аўтараў перакладу.' },
        { question: 'Я хачу зрабіць сваю лакалізацыю гульні X, але не ведаю, з чаго пачаць. Што мне рабіць?', answer: 'На пачатку, ацаніць свае сілы, магчымасці і вольны час :). Ствараць лакалізацыі не так проста, як здаецца. Калі ўсё ж гэта вас не пужае, то можаце смела звяртацца да нас па дапамогу праз наш тэлеграм-бот.' }
      ]
    }
  },
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

function t(path, vars = {}) {
  const value = path.split('.').reduce((acc, key) => acc?.[key], TEXTS);
  if (value == null) return path;
  if (typeof value !== 'string') return value;
  return value.replace(/\{(\w+)\}/g, (_, key) => vars[key] ?? '');
}

function applyI18n(root = document) {
  const vars = { year: new Date().getFullYear() };
  root.querySelectorAll('[data-i18n]').forEach(el => { el.innerHTML = t(el.dataset.i18n, vars); });
  root.querySelectorAll('[data-i18n-text]').forEach(el => { el.textContent = t(el.dataset.i18nText, vars); });
  root.querySelectorAll('[data-i18n-placeholder]').forEach(el => { el.placeholder = t(el.dataset.i18nPlaceholder); });
  root.querySelectorAll('[data-i18n-aria-label]').forEach(el => { el.setAttribute('aria-label', t(el.dataset.i18nAriaLabel)); });
  const titleKey = document.documentElement.dataset.titleKey;
  if (titleKey) document.title = t(titleKey);
}

document.addEventListener('DOMContentLoaded', () => applyI18n());

window.TEXTS = TEXTS;
window.t = t;
window.applyI18n = applyI18n;
