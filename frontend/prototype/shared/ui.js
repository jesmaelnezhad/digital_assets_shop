(function (global) {
  const root = () => global.PROTOTYPE_ROOT || "./";
  const api = () => global.Store4bots && Store4bots.api;
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
    },
    async byIdsAsync(ids) {
      await this.load();
      const out = [];
      for (const raw of ids || []) {
        const id = Number(raw);
        if (!id) continue;
        let p = this.byId[id];
        if (!p) {
          try {
            const r = await api().products.get(String(id));
            p = (r && r.product) || null;
            if (p && p.id) this.byId[p.id] = p;
          } catch (e) { p = null; }
        }
        if (p) out.push(this.decorate(p));
      }
      return out;
    }
  };

  const lists = {
    wish: new Set(),
    compare: new Set(),
    cart: new Map(),
    loaded: false,
    hasWish(id) { return this.wish.has(Number(id)); },
    hasCompare(id) { return this.compare.has(Number(id)); },
    inCart(id) { return this.cart.has(Number(id)); },
    setWish(id, on) { id = Number(id); if (on) this.wish.add(id); else this.wish.delete(id); },
    setCompare(id, on) { id = Number(id); if (on) this.compare.add(id); else this.compare.delete(id); },
    setCart(id, on, itemId) {
      id = Number(id);
      if (on) this.cart.set(id, itemId || true);
      else this.cart.delete(id);
    },
    async refresh() {
      const a = api();
      this.wish = new Set();
      this.compare = new Set();
      this.cart = new Map();
      if (!a) { this.loaded = true; return this; }
      try {
        const w = await a.wishlist.get();
        this.wish = new Set((w.products || []).map((x) => Number(x.product_id)).filter(Boolean));
      } catch (e) { /* guest */ }
      try {
        const c = await a.compare.list();
        this.compare = new Set((c.products || []).map((x) => Number(x.product_id)).filter(Boolean));
      } catch (e) { /* guest */ }
      try {
        const cart = await a.cart.get();
        (cart.items || []).forEach((i) => this.cart.set(Number(i.product_id), i.id));
      } catch (e) { /* guest */ }
      this.loaded = true;
      return this;
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
  async function readImageFromForm(form, fileName) {
    const input = form.querySelector("[name=" + (fileName || "photo") + "]");
    const file = input && input.files && input.files[0];
    if (!file) return "";
    if (file.size > 900000) throw new Error("Picture must be under 900KB in the prototype");
    if (file.type && file.type.indexOf("image/") !== 0) throw new Error("Choose an image file");
    return await new Promise((resolve, reject) => {
      const r = new FileReader();
      r.onload = () => resolve(r.result);
      r.onerror = () => reject(new Error("Could not read picture"));
      r.readAsDataURL(file);
    });
  }
  async function readAvatarFromForm(form) {
    const uploaded = await readImageFromForm(form, "avatar_file");
    if (uploaded) return uploaded;
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
  function postText(text) {
    const re = /https?:\/\/[^\s<]+/g;
    const s = String(text || "");
    let out = "", last = 0, m;
    while ((m = re.exec(s))) {
      out += esc(s.slice(last, m.index));
      let raw = m[0];
      raw = raw.replace(/[),.;!?]+$/, "");
      out += `<a class="post-url" href="${esc(raw)}" target="_blank" rel="noopener noreferrer">${esc(raw)}</a>`;
      last = m.index + raw.length;
      re.lastIndex = last;
    }
    return out + esc(s.slice(last));
  }
  function postMedia(p) {
    if (!p) return "";
    let html = "";
    if (p.image_url) html += `<img class="post-photo" src="${esc(p.image_url)}" alt="" />`;
    if (p.link_url) {
      let host = p.link_url;
      try { host = new URL(p.link_url).hostname.replace(/^www\./, ""); } catch (e) { /* keep */ }
      const thumb = p.link_image
        ? `<img src="${esc(p.link_image)}" alt="" />`
        : `<span class="post-link-mark">${esc(host.slice(0, 1).toUpperCase())}</span>`;
      html += `<a class="post-link" href="${esc(p.link_url)}" target="_blank" rel="noopener noreferrer">
        ${thumb}
        <span>
          <strong>${esc(p.link_title || host)}</strong>
          ${p.link_description ? `<em>${esc(p.link_description)}</em>` : ""}
          <small>${esc(host)}</small>
        </span>
      </a>`;
    }
    return html;
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
      ${p.content ? `<p>${postText(p.content)}</p>` : ""}
      ${postMedia(p)}
      <div class="post-actions">
        <button class="btn ${p.liked ? "is-on" : ""}" data-like="${p.id}" data-liked="${p.liked ? "1" : ""}">${p.like_count || 0} likes</button>
        <a class="btn" href="${root()}community-mfe/post.html?id=${p.id}">${p.comment_count || 0} comments</a>
      </div>
    </article>`;
  }
  function bindComposer(form, onPosted) {
    if (!form) return;
    const ta = form.querySelector("textarea");
    const cc = form.querySelector("#cc") || form.querySelector(".char-count span");
    const thumb = form.querySelector(".composer-thumb");
    const file = form.querySelector("[name=photo]");
    const drop = form.querySelector("[data-drop-photo]");
    if (ta && cc) ta.oninput = () => { cc.textContent = ta.value.length; };
    function clearPhoto() {
      if (file) file.value = "";
      if (thumb) {
        thumb.hidden = true;
        const img = thumb.querySelector("img");
        if (img) img.removeAttribute("src");
      }
    }
    if (file && thumb) {
      file.onchange = () => {
        const f = file.files && file.files[0];
        if (!f) { clearPhoto(); return; }
        const r = new FileReader();
        r.onload = () => {
          thumb.hidden = false;
          thumb.querySelector("img").src = r.result;
        };
        r.readAsDataURL(f);
      };
    }
    if (drop) drop.onclick = (e) => { e.preventDefault(); clearPhoto(); };
    form.onsubmit = async (e) => {
      e.preventDefault();
      const errEl = form.querySelector(".form-error");
      if (errEl) errEl.textContent = "";
      try {
        const image = await readImageFromForm(form, "photo");
        await api().community.createPost((ta && ta.value) || "", "post", image ? { image_url: image } : {});
        form.reset();
        if (cc) cc.textContent = "0";
        clearPhoto();
        if (onPosted) onPosted();
      } catch (ex) {
        if (ex && ex.status === 401) { toast("Log in to do that"); location.href = loginHref(); return; }
        if (errEl) errEl.textContent = ex.message;
        else toast(ex.message);
      }
    };
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
    const id = Number(p.id);
    const wished = lists.hasWish(id);
    const compared = lists.hasCompare(id);
    const inCart = lists.inCart(id);
    return `<article class="card">
      <a class="card-still" href="${productUrl(p.slug)}">
        ${pinned(p) ? '<span class="pin">Pinned</span>' : ""}
        <span class="file-tag">${esc(fileType(p))}</span>
        <img src="${img(p)}" alt="" />
        <div class="card-actions">
          <button type="button" class="pill${wished ? " is-on" : ""}" data-wish="${id}" aria-pressed="${wished}">${wished ? "Unsave" : "Save"}</button>
          <button type="button" class="pill${compared ? " is-on" : ""}" data-compare="${id}" aria-pressed="${compared}">${compared ? "Compared" : "Compare"}</button>
          <button type="button" class="pill${inCart ? " is-on" : ""}" data-cart="${id}" aria-pressed="${inCart}">${inCart ? "In cart" : "Add"}</button>
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
    if (!el || el._pawCardsBound) return;
    el._pawCardsBound = true;
    el.addEventListener("click", async (e) => {
      const wish = e.target.closest("[data-wish]");
      const cart = e.target.closest("[data-cart]");
      const cmp = e.target.closest("[data-compare]");
      if (!wish && !cart && !cmp) return;
      e.preventDefault();
      e.stopPropagation();
      try {
        if (wish) {
          const id = Number(wish.dataset.wish);
          const res = await api().wishlist.toggle(id);
          const on = !!(res && res.added);
          lists.setWish(id, on);
          wish.classList.toggle("is-on", on);
          wish.setAttribute("aria-pressed", on ? "true" : "false");
          wish.textContent = on ? "Unsave" : "Save";
          toast(on ? "Saved to wishlist" : "Removed from wishlist");
          if (!on && /wishlist/.test(location.pathname)) el.dispatchEvent(new Event("paw-refresh"));
        }
        if (cmp) {
          const id = Number(cmp.dataset.compare);
          if (!lists.hasCompare(id) && lists.compare.size >= 4) {
            toast("Compare up to 4 assets");
            return;
          }
          const res = await api().compare.toggle(id);
          const on = !!(res && res.added);
          lists.setCompare(id, on);
          cmp.classList.toggle("is-on", on);
          cmp.setAttribute("aria-pressed", on ? "true" : "false");
          cmp.textContent = on ? "Compared" : "Compare";
          toast(on ? "Added to compare" : "Removed from compare");
        }
        if (cart) {
          const id = Number(cart.dataset.cart);
          const res = await api().cart.toggle(id, 1);
          await lists.refresh();
          const on = lists.inCart(id);
          cart.classList.toggle("is-on", on);
          cart.setAttribute("aria-pressed", on ? "true" : "false");
          cart.textContent = on ? "In cart" : "Add";
          toast(res.added ? "Added to cart" : "Removed from cart");
        }
        if (global.PawChrome) PawChrome.mount();
      } catch (err) {
        if (err.status === 401) { toast("Log in to do that"); location.href = loginHref(); }
        else toast(err.message);
      }
    });
  }

  function syncBuyButtons(p) {
    if (!p) return;
    const id = Number(p.id);
    const inCart = lists.inCart(id);
    const wished = lists.hasWish(id);
    const compared = lists.hasCompare(id);
    const add = document.getElementById("add");
    const wish = document.getElementById("wish");
    const cmp = document.getElementById("cmp");
    if (add) {
      add.textContent = inCart ? "Remove from cart" : "Add to cart";
      add.classList.toggle("btn-accent", !inCart);
      add.classList.toggle("is-on", inCart);
    }
    if (wish) {
      wish.textContent = wished ? "Unsave" : "Save to wishlist";
      wish.classList.toggle("is-on", wished);
    }
    if (cmp) {
      cmp.textContent = compared ? "Compared" : "Compare";
      cmp.classList.toggle("is-on", compared);
    }
  }

  function bindBuy(p, opts) {
    opts = opts || {};
    syncBuyButtons(p);
    const add = document.getElementById("add");
    const wish = document.getElementById("wish");
    const cmp = document.getElementById("cmp");
    if (add) add.onclick = async () => {
      try {
        const extra = typeof opts.extra === "function" ? opts.extra() : (opts.extra || {});
        const tierId = typeof opts.tierId === "function" ? opts.tierId() : opts.tierId;
        const res = await api().cart.toggle(p.id, 1, tierId, extra);
        await lists.refresh();
        syncBuyButtons(p);
        toast(res.added ? "Added to cart" : "Removed from cart");
        if (global.PawChrome) PawChrome.mount();
      } catch (err) {
        if (err.status === 401) location.href = loginHref();
        else toast(err.message);
      }
    };
    if (wish) wish.onclick = async () => {
      try {
        const res = await api().wishlist.toggle(p.id);
        lists.setWish(p.id, !!(res && res.added));
        syncBuyButtons(p);
        toast(res.added ? "Saved to wishlist" : "Removed from wishlist");
      } catch (err) {
        if (err.status === 401) location.href = loginHref();
        else toast(err.message);
      }
    };
    if (cmp) cmp.onclick = async () => {
      try {
        if (!lists.hasCompare(p.id) && lists.compare.size >= 4) {
          toast("Compare up to 4 assets");
          return;
        }
        const res = await api().compare.toggle(p.id);
        lists.setCompare(p.id, !!(res && res.added));
        syncBuyButtons(p);
        toast(res.added ? "Added to compare" : "Removed from compare");
        if (global.PawChrome) PawChrome.mount();
      } catch (err) {
        if (err.status === 401) location.href = loginHref();
        else toast(err.message);
      }
    };
  }

  function compareBoard(products) {
    const specs = [
      ["Price", (p) => `<span class="price">${money(price(p))}</span>`],
      ["Format", (p) => esc(fileType(p) || "—")],
      ["Category", (p) => esc(catName(p) || "—")],
      ["PWYW", (p) => pwyw(p) ? "yes, min " + money(p.pwyw_min_price) : "no"]
    ];
    const still = products.map((p) => `<a class="compare-still" href="${productUrl(p.slug)}"><img src="${esc(img(p))}" alt=""></a>`).join("");
    const titles = products.map((p) => `<div class="compare-title"><a href="${productUrl(p.slug)}">${esc(p.title)}</a></div>`).join("");
    const specRows = specs.map(([label, fn]) =>
      `<div class="compare-label">${label}</div>` + products.map((p) => `<div class="compare-cell">${fn(p)}</div>`).join("")
    ).join("");
    const actions = products.map((p) => `<div class="compare-cell"><button type="button" class="btn" data-rm="${p.id}">Remove</button></div>`).join("");
    return `<div class="compare-board" style="--cols:${products.length}">
      <div class="compare-label compare-label-still"></div>${still}
      <div class="compare-label">Title</div>${titles}
      ${specRows}
      <div class="compare-label"></div>${actions}
    </div>`;
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
      const cached = JSON.parse(localStorage.getItem("store4bots_theme") || "null");
      if (cached) applyTheme(cached);
    } catch (e) { /* ignore */ }
    const client = apiObj || api();
    if (!client || !client.appearance) return;
    client.appearance.get().then((t) => {
      applyTheme(t);
      try { localStorage.setItem("store4bots_theme", JSON.stringify(t)); } catch (e) { /* ignore */ }
    }).catch(() => {});
  }

  global.PawUI = { esc, money, stars, card, bindCards, bindBuy, syncBuyButtons, compareBoard, productUrl, bundleUrl, orderUrl, toast, img, price, root, catalog, lists, couponOff, accountNav, communityNav, avatar, personRow, postCard, postText, postMedia, bindComposer, bindFollow, followControl, relationPills, loginHref, authGate, safeNext, readAvatarFromForm, readImageFromForm, avatarFields, mountHeroSlider, THEME, applyTheme, defaultTheme, bootTheme };
})(window);
