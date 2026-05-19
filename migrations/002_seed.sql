INSERT INTO genres (name, slug) VALUES
    ('Хорар',     'horror'),
    ('Трылер',    'thriller'),
    ('РПГ',       'rpg'),
    ('Экшн',      'action'),
    ('Прыгоды',   'adventure'),
    ('Стратэгія', 'strategy'),
    ('Сімулятар', 'simulator');

INSERT INTO games (title, slug, developer, publisher, release_date, description, cover_url, steamdb_url, steam_rating) VALUES
(
    'Silent Hill 2 (2025)',
    'silent-hill-2-2025',
    'Bloober Team',
    'Konami',
    '2024-10-08',
    'Атрымаўшы ліст ад сваёй мёртвай жонкі, Джэймс вяртаецца туды, дзе яны перажылі столькі ўспамінаў — у Сайлент-Хіл. Але замест гэтага ён знаходзіць горад-прывід, ахутаны густым туманам і напоўнены жахлівымі пачварамі. Супрацьстаяць гэтым істотам, вырашаць загадкі і шукаць сляды сваёй жонкі — усё гэта ў рэмейку культавага хорару.',
    '',
    'https://steamdb.info/app/2124490/',
    83
);

INSERT INTO game_genres (game_id, genre_id) VALUES (1, 1), (1, 2);

-- Two different translations for one game
INSERT INTO translations (game_id, studio_name, translator_names, type, coverage, external_url) VALUES
(
    1,
    'Студыя АБВ',
    '["Аляксей Краўчанка", "Марына Петрова", "Іван Сідараў"]',
    'manual',
    '["Субтытры", "Меню", "Інвентар", "Дакументы"]',
    'https://www.nexusmods.com/silenthill2/mods/1'
),
(
    1,
    'Аўтаперакладчык 9000',
    '["GPT-4o"]',
    'ai',
    '["Субтытры"]',
    'https://www.nexusmods.com/silenthill2/mods/2'
);

-- Reviews for the first translation (manual)
INSERT INTO reviews (translation_id, author_name, rating, body) VALUES
(1, 'Аляксей К.',  5, 'Выдатны пераклад! Тэкст гучыць натуральна па-беларуску, нідзе не ўзнікала адчування машыннасці. Асабліва спадабалася перадача атмасферы — жахлівыя сцэны яшчэ больш уражваюць на роднай мове.'),
(1, 'Марына П.',   4, 'У цэлым пераклад вельмі добры, але некаторыя назвы прадметаў у інвентары выглядаюць крыху няёмка. Субтытры ідуць сінхронна, праблем не заўважыла.'),
(1, 'Дзмітрый В.', 3, 'Заўважыў некалькі памылак у субтытрах у трэцяй главе. Некаторыя фразы гучаць занадта літаральна. Спадзяюся на выпраўленне.');

-- Review for the second translation (AI)
INSERT INTO reviews (translation_id, author_name, rating, body) VALUES
(2, 'Ганна Л.', 3, 'ШІ-пераклад ёсць ШІ-пераклад. Для разумення сюжэту сыдзе, але стыль кульгае і часам губляецца кантэкст. Лепш браць ручны ад Студыі АБВ.');
