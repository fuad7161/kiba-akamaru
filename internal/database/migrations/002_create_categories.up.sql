CREATE TABLE categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    name_bn     VARCHAR(100),
    slug        VARCHAR(100) NOT NULL UNIQUE,
    icon        VARCHAR(50),
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO categories (name, name_bn, slug, icon, sort_order) VALUES
  ('BCS',        'বিসিএস',      'bcs',        '⭐', 1),
  ('Bank Jobs',  'ব্যাংক চাকরি', 'bank-jobs',  '🏦', 2),
  ('Defense',    'সেনাবাহিনী',  'defense',    '🛡', 3),
  ('Police',     'পুলিশ',        'police',     '👮', 4),
  ('Education',  'শিক্ষা',       'education',  '🎓', 5),
  ('Health',     'স্বাস্থ্য',    'health',     '⚕',  6),
  ('Ministry',   'মন্ত্রণালয়', 'ministry',   '🏛', 7),
  ('Engineering','প্রকৌশল',     'engineering','⚙',  8),
  ('Railway',    'রেলওয়ে',      'railway',    '🚆', 9),
  ('Others',     'অন্যান্য',    'others',     '📋', 10);
