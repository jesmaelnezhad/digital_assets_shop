/* Pawradise prototype — in-browser backend. Persist in localStorage. */
(function (global) {
  const ROOT = () => global.PROTOTYPE_ROOT || "./";
  const media = (file) => {
    if (!file) return "";
    if (/^(data:|https?:|blob:)/i.test(file)) return file;
    return ROOT() + "shared/media/" + file;
  };
  const KEY = "pawradise-proto-v6";
  const defaultSettings = {
    referral_commission_percent: "5.0",
    payment_address: "0xPAWRADISE_WALLET_BSC",
    site_name: "Pawradise",
    site_title: "Pawradise — digital assets",
    site_description: "Buy once, download forever. Clay characters, UI kits, textures, and scenes.",
    site_keywords: "digital assets, 3d, ui kits, textures, environments",
    og_image: "",
    canonical_host: "https://pawradise.ir",
    download_policy: "Buy once, download forever. Unlimited re-downloads on paid orders.",
    default_currency: "USD",
    robots_index: "true"
  };

  const categorySeed = [
    { id: 1, slug: "3d", name: "3D & characters", desc: "Sculpts, kits, and game-ready meshes", parentId: 0, sort: 1, active: true },
    { id: 2, slug: "ui", name: "UI kits", desc: "Dashboards, components, tokens", parentId: 0, sort: 2, active: true },
    { id: 3, slug: "textures", name: "Textures", desc: "Materials, HDRI, surfaces", parentId: 0, sort: 3, active: true },
    { id: 4, slug: "audio", name: "Audio", desc: "Loops, kits, foley", parentId: 0, sort: 4, active: true },
    { id: 5, slug: "environments", name: "Environments", desc: "Scenes and world kits", parentId: 0, sort: 5, active: true },
    { id: 6, slug: "icons", name: "Icons & type", desc: "Icon sets and specimens", parentId: 0, sort: 6, active: true },
    { id: 7, slug: "vfx", name: "VFX", desc: "Sprites, volumes, sequences", parentId: 0, sort: 7, active: true },
    { id: 8, slug: "motion", name: "Motion", desc: "Loops, titles, and boards", parentId: 0, sort: 8, active: true },
    { id: 9, slug: "print", name: "Print", desc: "Posters, editorial, layouts", parentId: 0, sort: 9, active: true },
    { id: 10, slug: "photos", name: "Photos", desc: "Stills and plates", parentId: 0, sort: 10, active: true }
  ];
  const volumeCats = ["3d", "ui", "textures", "audio", "environments", "icons", "vfx", "motion", "print", "photos"];
  function volumeProducts() {
    const out = [];
    for (let i = 0; i < 36; i++) {
      const b = catalogSeed[i % catalogSeed.length];
      const n = 13 + i;
      const price = Math.max(6, ((b.price + (i % 7) * 3) % 90) || 9);
      out.push({
        id: n,
        slug: b.slug + "-vol-" + (i + 1),
        title: b.title + " Vol. " + (i + 1),
        category: volumeCats[i % volumeCats.length],
        price,
        fileType: b.fileType,
        rating: Math.round((4 + (i % 10) / 10) * 10) / 10,
        reviews: 8 + i,
        pinned: false,
        popular: 40 + i * 11,
        created: "2026-" + String((i % 9) + 1).padStart(2, "0") + "-" + String((i % 27) + 1).padStart(2, "0"),
        status: "active",
        pwyw: false,
        description: b.description,
        images: b.images,
        tiers: [{ id: "std-" + n, name: "Standard", price, note: "Full download" }]
      });
    }
    return out;
  }
  function volumeUsers() {
    const first = ["Asha", "Ben", "Cora", "Drew", "Eve", "Finn", "Gia", "Hugo", "Ivy", "Jules"];
    const last = ["Okoye", "Singh", "Adler", "Ng", "Kovacs", "Berg", "Diaz", "Sato", "Khan", "Walsh"];
    const avatars = ["p01.jpg", "p02.jpg", "p03.jpg", "p04.jpg", "p05.jpg", "p06.jpg", "p07.jpg", "p08.jpg", "p09.jpg", "p10.jpg", "p11.jpg", "p12.jpg"];
    const out = [];
    for (let i = 0; i < 76; i++) {
      const id = 5 + i;
      out.push({
        id,
        name: first[i % first.length] + " " + last[Math.floor(i / 10) % last.length],
        email: "collector" + id + "@example.com",
        password: "demo",
        bio: "Collector #" + id + ". " + volumeCats[i % volumeCats.length] + " library.",
        wallet: "",
        avatar: avatars[i % avatars.length],
        referral: "COL-" + id,
        referredBy: (i % 4) + 1,
        role: "customer",
        staff_tabs: ""
      });
    }
    return out;
  }
  function volumeSocial() {
    const posts = [];
    const comments = [];
    const likes = [];
    const follows = [];
    let pid = 7;
    let cid = 5;
    const notes = [
      "Just dropped a still from the courtyard.",
      "Anyone pairing this with Glyph Factory?",
      "The 8k chips hold up.",
      "Jam weekend material.",
      "Pinned on my board.",
      "Need the studio tier.",
      "Rain stems under marble.",
      "HUD from Northline."
    ];
    for (let i = 0; i < 54; i++) {
      const userId = 1 + (i % 40);
      posts.push({
        id: pid,
        userId,
        content: notes[i % notes.length] + " #" + pid,
        created: "2026-09-" + String((i % 17) + 1).padStart(2, "0") + "T" + String(10 + (i % 8)).padStart(2, "0") + ":00:00Z"
      });
      if (i % 2 === 0) comments.push({ id: cid++, postId: pid, userId: 1 + ((i + 1) % 20), content: "Noted on #" + pid });
      likes.push({ userId: 1 + ((i + 2) % 20), postId: pid });
      if (i % 3 === 0) likes.push({ userId: 1 + ((i + 5) % 20), postId: pid });
      pid++;
    }
    for (let i = 0; i < 16; i++) {
      comments.push({ id: cid++, postId: 1, userId: 2 + (i % 10), content: "Thread note " + (i + 1) + " on the clay pin." });
    }
    for (let u = 5; u <= 80; u++) {
      follows.push({ followerId: u, followingId: 1 + (u % 4) });
      if (u % 2 === 0) follows.push({ followerId: u, followingId: 2 });
      if (u % 5 === 0) follows.push({ followerId: 1, followingId: u });
    }
    return { posts, comments, likes, follows, nextPost: pid, nextComment: cid };
  }

  const catalogSeed = [
    { id: 1, slug: "lunar-clay-characters", title: "Lunar Clay Characters", category: "3d", price: 48, fileType: "FBX", rating: 4.8, reviews: 126, pinned: true, popular: 910, created: "2026-08-02", status: "active", pwyw: false, description: "Twelve stylized clay figures, game-ready, 8k texture set, and a small turntable scene.", images: ["p01.jpg", "p11.jpg", "p12.jpg"], tiers: [{ id: "std", name: "Game-ready", price: 48, note: "FBX + 4k maps" }, { id: "studio", name: "Studio", price: 72, note: "Blend + 8k + turntable" }] },
    { id: 2, slug: "northline-ui-kit", title: "Northline UI Kit", category: "ui", price: 36, fileType: "Figma", rating: 4.6, reviews: 88, pinned: true, popular: 640, created: "2026-07-18", status: "active", pwyw: false, description: "80 frames of dense product UI: tables, filters, empty states. Dark and light.", images: ["p02.jpg", "p02b.jpg"], tiers: [{ id: "fig", name: "Figma file", price: 36, note: "Variables + components" }] },
    { id: 3, slug: "obsidian-marble-pack", title: "Obsidian Marble Pack", category: "textures", price: 19, fileType: "PNG", rating: 4.9, reviews: 201, pinned: false, popular: 1200, created: "2026-05-03", status: "active", pwyw: false, description: "Fourteen seamless marble and stone materials with displacement and roughness.", images: ["p03.jpg", "p03b.jpg"], tiers: [{ id: "all", name: "Full pack", price: 19, note: "4k + 8k" }] },
    { id: 4, slug: "night-bus-lofi", title: "Night Bus Lo-fi Kit", category: "audio", price: 15, fileType: "WAV", rating: 4.4, reviews: 54, pinned: false, popular: 330, created: "2026-09-01", status: "active", pwyw: false, description: "Tape-worn keys, bus interiors, rain on glass. 48 stems, tempo-labeled.", images: ["p04.jpg", "p04b.jpg"], tiers: [{ id: "wav", name: "WAV pack", price: 15, note: "48/24" }] },
    { id: 5, slug: "concrete-atlas", title: "Concrete Atlas", category: "environments", price: 64, fileType: "USD", rating: 4.7, reviews: 41, pinned: false, popular: 210, created: "2026-04-21", status: "active", pwyw: false, description: "A brutalist courtyard and tower. Modular walls, hero camera, dusk HDRI.", images: ["p05.jpg", "p05b.jpg"], tiers: [{ id: "viz", name: "Lookdev", price: 64, note: "Cameras + lights" }, { id: "game", name: "Real-time", price: 88, note: "Game-ready meshes" }] },
    { id: 6, slug: "glyph-factory", title: "Glyph Factory Icons", category: "icons", price: 12, fileType: "SVG", rating: 4.5, reviews: 310, pinned: false, popular: 1500, created: "2026-03-12", status: "active", pwyw: false, description: "420 stroke icons on a 1.5px optical grid. Figma + SVG.", images: ["p06.jpg"], tiers: [{ id: "svg", name: "SVG + Figma", price: 12, note: "Variable stroke" }] },
    { id: 7, slug: "drift-specimen", title: "Drift Type Specimen", category: "icons", price: 29, fileType: "OTF", rating: 4.3, reviews: 19, pinned: false, popular: 90, created: "2026-06-09", status: "active", pwyw: false, description: "A display serif with a slightly broken italic. Two optical sizes.", images: ["p07.jpg"], tiers: [{ id: "desk", name: "Desktop + web", price: 29, note: "OTF / WOFF2" }] },
    { id: 8, slug: "canopy-environment", title: "Canopy Environment", category: "environments", price: 42, fileType: "USD", rating: 4.8, reviews: 67, pinned: false, popular: 400, created: "2026-08-20", status: "active", pwyw: false, description: "Dappled forest floor, wind-ready foliage, and a path that leads somewhere.", images: ["p08.jpg"], tiers: [{ id: "usd", name: "USD scene", price: 42, note: "Scatter + cameras" }] },
    { id: 9, slug: "signal-vfx", title: "Signal VFX Pack", category: "vfx", price: 27, fileType: "EXR", rating: 4.2, reviews: 22, pinned: false, popular: 140, created: "2026-02-02", status: "active", pwyw: false, description: "Lens dirt, hologram tiles, energy bursts. 32-bit EXR sequences.", images: ["p09.jpg"], tiers: [{ id: "exr", name: "Sequences", price: 27, note: "2k plates" }] },
    { id: 10, slug: "arcade-tile-kit", title: "Arcade Tile Kit", category: "3d", price: 9, fileType: "PNG", rating: 4.6, reviews: 480, pinned: false, popular: 2200, created: "2026-01-15", status: "active", pwyw: false, description: "16×16 dungeon and shop tiles with a matching character sheet.", images: ["p10.jpg"], tiers: [{ id: "px", name: "Pixel pack", price: 9, note: "Aseprite + PNG" }] },
    { id: 11, slug: "studio-light-hdris", title: "Studio Light HDRIs", category: "textures", price: 22, fileType: "HDR", rating: 4.7, reviews: 73, pinned: false, popular: 500, created: "2026-07-01", status: "active", pwyw: false, description: "Six studio wraps, from hard beauty to dusty warehouse.", images: ["p11.jpg"], tiers: [{ id: "hdr", name: "HDR set", price: 22, note: "4k + 8k" }] },
    { id: 12, slug: "sketch-brushes", title: "Sketchbook Brushes", category: "ui", price: 8, fileType: "ABR", rating: 4.1, reviews: 18, pinned: false, popular: 70, created: "2026-09-10", status: "active", pwyw: true, pwywMin: 4, description: "Pay what you want, floor $4. Graphite and wash brushes.", images: ["p12.jpg"], tiers: [{ id: "abr", name: "Brush set", price: 8, note: "Minimum $4" }] }
  ];

  function defaultOrderSteps() {
    return [
      { id: 1, slug: "created", label: "Created", sort_order: 10, is_system: true, is_terminal: false },
      { id: 2, slug: "awaiting_payment", label: "Waiting for payment", sort_order: 20, is_system: true, is_terminal: false },
      { id: 3, slug: "paid", label: "Paid", sort_order: 30, is_system: true, is_terminal: false },
      { id: 4, slug: "preparation", label: "Preparation", sort_order: 40, is_system: false, is_terminal: false },
      { id: 5, slug: "delivered", label: "Delivered", sort_order: 50, is_system: false, is_terminal: false },
      { id: 6, slug: "cancelled", label: "Cancelled", sort_order: 90, is_system: true, is_terminal: true },
      { id: 7, slug: "refunded", label: "Refunded", sort_order: 91, is_system: true, is_terminal: true },
      { id: 8, slug: "failed", label: "Failed", sort_order: 92, is_system: true, is_terminal: true }
    ];
  }

  function orderUnlocked(status) {
    return !({ created: 1, awaiting_payment: 1, pending: 1, processing: 1, cancelled: 1, refunded: 1, failed: 1 })[status];
  }

  function seedDb() {
    const extraUsers = volumeUsers();
    const social = volumeSocial();
    return {
      v: 6,
      session: null,
      admin: false,
      appearance: { palette: "clay", font: "system", radius: "soft", density: "comfortable" },
      cart: [],
      wishlist: [],
      compare: [],
      recentlyViewed: [],
      banner: [1, 2, 5],
      orderSteps: defaultOrderSteps(),
      products: JSON.parse(JSON.stringify(catalogSeed.concat(volumeProducts()))),
      categories: JSON.parse(JSON.stringify(categorySeed)),
      users: [
        { id: 1, name: "Nia Voss", email: "nia@example.com", password: "nia", bio: "Lookdev and clay. Buying once, downloading forever.", wallet: "0xNIA00…a4f", avatar: "avatar.jpg", referral: "NIA-STUDIO", referredBy: null, role: "admin", staff_tabs: "" },
        { id: 2, name: "Leo Park", email: "leo@example.com", password: "leo", bio: "Ships small games on weekends.", wallet: "", avatar: "p10.jpg", referral: "LEO-BUSY", referredBy: 1, role: "staff", staff_tabs: "Products,Banner,Orders,Community" },
        { id: 3, name: "Maya Chen", email: "maya@example.com", password: "maya", bio: "Environment lighting. Collects HDRIs and marble packs.", wallet: "0xMAYA…c2e", avatar: "p08.jpg", referral: "MAYA-LIGHT", referredBy: 1, role: "customer", staff_tabs: "" },
        { id: 4, name: "Owen Reid", email: "owen@example.com", password: "owen", bio: "UI kits into small tools. Quiet in the feed, loud in Figma.", wallet: "", avatar: "p02.jpg", referral: "OWEN-GRID", referredBy: 2, role: "customer", staff_tabs: "" }
      ].concat(extraUsers),
      posts: [
        { id: 1, userId: 1, content: "Pinned the clay set to the top of the library. If you render stills, use the studio tier — the turntable lights actually match the HDRIs.", created: "2026-09-12T10:00:00Z" },
        { id: 2, userId: 2, content: "Arcade Tile Kit + Glyph Factory is an entire jam weekend. Anyone bundling those?", created: "2026-09-14T16:20:00Z" },
        { id: 3, userId: 1, content: "Night Bus stems sit under a marble lookdev surprisingly well. Film the courtyard, add rain.", created: "2026-09-16T08:11:00Z" },
        { id: 4, userId: 3, content: "If you pin marble + clay, the stills stack is basically the Studio Kit. Following Nia's notes on the 8k chips.", created: "2026-09-16T14:40:00Z" },
        { id: 5, userId: 4, content: "Northline's empty states are the reason I bought it. Anyone pairing it with Glyph Factory as a HUD?", created: "2026-09-17T09:05:00Z" },
        { id: 6, userId: 3, content: "Canopy + Concrete Atlas is a whole short. Dusk HDRI, then rain from Night Bus.", created: "2026-09-17T18:22:00Z" }
      ].concat(social.posts),
      comments: [
        { id: 1, postId: 2, userId: 1, content: "Studio Kit bundle already does that — 15% under buying separate." },
        { id: 2, postId: 1, userId: 2, content: "Grabbed studio. The 8k clay chips hold up at 300mm." },
        { id: 3, postId: 4, userId: 1, content: "Yes — that's why it's pinned." },
        { id: 4, postId: 5, userId: 3, content: "Do that. Owen, also follow Leo — he already shipped a jam that way." }
      ].concat(social.comments),
      likes: [{ userId: 2, postId: 1 }, { userId: 1, postId: 2 }, { userId: 3, postId: 1 }, { userId: 3, postId: 2 }, { userId: 4, postId: 5 }, { userId: 1, postId: 4 }].concat(social.likes),
      follows: [
        { followerId: 2, followingId: 1 },
        { followerId: 3, followingId: 1 },
        { followerId: 3, followingId: 2 },
        { followerId: 1, followingId: 3 },
        { followerId: 4, followingId: 3 },
        { followerId: 4, followingId: 1 },
        { followerId: 2, followingId: 4 }
      ].concat(social.follows),
      bundles: [
        { id: 1, slug: "studio-kit", title: "Studio Kit", description: "Clay characters, marble, and HDRIs — the stills stack.", price: 72, productIds: [1, 3, 11], status: "active" },
        { id: 2, slug: "jam-pack", title: "Jam Pack", description: "Tiles, icons, and lo-fi for a weekend ship.", price: 28, productIds: [10, 6, 4], status: "active" },
        { id: 3, slug: "lookdev-stack", title: "Lookdev Stack", description: "Courtyard, canopy, and marble for stills.", price: 96, productIds: [5, 8, 3], status: "active" },
        { id: 4, slug: "weekend-hud", title: "Weekend HUD", description: "UI kit, icons, and type for a HUD jam.", price: 58, productIds: [2, 6, 7], status: "active" },
        { id: 5, slug: "forest-set", title: "Forest Set", description: "Canopy plus VFX plates.", price: 54, productIds: [8, 9], status: "active" }
      ],
      coupons: [
        { id: 1, code: "SAVE12", type: "percentage", value: 12, min: 20, usageLimit: 200, used: 14, expires: "2027-01-01", productId: null, active: true },
        { id: 2, code: "WELCOME", type: "fixed", value: 5, min: 0, usageLimit: 999, used: 40, expires: "2027-12-31", productId: null, active: true },
        { id: 3, code: "MARBLE", type: "percentage", value: 25, min: 0, usageLimit: 50, used: 3, expires: "2026-12-01", productId: 3, active: true }
      ],
      orders: [
        { id: 101, userId: 1, email: "nia@example.com", status: "paid", total: 19, discount: 0, coupon: null, created: "2026-08-20T12:00:00Z", paidAt: "2026-08-20T12:04:00Z", chain: "BSC", crypto: 0.031, address: "0xPAWR…BSC", items: [{ productId: 3, title: "Obsidian Marble Pack", slug: "obsidian-marble-pack", tier: "Full pack", qty: 1, unit: 19 }] },
        { id: 102, userId: 2, email: "leo@example.com", status: "preparation", total: 9, discount: 0, coupon: null, created: "2026-09-02T09:00:00Z", paidAt: "2026-09-02T09:03:00Z", chain: "BSC", crypto: 0.015, address: "0xPAWR…BSC", items: [{ productId: 10, title: "Arcade Tile Kit", slug: "arcade-tile-kit", tier: "Pixel pack", qty: 1, unit: 9 }] },
        { id: 103, userId: 3, email: "maya@example.com", status: "awaiting_payment", total: 12, discount: 0, coupon: null, created: "2026-09-12T11:00:00Z", paidAt: null, chain: "BSC", crypto: 0.02, address: "0xPAWR…BSC", items: [{ productId: 8, title: "Clay Fox Bust", slug: "clay-fox-bust", tier: "Bust", qty: 1, unit: 12 }] },
        { id: 104, userId: 4, email: "owen@example.com", status: "delivered", total: 14, discount: 0, coupon: null, created: "2026-09-05T16:00:00Z", paidAt: "2026-09-05T16:08:00Z", chain: "ETH", crypto: 0.004, address: "0xPAWR…ETH", items: [{ productId: 2, title: "Studio HDRI Set", slug: "studio-hdri-set", tier: "Full", qty: 1, unit: 14 }] }
      ],
      guestOrders: [],
      reviews: [
        { productId: 3, userId: 1, rating: 5 },
        { productId: 10, userId: 2, rating: 4 }
      ],
      commissions: [{ id: 1, orderId: 102, referrerId: 1, usd: 0.45, status: "owed" }],
      rates: [
        { chain: "BSC", symbol: "BNB", rate: 612 },
        { chain: "ETH", symbol: "ETH", rate: 3480 },
        { chain: "SOL", symbol: "SOL", rate: 168 }
      ],
      settings: Object.assign({}, defaultSettings),
      requests: [
        { id: 1, title: "Custom lunar fox", description: "A fox in the clay character style, game-ready.", category: "3d", budget_usd: 80, email: "buyer@example.com", status: "open", created: "2026-09-10T12:00:00Z" }
      ],
      nextIds: { user: 81, post: social.nextPost, comment: social.nextComment, order: 200, coupon: 4, bundle: 6, product: 49, cart: 1, request: 2, tier: 80, category: 11, step: 20 }
    };
  }

  function load() {
    try {
      const raw = JSON.parse(localStorage.getItem(KEY));
      if (raw && raw.v === 6) {
        raw.settings = Object.assign({}, defaultSettings, raw.settings || {});
        if (!raw.appearance) raw.appearance = { palette: "clay", font: "system", radius: "soft", density: "comfortable" };
        (raw.users || []).forEach((u) => {
          if (!u.role) u.role = u.id === 1 ? "admin" : (u.id === 2 ? "staff" : "customer");
          if (u.staff_tabs == null) u.staff_tabs = u.role === "staff" ? "Products,Banner,Orders,Community" : "";
        });
        raw.settings = Object.assign({}, defaultSettings, raw.settings || {});
        if (!raw.requests) raw.requests = [];
        if (!raw.nextIds) raw.nextIds = {};
        if (!raw.nextIds.request) raw.nextIds.request = 1;
        if (!raw.nextIds.tier) raw.nextIds.tier = 40;
        if (!raw.categories) raw.categories = JSON.parse(JSON.stringify(categorySeed));
        if (!raw.nextIds.category) raw.nextIds.category = (raw.categories.reduce((n, c) => Math.max(n, c.id), 0) || 0) + 1;
        if (!raw.orderSteps || !raw.orderSteps.length) raw.orderSteps = defaultOrderSteps();
        if (!raw.nextIds.step) raw.nextIds.step = (raw.orderSteps.reduce((n, s) => Math.max(n, s.id || 0), 0) || 0) + 1;
        const statusMap = { pending: "awaiting_payment", processing: "awaiting_payment", confirmed: "paid", shipped: "delivered", completed: "delivered" };
        (raw.orders || []).forEach((o) => { if (statusMap[o.status]) o.status = statusMap[o.status]; });
        (raw.categories || []).forEach((c) => {
          if (c.parentId == null) c.parentId = 0;
          if (c.sort == null) c.sort = c.id;
          if (c.active == null) c.active = true;
        });
        const avatars = { 1: "avatar.jpg", 2: "p10.jpg", 3: "p08.jpg", 4: "p02.jpg" };
        (raw.users || []).forEach((u) => {
          if (avatars[u.id] && !u.avatar) u.avatar = avatars[u.id];
        });
        return raw;
      }
    } catch (e) { /* ignore */ }
    const fresh = seedDb();
    localStorage.setItem(KEY, JSON.stringify(fresh));
    return fresh;
  }
  function save(db) {
    localStorage.setItem(KEY, JSON.stringify(db));
    return db;
  }
  function patch(fn) {
    const db = load();
    fn(db);
    return save(db);
  }

  function listCategories(all) {
    const rows = (load().categories || []).slice();
    const visible = all ? rows : rows.filter((c) => c.active !== false);
    visible.sort((a, b) => (a.sort || a.id) - (b.sort || b.id) || String(a.name).localeCompare(String(b.name)));
    return visible;
  }
  function findCategory(idOrSlug) {
    return listCategories(true).find((c) => c.id === Number(idOrSlug) || c.slug === idOrSlug) || null;
  }
  function slugify(s) {
    return String(s || "").toLowerCase().trim().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "category";
  }

  function hydrate(p) {
    if (!p) return null;
    const cat = findCategory(p.category);
    const revs = load().reviews.filter((r) => r.productId === p.id);
    const avg = revs.length ? revs.reduce((n, r) => n + r.rating, 0) / revs.length : p.rating;
    return Object.assign({}, p, {
      categoryName: cat ? cat.name : p.category,
      imageUrls: (p.images || []).map(media),
      thumb: media((p.images && p.images[0]) || "p01.jpg"),
      rating: Math.round(avg * 10) / 10,
      reviews: Math.max(p.reviews || 0, revs.length)
    });
  }
  function activeProducts(db) {
    return (db || load()).products.filter((p) => p.status !== "archived");
  }
  function userById(id) {
    return load().users.find((u) => u.id === Number(id));
  }
  function publicUser(u) {
    if (!u) return null;
    const db = load();
    const posts = db.posts.filter((p) => p.userId === u.id);
    return {
      id: u.id, name: u.name, bio: u.bio, wallet: u.wallet,
      avatar: u.avatar ? media(u.avatar) : null,
      postCount: posts.length,
      followerCount: db.follows.filter((f) => f.followingId === u.id).length,
      followingCount: db.follows.filter((f) => f.followerId === u.id).length
    };
  }
  function me() {
    const db = load();
    if (!db.session) return null;
    return userById(db.session.userId);
  }
  function deny(status, message) {
    const err = new Error(message);
    err.status = status;
    err.code = status;
    throw err;
  }
  function requireUser() {
    const u = me();
    if (!u) deny(401, "unauthorized");
    return u;
  }
  function adoptSession(userId) {
    if (!userId) return;
    const db = load();
    if (!db.session || db.session.userId !== Number(userId)) {
      patch((d) => { d.session = { userId: Number(userId) }; });
    }
  }

  function listProducts(query) {
    const q = query || {};
    const db = load();
    let rows = activeProducts(db).filter((p) => p.status === "active" || q.admin);
    if (q.search) {
      const s = q.search.toLowerCase();
      rows = rows.filter((p) => (p.title + " " + p.description).toLowerCase().includes(s));
    }
    if (q.category) rows = rows.filter((p) => p.category === q.category);
    if (q.file_type) rows = rows.filter((p) => p.fileType.toLowerCase() === String(q.file_type).toLowerCase());
    if (q.rating) rows = rows.filter((p) => hydrate(p).rating >= Number(q.rating));
    if (q.price_min) rows = rows.filter((p) => p.price >= Number(q.price_min));
    if (q.price_max) rows = rows.filter((p) => p.price <= Number(q.price_max));
    const bannerOn = q.banner === 1 || q.banner === "1" || q.banner === true;
    if (bannerOn) {
      const order = db.banner || [];
      rows = order.map((id) => rows.find((p) => p.id === Number(id))).filter(Boolean);
    } else {
      const sort = q.sort || "newest";
      rows.sort((a, b) => {
        if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
        if (sort === "price_asc") return a.price - b.price;
        if (sort === "price_desc") return b.price - a.price;
        if (sort === "popular") return b.popular - a.popular;
        return String(b.created).localeCompare(String(a.created));
      });
    }
    const total = rows.length;
    const page = Math.max(1, Number(q.page) || 1);
    const per = Math.min(200, Math.max(1, Number(q.per_page) || 9));
    const start = (page - 1) * per;
    return { products: rows.slice(start, start + per).map(hydrate), total, page, per_page: per };
  }

  function getProduct(slug) {
    const db = load();
    const p = db.products.find((x) => x.slug === slug || String(x.id) === String(slug));
    if (!p || p.status === "archived") return null;
    const related = activeProducts(db).filter((x) => x.category === p.category && x.id !== p.id && x.status === "active").slice(0, 3).map(hydrate);
    const reviews = db.reviews.filter((r) => r.productId === p.id).map((r) => ({ ...r, user: publicUser(userById(r.userId)) }));
    return Object.assign(hydrate(p), { related, reviewList: reviews });
  }

  function recordView(productId) {
    patch((db) => {
      db.recentlyViewed = [productId].concat(db.recentlyViewed.filter((id) => id !== productId)).slice(0, 12);
    });
  }

  function addToCart(productId, tierId, qty, pwywPrice) {
    const db = load();
    const p = db.products.find((x) => x.id === productId && x.status === "active");
    if (!p) throw new Error("Product unavailable");
    const tier = (p.tiers || []).find((t) => t.id === tierId) || p.tiers[0];
    let unit = tier.price;
    if (p.pwyw) {
      unit = Number(pwywPrice);
      if (!(unit >= p.pwywMin)) throw new Error("Price is below the minimum");
    }
    patch((d) => {
      if (!d.nextIds.cart) d.nextIds.cart = 1;
      const existing = d.cart.find((i) => i.productId === productId && i.tierId === tier.id);
      if (existing) existing.qty += qty || 1;
      else d.cart.push({ id: d.nextIds.cart++, productId, tierId: tier.id, qty: qty || 1, unit });
    });
  }

  function cartView() {
    const db = load();
    const items = db.cart.map((i) => {
      const p = hydrate(db.products.find((x) => x.id === i.productId));
      const tier = (p && p.tiers || []).find((t) => t.id === i.tierId);
      return { ...i, product: p, tier, line: i.unit * i.qty };
    }).filter((i) => i.product);
    const subtotal = items.reduce((n, i) => n + i.line, 0);
    return { items, subtotal, total: subtotal };
  }

  function validateCoupon(code, subtotal, items) {
    const db = load();
    const c = db.coupons.find((x) => x.code.toUpperCase() === String(code || "").toUpperCase() && x.active);
    if (!c) return { ok: false, error: "Coupon not found" };
    if (new Date(c.expires) < new Date()) return { ok: false, error: "Coupon expired" };
    if (c.used >= c.usageLimit) return { ok: false, error: "Usage limit reached" };
    if (subtotal < c.min) return { ok: false, error: "Minimum purchase not met" };
    let base = subtotal;
    if (c.productId) {
      const hit = (items || []).filter((i) => i.productId === c.productId);
      if (!hit.length) return { ok: false, error: "Coupon does not apply to this cart" };
      base = hit.reduce((n, i) => n + i.line, 0);
    }
    const discount = c.type === "percentage" ? Math.round(base * c.value) / 100 : Math.min(c.value, base);
    return { ok: true, coupon: c, discount };
  }

  function checkout({ couponCode, guestEmail, chain }) {
    const db = load();
    const cart = cartView();
    if (!cart.items.length) throw new Error("Cart is empty");
    const u = me();
    const email = u ? u.email : (guestEmail || "").trim();
    if (!email || !email.includes("@")) throw new Error("Email required");
    let discount = 0, coupon = null;
    if (couponCode) {
      const v = validateCoupon(couponCode, cart.subtotal, cart.items);
      if (!v.ok) throw new Error(v.error);
      discount = v.discount;
      coupon = v.coupon.code;
    }
    const total = Math.max(0, Math.round((cart.subtotal - discount) * 100) / 100);
    const rate = (db.rates.find((r) => r.chain === (chain || "BSC")) || db.rates[0]);
    const crypto = rate ? Math.round((total / rate.rate) * 10000) / 10000 : 0;
    const order = {
      id: db.nextIds.order++,
      userId: u ? u.id : null,
      email,
      status: "awaiting_payment",
      total, discount, coupon,
      created: new Date().toISOString(),
      paidAt: null,
      chain: rate.chain,
      crypto,
      address: db.settings.payment_address,
      items: cart.items.map((i) => ({
        productId: i.productId, title: i.product.title, slug: i.product.slug,
        tier: i.tier.name, qty: i.qty, unit: i.unit
      }))
    };
    patch((d) => {
      d.orders.push(order);
      if (!u) d.guestOrders.push(order.id);
      if (coupon) {
        const c = d.coupons.find((x) => x.code === coupon);
        if (c) c.used += 1;
      }
      d.cart = [];
      d.nextIds.order = db.nextIds.order;
    });
    return getOrder(order.id, email);
  }

  function confirmPay(orderId, email) {
    const order = getOrder(orderId, email);
    if (!order) throw new Error("Order not found");
    patch((d) => {
      const o = d.orders.find((x) => x.id === order.id);
      if (o.status === "paid") return;
      o.status = "paid";
      o.paidAt = new Date().toISOString();
      const buyer = o.userId && d.users.find((u) => u.id === o.userId);
      if (buyer && buyer.referredBy) {
        const pct = Number(d.settings.referral_commission_percent) / 100;
        d.commissions.push({ id: Date.now(), orderId: o.id, referrerId: buyer.referredBy, usd: Math.round(o.total * pct * 100) / 100, status: "owed" });
      }
    });
    return getOrder(orderId, email);
  }

  function getOrder(id, email) {
    const db = load();
    const o = db.orders.find((x) => x.id === Number(id));
    if (!o) return null;
    const u = me();
    if (u && o.userId === u.id) return o;
    if (email && o.email === email) return o;
    if (db.admin) return o;
    if (u && o.userId !== u.id) return null;
    if (!email) return null;
    return o;
  }

  function myOrders() {
    const u = me();
    if (!u) return [];
    return load().orders.filter((o) => o.userId === u.id).sort((a, b) => b.id - a.id);
  }

  function hasPurchased(productId) {
    const u = me();
    if (!u) return false;
    return load().orders.some((o) => o.userId === u.id && orderUnlocked(o.status) && o.items.some((i) => i.productId === productId));
  }

  function downloadBlob(orderId, itemIndex) {
    const o = getOrder(orderId, me() && me().email);
    if (!o || !orderUnlocked(o.status)) throw new Error("Not available");
    const item = o.items[itemIndex];
    if (!item) throw new Error("Missing item");
    const blob = new Blob(["Pawradise prototype file for " + item.title + " / " + item.tier + "\nBuy once, download forever.\n"], { type: "text/plain" });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = item.slug + ".txt";
    a.click();
  }

  function login(email, password) {
    const u = load().users.find((x) => x.email.toLowerCase() === String(email).toLowerCase());
    if (!u || u.password !== password) {
      const e = new Error("invalid credentials");
      e.status = 401;
      throw e;
    }
    patch((d) => { d.session = { userId: u.id }; });
    return publicUser(u);
  }
  function register({ name, email, password, referral }) {
    if (!email || !email.includes("@")) throw new Error("Valid email required");
    if (!password || password.length < 3) throw new Error("Password too short");
    const db = load();
    if (db.users.some((u) => u.email.toLowerCase() === email.toLowerCase())) {
      const e = new Error("email already registered");
      e.status = 409;
      throw e;
    }
    let referredBy = null;
    if (referral) {
      const ref = db.users.find((u) => u.referral.toLowerCase() === String(referral).toLowerCase());
      if (ref) referredBy = ref.id;
    }
    const u = { id: db.nextIds.user, name: name || email.split("@")[0], email, password, bio: "", wallet: "", avatar: null, referral: (name || "USER").slice(0, 3).toUpperCase() + "-" + db.nextIds.user, referredBy, role: "customer", staff_tabs: "" };
    patch((d) => { d.users.push(u); d.session = { userId: u.id }; d.nextIds.user += 1; });
    return publicUser(u);
  }
  function logout() {
    try { localStorage.removeItem("pawradise_proto_token"); } catch (e) { /* ignore */ }
    patch((d) => { d.session = null; d.admin = false; });
  }
  function updateMe(fields) {
    const u = requireUser();
    patch((d) => {
      const row = d.users.find((x) => x.id === u.id);
      if (fields.name != null) row.name = fields.name;
      if (fields.bio != null) row.bio = fields.bio;
      if (fields.wallet != null) row.wallet = fields.wallet;
      if (fields.avatar !== undefined) row.avatar = fields.avatar || null;
    });
    return publicUser(userById(u.id));
  }

  function listPosts() {
    const db = load();
    return db.posts.slice().sort((a, b) => b.id - a.id).map((p) => decoratePost(p, db));
  }
  function decoratePost(p, db) {
    db = db || load();
    const u = me();
    return {
      ...p,
      author: publicUser(userById(p.userId)),
      likes: db.likes.filter((l) => l.postId === p.id).length,
      liked: !!(u && db.likes.some((l) => l.postId === p.id && l.userId === u.id)),
      comments: db.comments.filter((c) => c.postId === p.id).map((c) => ({ ...c, author: publicUser(userById(c.userId)) }))
    };
  }
  function addPost(content) {
    const u = requireUser();
    if (!content || content.length > 500) throw new Error("Posts are 1–500 characters");
    let post;
    patch((d) => {
      post = { id: d.nextIds.post++, userId: u.id, content, created: new Date().toISOString() };
      d.posts.unshift(post);
    });
    return decoratePost(post);
  }
  function toggleLike(postId) {
    const u = requireUser();
    patch((d) => {
      const i = d.likes.findIndex((l) => l.postId === postId && l.userId === u.id);
      if (i >= 0) d.likes.splice(i, 1);
      else d.likes.push({ postId, userId: u.id });
    });
    return decoratePost(load().posts.find((p) => p.id === postId));
  }
  function addComment(postId, content) {
    const u = requireUser();
    if (!content) throw new Error("Comment required");
    patch((d) => { d.comments.push({ id: d.nextIds.comment++, postId, userId: u.id, content }); });
    return decoratePost(load().posts.find((p) => p.id === postId));
  }
  function deleteComment(id) {
    const u = requireUser();
    patch((d) => { d.comments = d.comments.filter((c) => !(c.id === id && c.userId === u.id)); });
  }
  function deleteOwnPost(id) {
    const u = requireUser();
    patch((d) => { d.posts = d.posts.filter((p) => !(p.id === id && p.userId === u.id)); });
  }
  function isFollowing(id) {
    const u = me();
    return !!(u && load().follows.some((f) => f.followerId === u.id && f.followingId === Number(id)));
  }
  function follow(userId) {
    const u = requireUser();
    userId = Number(userId);
    if (u.id === userId) {
      const e = new Error("cannot follow yourself");
      e.status = 400;
      throw e;
    }
    if (isFollowing(userId)) {
      const e = new Error("already following");
      e.status = 409;
      throw e;
    }
    if (!userById(userId)) {
      const e = new Error("user not found");
      e.status = 404;
      throw e;
    }
    patch((d) => { d.follows.push({ followerId: u.id, followingId: userId }); });
    return publicUser(userById(userId));
  }
  function unfollow(userId) {
    const u = requireUser();
    userId = Number(userId);
    if (!isFollowing(userId)) {
      const e = new Error("not following");
      e.status = 404;
      throw e;
    }
    patch((d) => { d.follows = d.follows.filter((f) => !(f.followerId === u.id && f.followingId === userId)); });
    return publicUser(userById(userId));
  }
  function toggleFollow(userId) {
    if (isFollowing(userId)) return unfollow(userId);
    return follow(userId);
  }
  function listFollowGraph(userId, dir) {
    const db = load();
    const id = Number(userId);
    const ids = dir === "followers"
      ? db.follows.filter((f) => f.followingId === id).map((f) => f.followerId)
      : db.follows.filter((f) => f.followerId === id).map((f) => f.followingId);
    return ids.map((x) => userById(x)).filter(Boolean);
  }
  function suggestions() {
    const u = me();
    const db = load();
    const following = new Set(u ? db.follows.filter((f) => f.followerId === u.id).map((f) => f.followingId) : []);
    return db.users
      .filter((x) => !u || (x.id !== u.id && !following.has(x.id)))
      .map((x) => publicUser(x))
      .sort((a, b) => b.followerCount - a.followerCount);
  }

  function addReview(productId, rating) {
    const u = requireUser();
    rating = Number(rating);
    if (rating < 1 || rating > 5) throw new Error("Rating must be 1–5");
    if (!hasPurchased(productId)) throw new Error("Verified purchase required");
    const db = load();
    if (db.reviews.some((r) => r.productId === productId && r.userId === u.id)) throw new Error("Already reviewed");
    patch((d) => { d.reviews.push({ productId, userId: u.id, rating }); });
  }

  function referrals() {
    const u = requireUser();
    const db = load();
    const referred = db.users.filter((x) => x.referredBy === u.id).map((x) => ({ id: x.id, name: x.name, email: x.email }));
    const earnings = db.commissions.filter((c) => c.referrerId === u.id);
    return {
      code: u.referral,
      link: location.origin + location.pathname.replace(/[^/]+$/, "") + "../auth-mfe/register.html?ref=" + encodeURIComponent(u.referral),
      referred,
      earnings,
      total: earnings.reduce((n, c) => n + c.usd, 0)
    };
  }

  function adminUnlock(token) {
    if (token !== "studio-admin" && token !== "admin_secret_staging_2026") throw new Error("Bad admin token");
    patch((d) => { d.admin = true; });
  }
  function requireAdmin() {
    if (load().admin) return;
    const u = me();
    if (u && (u.role === "admin" || u.role === "staff")) return;
    deny(401, "unauthorized");
  }
  function requireOperator() {
    if (!load().admin) deny(403, "Operator token required");
    const u = me();
    if (u && u.role !== "admin") deny(403, "Operator token required");
  }
  function setUserAccess(id, role, tabs) {
    requireOperator();
    role = role === "admin" || role === "staff" ? role : "customer";
    const list = Array.isArray(tabs) ? tabs : String(tabs || "").split(",");
    const allow = ["Stats","Users","Products","Banner","Categories","Bundles","Coupons","Orders","Guest","Community","Referrals","Rates","Settings","SEO","Export","Requests","Appearance"];
    const clean = [];
    list.forEach((t) => {
      t = String(t || "").trim();
      if (allow.indexOf(t) >= 0 && clean.indexOf(t) < 0) clean.push(t);
    });
    const staffTabs = role === "staff" ? clean.join(",") : "";
    patch((d) => {
      const admins = d.users.filter((x) => x.role === "admin");
      const u = d.users.find((x) => x.id === Number(id));
      if (!u) deny(404, "not found");
      if (u.role === "admin" && role !== "admin" && admins.length <= 1) {
        deny(400, "cannot demote the last admin");
      }
      u.role = role;
      u.staff_tabs = staffTabs;
    });
    const u = userById(id);
    return { id: u.id, role: u.role, staff_tabs: u.staff_tabs };
  }
  function appearance() {
    return Object.assign({ palette: "clay", font: "system", radius: "soft", density: "comfortable" }, load().appearance || {});
  }
  function saveAppearance(t) {
    requireAdmin();
    const next = appearance();
    if (t.palette) next.palette = t.palette;
    if (t.font) next.font = t.font;
    if (t.radius) next.radius = t.radius;
    if (t.density) next.density = t.density;
    patch((d) => { d.appearance = next; });
    return next;
  }

  function uniqueSlug(rows, slug, exceptId, strict) {
    const base = slugify(slug);
    const clash = rows.some((c) => c.slug === base && c.id !== exceptId);
    if (!clash) return base;
    if (strict) throw new Error("slug already exists");
    let n = 2, s = base + "-" + n;
    while (rows.some((c) => c.slug === s && c.id !== exceptId)) s = base + "-" + (++n);
    return s;
  }
  function saveCategory(input) {
    requireAdmin();
    input = input || {};
    let saved = null;
    patch((d) => {
      const name = String(input.name || "").trim();
      if (!name) throw new Error("name is required");
      const parentId = Number(input.parent_id != null ? input.parent_id : input.parentId) || 0;
      if (parentId && !d.categories.some((c) => c.id === parentId)) throw new Error("parent not found");
      const now = new Date().toISOString();
      const activeVal = input.is_active != null ? !!input.is_active : (input.active != null ? !!input.active : true);
      if (input.id) {
        const hit = d.categories.find((c) => c.id === Number(input.id));
        if (!hit) deny(404, "not found");
        if (parentId === hit.id) throw new Error("category cannot parent itself");
        let walk = parentId, guard = 0;
        while (walk && guard++ < 24) {
          if (walk === hit.id) throw new Error("circular parent");
          const p = d.categories.find((c) => c.id === walk);
          walk = p ? (p.parentId || 0) : 0;
        }
        const oldSlug = hit.slug;
        hit.name = name;
        const nextSlug = uniqueSlug(d.categories, input.slug || name, hit.id, !!input.slug);
        hit.slug = nextSlug;
        if (input.description != null || input.desc != null) hit.desc = String(input.description != null ? input.description : input.desc);
        hit.parentId = parentId;
        if (input.sort_order != null || input.sort != null) {
          const s = Number(input.sort_order != null ? input.sort_order : input.sort);
          hit.sort = Number.isFinite(s) ? s : hit.id;
        }
        if (input.is_active != null || input.active != null) hit.active = activeVal;
        if (input.image_url != null || input.imageUrl != null) hit.imageUrl = input.image_url || input.imageUrl || "";
        hit.updated = now;
        if (oldSlug !== hit.slug) {
          d.products.forEach((p) => { if (p.category === oldSlug) p.category = hit.slug; });
          (d.requests || []).forEach((r) => { if (r.category === oldSlug) r.category = hit.slug; });
        }
        saved = hit;
      } else {
        saved = {
          id: d.nextIds.category++,
          name,
          slug: uniqueSlug(d.categories, input.slug || name, 0, !!input.slug),
          desc: String(input.description || input.desc || ""),
          parentId,
          sort: (function () {
            const s = Number(input.sort_order != null ? input.sort_order : input.sort);
            return Number.isFinite(s) && String(input.sort_order || input.sort || "") !== "" ? s : (d.categories.length + 1);
          }()),
          active: activeVal,
          imageUrl: input.image_url || input.imageUrl || "",
          created: now,
          updated: now
        };
        d.categories.push(saved);
      }
    });
    return saved;
  }
  function deleteCategory(id) {
    requireAdmin();
    patch((d) => {
      const hit = d.categories.find((c) => c.id === Number(id));
      if (!hit) deny(404, "not found");
      const n = d.products.filter((p) => p.category === hit.slug).length;
      if (n) throw new Error(n + " products still in this category");
      const kids = d.categories.filter((c) => c.parentId === hit.id).length;
      if (kids) throw new Error("Move child categories first");
      d.categories = d.categories.filter((c) => c.id !== hit.id);
    });
  }

  function stats() {
    const db = load();
    const paid = db.orders.filter((o) => orderUnlocked(o.status));
    const views = db.products.reduce((n, p) => n + (p.popular || 0), 0) || 1;
    const carts = Math.max(db.cart.length, paid.length + 1);
    const checkouts = db.orders.length || 1;
    const purchases = paid.length;
    const byProduct = {};
    paid.forEach((o) => (o.items || []).forEach((it) => {
      byProduct[it.productId] = byProduct[it.productId] || { id: it.productId, title: it.title, units: 0, revenue: 0 };
      byProduct[it.productId].units += it.qty || 1;
      byProduct[it.productId].revenue += (it.unit || 0) * (it.qty || 1);
    }));
    const top = Object.values(byProduct).sort((a, b) => b.revenue - a.revenue).slice(0, 5);
    const days = [];
    for (let i = 13; i >= 0; i--) {
      const d = new Date();
      d.setDate(d.getDate() - i);
      const key = d.toISOString().slice(0, 10);
      const dayRev = paid.filter((o) => String(o.created).slice(0, 10) === key).reduce((n, o) => n + o.total, 0);
      days.push({ date: key, revenue: dayRev });
    }
    if (!days.some((x) => x.revenue)) {
      days[days.length - 5].revenue = 19;
      days[days.length - 2].revenue = 9;
    }
    return {
      users: db.users.length,
      products: db.products.filter((p) => p.status === "active").length,
      orders: db.orders.length,
      revenue: paid.reduce((n, o) => n + o.total, 0),
      categories: listCategories(true).length,
      pinned: db.products.filter((p) => p.pinned).length,
      downloads: paid.reduce((n, o) => n + (o.items || []).length, 0),
      revenue_daily: days,
      top_products: top,
      conversions: {
        views,
        carts,
        checkouts,
        purchases,
        visitor_to_view: 62,
        view_to_cart: Math.round((carts / views) * 1000) / 10,
        cart_to_checkout: Math.round((checkouts / carts) * 1000) / 10,
        checkout_to_purchase: Math.round((purchases / checkouts) * 1000) / 10
      },
      order_by_step: (db.orderSteps || defaultOrderSteps()).map((s) => ({
        slug: s.slug, label: s.label, count: db.orders.filter((o) => o.status === s.slug).length, is_terminal: !!s.is_terminal
      }))
    };
  }

  function setOrderStatus(id, status) {
    requireAdmin();
    const steps = load().orderSteps || defaultOrderSteps();
    if (!steps.some((s) => s.slug === status)) throw new Error("unknown step");
    patch((d) => {
      const o = d.orders.find((x) => x.id === Number(id));
      if (!o) throw new Error("Order not found");
      o.status = status;
    });
  }

  function listOrderSteps() {
    requireAdmin();
    return (load().orderSteps || defaultOrderSteps()).slice().sort((a, b) => a.sort_order - b.sort_order);
  }

  function createOrderStep(label) {
    requireAdmin();
    const slug = String(label || "").toLowerCase().trim().replace(/[^a-z0-9]+/g, "_").replace(/^_|_$/g, "");
    if (!slug) throw new Error("label required");
    if (listOrderSteps().some((s) => s.slug === slug)) throw new Error("step already exists");
    let saved;
    patch((d) => {
      const steps = d.orderSteps || defaultOrderSteps();
      const term = steps.filter((s) => s.is_terminal).reduce((n, s) => Math.min(n, s.sort_order), 80);
      saved = { id: d.nextIds.step++, slug, label: String(label).trim(), sort_order: Math.max(31, term - 1), is_system: false, is_terminal: false };
      steps.push(saved);
      d.orderSteps = steps;
    });
    return saved;
  }

  function saveOrderSteps(rows) {
    requireAdmin();
    patch((d) => {
      const cur = d.orderSteps || defaultOrderSteps();
      const byId = {};
      cur.forEach((s) => { byId[s.id] = s; });
      rows.forEach((row, i) => {
        const hit = byId[row.id];
        if (!hit) return;
        if (row.label) hit.label = row.label;
        hit.sort_order = i + 1;
      });
      d.orderSteps = cur;
    });
    return listOrderSteps();
  }

  function renameOrderStep(id, label) {
    requireAdmin();
    patch((d) => {
      const hit = (d.orderSteps || []).find((s) => s.id === Number(id));
      if (!hit) throw new Error("step not found");
      hit.label = String(label).trim();
    });
  }

  function deleteOrderStep(id) {
    requireAdmin();
    patch((d) => {
      const hit = (d.orderSteps || []).find((s) => s.id === Number(id));
      if (!hit) throw new Error("step not found");
      if (hit.is_system) throw new Error("cannot delete a system step");
      const n = d.orders.filter((o) => o.status === hit.slug).length;
      if (n) throw new Error(n + " orders still use this step");
      d.orderSteps = d.orderSteps.filter((s) => s.id !== Number(id));
    });
  }

  global.Paw = {
    media,
    get categories() { return listCategories(false); },
    listCategories, findCategory, saveCategory, deleteCategory,
    get products() { return activeProducts().map(hydrate); },
    db: load,
    listProducts, getProduct, recordView,
    categoryCounts() {
      const rows = activeProducts().filter((p) => p.status === "active");
      return listCategories(false).map((c) => ({ ...c, count: rows.filter((p) => p.category === c.slug).length }));
    },
    featured() { return hydrate(activeProducts().find((p) => p.pinned && p.status === "active") || activeProducts()[0]); },
    fileTypes() { return [...new Set(activeProducts().map((p) => p.fileType))]; },
    addToCart, cartView,
    setCartQty(productId, tierId, qty) {
      patch((d) => {
        d.cart = d.cart.filter((i) => {
          if (i.productId === productId && i.tierId === tierId) { i.qty = qty; return qty > 0; }
          return true;
        });
      });
    },
    removeCartItem(id) {
      patch((d) => { d.cart = d.cart.filter((i) => i.id !== Number(id)); });
    },
    setCartItemQty(id, qty) {
      qty = Number(qty);
      patch((d) => {
        d.cart = d.cart.filter((i) => {
          if (i.id !== Number(id)) return true;
          if (qty <= 0) return false;
          i.qty = qty;
          return true;
        });
      });
    },
    clearCart() { patch((d) => { d.cart = []; }); },
    cartCount() { return load().cart.reduce((n, i) => n + i.qty, 0); },
    toggleWishlist(productId) {
      patch((d) => {
        const i = d.wishlist.indexOf(productId);
        if (i >= 0) d.wishlist.splice(i, 1); else d.wishlist.push(productId);
      });
      return load().wishlist;
    },
    wishlist() { return load().wishlist; },
    wishlistProducts() { return load().wishlist.map((id) => hydrate(load().products.find((p) => p.id === id))).filter(Boolean); },
    toggleCompare(productId) {
      patch((d) => {
        const i = d.compare.indexOf(productId);
        if (i >= 0) d.compare.splice(i, 1);
        else {
          if (d.compare.length >= 4) d.compare.shift();
          d.compare.push(productId);
        }
      });
      return load().compare;
    },
    compare() { return load().compare.map((id) => hydrate(load().products.find((p) => p.id === id))).filter(Boolean); },
    recentlyViewed() { return load().recentlyViewed.map((id) => hydrate(load().products.find((p) => p.id === id))).filter(Boolean); },
    bundles() {
      return load().bundles.filter((b) => b.status === "active").map((b) => {
        const items = b.productIds.map((id) => hydrate(load().products.find((p) => p.id === id))).filter(Boolean);
        const separate = items.reduce((n, p) => n + p.price, 0);
        return { ...b, items, separate, savings: Math.max(0, separate - b.price), cover: items[0] && items[0].thumb };
      });
    },
    getBundle(id) { return this.bundles().find((b) => b.id === Number(id) || b.slug === id); },
    addBundleToCart(id) {
      const b = this.getBundle(id);
      if (!b) throw new Error("Bundle not found");
      b.items.forEach((p) => { try { addToCart(p.id, p.tiers[0].id, 1); } catch (e) { /* skip */ } });
    },
    validateCoupon, checkout, confirmPay, getOrder, myOrders, hasPurchased, downloadBlob,
    login, register, logout, updateMe, adoptSession,
    session() { return load().session; },
    me() { const u = me(); return u ? Object.assign(publicUser(u), { email: u.email, referral: u.referral }) : null; },
    isAdmin() { return !!load().admin; },
    profile(id) { return publicUser(userById(id)); },
    listPosts, getPost(id) { const p = load().posts.find((x) => x.id === Number(id)); return p ? decoratePost(p) : null; },
    addPost, toggleLike, addComment, deleteComment, deleteOwnPost, follow, unfollow, toggleFollow, isFollowing,
    listFollowers(id) { return listFollowGraph(id, "followers"); },
    listFollowing(id) { return listFollowGraph(id, "following"); },
    suggestions, members() { return load().users.map((u) => publicUser(u)); },
    userPosts(id) { return load().posts.filter((p) => p.userId === Number(id)).map((p) => decoratePost(p)); },
    addReview, referrals, rates() { return load().rates.slice(); },
    setting(key) { return load().settings[key]; },
    settings() { return Object.assign({}, load().settings); },
    stats, adminUnlock, adminLock() { patch((d) => { d.admin = false; }); },
    setUserAccess, appearance, saveAppearance,
    users() { requireAdmin(); return load().users.map((u) => publicUser(u)); },
    deleteUser(id) { requireAdmin(); patch((d) => { d.users = d.users.filter((u) => u.id !== Number(id)); }); },
    resetPassword(id) { requireAdmin(); patch((d) => { const u = d.users.find((x) => x.id === Number(id)); if (u) u.password = "reset123"; }); return "reset123"; },
    allOrders() { requireAdmin(); return load().orders.slice().sort((a, b) => b.id - a.id); },
    setOrderStatus,
    listOrderSteps, createOrderStep, saveOrderSteps, renameOrderStep, deleteOrderStep,
    allPosts() { requireAdmin(); return listPosts(); },
    adminDeletePost(id) { requireAdmin(); patch((d) => { d.posts = d.posts.filter((p) => p.id !== Number(id)); }); },
    coupons() { return load().coupons.slice(); },
    saveCoupon(c) {
      requireAdmin();
      patch((d) => {
        const apply = (hit) => {
          if (c.code != null) hit.code = c.code;
          if (c.type) hit.type = c.type;
          if (c.value != null) hit.value = Number(c.value);
          if (c.min != null) hit.min = Number(c.min);
          if (c.usageLimit != null) hit.usageLimit = Number(c.usageLimit);
          if (c.expires != null) hit.expires = c.expires;
          if (c.productId !== undefined) hit.productId = c.productId || null;
          if (c.active != null) hit.active = !!c.active;
        };
        if (c.id) {
          const hit = d.coupons.find((x) => x.id === Number(c.id));
          if (hit) apply(hit);
        } else {
          d.coupons.push({ id: d.nextIds.coupon++, code: c.code, type: c.type || "percentage", value: Number(c.value), min: Number(c.min) || 0, usageLimit: Number(c.usageLimit) || 100, used: 0, expires: c.expires, productId: c.productId || null, active: c.active !== false });
        }
      });
    },
    pinProduct(id, pinned) {
      requireAdmin();
      patch((d) => {
        const p = d.products.find((x) => x.id === Number(id));
        if (!p) return;
        p.pinned = !!pinned;
        d.banner = d.banner || [];
        if (pinned) {
          if (d.banner.indexOf(p.id) < 0) d.banner.push(p.id);
        } else {
          d.banner = d.banner.filter((x) => x !== p.id);
        }
      });
    },
    setBanner(ids) {
      requireAdmin();
      patch((d) => {
        const wanted = (ids || []).map(Number).filter((n) => d.products.some((p) => p.id === n));
        d.banner = wanted;
        d.products.forEach((p) => { p.pinned = wanted.indexOf(p.id) >= 0; });
      });
    },
    saveProduct(p) {
      requireAdmin();
      patch((d) => {
        const apply = (hit) => {
          if (p.title != null) hit.title = p.title;
          if (p.slug) hit.slug = p.slug;
          if (p.description != null) hit.description = p.description;
          if (p.price != null) hit.price = Number(p.price);
          if (p.category) hit.category = p.category;
          if (p.status) hit.status = p.status;
          if (p.pwyw != null) hit.pwyw = !!p.pwyw;
          if (p.pwywMin != null) hit.pwywMin = Number(p.pwywMin);
          if (p.fileType) hit.fileType = p.fileType;
          if (p.pinned != null) hit.pinned = !!p.pinned;
          if (p.seoTitle != null) hit.seoTitle = p.seoTitle;
          if (p.seoDescription != null) hit.seoDescription = p.seoDescription;
          if (p.ogImage != null) hit.ogImage = p.ogImage;
        };
        if (p.id) {
          const hit = d.products.find((x) => x.id === Number(p.id));
          if (hit) apply(hit);
        } else {
          const id = d.nextIds.product++;
          const row = {
            id,
            slug: p.slug || ("asset-" + id),
            title: p.title,
            category: p.category || "3d",
            price: Number(p.price) || 0,
            fileType: p.fileType || "ZIP",
            rating: 0,
            reviews: 0,
            pinned: !!p.pinned,
            popular: 0,
            created: new Date().toISOString().slice(0, 10),
            status: p.status || "active",
            pwyw: !!p.pwyw,
            pwywMin: Number(p.pwywMin) || 0,
            description: p.description || "",
            images: p.images || ["p12.jpg"],
            tiers: p.tiers || [{ id: "std", name: "Standard", price: Number(p.price) || 0, note: "" }],
            seoTitle: p.seoTitle || "",
            seoDescription: p.seoDescription || "",
            ogImage: p.ogImage || ""
          };
          d.products.push(row);
        }
      });
    },
    archiveProduct(id) { requireAdmin(); patch((d) => { const p = d.products.find((x) => x.id === Number(id)); if (p) p.status = "archived"; }); },
    saveBundle(b) {
      requireAdmin();
      patch((d) => {
        const slug = b.slug || String(b.title || "bundle").toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
        if (b.id) {
          const hit = d.bundles.find((x) => x.id === Number(b.id));
          if (hit) {
            if (b.title != null) hit.title = b.title;
            if (b.description != null) hit.description = b.description;
            if (b.price != null) hit.price = Number(b.price);
            if (b.status) hit.status = b.status;
            if (b.productIds) hit.productIds = b.productIds;
            if (b.slug) hit.slug = b.slug;
          }
        } else {
          d.bundles.push({ id: d.nextIds.bundle++, slug, title: b.title, description: b.description || "", price: Number(b.price), productIds: b.productIds || [], status: b.status || "active" });
        }
      });
    },
    deleteBundle(id) { requireAdmin(); patch((d) => { d.bundles = d.bundles.filter((b) => b.id !== Number(id)); }); },
    deleteCoupon(id) { requireAdmin(); patch((d) => { d.coupons = d.coupons.filter((c) => c.id !== Number(id)); }); },
    addTier(productId, t) {
      requireAdmin();
      let id;
      patch((d) => {
        const p = d.products.find((x) => x.id === Number(productId));
        if (!p) throw new Error("not found");
        if (!p.tiers) p.tiers = [];
        id = d.nextIds.tier++;
        p.tiers.push({ id: "t" + id, name: t.tier_name || t.name, price: Number(t.price_usd != null ? t.price_usd : t.price) || 0, note: t.note || "" });
      });
      return id;
    },
    updateTier(productId, numericId, t) {
      requireAdmin();
      patch((d) => {
        const p = d.products.find((x) => x.id === Number(productId));
        if (!p) throw new Error("not found");
        const i = Number(numericId) - p.id * 10 - 1;
        const tier = p.tiers[i];
        if (!tier) throw new Error("not found");
        if (t.tier_name) tier.name = t.tier_name;
        if (t.price_usd != null) tier.price = Number(t.price_usd);
      });
    },
    deleteTier(productId, numericId) {
      requireAdmin();
      patch((d) => {
        const p = d.products.find((x) => x.id === Number(productId));
        if (!p) throw new Error("not found");
        const i = Number(numericId) - p.id * 10 - 1;
        if (i >= 0) p.tiers.splice(i, 1);
      });
    },
    addImage(productId, url) {
      requireAdmin();
      patch((d) => {
        const p = d.products.find((x) => x.id === Number(productId));
        if (!p) throw new Error("not found");
        const file = String(url || "").split("/").pop() || "p12.jpg";
        p.images = (p.images || []).concat([file.replace(/\?.*$/, "")]);
      });
    },
    deleteImage(productId, imageId) {
      requireAdmin();
      patch((d) => {
        const p = d.products.find((x) => x.id === Number(productId));
        if (!p || !p.images) return;
        const i = Number(imageId) - p.id * 10 - 1;
        if (i >= 0) p.images.splice(i, 1);
      });
    },
    bulkProducts(ids, fields) {
      requireAdmin();
      patch((d) => {
        d.products.forEach((p) => {
          if (ids.indexOf(p.id) < 0) return;
          if (fields.status) p.status = fields.status;
          if (fields.category) p.category = fields.category;
        });
      });
    },
    addRequest(r) {
      let row;
      patch((d) => {
        if (!d.requests) d.requests = [];
        if (!d.nextIds.request) d.nextIds.request = 1;
        row = { id: d.nextIds.request++, title: r.title, description: r.description || "", category: r.category || "", budget_usd: Number(r.budget_usd) || 0, email: r.email || "", status: "open", created: new Date().toISOString() };
        d.requests.push(row);
      });
      return row;
    },
    listRequests() { return (load().requests || []).slice(); },
    setRequestStatus(id, status) {
      requireAdmin();
      patch((d) => {
        const r = (d.requests || []).find((x) => x.id === Number(id));
        if (r) r.status = status;
      });
    },
    saveRate(r) { requireAdmin(); patch((d) => { const x = d.rates.find((i) => i.chain === r.chain); if (x) x.rate = Number(r.rate); else d.rates.push({ chain: r.chain, symbol: r.symbol || r.chain, rate: Number(r.rate) }); }); },
    deleteRate(chain) { requireAdmin(); patch((d) => { d.rates = d.rates.filter((r) => r.chain !== chain); }); },
    saveSetting(k, v) { requireAdmin(); patch((d) => { d.settings[k] = v; }); },
    exportEmails(filters) {
      requireAdmin();
      filters = filters || {};
      const db = load();
      const paid = db.orders.filter((o) => {
        if (o.status !== "paid") return false;
        if (filters.from && String(o.created).slice(0, 10) < filters.from) return false;
        if (filters.to && String(o.created).slice(0, 10) > filters.to) return false;
        if (filters.product_id && !(o.items || []).some((it) => it.productId === Number(filters.product_id))) return false;
        if (filters.category) {
          const ids = (o.items || []).map((it) => it.productId);
          const hit = db.products.some((p) => ids.indexOf(p.id) >= 0 && p.category === filters.category);
          if (!hit) return false;
        }
        return true;
      });
      const map = {};
      paid.forEach((o) => {
        const u = db.users.find((x) => x.email === o.email);
        map[o.email] = map[o.email] || {
          email: o.email,
          name: u ? u.name : "",
          orders: 0,
          spent: 0,
          first_purchase: o.created,
          last_purchase: o.created
        };
        map[o.email].orders += 1;
        map[o.email].spent += o.total;
        if (o.created < map[o.email].first_purchase) map[o.email].first_purchase = o.created;
        if (o.created > map[o.email].last_purchase) map[o.email].last_purchase = o.created;
      });
      return Object.values(map);
    },
    guestOrders() {
      requireAdmin();
      return load().orders.filter((o) => !o.userId).slice().sort((a, b) => b.id - a.id);
    },
    generatePreviews(id) {
      requireAdmin();
      patch((d) => {
        const p = d.products.find((x) => x.id === Number(id));
        if (p) p.previewGenerated = new Date().toISOString();
      });
    },
    allReferrals() {
      requireAdmin();
      return load().users.filter((u) => u.referredBy).map((u) => ({ user: publicUser(u), referrer: publicUser(userById(u.referredBy)) }));
    },
    resetDemo() {
      try { localStorage.removeItem("pawradise_proto_token"); } catch (e) { /* ignore */ }
      localStorage.removeItem(KEY);
      location.reload();
    }
  };
})(window);
