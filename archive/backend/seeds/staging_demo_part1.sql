-- Pawradise STAGING demo seed, part 1: users, profiles, products, images.
-- STAGING ONLY (appdb_staging). Never apply to production.
-- Idempotent-ish: cleans prior demo rows first (demo emails + e2e junk).

BEGIN;

-- cleanup: e2e junk products and demo/test leftovers
DELETE FROM products WHERE slug LIKE 'e2e-widget-%';
DELETE FROM users WHERE email LIKE '%@pawradise.demo' OR email IN ('1@1.com');

-- demo users (ids 10-25), password for all: demo1234
-- bcrypt hash of 'demo1234' (cost 12)
INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES
(10, 'SatoshiFan',    'satoshifan@pawradise.demo',    '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(11, 'KernelPanic',   'kernelpanic@pawradise.demo',   '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(12, 'ByteMe',        'byteme@pawradise.demo',        '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(13, 'NullPointer',   'nullpointer@pawradise.demo',   '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(14, 'SegFaultSam',   'segfaultsam@pawradise.demo',   '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(15, 'CryptoCat',     'cryptocat@pawradise.demo',     '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(16, 'GasGuzzler',    'gasguzzler@pawradise.demo',    '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(17, 'HODLer99',      'hodler99@pawradise.demo',      '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(18, 'MergeConflict', 'mergeconflict@pawradise.demo', '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(19, 'RubberDuck',    'rubberduck@pawradise.demo',    '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(20, 'StackSmasher',  'stacksmasher@pawradise.demo',  '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(21, 'TestNetTess',   'testnettess@pawradise.demo',   '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(22, 'WeiWanderer',   'weiwanderer@pawradise.demo',   '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(23, 'ForkYeah',      'forkyeah@pawradise.demo',      '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(24, 'MainnetMara',   'mainnetmara@pawradise.demo',   '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now()),
(25, 'HexHound',      'hexhound@pawradise.demo',      '$2b$12$riOdVxDxrXzmWLtxyZKXVunIbciu9wncQYlflFORvjiabb3wl6A5W', now(), now())
ON CONFLICT (id) DO NOTHING;

-- profiles for demo users
INSERT INTO user_profiles (user_id, bio, avatar_url, wallet_address, created_at, updated_at) VALUES
(10, 'Full-stack dev. I buy icon packs like trading cards.', 'https://placehold.co/100x100/0a0a0f/00d4ff?text=SF', '0x8f3a4b2c9d1e5f60718293a4b5c6d7e8f90a1b2', now(), now()),
(11, 'Systems programmer. My uptime is longer than my sleep.', 'https://placehold.co/100x100/0a0a0f/ff0066?text=KP', '0x1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4', now(), now()),
(12, 'Frontend gremlin. CSS is my love language.', 'https://placehold.co/100x100/0a0a0f/00ff88?text=BM', '0xabcdef1234567890abcdef1234567890abcdef12', now(), now()),
(13, 'Backend engineer. It works on my machine, ships anyway.', 'https://placehold.co/100x100/0a0a0f/00d4ff?text=NP', '0x1234567890abcdef1234567890abcdef12345678', now(), now()),
(14, 'Game jam veteran. 48 hours, 0 regrets.', 'https://placehold.co/100x100/0a0a0f/ff0066?text=SS', '0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef', now(), now()),
(15, 'Pixel artist and chiptune enjoyer.', 'https://placehold.co/100x100/0a0a0f/00ff88?text=CC', '0xcafecafecafecafecafecafecafecafecafecafe', now(), now()),
(16, 'DeFi degen with a hardware wallet and trust issues.', 'https://placehold.co/100x100/0a0a0f/00d4ff?text=GG', '0x9999999999999999999999999999999999999999', now(), now()),
(17, 'Never sold. Buying assets since block 12M.', 'https://placehold.co/100x100/0a0a0f/ff0066?text=H9', '0x7777777777777777777777777777777777777777', now(), now()),
(18, 'Git wrangler. Rebase > merge, fight me.', 'https://placehold.co/100x100/0a0a0f/00ff88?text=MC', '0x5555555555555555555555555555555555555555', now(), now()),
(19, 'Rubber duck debugger for hire. Quack-driven development.', 'https://placehold.co/100x100/0a0a0f/00d4ff?text=RD', '0x3333333333333333333333333333333333333333', now(), now()),
(20, 'Security researcher. I read CVEs for fun.', 'https://placehold.co/100x100/0a0a0f/ff0066?text=SM', '0x1111111111111111111111111111111111111111', now(), now()),
(21, 'Testnet maximalist. Break it where it is free.', 'https://placehold.co/100x100/0a0a0f/00ff88?text=TT', '0x2222222222222222222222222222222222222222', now(), now()),
(22, 'Solidity student. My contracts compile on the third try.', 'https://placehold.co/100x100/0a0a0f/00d4ff?text=WW', '0x4444444444444444444444444444444444444444', now(), now()),
(23, 'Open source maintainer. PRs welcome, issues ignored.', 'https://placehold.co/100x100/0a0a0f/ff0066?text=FY', '0x6666666666666666666666666666666666666666', now(), now()),
(24, 'Mainnet only. No testnet energy in this house.', 'https://placehold.co/100x100/0a0a0f/00ff88?text=MM', '0x8888888888888888888888888888888888888888', now(), now()),
(25, 'Low-level tinkerer. I dream in hex dumps.', 'https://placehold.co/100x100/0a0a0f/00d4ff?text=HH', '0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', now(), now())
ON CONFLICT (user_id) DO NOTHING;

-- 25 new products (ids 10-34) across all 8 categories
INSERT INTO products (id, title, slug, description, category_id, price_usd, asset_path, asset_hash, status, download_count_limit, max_downloads_per_user, file_size_bytes, file_mime_type, created_at, updated_at) VALUES
(10, 'Neon Terminal Wallpaper Pack', 'neon-terminal-wallpaper-pack', '40 hand-crafted 4K wallpapers: terminal grids, phosphor glow, synthwave skylines. PNG + JPG.', 1, 14.00, '', '', 'active', 0, 0, 48200000, 'application/zip', now(), now()),
(11, 'Cyberpunk UI Kit Dark', 'cyberpunk-ui-kit-dark', '120+ dark-mode components for dashboards: cards, modals, charts, tables. Figma + HTML/CSS.', 1, 29.00, '', '', 'active', 0, 0, 96500000, 'application/zip', now(), now()),
(12, 'Glitch Logo Templates', 'glitch-logo-templates', '25 editable glitch-effect logo templates. SVG sources plus After Effects project files.', 1, 19.00, '', '', 'active', 0, 0, 31400000, 'application/zip', now(), now()),
(13, 'React Admin Dashboard Pro', 'react-admin-dashboard-pro', 'Production-ready admin dashboard: auth pages, charts, data tables, dark mode. React + Vite + TS.', 2, 39.00, '', '', 'active', 0, 0, 128000000, 'application/zip', now(), now()),
(14, 'Python CLI Starter Kit', 'python-cli-starter-kit', 'Opinionated CLI boilerplate: argparse setup, logging, config files, packaging, tests included.', 2, 15.00, '', '', 'active', 0, 0, 8400000, 'application/zip', now(), now()),
(15, 'Go Microservice Boilerplate', 'go-microservice-boilerplate', 'Gin + Postgres + Docker Compose starter with health checks, migrations, and graceful shutdown.', 2, 24.00, '', '', 'active', 0, 0, 12600000, 'application/zip', now(), now()),
(16, 'VSCode Neon Theme Pack', 'vscode-neon-theme-pack', '6 glowing editor themes with matching terminal palettes and icon tweaks.', 2, 7.00, '', '', 'active', 0, 0, 3200000, 'application/zip', now(), now()),
(17, 'Low-Poly Spaceship Fleet', 'lowpoly-spaceship-fleet', '12 game-ready spaceships, under 2k tris each. FBX + OBJ + Blend with PBR textures.', 3, 22.00, '', '', 'active', 0, 0, 210000000, 'application/zip', now(), now()),
(18, 'Voxel City Blocks', 'voxel-city-blocks', '60 modular voxel buildings, roads, and props. MagicaVoxel sources plus Unity prefabs.', 3, 16.00, '', '', 'active', 0, 0, 88000000, 'application/zip', now(), now()),
(19, 'Chiptune Loop Library', 'chiptune-loop-library', '200 royalty-free chiptune loops at 140 BPM. WAV + MP3, key and tempo tagged.', 4, 11.00, '', '', 'active', 0, 0, 640000000, 'application/zip', now(), now()),
(20, 'Dark Synth Presets 100', 'dark-synth-presets-100', '100 presets for Vital and Serum: basses, pads, leads, arps. All macros mapped.', 4, 13.00, '', '', 'active', 0, 0, 45000000, 'application/zip', now(), now()),
(21, 'Dev Portfolio Template', 'dev-portfolio-template', 'Single-page portfolio with project grid, blog section, and contact form. Plain HTML/CSS/JS.', 5, 12.00, '', '', 'active', 0, 0, 9800000, 'application/zip', now(), now()),
(22, 'Freelance Contract Bundle', 'freelance-contract-bundle', 'MSA, SOW, NDA, and invoice templates reviewed by a tech lawyer. DOCX + PDF.', 5, 9.00, '', '', 'active', 0, 0, 4100000, 'application/pdf', now(), now()),
(23, 'Notion Dev Journal', 'notion-dev-journal', 'Daily standup, sprint retro, and learning-log templates. Duplicate-and-go Notion setup.', 5, 6.00, '', '', 'active', 0, 0, 1200000, 'application/zip', now(), now()),
(24, 'Solidity Basics Course', 'solidity-basics-course', '6-hour video course plus exercises: from first contract to testnet deployment. MP4 + repo.', 6, 49.00, '', '', 'active', 0, 0, 2400000000, 'application/zip', now(), now()),
(25, 'Rust for JS Devs Course', 'rust-for-js-devs-course', 'Ownership, lifetimes, and async explained through JavaScript analogies. 4 hours + katas.', 6, 45.00, '', '', 'active', 0, 0, 1900000000, 'application/zip', now(), now()),
(26, 'Git Wizardry Workshop', 'git-wizardry-workshop', 'Rebase, bisect, worktrees, and reflog rescue missions. Slides + exercise repos.', 6, 19.00, '', '', 'active', 0, 0, 67000000, 'application/zip', now(), now()),
(27, 'Pixel Platformer Tileset', 'pixel-platformer-tileset', '480 tiles at 16x16: grass, caves, ice, lava worlds plus animated coins and doors. PNG.', 7, 17.00, '', '', 'active', 0, 0, 23000000, 'application/zip', now(), now()),
(28, 'Retro RPG Sprite Pack', 'retro-rpg-sprite-pack', '80 characters with 4-direction walk cycles, 30 enemies, tiles for two towns. Aseprite files.', 7, 21.00, '', '', 'active', 0, 0, 39000000, 'application/zip', now(), now()),
(29, 'Tower Defense Starter Godot', 'tower-defense-starter-godot', 'Complete Godot 4 project: waves, towers, upgrades, save system. MIT-licensed code.', 7, 27.00, '', '', 'active', 0, 0, 54000000, 'application/zip', now(), now()),
(30, 'ASCII Art Bundle 500', 'ascii-art-bundle-500', '500 fruity ASCII banners, dividers, and logos for READMEs and commit messages. TXT.', 8, 8.00, '', '', 'active', 0, 0, 900000, 'text/plain', now(), now()),
(31, 'Terminal Cheatsheet Posters', 'terminal-cheatsheet-posters', 'Git, Docker, kubectl, and vim cheatsheets as print-ready A2 posters. PDF.', 8, 6.00, '', '', 'active', 0, 0, 18000000, 'application/pdf', now(), now()),
(32, 'Dotfiles Pro Pack', 'dotfiles-pro-pack', 'Bash, zsh, tmux, and nvim configs with install script. Documented line by line.', 8, 10.00, '', '', 'active', 0, 0, 5600000, 'application/zip', now(), now()),
(33, 'API Mock Server Toolkit', 'api-mock-server-toolkit', 'Dockerized mock API with scenario files, latency simulation, and webhook playback.', 8, 18.00, '', '', 'active', 0, 0, 31000000, 'application/zip', now(), now()),
(34, 'Markdown Blog Engine', 'markdown-blog-engine', 'Zero-dependency static blog: markdown in, RSS + sitemap out. Single Go binary.', 8, 23.00, '', '', 'active', 0, 0, 15000000, 'application/zip', now(), now())
ON CONFLICT (id) DO NOTHING;

-- primary images for new products
INSERT INTO product_images (product_id, url, is_primary, created_at)
SELECT p.id, 'https://placehold.co/600x400/0a0a0f/00d4ff?text=' || replace(p.slug, '-', '+'), true, now()
FROM products p WHERE p.id BETWEEN 10 AND 34
ON CONFLICT DO NOTHING;

SELECT setval('users_id_seq', (SELECT max(id) FROM users));
SELECT setval('products_id_seq', (SELECT max(id) FROM products));
SELECT setval('product_images_id_seq', (SELECT max(id) FROM product_images));

COMMIT;
