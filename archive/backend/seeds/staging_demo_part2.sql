-- Pawradise STAGING demo seed, part 2: community + orders.
-- STAGING ONLY (appdb_staging). Never apply to production.

BEGIN;

-- drop the leftover junk test post
DELETE FROM community_posts WHERE id = 3;

-- 60 community posts (ids 10-69), spread over the last 10 days
INSERT INTO community_posts (id, user_id, content, created_at, updated_at) VALUES
(10, 10, 'Just grabbed the pixel icon pack and my side project finally looks legit. Best $12 I spent this month.', now() - interval '236 hours', now() - interval '236 hours'),
(11, 11, 'Does anyone have a good flow for verifying BSC payments automatically? Polling BscScan every 30s feels hacky.', now() - interval '232 hours', now() - interval '232 hours'),
(12, 14, 'Shipped my first godot prototype using the tower defense starter. Pathfinding still cursed but it runs.', now() - interval '228 hours', now() - interval '228 hours'),
(13, 12, 'Unpopular opinion: README-driven development saves more time than any framework.', now() - interval '224 hours', now() - interval '224 hours'),
(14, 15, 'The synthwave sound pack slaps. My stream intro finally has an identity.', now() - interval '220 hours', now() - interval '220 hours'),
(15, 22, 'Day 47 of learning Rust. The borrow checker and I are in couples therapy.', now() - interval '216 hours', now() - interval '216 hours'),
(16, 16, 'PSA: rotate your hot wallet keys. Ask me how I know. Actually do not ask.', now() - interval '212 hours', now() - interval '212 hours'),
(17, 14, 'Pixel platformer tileset plus three coffees equals finished game jam entry. Links in bio.', now() - interval '208 hours', now() - interval '208 hours'),
(18, 13, 'Is 500 icons enough? Asking for a dashboard with 501 menu items.', now() - interval '204 hours', now() - interval '204 hours'),
(19, 10, 'Just connected my wallet to Pawradise. The future is sending BNB into the void and hoping.', now() - interval '200 hours', now() - interval '200 hours'),
(20, 18, 'Code review tip: if the diff is bigger than your terminal scrollback, split the PR.', now() - interval '196 hours', now() - interval '196 hours'),
(21, 23, 'My dotfiles finally sync across three machines. I feel like a wizard.', now() - interval '192 hours', now() - interval '192 hours'),
(22, 15, 'Anyone tried the voxel city blocks in Blender? Import was painless.', now() - interval '188 hours', now() - interval '188 hours'),
(23, 17, 'Gas fees on BSC are so low I keep double-checking I actually paid.', now() - interval '184 hours', now() - interval '184 hours'),
(24, 19, 'Writing docs is just leaving messages for your future self who will ignore them.', now() - interval '180 hours', now() - interval '180 hours'),
(25, 15, 'The chiptune loops are perfect for coding sessions. 140 BPM focus mode.', now() - interval '176 hours', now() - interval '176 hours'),
(26, 17, 'Sold my first digital asset today. Cozy little payout, cozy little dance.', now() - interval '172 hours', now() - interval '172 hours'),
(27, 12, 'Hot take: dark mode is not a theme, it is a lifestyle.', now() - interval '168 hours', now() - interval '168 hours'),
(28, 13, 'My portfolio template deployment took 4 minutes. Four. I remember FTP.', now() - interval '164 hours', now() - interval '164 hours'),
(29, 11, 'Segfault at 3am hits different when the tests were green at midnight.', now() - interval '160 hours', now() - interval '160 hours'),
(30, 25, 'The ASCII art bundle is unreasonably fun. My commit messages have never looked better.', now() - interval '156 hours', now() - interval '156 hours'),
(31, 20, 'Question for sellers: do you price in USD and convert, or price natively in crypto?', now() - interval '152 hours', now() - interval '152 hours'),
(32, 23, 'Just followed half the community feed. My following list is now a tech conference.', now() - interval '148 hours', now() - interval '148 hours'),
(33, 14, 'Low-poly trees look great with flat shading. My diorama scene is coming together.', now() - interval '144 hours', now() - interval '144 hours'),
(34, 16, 'Reminder: your seed phrase is not a password. Write it on paper, not in a notes app.', now() - interval '140 hours', now() - interval '140 hours'),
(35, 13, 'The Go microservice boilerplate saved me a full weekend. Health checks included!', now() - interval '136 hours', now() - interval '136 hours'),
(36, 12, 'Anyone else collect terminal themes like trading cards? The neon pack is my new main.', now() - interval '132 hours', now() - interval '132 hours'),
(37, 23, 'Deployed the markdown blog engine on a $4 VPS. 40ms response times. Chef kiss.', now() - interval '128 hours', now() - interval '128 hours'),
(38, 22, 'Smart contract course starts Monday. My testnet wallet is ready and empty.', now() - interval '124 hours', now() - interval '124 hours'),
(39, 14, 'Retro RPG sprites plus dialogue system equals my weekend is gone and I regret nothing.', now() - interval '120 hours', now() - interval '120 hours'),
(40, 20, 'Pro tip: name your orders with the memo. Future-you doing taxes says thanks.', now() - interval '116 hours', now() - interval '116 hours'),
(41, 23, 'The freelance contract bundle paid for itself with the first client. Twice over.', now() - interval '112 hours', now() - interval '112 hours'),
(42, 18, 'Vim vs Emacs thread in 3... 2... 1... (I use VS Code with vim bindings, the coward way out)', now() - interval '108 hours', now() - interval '108 hours'),
(43, 13, 'API mock toolkit is genuinely great for frontend devs waiting on backend. Which is always.', now() - interval '104 hours', now() - interval '104 hours'),
(44, 10, 'Just hit 12 confirmations on my first order. Downloads unlocked. The dopamine is real.', now() - interval '100 hours', now() - interval '100 hours'),
(45, 17, 'Nobody talks about how nice instant delivery is. No shipping, no tracking number, just files.', now() - interval '96 hours', now() - interval '96 hours'),
(46, 14, 'My game assets folder is now 4GB of beautifully organized chaos.', now() - interval '92 hours', now() - interval '92 hours'),
(47, 12, 'CSS grid finally clicked for me after the dashboard template. It is just tables with extra steps.', now() - interval '88 hours', now() - interval '88 hours'),
(48, 18, 'The workshop on git wizardry taught me rebase onto. I feel dangerous now.', now() - interval '84 hours', now() - interval '84 hours'),
(49, 20, 'Do people actually read the memo field instructions? Asking because I forgot mine twice.', now() - interval '80 hours', now() - interval '80 hours'),
(50, 15, 'Chiptune plus mechanical keyboard ASMR equals peak cozy coding stream.', now() - interval '76 hours', now() - interval '76 hours'),
(51, 22, 'Bought the solidity course. Day 1: it compiles. I am basically a blockchain engineer now.', now() - interval '72 hours', now() - interval '72 hours'),
(52, 15, 'Pixel art is just spreadsheets with commitment issues.', now() - interval '68 hours', now() - interval '68 hours'),
(53, 23, 'My community feed is 50 percent shipping announcements and I love it.', now() - interval '64 hours', now() - interval '64 hours'),
(54, 23, 'The contract bundle has an NDA template that does not read like a curse scroll.', now() - interval '60 hours', now() - interval '60 hours'),
(55, 23, 'Started journaling dev progress in Notion. Streak: 6 days. The template helps.', now() - interval '56 hours', now() - interval '56 hours'),
(56, 14, 'Three.js plus low-poly fleet equals my portfolio hero section finally moves.', now() - interval '52 hours', now() - interval '52 hours'),
(57, 18, 'Why is naming things still the hardest problem? My slugs are poetry though.', now() - interval '48 hours', now() - interval '48 hours'),
(58, 25, 'The poster cheatsheets look great above my desk. Git rebase diagram equals wall art.', now() - interval '44 hours', now() - interval '44 hours'),
(59, 21, 'Testnet Tess is my new role model. Break things where it is free, folks.', now() - interval '40 hours', now() - interval '40 hours'),
(60, 17, 'Downloaded the same asset twice because unlimited downloads. No regrets, just backups.', now() - interval '36 hours', now() - interval '36 hours'),
(61, 17, 'Small shops like this beat marketplaces. No algorithm, just good packs.', now() - interval '32 hours', now() - interval '32 hours'),
(62, 10, 'My wallet connect took two clicks. Two! The UX bar has been raised.', now() - interval '28 hours', now() - interval '28 hours'),
(63, 15, 'Anyone pairing the dark synth presets with the chiptune loops? Genre fusion time.', now() - interval '24 hours', now() - interval '24 hours'),
(64, 13, 'Refactored my whole side project to use the CLI starter kit. Argparse never felt so good.', now() - interval '20 hours', now() - interval '20 hours'),
(65, 10, 'Order history page is clean. I can see exactly what I bought and when.', now() - interval '16 hours', now() - interval '16 hours'),
(66, 14, 'The spaceship fleet renders beautifully at low poly counts. Performance for days.', now() - interval '12 hours', now() - interval '12 hours'),
(67, 10, 'Writing a review for the icon pack: crisp at 16px, which is witchcraft.', now() - interval '8 hours', now() - interval '8 hours'),
(68, 19, 'Community question: what is everyone building this weekend?', now() - interval '4 hours', now() - interval '4 hours'),
(69, 21, 'First post here. Be kind. Currently deciding between three dashboard templates.', now() - interval '1 hours', now() - interval '1 hours')
ON CONFLICT (id) DO NOTHING;

-- likes: 3 deterministic likers per post (180 rows, deduped)
INSERT INTO post_likes (user_id, post_id, created_at)
SELECT ((p.id * 7 + i) % 16) + 10, p.id, p.created_at + (i || ' hours')::interval
FROM community_posts p CROSS JOIN generate_series(0, 2) AS i
WHERE p.id BETWEEN 10 AND 69
ON CONFLICT (user_id, post_id) DO NOTHING;

-- follows: each demo user follows the next 3 cyclically (~48 rows)
INSERT INTO follows (follower_id, following_id, created_at)
SELECT f, ((f - 10 + k) % 16) + 10, now() - ((k * 9) || ' hours')::interval
FROM generate_series(10, 25) AS f CROSS JOIN generate_series(1, 3) AS k
WHERE ((f - 10 + k) % 16) + 10 <> f
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- hand-written comments across popular posts
INSERT INTO post_comments (user_id, post_id, content, created_at) VALUES
(11, 19, 'Same, the two-click connect got me.', now() - interval '190 hours'),
(21, 19, 'Welcome to the void, we have downloads.', now() - interval '189 hours'),
(10, 44, 'That unlock moment is the best part of crypto commerce.', now() - interval '90 hours'),
(13, 11, 'BscScan API plus backoff works fine for low volume. Cache aggressively.', now() - interval '225 hours'),
(16, 11, 'We run a cron every minute. Good enough until real webhooks.', now() - interval '224 hours'),
(19, 15, 'It gets better. Then lifetimes happen and it gets worse again. Worth it.', now() - interval '210 hours'),
(11, 26, 'Congrats! First sale energy is unmatched.', now() - interval '165 hours'),
(21, 51, 'Day 30 me says: slow down, test everything twice.', now() - interval '65 hours'),
(13, 37, '40ms on $4 hardware is the dream. What stack?', now() - interval '120 hours'),
(23, 37, 'Markdown plus SQLite, probably. Classic cozy combo.', now() - interval '119 hours'),
(19, 42, 'Coward here too. Vim bindings everywhere.', now() - interval '100 hours'),
(12, 49, 'I set a calendar reminder named MEMO now.', now() - interval '70 hours'),
(15, 29, 'Green at midnight, red at 3am. The circle of dev life.', now() - interval '150 hours'),
(14, 68, 'Building a voxel diorama. You?', now() - interval '3 hours'),
(22, 68, 'Smart contract bingo card. Free space: out of gas.', now() - interval '2 hours'),
(13, 68, 'Polishing my portfolio. Again. Forever.', now() - interval '2 hours'),
(14, 56, 'Performance for days indeed. What frame budget?', now() - interval '45 hours'),
(23, 21, 'Share the sync method! Symlinks or a bare repo?', now() - interval '185 hours'),
(21, 21, 'Bare repo plus alias. Changed my life.', now() - interval '184 hours'),
(24, 34, 'Paper gang. Laminated, even.', now() - interval '130 hours'),
(17, 61, 'Exactly. Curation over algorithm.', now() - interval '25 hours'),
(10, 45, 'Instant delivery converted me. Never going back to shipping queues.', now() - interval '88 hours'),
(15, 25, '140 BPM is the optimal commit tempo, science says.', now() - interval '168 hours'),
(18, 57, 'Slug poetry should be a conference talk.', now() - interval '40 hours'),
(13, 64, 'Argparse supremacy. Click users stay quiet.', now() - interval '15 hours'),
(23, 17, 'Link? Would love to try the jam entry.', now() - interval '200 hours'),
(14, 17, 'Dropped it in my profile bio!', now() - interval '199 hours'),
(14, 39, 'No regrets is the correct gamedev mindset.', now() - interval '112 hours'),
(23, 53, 'Shipping announcements are my love language.', now() - interval '55 hours'),
(13, 35, 'Health checks included?? Sold.', now() - interval '128 hours');

-- 8 sample orders (mix of paid + pending) with items
INSERT INTO orders (id, user_id, status, total_usd, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations, payment_confirmed_at, paid_at, created_at, updated_at) VALUES
(1, 10, 'paid', 37.00, '0.10571428', 'bsc', '0xDemoSellerWallet000000000000000000000001', '0xaaa111aaa111aaa111aaa111aaa111aaa111aaa111aaa111aaa111aaa1', 15, now() - interval '96 hours', now() - interval '96 hours', now() - interval '100 hours', now() - interval '96 hours'),
(2, 11, 'pending', 18.00, '0.05142857', 'bsc', '0xDemoSellerWallet000000000000000000000001', NULL, 1, NULL, NULL, now() - interval '30 hours', now() - interval '30 hours'),
(3, 12, 'paid', 39.00, '0.11142857', 'bsc', '0xDemoSellerWallet000000000000000000000001', '0xbbb222bbb222bbb222bbb222bbb222bbb222bbb222bbb222bbb222bbb2', 20, now() - interval '70 hours', now() - interval '70 hours', now() - interval '74 hours', now() - interval '70 hours'),
(4, 13, 'pending', 43.00, '0.12285714', 'bsc', '0xDemoSellerWallet000000000000000000000001', NULL, 0, NULL, NULL, now() - interval '10 hours', now() - interval '10 hours'),
(5, 14, 'paid', 49.00, '0.14000000', 'bsc', '0xDemoSellerWallet000000000000000000000001', '0xccc333ccc333ccc333ccc333ccc333ccc333ccc333ccc333ccc333ccc3', 25, now() - interval '50 hours', now() - interval '50 hours', now() - interval '55 hours', now() - interval '50 hours'),
(6, 15, 'pending', 34.00, '0.09714285', 'bsc', '0xDemoSellerWallet000000000000000000000001', NULL, 2, NULL, NULL, now() - interval '5 hours', now() - interval '5 hours'),
(7, 16, 'paid', 20.00, '0.05714285', 'bsc', '0xDemoSellerWallet000000000000000000000001', '0xddd444ddd444ddd444ddd444ddd444ddd444ddd444ddd444ddd444ddd4', 18, now() - interval '26 hours', now() - interval '26 hours', now() - interval '28 hours', now() - interval '26 hours'),
(8, 17, 'pending', 6.00, '0.01714285', 'bsc', '0xDemoSellerWallet000000000000000000000001', NULL, 0, NULL, NULL, now() - interval '2 hours', now() - interval '2 hours')
ON CONFLICT (id) DO NOTHING;

INSERT INTO order_items (order_id, product_id, quantity, price_usd, price_crypto, download_count, max_downloads, created_at) VALUES
(1, 1, 1, 12.00, '0.03428571', 2, 0, now() - interval '100 hours'),
(1, 2, 1, 25.00, '0.07142857', 1, 0, now() - interval '100 hours'),
(2, 3, 1, 18.00, '0.05142857', 0, 0, now() - interval '30 hours'),
(3, 13, 1, 39.00, '0.11142857', 3, 0, now() - interval '74 hours'),
(4, 10, 1, 14.00, '0.04000000', 0, 0, now() - interval '10 hours'),
(4, 11, 1, 29.00, '0.08285714', 0, 0, now() - interval '10 hours'),
(5, 24, 1, 49.00, '0.14000000', 1, 0, now() - interval '55 hours'),
(6, 27, 2, 34.00, '0.09714285', 0, 0, now() - interval '5 hours'),
(7, 4, 1, 9.00, '0.02571428', 4, 0, now() - interval '28 hours'),
(7, 19, 1, 11.00, '0.03142857', 2, 0, now() - interval '28 hours'),
(8, 31, 1, 6.00, '0.01714285', 0, 0, now() - interval '2 hours')
ON CONFLICT DO NOTHING;

SELECT setval('community_posts_id_seq', (SELECT max(id) FROM community_posts));
SELECT setval('post_likes_id_seq', (SELECT max(id) FROM post_likes));
SELECT setval('post_comments_id_seq', (SELECT max(id) FROM post_comments));
SELECT setval('follows_id_seq', (SELECT max(id) FROM follows));
SELECT setval('orders_id_seq', (SELECT max(id) FROM orders));
SELECT setval('order_items_id_seq', (SELECT max(id) FROM order_items));

COMMIT;
