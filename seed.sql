INSERT INTO sites (name, url, enabled) VALUES
    ('СберБанк', 'https://sberbank.ru', TRUE),
    ('ВТБ', 'https://vtb.ru', TRUE),
    ('Альфа-Банк', 'https://alfabank.ru', TRUE),
    ('Т-Банк', 'https://tbank.ru', TRUE),
    ('Газпромбанк', 'https://gazprombank.ru', TRUE),
    ('Банки.ру', 'https://banki.ru', TRUE),

    ('Ozon', 'https://ozon.ru', TRUE),
    ('Wildberries', 'https://wildberries.ru', TRUE),
    ('Яндекс Маркет', 'https://market.yandex.ru', TRUE),
    ('Авито', 'https://avito.ru', TRUE),
    ('DNS', 'https://dns-shop.ru', TRUE),
    ('М.Видео', 'https://mvideo.ru', TRUE),

    ('Яндекс Музыка', 'https://music.yandex.ru', TRUE),
    ('VK Музыка', 'https://vk.com/music', TRUE),
    ('Звук', 'https://zvuk.com', TRUE),
    ('Кинопоиск', 'https://kinopoisk.ru', TRUE),
    ('Иви', 'https://ivi.ru', TRUE),
    ('Rutube', 'https://rutube.ru', TRUE),

    ('МТС', 'https://mts.ru', TRUE),
    ('МегаФон', 'https://megafon.ru', TRUE),
    ('Билайн', 'https://beeline.ru', TRUE),
    ('Ростелеком', 'https://rt.ru', TRUE),
    ('T2', 'https://t2.ru', TRUE),

    ('СДЭК', 'https://cdek.ru', TRUE),
    ('Почта России', 'https://pochta.ru', TRUE),
    ('HeadHunter', 'https://hh.ru', TRUE),
    ('Госуслуги', 'https://gosuslugi.ru', TRUE),
    ('РЖД', 'https://rzd.ru', TRUE),
    ('Аэрофлот', 'https://aeroflot.ru', TRUE),
    ('VK', 'https://vk.com', TRUE)
ON CONFLICT DO NOTHING;
