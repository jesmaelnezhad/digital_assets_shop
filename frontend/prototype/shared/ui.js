(function (global) {
  const root = () => global.PROTOTYPE_ROOT || "./";
  const api = () => global.Pawradise && Pawradise.api;
  function esc(s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  }
  function money(n) {
    n = Number(n) || 0;
    return "$" + (Math.round(n * 100) / 100).toFixed(n % 1 ? 2 : 0).replace(/\.00$/, "");
  }
  function stars(n) {
    const r = Math.round(Number(n) || 0);
    return `<span class="stars">${"★".repeat(Math.min(5, r))}${"☆".repeat(Math.max(0, 5 - r))}</span>`;
  }
  function productUrl(slug) { return root() + "product-mfe/index.html?slug=" + encodeURIComponent(slug); }
  function bundleUrl(id) { return root() + "product-mfe/bundle.html?id=" + encodeURIComponent(id); }
  function orderUrl(id) { return root() + "account-mfe/order.html?id=" + encodeURIComponent(id); }
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
    return root() + "auth-mfe/login.html?next=" + encodeURIComponent(n);
  }
  function authGate(job) {
    return `<p class="note">${esc(job)} <a href="${loginHref()}">Log in</a></p>`;
  }
  function safeNext(raw) {
    if (!raw) return root() + "shop-mfe/index.html";
    try { raw = decodeURIComponent(raw); } catch (e) { /* keep */ }
    if (/^https?:|^\/\/|^javascript:/i.test(raw)) return root() + "shop-mfe/index.html";
    return raw;
  }
  async function readAvatarFromForm(form) {
    const file = form.querySelector("[name=avatar_file]") && form.querySelector("[name=avatar_file]").files[0];
    if (file) {
      if (file.size > 900000) throw new Error("Picture must be under 900KB in the prototype");
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
      <a class="${on === "feed" ? "is-on" : ""}" href="${r}community-mfe/index.html">Feed</a>
      <a class="${on === "people" ? "is-on" : ""}" href="${r}community-mfe/people.html">People</a>
      ${meId ? `<a class="${on === "profile" ? "is-on" : ""}" href="${r}community-mfe/profile.html?id=${meId}">Your profile</a>` : ""}
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
        <a href="${root()}community-mfe/profile.html?id=${u.id}"><strong>${esc(u.name)}</strong></a>
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
          <a href="${root()}community-mfe/profile.html?id=${p.user_id}">${esc(a.name || "User")}</a>
          ${relationPills(a, meId)}
          <span>${new Date(p.created_at).toLocaleDateString()}</span>
        </div>
        ${followControl(a, meId)}
      </div>
      <p>${esc(p.content)}</p>
      <div class="post-actions">
        <button class="btn ${p.liked ? "is-on" : ""}" data-like="${p.id}" data-liked="${p.liked ? "1" : ""}">${p.like_count || 0} likes</button>
        <a class="btn" href="${root()}community-mfe/post.html?id=${p.id}">${p.comment_count || 0} comments</a>
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
          <button class="pill" data-wish="${p.id}">Save</button>
          <button class="pill" data-compare="${p.id}">Compare</button>
          <button class="pill" data-cart="${p.id}">Add</button>
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
      ["account.html", "Orders", "account"],
      ["cart.html", "Cart", "cart"],
      ["wishlist.html", "Wishlist", "wishlist"],
      ["compare.html", "Compare", "compare"],
      ["recent.html", "Recent", "recent"],
      ["referrals.html", "Referrals", "referrals"]
    ];
    return `<nav class="subnav">${links.map(([h, l, p]) =>
      `<a class="${on === p ? "is-on" : ""}" href="${root()}account-mfe/${h}">${l}</a>`).join("")}</nav>`;
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
      { id: "night", name: "Night", note: "Ice on ink", swatch: ["#07080d", "#7eb6ff", "#e8edf8"] },
      { id: "moss", name: "Moss", note: "Forest stills", swatch: ["#0c110e", "#9cbf7a", "#e6eee4"] },
      { id: "ink", name: "Ink", note: "Navy and gold", swatch: ["#0a0e18", "#d4ba6e", "#f3ead4"] },
      { id: "ember", name: "Ember", note: "Warm charcoal", swatch: ["#120c0b", "#e07a4a", "#f6ece6"] },
      { id: "dune", name: "Dune", note: "Sand and bronze", swatch: ["#16120e", "#c4a06a", "#f3eadc"] },
      { id: "frost", name: "Frost", note: "Cool cyan", swatch: ["#0c1014", "#6ec8d4", "#e6eef2"] }
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
    ]
  };
  function defaultTheme() {
    return { palette: "clay", font: "system", radius: "soft", density: "comfortable" };
  }
  function applyTheme(t) {
    t = Object.assign(defaultTheme(), t || {});
    const html = document.documentElement;
    html.dataset.palette = t.palette;
    html.dataset.font = t.font;
    html.dataset.radius = t.radius;
    html.dataset.density = t.density;
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

  global.PawUI = { esc, money, stars, card, bindCards, productUrl, bundleUrl, orderUrl, toast, img, price, root, catalog, couponOff, accountNav, communityNav, avatar, personRow, postCard, bindFollow, followControl, relationPills, loginHref, authGate, safeNext, readAvatarFromForm, avatarFields, mountHeroSlider, THEME, applyTheme, defaultTheme, bootTheme };
})(window);
