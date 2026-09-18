(function (global) {
  const root = () => "/";
  const api = () => global.Pawradise && Pawradise.api;
  function esc(s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  }
  function money(n) {
    n = Number(n) || 0;
    return "$" + (Math.round(n * 100) / 100).toFixed(n % 1 ? 2 : 0).replace(/\.00$/, "");
  }
  function icon(name) {
    const paths = {
      shop: '<path d="M4 10.5L12 4l8 6.5"/><path d="M6 10v10h12V10"/>',
      bundles: '<rect x="3" y="8" width="18" height="12" rx="2"/><path d="M8 8V6a4 4 0 0 1 8 0v2"/>',
      community: '<path d="M8 19v-1a4 4 0 0 1 4-4h0a4 4 0 0 1 4 4v1"/><circle cx="12" cy="8" r="3"/><path d="M4 19v-1a3 3 0 0 1 2-2.8"/><circle cx="5" cy="9" r="2"/><path d="M20 19v-1a3 3 0 0 0-2-2.8"/><circle cx="19" cy="9" r="2"/>',
      cart: '<circle cx="9" cy="20" r="1.4"/><circle cx="18" cy="20" r="1.4"/><path d="M3 4h2l2.2 11h11.3l2-7H7"/>',
      account: '<circle cx="12" cy="8" r="3.2"/><path d="M5 19c1.2-3.2 3.6-5 7-5s5.8 1.8 7 5"/>',
      admin: '<circle cx="12" cy="12" r="3"/><path d="M12 3v2M12 19v2M4.9 6.5l1.5 1.5M17.6 16l1.5 1.5M3 12h2M19 12h2M4.9 17.5l1.5-1.5M17.6 8l1.5-1.5"/>',
      search: '<circle cx="11" cy="11" r="6.5"/><path d="M20 20l-3.2-3.2"/>',
      orders: '<path d="M7 4h10l2 4H5z"/><path d="M5 8h14v12H5z"/><path d="M9 12h6"/>',
      wishlist: '<path d="M12 19s-7-4.4-7-9.2A4 4 0 0 1 12 7a4 4 0 0 1 7 2.8C19 14.6 12 19 12 19z"/>',
      compare: '<path d="M8 4v16M16 4v16M4 8h4M16 16h4"/>',
      recent: '<circle cx="12" cy="12" r="8"/><path d="M12 8v5l3 2"/>',
      referrals: '<path d="M8 12h8"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="12" r="3"/>',
      save: '<path d="M12 19s-7-4.4-7-9.2A4 4 0 0 1 12 7a4 4 0 0 1 7 2.8C19 14.6 12 19 12 19z"/>',
      add: '<path d="M12 5v14M5 12h14"/>',
      pay: '<rect x="3" y="6" width="18" height="12" rx="2"/><path d="M3 10h18"/>',
      menu: '<path d="M4 7h16M4 12h16M4 17h16"/>',
      wallet: '<rect x="3" y="7" width="18" height="12" rx="2"/><path d="M16 13h4"/>',
      people: '<circle cx="9" cy="8" r="3"/><path d="M3 19c.8-3 3-5 6-5s5.2 2 6 5"/><circle cx="17" cy="9" r="2.2"/><path d="M16 19c.4-1.6 1.6-3 3.4-3.6"/>',
      pin: '<path d="M12 21s7-6.2 7-11a7 7 0 1 0-14 0c0 4.8 7 11 7 11z"/><circle cx="12" cy="10" r="2"/>',
      logout: '<path d="M10 7V5a2 2 0 0 1 2-2h7v18h-7a2 2 0 0 1-2-2v-2"/><path d="M4 12h11M12 8l4 4-4 4"/>',
      login: '<path d="M14 7V5a2 2 0 0 0-2-2H5v18h7a2 2 0 0 0 2-2v-2"/><path d="M10 12h11M17 8l4 4-4 4"/>',
      coupon: '<path d="M4 9a3 3 0 0 0 0 6v4h16v-4a3 3 0 0 0 0-6V5H4z"/><path d="M14 5v14"/>',
      check: '<path d="M5 12l5 5L20 7"/>'
    };
    const glyphs = {
      shop: "⌂", bundles: "▣", community: "☺", cart: "🛒", account: "●", admin: "✶",
      search: "⌕", orders: "☰", wishlist: "♥", compare: "⇄", recent: "⏱", referrals: "↔",
      save: "♥", add: "+", pay: "$", menu: "☰", wallet: "▭", people: "☺", pin: "📍",
      logout: "→", login: "←", coupon: "%", check: "✓"
    };
    const d = paths[name] || paths.shop;
    const g = glyphs[name] || "•";
    return `<span class="ico" data-ico="${esc(name)}" data-glyph="${g}" aria-hidden="true"><svg class="ico-svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">${d}</svg></span>`;
  }
  function stars(n) {
    const r = Math.round(Number(n) || 0);
    return `<span class="stars">${"★".repeat(Math.min(5, r))}${"☆".repeat(Math.max(0, 5 - r))}</span>`;
  }
  function productUrl(slug) { return "/product/" + encodeURIComponent(slug); }
  function bundleUrl(id) { return "/product/bundle.html?id=" + encodeURIComponent(id); }
  function orderUrl(id) { return "/account/order.html?id=" + encodeURIComponent(id); }
  function img(p) { return (p && (p.image_url || p.thumb)) || ""; }
  function price(p) { return p && p.price_usd != null ? p.price_usd : (p && p.price) || 0; }
  function catName(p) { return (p && (p.category_name || p.categoryName)) || ""; }
  function pinned(p) { return !!(p && (p.is_pinned || p.pinned)); }
  function pwyw(p) { return !!(p && (p.is_pwyw || p.pwyw)); }
  function fileType(p) { return (p && (p.digital_formats || p.file_type)) || ""; }

  const catalog = {
    products: null,
    cats: null,
    byId: {},
    catById: {},
    catBySlug: {},
    async load() {
      if (this.products) return this;
      const a = api();
      const [list, cats] = await Promise.all([
        a.products.list({ per_page: 200 }),
        a.products.getCategories()
      ]);
      this.products = list.products || [];
      this.cats = cats.categories || [];
      this.byId = {};
      this.products.forEach((p) => { this.byId[p.id] = p; });
      this.catById = {};
      this.catBySlug = {};
      this.cats.forEach((c) => { this.catById[c.id] = c; this.catBySlug[c.slug] = c; });
      return this;
    },
    invalidate() { this.products = null; this.cats = null; },
    decorate(p) {
      if (!p) return p;
      const cat = this.catById[p.category_id];
      return Object.assign({}, p, {
        category_name: p.category_name || (cat && cat.name) || "",
        category_slug: p.category_slug || (cat && cat.slug) || ""
      });
    },
    byIds(ids) {
      return (ids || []).map((id) => this.decorate(this.byId[id])).filter(Boolean);
    }
  };

  function loginHref(next) {
    const n = next || (location.pathname + location.search);
    return "/login?next=" + encodeURIComponent(n);
  }
  function authGate(job) {
    return `<p class="note">${esc(job)} <a href="${loginHref()}">Log in</a></p>`;
  }
  function safeNext(raw) {
    if (!raw) return "/";
    try { raw = decodeURIComponent(raw); } catch (e) { /* keep */ }
    if (/^https?:|^\/\/|^javascript:/i.test(raw)) return "/";
    return raw;
  }
  async function readAvatarFromForm(form) {
    const file = form.querySelector("[name=avatar_file]") && form.querySelector("[name=avatar_file]").files[0];
    if (file) {
      if (file.size > 900000) throw new Error("Picture must be under 900KB");
      return await new Promise((resolve, reject) => {
        const r = new FileReader();
        r.onload = () => resolve(r.result);
        r.onerror = () => reject(new Error("Could not read picture"));
        r.readAsDataURL(file);
      });
    }
    const url = form.querySelector("[name=avatar_url]");
    return url ? String(url.value || "").trim() : "";
  }
  function avatarFields(currentUrl) {
    return `<label class="field-col">Photo URL <input name="avatar_url" value="${esc(currentUrl || "")}" placeholder="https://… or leave blank" /></label>
      <label class="field-col">Or upload a picture <input name="avatar_file" type="file" accept="image/*" /></label>`;
  }
  function communityNav(on, meId) {
    const r = root();
    return `<nav class="subnav">
      <a class="${on === "feed" ? "is-on" : ""}" href="/community">${icon("community")} Feed</a>
      <a class="${on === "people" ? "is-on" : ""}" href="/people">${icon("people")} People</a>
      ${meId ? `<a class="${on === "profile" ? "is-on" : ""}" href="/profile/${meId}">${icon("account")} Your profile</a>` : ""}
    </nav>`;
  }
  function avatar(u, size) {
    const url = u && (u.avatar_url || u.avatar);
    const name = (u && (u.name || u.display_name)) || "?";
    const cls = size === "lg" ? "avatar avatar-lg" : "avatar";
    if (url) return `<img class="${cls}" src="${esc(url)}" alt="" />`;
    return `<span class="${cls}">${esc(name.slice(0, 1).toUpperCase())}</span>`;
  }
  function followControl(u, meId) {
    if (!u) return "";
    if (meId && meId === u.id) return "";
    if (!meId) return `<a class="btn btn-accent" href="${loginHref()}">Follow</a>`;
    if (u.following) return `<button class="btn" data-unfollow="${u.id}">Unfollow</button>`;
    return `<button class="btn btn-accent" data-follow="${u.id}">${u.follows_you ? "Follow back" : "Follow"}</button>`;
  }
  function relationPills(u, meId) {
    if (!u) return "";
    if (meId && meId === u.id) return '<span class="pill-you">You</span>';
    if (u.following && u.follows_you) return '<span class="pill-you">Mutual</span>';
    if (u.follows_you) return '<span class="pill-you">Follows you</span>';
    return "";
  }
  function personRow(u, meId, opts) {
    opts = opts || {};
    return `<article class="person${opts.compact ? " person-compact" : ""}">
      ${avatar(u)}
      <div>
        <a href="/profile/${u.id}"><strong>${esc(u.name)}</strong></a>
        ${relationPills(u, meId)}
        ${opts.compact ? "" : `<p>${esc(u.bio || "")}</p>
        <p class="person-meta">${u.follower_count || 0} followers · ${u.following_count || 0} following</p>`}
      </div>
      ${followControl(u, meId)}
    </article>`;
  }
  function postCard(p, meId) {
    const a = p.author || {};
    return `<article class="post">
      <div class="post-head">
        ${avatar(a)}
        <div>
          <a href="/profile/${p.user_id}">${esc(a.name || "User")}</a>
          ${relationPills(a, meId)}
          <span>${new Date(p.created_at).toLocaleDateString()}</span>
        </div>
        ${followControl(a, meId)}
      </div>
      <p>${esc(p.content)}</p>
      <div class="post-actions">
        <button class="btn ${p.liked ? "is-on" : ""}" data-like="${p.id}" data-liked="${p.liked ? "1" : ""}">${p.like_count || 0} likes</button>
        <a class="btn" href="/post/${p.id}">${p.comment_count || 0} comments</a>
      </div>
    </article>`;
  }
  async function bindFollow(el, after) {
    el._pawAfter = after;
    if (el._pawFollowBound) return;
    el._pawFollowBound = true;
    el.addEventListener("click", async (e) => {
      const follow = e.target.closest("[data-follow]");
      const unf = e.target.closest("[data-unfollow]");
      const like = e.target.closest("[data-like]");
      try {
        if (follow) { e.preventDefault(); await api().community.follow(follow.dataset.follow); if (el._pawAfter) el._pawAfter(); }
        else if (unf) { e.preventDefault(); await api().community.unfollow(unf.dataset.unfollow); if (el._pawAfter) el._pawAfter(); }
        else if (like) {
          e.preventDefault();
          if (like.dataset.liked) await api().community.unlikePost(like.dataset.like);
          else await api().community.likePost(like.dataset.like);
          if (el._pawAfter) el._pawAfter();
        }
      } catch (err) {
        if (err.status === 401) { toast("Log in to do that"); location.href = loginHref(); }
        else toast(err.message);
      }
    });
  }
  function couponOff(coupon, sub) {
    if (!coupon || !coupon.valid) return 0;
    const v = Number(coupon.discount_value);
    if (coupon.discount_type === "percentage") return Math.round(sub * v) / 100;
    return Math.min(v, sub);
  }

  function card(p) {
    p = catalog.decorate(p);
    return `<article class="card">
      <a class="card-still" href="${productUrl(p.slug)}">
        ${pinned(p) ? '<span class="pin">Pinned</span>' : ""}
        <span class="file-tag">${esc(fileType(p))}</span>
        <img src="${img(p)}" alt="" />
        <div class="card-actions">
          <button class="pill" data-wish="${p.id}">${icon("save")} Save</button>
          <button class="pill" data-compare="${p.id}">${icon("compare")} Compare</button>
          <button class="pill" data-cart="${p.id}">${icon("add")} Add</button>
        </div>
      </a>
      <div class="card-body">
        <h3><a href="${productUrl(p.slug)}">${esc(p.title)}</a></h3>
        <div class="card-sub">
          <span>${esc(catName(p))}</span>
          <span class="price">${pwyw(p) ? "from " + money(p.pwyw_min_price) : money(price(p))}</span>
        </div>
      </div>
    </article>`;
  }

  function toast(msg) {
    let t = document.querySelector(".toast");
    if (!t) { t = document.createElement("div"); t.className = "toast"; document.body.appendChild(t); }
    t.textContent = msg;
    t.classList.add("show");
    clearTimeout(t._id);
    t._id = setTimeout(() => t.classList.remove("show"), 1800);
  }

  function bindCards(el) {
    el.addEventListener("click", async (e) => {
      const wish = e.target.closest("[data-wish]");
      const cart = e.target.closest("[data-cart]");
      const cmp = e.target.closest("[data-compare]");
      if (!wish && !cart && !cmp) return;
      e.preventDefault();
      try {
        if (wish) await api().wishlist.toggle(Number(wish.dataset.wish));
        if (cmp) await api().compare.toggle(Number(cmp.dataset.compare));
        if (cart) { await api().cart.add(Number(cart.dataset.cart), 1); toast("Added to cart"); }
        if (global.PawChrome) PawChrome.mount();
        el.dispatchEvent(new Event("paw-refresh"));
      } catch (err) {
        if (err.status === 401) { toast("Log in to do that"); location.href = loginHref(); }
        else toast(err.message);
      }
    });
  }

  function accountNav(on) {
    const links = [
      ["/account", "Orders", "account", "orders"],
      ["/cart", "Cart", "cart", "cart"],
      ["/wishlist", "Wishlist", "wishlist", "wishlist"],
      ["/compare", "Compare", "compare", "compare"],
      ["/recent", "Recent", "recent", "recent"],
      ["/referrals", "Referrals", "referrals", "referrals"]
    ];
    return `<nav class="subnav">${links.map(([h, l, p, ic]) =>
      `<a class="${on === p ? "is-on" : ""}" href="${h}">${icon(ic)} ${l}</a>`).join("")}</nav>`;
  }
  function moreBtn(id, show) {
    if (!show) return "";
    return `<p class="more-wrap"><button type="button" class="btn" id="${esc(id)}">Show more</button></p>`;
  }

  function mountHeroSlider(el, products) {
    if (!el) return;
    const slides = (products || []).map((p) => catalog.decorate(p)).filter(Boolean);
    if (!slides.length) { el.innerHTML = ""; return; }
    let i = 0;
    let timer = null;
    function paint() {
      el.className = "hero";
      el.setAttribute("aria-roledescription", "carousel");
      el.innerHTML = slides.map((p, n) => `
        <article class="hero-slide${n === i ? " is-on" : ""}" data-slide="${n}">
          <img src="${img(p)}" alt="" />
          <div class="hero-veil"></div>
          <div class="hero-copy">
            <div class="kicker">Banner · ${n + 1} / ${slides.length} · ${esc(p.category_name || "")}</div>
            <h1>${esc(p.title)}</h1>
            <p>${esc(p.description)}</p>
            <div class="hero-meta">
              <span class="price">${money(p.price_usd)}</span>
              <span>${esc(p.digital_formats || "")}</span>
            </div>
            <a class="btn btn-accent" href="${productUrl(p.slug)}">View asset</a>
          </div>
        </article>`).join("") + (slides.length > 1 ? `
          <button type="button" class="hero-arrow hero-prev" aria-label="Previous banner">‹</button>
          <button type="button" class="hero-arrow hero-next" aria-label="Next banner">›</button>
          <div class="hero-dots">${slides.map((_, n) => `<button type="button" class="${n === i ? "is-on" : ""}" data-dot="${n}" aria-label="Banner ${n + 1}"></button>`).join("")}</div>
        ` : "");
    }
    function go(n) {
      i = (n + slides.length) % slides.length;
      paint();
      arm();
    }
    function arm() {
      if (timer) clearInterval(timer);
      if (slides.length < 2) return;
      timer = setInterval(() => go(i + 1), 6500);
    }
    el.onclick = (e) => {
      if (e.target.closest(".hero-next")) { e.preventDefault(); go(i + 1); }
      else if (e.target.closest(".hero-prev")) { e.preventDefault(); go(i - 1); }
      const dot = e.target.closest("[data-dot]");
      if (dot) { e.preventDefault(); go(Number(dot.dataset.dot)); }
    };
    let sx = 0;
    el.ontouchstart = (e) => { sx = e.changedTouches[0].clientX; };
    el.ontouchend = (e) => {
      const dx = e.changedTouches[0].clientX - sx;
      if (Math.abs(dx) < 40) return;
      go(dx > 0 ? i - 1 : i + 1);
    };
    paint();
    arm();
  }

  const THEME = {
    palettes: [
      { id: "clay", name: "Clay", note: "Default gallery gold", swatch: ["#0c0c0e", "#e8b86d", "#ecece8"] },
      { id: "marble", name: "Marble", note: "Light cream studio", swatch: ["#f3eee6", "#9a5b2f", "#1b1712"] },
      { id: "paper", name: "Paper", note: "Warm white desk", swatch: ["#f7f2ea", "#c45c26", "#1c1610"] },
      { id: "chalk", name: "Chalk", note: "Cool daylight", swatch: ["#f4f6f8", "#3d6b9a", "#15202b"] },
      { id: "linen", name: "Linen", note: "Beige gallery", swatch: ["#f3ead9", "#7a5c38", "#1a1510"] },
      { id: "mist", name: "Mist", note: "Pale blue-gray", swatch: ["#eef2f5", "#4a7d8c", "#142026"] },
      { id: "petal", name: "Petal", note: "Blush studio", swatch: ["#f8eef1", "#b85a72", "#1f1418"] },
      { id: "foam", name: "Foam", note: "Mint light", swatch: ["#eef6f1", "#2f7a5a", "#102018"] },
      { id: "porcelain", name: "Porcelain", note: "Blue-white", swatch: ["#f3f6fb", "#4b5f9a", "#141824"] },
      { id: "sage", name: "Sage", note: "Light garden", swatch: ["#eef3e8", "#5a7a48", "#14180f"] },
      { id: "snow", name: "Snow", note: "High-contrast white", swatch: ["#fbfbfc", "#2563a8", "#111318"] },
      { id: "honey", name: "Honey", note: "Cream gold", swatch: ["#f8f0d8", "#b8860b", "#1a1508"] },
      { id: "bone", name: "Bone", note: "Warm off-white", swatch: ["#f4efe6", "#6b4f3a", "#1a1510"] },
      { id: "cloud", name: "Cloud", note: "Lavender gray", swatch: ["#f1eef6", "#6b5b95", "#16141c"] },
      { id: "night", name: "Night", note: "Ice on ink", swatch: ["#07080d", "#7eb6ff", "#e8edf8"] },
      { id: "moss", name: "Moss", note: "Forest stills", swatch: ["#0c110e", "#9cbf7a", "#e6eee4"] },
      { id: "ink", name: "Ink", note: "Navy and gold", swatch: ["#0a0e18", "#d4ba6e", "#f3ead4"] },
      { id: "ember", name: "Ember", note: "Warm charcoal", swatch: ["#120c0b", "#e07a4a", "#f6ece6"] },
      { id: "dune", name: "Dune", note: "Sand and bronze", swatch: ["#16120e", "#c4a06a", "#f3eadc"] },
      { id: "frost", name: "Frost", note: "Cool cyan", swatch: ["#0c1014", "#6ec8d4", "#e6eef2"] },
      { id: "wine", name: "Wine", note: "Burgundy night", swatch: ["#1a0d12", "#d4a0b0", "#f6e8ee"] },
      { id: "violet", name: "Violet", note: "Royal dark", swatch: ["#120e1a", "#b89ae8", "#efe8f8"] },
      { id: "ocean", name: "Ocean", note: "Deep teal", swatch: ["#071218", "#3db8c8", "#e4f4f6"] },
      { id: "slate", name: "Slate", note: "Mid gray studio", swatch: ["#1a1d22", "#8aa0b8", "#e8edf2"] }
    ],
    fonts: [
      { id: "system", name: "System" },
      { id: "humanist", name: "Humanist" },
      { id: "serif", name: "Serif" },
      { id: "mono", name: "Mono" },
      { id: "display", name: "Display" }
    ],
    radii: [
      { id: "sharp", name: "Sharp" },
      { id: "soft", name: "Soft" },
      { id: "round", name: "Round" }
    ],
    densities: [
      { id: "compact", name: "Compact" },
      { id: "comfortable", name: "Comfortable" },
      { id: "roomy", name: "Roomy" }
    ],
    icons: [
      { id: "line", name: "Line", note: "Lucide-style outline" },
      { id: "bold", name: "Bold", note: "Heavier stroke" },
      { id: "filled", name: "Filled", note: "Duotone fill" },
      { id: "glyph", name: "Glyph", note: "Plain symbols" }
    ],
    contrasts: [
      { id: "soft", name: "Soft" },
      { id: "standard", name: "Standard" },
      { id: "punchy", name: "Punchy" }
    ],
    grains: [
      { id: "off", name: "Off" },
      { id: "light", name: "Light" },
      { id: "heavy", name: "Heavy" }
    ],
    glows: [
      { id: "none", name: "None" },
      { id: "halo", name: "Halo" },
      { id: "bloom", name: "Bloom" }
    ],
    motions: [
      { id: "still", name: "Still" },
      { id: "gentle", name: "Gentle" }
    ],
    trackings: [
      { id: "tight", name: "Tight" },
      { id: "normal", name: "Normal" },
      { id: "wide", name: "Wide" }
    ]
  };
  function defaultTheme() {
    return {
      palette: "clay", font: "system", radius: "soft", density: "comfortable",
      icons: "line", contrast: "standard", grain: "light", glow: "halo",
      motion: "gentle", tracking: "normal"
    };
  }
  function applyTheme(t) {
    t = Object.assign(defaultTheme(), t || {});
    const html = document.documentElement;
    html.dataset.palette = t.palette;
    html.dataset.font = t.font;
    html.dataset.radius = t.radius;
    html.dataset.density = t.density;
    html.dataset.icons = t.icons;
    html.dataset.contrast = t.contrast;
    html.dataset.grain = t.grain;
    html.dataset.glow = t.glow;
    html.dataset.motion = t.motion;
    html.dataset.tracking = t.tracking;
    return t;
  }
  function bootTheme(apiObj) {
    try {
      const cached = JSON.parse(localStorage.getItem("pawradise_theme") || "null");
      if (cached) applyTheme(cached);
    } catch (e) { /* ignore */ }
    const client = apiObj || api();
    if (!client || !client.appearance) return;
    client.appearance.get().then((t) => {
      applyTheme(t);
      try { localStorage.setItem("pawradise_theme", JSON.stringify(t)); } catch (e) { /* ignore */ }
    }).catch(() => {});
  }

  global.PawUI = { esc, money, stars, card, bindCards, productUrl, bundleUrl, orderUrl, toast, img, price, root, catalog, couponOff, accountNav, communityNav, avatar, personRow, postCard, bindFollow, followControl, relationPills, loginHref, authGate, safeNext, readAvatarFromForm, avatarFields, mountHeroSlider, THEME, applyTheme, defaultTheme, bootTheme, icon, moreBtn };
})(window);
