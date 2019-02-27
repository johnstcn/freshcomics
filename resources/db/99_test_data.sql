INSERT INTO site_defs (
  name,
  active,
  nsfw,
  start_url,
  url_template,
  ref_xpath,
  ref_regexp,
  title_xpath,
  title_regexp
  ) VALUES (
  'Emma & The Granny Fairies',
  TRUE,
  FALSE,
  'http://grannyfairies.com/p1-once.html',
  'http://grannyfairies.com/%s.html',
  '//p[@class="nav"]/a[@class="on"]/@href',
  '([^/]+)\.html$',
  '//img[@class="comic"]/@alt',
  '(.+)'
);