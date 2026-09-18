/**
 * Intercepts fetch() for /api/v1 so api-client.js can run against the in-browser store.
 * Response bodies match the live Go handlers (json tags + gin.H keys), not the internal store shape.
 *
 * Load order: store.js → mock-api.js → api-client.js
 */
(function (global) {
  const realFetch = global.fetch ? global.fetch.bind(global) : null;
  const TOKEN_KEY = "pawradise_proto_token";
  const ADMIN_TOKENS = ["admin_secret_staging_2026", "studio-admin"];
  let currentMethod = "GET";
  let currentPath = "";

  const eventLog = { ttl: 3600, rows: [] };

  function json(status, body) {
    return new Response(body == null ? "" : JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" }
    });
  }
  function err(status, message) { return json(status, { error: message }); }

  function headersOf(init, input) {
    const h = {};
    const raw = (init && init.headers) || (input && input.headers);
    if (!raw) return h;
    if (typeof raw.forEach === "function") raw.forEach((v, k) => { h[k.toLowerCase()] = v; });
    else Object.keys(raw).forEach((k) => { h[k.toLowerCase()] = raw[k]; });
    return h;
  }

  function bearer(init, input) {
    const h = headersOf(init, input);
    const a = h.authorization || "";
    const m = a.match(/^Bearer\s+(.+)/i);
    if (m) return m[1];
    try { return localStorage.getItem(TOKEN_KEY) || ""; } catch (e) { return ""; }
  }

  function authUser(init, input) {
    const t = bearer(init, input);
    const db = Paw.db();
    if (t && ADMIN_TOKENS.indexOf(t) === -1) {
      const m = String(t).match(/^mock\.(\d+)\./);
      if (m) return db.users.find((u) => u.id === Number(m[1])) || null;
    }
    if (db.session && db.session.userId) return db.users.find((u) => u.id === db.session.userId) || null;
    return null;
  }

  function operatorUnlocked(init, input) {
    const t = bearer(init, input);
    const x = headersOf(init, input)["x-admin-token"] || "";
    if (ADMIN_TOKENS.indexOf(t) >= 0 || ADMIN_TOKENS.indexOf(x) >= 0) return true;
    return !!Paw.isAdmin();
  }

  function tabForPath(method, path) {
    const p = String(path || "").toLowerCase();
    const m = String(method || "GET").toUpperCase();
    if (p.indexOf("/products/appearance") >= 0 || /\/appearance$/.test(p)) return "Appearance";
    if (p.indexOf("/products/banner") >= 0) return "Banner";
    if (p.indexOf("/admin/product-requests") >= 0) return "Requests";
    if (p.indexOf("/admin/guest") >= 0) return "Guest";
    if (p.indexOf("/admin/order-steps") >= 0) return m === "GET" ? "OrderStepsRead" : "Steps";
    if (p.indexOf("/admin/orders") >= 0) return "Orders";
    if (p.indexOf("/admin/community") >= 0) return "Community";
    if (p.indexOf("/admin/referrals") >= 0) return "Referrals";
    if (p.indexOf("/exchange-rates") >= 0) return "Rates";
    if (p.indexOf("/admin/coupons") >= 0) return "Coupons";
    if (p.indexOf("/admin/bundles") >= 0 || (p.indexOf("/bundles") >= 0 && m !== "GET")) return "Bundles";
    if (p.indexOf("/admin/categories") >= 0 || (p.indexOf("/categories") >= 0 && m !== "GET")) return "Categories";
    if (p.indexOf("/admin/export") >= 0) return "Export";
    if (p.indexOf("/admin/stats") >= 0 || p.indexOf("/products/stats") >= 0) return "Stats";
    if (p.indexOf("/admin/users") >= 0) return "Users";
    if (p.indexOf("/admin/settings") >= 0) {
      if (m === "GET") return "SettingsRead";
      return "Settings";
    }
    if (p.indexOf("/admin/products") >= 0 || (p.indexOf("/products") >= 0 && m !== "GET")) return "Products";
    return "";
  }

  function staffHas(u, want) {
    const tabs = String((u && u.staff_tabs) || "").split(",").map((s) => s.trim()).filter(Boolean);
    if (want === "SettingsRead") return tabs.some((t) => t === "Settings" || t === "SEO" || t === "Appearance");
    if (want === "OrderStepsRead") return tabs.indexOf("Orders") >= 0 || tabs.indexOf("Steps") >= 0;
    return tabs.indexOf(want) >= 0;
  }

  function isAdmin(init, input) {
    const t = bearer(init, input);
    if (ADMIN_TOKENS.indexOf(t) >= 0) return true;
    const u = authUser(init, input);
    return !!(u && (u.role === "admin" || u.role === "staff"));
  }

  function needUser(init, input) {
    const u = authUser(init, input);
    if (!u) throw Object.assign(new Error("unauthorized"), { status: 401, code: 401 });
    if (Paw.adoptSession) Paw.adoptSession(u.id);
    return u;
  }
  function needAdmin(init, input) {
    const t = bearer(init, input);
    if (ADMIN_TOKENS.indexOf(t) >= 0) return;
    const u = authUser(init, input);
    if (!u) throw Object.assign(new Error("unauthorized"), { status: 401, code: 401 });
    if (Paw.adoptSession) Paw.adoptSession(u.id);
    const role = u.role === "admin" || u.role === "staff" ? u.role : "customer";
    if (role === "customer") throw Object.assign(new Error("Staff or admin access required"), { status: 403 });
    const privileged = /\/admin\/users\/\d+\/(access|role)$/.test(currentPath || "");
    if (privileged) {
      if (role !== "admin" || !operatorUnlocked(init, input)) {
        throw Object.assign(new Error("Operator token required"), { status: 403 });
      }
      return;
    }
    if (role === "admin") return;
    const tab = tabForPath(currentMethod, currentPath);
    if (!staffHas(u, tab)) {
      throw Object.assign(new Error("This section is not assigned to your account"), { status: 403 });
    }
  }

  function bodyOf(init) {
    if (!init || init.body == null || init.body === "") return {};
    if (typeof init.body === "string") {
      try { return JSON.parse(init.body); } catch (e) { return {}; }
    }
    return init.body;
  }

  function iso(d) {
    if (!d) return new Date().toISOString();
    const s = String(d);
    if (/^\d{4}-\d{2}-\d{2}$/.test(s)) return s + "T00:00:00Z";
    return s;
  }

  function catBySlug(slug) { return Paw.findCategory(slug); }
  function catById(id) { return Paw.findCategory(id); }
  function catId(slug) { const c = catBySlug(slug); return c ? c.id : 0; }

  /** product-service models.Product — list + get */
  function productJSON(p) {
    if (!p) return null;
    const created = iso(p.created);
    return {
      id: p.id,
      title: p.title,
      slug: p.slug,
      description: p.description,
      image_url: p.thumb || Paw.media((p.images && p.images[0]) || "p01.jpg"),
      stock_count: 0,
      category_id: catId(p.category),
      price_usd: p.price,
      status: p.status || "active",
      stock_quantity: 0,
      max_downloads_per_user: 0,
      preview_images: (p.images || []).length,
      download_window_hours: 0,
      free_download: false,
      sort_order: 0,
      digital_formats: p.fileType || "",
      is_pinned: !!p.pinned,
      pinned: !!p.pinned,
      views: p.popular || 0,
      downloads: 0,
      purchase_count: p.popular || 0,
      tags: p.category || "",
      is_pwyw: !!p.pwyw,
      pwyw_min_price: p.pwywMin || 0,
      seo_title: p.seoTitle || "",
      seo_description: p.seoDescription || "",
      og_image: p.ogImage || "",
      created_at: created,
      updated_at: created
    };
  }

  function imagesJSON(p) {
    const created = iso(p.created);
    return (p.imageUrls || (p.images || []).map((f) => Paw.media(f))).map((url, i) => ({
      id: p.id * 10 + i + 1,
      product_id: p.id,
      url,
      mime_type: "image/jpeg",
      alt_text: p.title,
      is_primary: i === 0,
      image_type: i === 0 ? "full" : "preview",
      width: 0,
      height: 0,
      file_size_bytes: 0,
      storage_path: "",
      sort_order: i,
      created_at: created
    }));
  }

  function tiersJSON(p) {
    const created = iso(p.created);
    return (p.tiers || []).map((t, i) => ({
      id: p.id * 10 + i + 1,
      product_id: p.id,
      tier_name: t.name,
      price_usd: t.price,
      download_count: 0,
      download_limit: 0,
      is_active: true,
      created_at: created,
      updated_at: created
    }));
  }

  function categoryJSON(c) {
    return {
      id: c.id,
      name: c.name,
      slug: c.slug,
      description: c.desc || c.description || "",
      parent_id: c.parentId || 0,
      image_url: c.imageUrl || "",
      is_active: c.active !== false,
      sort_order: c.sort || c.id,
      created_at: iso(c.created) || "2026-01-01T00:00:00Z",
      updated_at: iso(c.updated || c.created) || "2026-01-01T00:00:00Z"
    };
  }

  function bundleJSON(b) {
    const created = "2026-01-01T00:00:00Z";
    return {
      id: b.id,
      title: b.title,
      slug: b.slug,
      description: b.description,
      price_usd: b.price,
      status: b.status || "active",
      sort_order: b.id,
      created_at: created,
      updated_at: created,
      items: (b.productIds || []).join(",")
    };
  }

  function userPublic(u, withTimes) {
    if (!u) return null;
    const out = { id: u.id, email: u.email, name: u.name, role: u.role || "customer", staff_tabs: u.staff_tabs || "" };
    if (withTimes) {
      out.created_at = "2026-01-01T00:00:00Z";
      out.updated_at = "2026-01-01T00:00:00Z";
    }
    return out;
  }

  function userMe(u) {
    const out = userPublic(u, true);
    out.profile = {
      user_id: u.id,
      display_name: u.name,
      avatar_url: u.avatar ? Paw.media(u.avatar) : "",
      bio: u.bio || "",
      wallet_address: u.wallet || "",
      social_links: "",
      preferred_currency: "USD",
      newsletter_enabled: true
    };
    return out;
  }

  function authorJSON(u, following) {
    u = u || {};
    const me = Paw.me();
    return {
      id: u.id || 0,
      email: u.email || "",
      name: u.name || "",
      display_name: u.name || "",
      avatar_url: u.avatar ? Paw.media(u.avatar) : (u.avatar_url || ""),
      following: !!following,
      follows_you: !!(me && u.id && Paw.db().follows.some((f) => f.followerId === u.id && f.followingId === me.id))
    };
  }

  function memberJSON(u) {
    if (!u) return null;
    const pub = Paw.profile(u.id) || {};
    const me = Paw.me();
    return {
      id: u.id,
      email: u.email || "",
      name: u.name,
      display_name: u.name,
      avatar_url: pub.avatar || "",
      bio: u.bio || "",
      follower_count: pub.followerCount || 0,
      following_count: pub.followingCount || 0,
      following: !!(me && Paw.isFollowing(u.id)),
      follows_you: !!(me && Paw.db().follows.some((f) => f.followerId === u.id && f.followingId === me.id))
    };
  }

  function feedPostJSON(p) {
    const u = Paw.db().users.find((x) => x.id === p.userId) || (p.author && { id: p.author.id, name: p.author.name, avatar: p.author.avatar });
    const me = Paw.me();
    const following = !!(me && Paw.isFollowing(p.userId));
    return {
      id: p.id,
      user_id: p.userId,
      content: p.content,
      type: "post",
      is_public: true,
      is_pinned: false,
      like_count: p.likes,
      comment_count: (p.comments || []).length,
      liked: !!p.liked,
      created_at: iso(p.created),
      updated_at: iso(p.created),
      author: authorJSON(u, following)
    };
  }

  function communityPostJSON(p) {
    return {
      id: p.id,
      user_id: p.userId,
      content: p.content,
      type: "post",
      is_public: true,
      is_pinned: false,
      like_count: p.likes,
      comment_count: (p.comments || []).length,
      created_at: iso(p.created),
      updated_at: iso(p.created)
    };
  }

  function orderListJSON(o) {
    return { id: o.id, total_usd: o.total, status: o.status, created_at: iso(o.created) };
  }

  function orderGetJSON(o) {
    return {
      id: o.id,
      status: o.status,
      total_usd: String(o.total),
      total_crypto: o.crypto != null ? String(o.crypto) : "",
      crypto_chain: o.chain || "BSC",
      payment_address: o.address || "",
      payment_tx_hash: o.paidAt ? "0xMOCK" : "",
      memo: String(o.id),
      payment_confirmations: o.status === "paid" ? 12 : 0,
      paid_at: o.paidAt || undefined,
      coupon_id: undefined,
      discount_usd: o.discount ? String(o.discount) : undefined,
      created_at: iso(o.created),
      updated_at: iso(o.paidAt || o.created)
    };
  }

  function ratePublic(r, i) {
    return {
      id: i + 1,
      chain: r.chain,
      symbol: r.symbol,
      rate_to_usd: String(r.rate),
      updated_at: "2026-01-01T00:00:00Z"
    };
  }

  function couponAdmin(c) {
    return {
      id: c.id,
      code: c.code,
      discount_type: c.type,
      discount_value: String(c.value),
      min_purchase_usd: String(c.min || 0),
      usage_limit: c.usageLimit,
      times_used: c.used,
      product_id: c.productId || undefined,
      is_active: !!c.active,
      expires_at: c.expires ? iso(c.expires) : undefined,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z"
    };
  }

  async function route(method, pathname, search, init, input) {
    const path = pathname.replace(/\/$/, "") || "/";
    const q = Object.fromEntries(new URLSearchParams(search));
    const parts = path.split("/").filter(Boolean);
    const p = parts.slice(2);
    const seg = (i) => p[i] || "";
    const M = method.toUpperCase();
    currentMethod = M;
    currentPath = path;

    if (M === "GET" && (path === "/api/v1/health" || path === "/health")) {
      return json(200, { status: "ok", service: "prototype-mock" });
    }

    // ---- identity ----
    if (M === "POST" && path === "/api/v1/register") {
      const b = bodyOf(init);
      const u = Paw.register({ name: b.name, email: b.email, password: b.password, referral: b.referral_code });
      const full = Paw.db().users.find((x) => x.id === u.id);
      const token = "mock." + u.id + "." + ((full && full.role) || "customer");
      localStorage.setItem(TOKEN_KEY, token);
      return json(201, { user: userPublic(full, true), token });
    }
    if (M === "POST" && path === "/api/v1/login") {
      const b = bodyOf(init);
      const pub = Paw.login(b.email, b.password);
      const u = Paw.db().users.find((x) => x.id === pub.id);
      const token = "mock." + pub.id + "." + ((u && u.role) || "customer");
      localStorage.setItem(TOKEN_KEY, token);
      return json(200, { user: userPublic(u, false), token });
    }
    if (M === "POST" && path === "/api/v1/logout") {
      localStorage.removeItem(TOKEN_KEY);
      Paw.logout();
      return json(200, { message: "logged out" });
    }
    if (M === "GET" && path === "/api/v1/me") {
      return json(200, userMe(needUser(init, input)));
    }
    if (M === "PUT" && path === "/api/v1/me") {
      needUser(init, input);
      const b = bodyOf(init);
      Paw.updateMe({ name: b.name, bio: b.bio, wallet: b.wallet_address, avatar: b.avatar_url });
      return json(200, { name: b.name || "", bio: b.bio || "", wallet_address: b.wallet_address || "", avatar_url: b.avatar_url || "" });
    }
    if (M === "GET" && seg(0) === "profile" && seg(1) && !seg(2)) {
      const u = Paw.db().users.find((x) => x.id === Number(seg(1)));
      if (!u) return err(404, "not found");
      return json(200, { user: userPublic(u, true) });
    }

    if (path === "/api/v1/referrals" && M === "GET") {
      const u = needUser(init, input);
      const r = Paw.referrals();
      const refs = (r.referred || []).map((x, i) => ({
        id: i + 1,
        code: r.code,
        created_at: "2026-01-01T00:00:00Z",
        referred_user_id: x.id,
        name: x.name,
        email: x.email,
        referral_created_at: "2026-01-01T00:00:00Z"
      }));
      return json(200, {
        referral_link: r.code,
        referrals: refs,
        total_earnings: r.total,
        total_commissions: r.total,
        total_referrals: refs.length
      });
    }
    if ((path === "/api/v1/referrals/earnings" || path === "/api/v1/commissions") && M === "GET") {
      needUser(init, input);
      const r = Paw.referrals();
      return json(200, {
        earnings: (r.earnings || []).map((c) => ({
          id: c.id,
          user_id: c.referrerId,
          referral_link_id: 1,
          order_id: c.orderId,
          amount: c.usd,
          created_at: "2026-01-01T00:00:00Z"
        }))
      });
    }
    if (M === "POST" && path === "/api/v1/referrals/track") {
      needUser(init, input);
      return json(200, { message: "tracked" });
    }

    // ---- products ----
    if (M === "GET" && path === "/api/v1/products/appearance") {
      return json(200, Paw.appearance());
    }
    if (M === "GET" && path === "/api/v1/products") {
      const res = Paw.listProducts({
        search: q.search || q.q,
        category: q.category || (q.category_id && (catById(q.category_id) || {}).slug),
        sort: q.sort,
        rating: q.rating,
        price_min: q.price_min,
        price_max: q.price_max,
        file_type: q.file_type,
        page: q.page,
        per_page: q.per_page || 20,
        banner: q.banner
      });
      return json(200, {
        products: res.products.map((x) => productJSON(x)),
        total: res.total,
        page: res.page,
        per_page: res.per_page
      });
    }
    if (M === "PUT" && path === "/api/v1/products/appearance") {
      needAdmin(init, input);
      return json(200, Paw.saveAppearance(bodyOf(init)));
    }
    if (M === "PUT" && path === "/api/v1/products/banner") {
      needAdmin(init, input);
      const ids = bodyOf(init).product_ids || [];
      Paw.setBanner(ids);
      return json(200, { message: "banner updated", product_ids: ids });
    }
    if (M === "GET" && seg(0) === "products" && seg(2) === "tiers") {
      const prod = Paw.products.find((x) => x.id === Number(seg(1))) || Paw.getProduct(seg(1));
      if (!prod) return err(404, "not found");
      const full = Paw.getProduct(prod.slug);
      return json(200, { tiers: tiersJSON(full) });
    }
    if (M === "GET" && seg(0) === "products" && seg(1) && !seg(2)) {
      const prod = Paw.getProduct(seg(1));
      if (!prod) return err(404, "product not found");
      const u = authUser(init, input);
      if (u) Paw.recordView(prod.id);
      return json(200, { product: productJSON(prod), images: imagesJSON(prod) });
    }
    if (M === "GET" && path === "/api/v1/categories") {
      return json(200, { categories: Paw.listCategories(false).map(categoryJSON) });
    }
    if (M === "POST" && path === "/api/v1/categories") {
      needAdmin(init, input);
      const b = bodyOf(init);
      if (!b.name) return err(400, "name is required");
      Paw.saveCategory({
        name: b.name,
        slug: b.slug,
        description: b.description,
        parent_id: b.parent_id,
        sort_order: b.sort_order,
        is_active: b.is_active,
        image_url: b.image_url
      });
      return json(201, { message: "category created" });
    }
    if (M === "GET" && seg(0) === "categories" && seg(1)) {
      const c = catById(seg(1)) || catBySlug(seg(1));
      if (!c) return err(404, "not found");
      return json(200, { category: categoryJSON(c) });
    }
    if ((M === "PUT" || M === "PATCH") && seg(0) === "categories" && seg(1) && !seg(2)) {
      needAdmin(init, input);
      const b = bodyOf(init);
      Paw.saveCategory({
        id: Number(seg(1)),
        name: b.name,
        slug: b.slug,
        description: b.description,
        parent_id: b.parent_id,
        sort_order: b.sort_order,
        is_active: b.is_active,
        image_url: b.image_url
      });
      return json(200, { message: "category updated" });
    }
    if (M === "DELETE" && seg(0) === "categories" && seg(1) && !seg(2)) {
      needAdmin(init, input);
      Paw.deleteCategory(seg(1));
      return json(200, { message: "category deleted" });
    }
    if (M === "GET" && path === "/api/v1/bundles") {
      return json(200, { bundles: Paw.bundles().map(bundleJSON) });
    }
    if (M === "GET" && seg(0) === "bundles" && seg(1)) {
      const b = Paw.getBundle(seg(1));
      if (!b) return err(404, "bundle not found");
      return json(200, { bundle: bundleJSON(b) });
    }
    if (M === "GET" && seg(0) === "recommendations" && seg(1)) {
      const prod = Paw.products.find((x) => x.id === Number(seg(1))) || Paw.getProduct(seg(1));
      const related = prod ? (Paw.getProduct(prod.slug).related || []) : [];
      return json(200, { products: related.map((x) => productJSON(x)) });
    }

    // ---- cart / commerce ----
    if (seg(0) === "cart") {
      const u = needUser(init, input);
      void u;
      if (M === "GET" && path === "/api/v1/cart") {
        const c = Paw.cartView();
        return json(200, {
          items: c.items.map((i, idx) => ({
            id: i.id || idx + 1,
            product_id: i.productId,
            quantity: i.qty
          }))
        });
      }
      if (M === "POST" && path === "/api/v1/cart/items") {
        const b = bodyOf(init);
        const prod = Paw.products.find((x) => x.id === Number(b.product_id));
        const tier = b.tier_id || (prod && prod.tiers[0] && prod.tiers[0].id);
        Paw.addToCart(Number(b.product_id), tier, b.quantity || 1, b.pwyw_price);
        return json(200, { message: "added to cart" });
      }
      if (M === "DELETE" && seg(1) === "items" && seg(2)) {
        Paw.removeCartItem(seg(2));
        return json(200, { message: "removed from cart" });
      }
      if (M === "PUT" && seg(1) === "items" && seg(2)) {
        Paw.setCartItemQty(seg(2), bodyOf(init).quantity);
        return json(200, { message: "quantity updated" });
      }
    }

    if (path === "/api/v1/wishlist") {
      needUser(init, input);
      if (M === "GET") {
        return json(200, {
          products: Paw.wishlist().map((id, i) => ({ id: i + 1, product_id: id }))
        });
      }
    }
    if (M === "POST" && path === "/api/v1/wishlist/toggle") {
      needUser(init, input);
      const id = Number(bodyOf(init).product_id || bodyOf(init).id);
      const before = Paw.wishlist().includes(id);
      Paw.toggleWishlist(id);
      return json(200, {
        message: before ? "removed from wishlist" : "added to wishlist",
        added: !before
      });
    }

    if (path === "/api/v1/recently-viewed") {
      needUser(init, input);
      if (M === "GET") {
        return json(200, {
          products: Paw.db().recentlyViewed.map((id, i) => ({
            id: i + 1,
            product_id: id,
            viewed_at: new Date().toISOString()
          }))
        });
      }
      if (M === "POST") {
        const id = Number(bodyOf(init).product_id || seg(1));
        if (id) Paw.recordView(id);
        return json(200, { message: "view recorded" });
      }
    }
    if (path === "/api/v1/compare") {
      needUser(init, input);
      if (M === "GET") {
        return json(200, {
          products: Paw.db().compare.map((id, i) => ({ id: i + 1, product_id: id }))
        });
      }
    }
    if (M === "POST" && path === "/api/v1/compare/toggle") {
      needUser(init, input);
      const id = Number(bodyOf(init).product_id);
      const before = Paw.db().compare.includes(id);
      Paw.toggleCompare(id);
      return json(200, { message: before ? "removed from compare" : "added to compare" });
    }

    if (M === "POST" && path === "/api/v1/events") {
      const b = bodyOf(init);
      if (b.name !== "product_view" && b.name !== "checkout_click") return err(400, "unknown event");
      eventLog.rows.unshift({
        id: String(eventLog.rows.length + 1),
        name: b.name,
        session_id: b.session_id || "",
        path: b.path || "",
        properties: b.properties || {},
        occurred_at: b.occurred_at || new Date().toISOString(),
        received_at: new Date().toISOString(),
        expire_at: new Date(Date.now() + eventLog.ttl * 1000).toISOString()
      });
      return json(202, { ok: true, name: b.name });
    }
    if (path === "/api/v1/admin/events/ttl") {
      if (M === "GET") return json(200, { seconds: eventLog.ttl, hours: eventLog.ttl / 3600 });
      const b = bodyOf(init);
      eventLog.ttl = Math.max(60, Number(b.seconds || b.hours * 3600) || 3600);
      return json(200, { seconds: eventLog.ttl, hours: eventLog.ttl / 3600 });
    }
    if (M === "GET" && path === "/api/v1/admin/events") {
      return json(200, { events: eventLog.rows.slice(0, Number(q.limit) || 20), total: eventLog.rows.length });
    }

    if (M === "POST" && path === "/api/v1/coupons/validate") {
      const b = bodyOf(init);
      const cart = Paw.cartView();
      const v = Paw.validateCoupon(b.code, b.cart_total != null ? b.cart_total : cart.subtotal, cart.items);
      if (!v.ok) {
        const msg = v.error === "Coupon not found" ? "coupon not found" : v.error;
        const status = msg === "coupon not found" ? 404 : 400;
        return err(status, msg.toLowerCase().includes("not found") ? "coupon not found" : msg.charAt(0).toLowerCase() + msg.slice(1).replace(/^C/, "c"));
      }
      return json(200, {
        valid: true,
        discount_type: v.coupon.type,
        discount_value: String(v.coupon.value)
      });
    }

    if (seg(0) === "orders") {
      if (M === "POST" && path === "/api/v1/orders") {
        needUser(init, input);
        const b = bodyOf(init);
        const order = Paw.checkout({ couponCode: b.coupon_code, chain: b.crypto_chain || "BSC" });
        return json(201, {
          order: { id: order.id },
          total_usd: order.total,
          payment_address: order.address,
          total_crypto: 0,
          crypto_chain: order.chain || "BSC",
          memo: order.id
        });
      }
      if (M === "GET" && path === "/api/v1/orders") {
        needUser(init, input);
        return json(200, { orders: Paw.myOrders().map(orderListJSON) });
      }
      if (seg(1) && M === "GET" && !seg(2)) {
        const u = authUser(init, input);
        const o = Paw.getOrder(seg(1), u && u.email);
        if (!o) return err(404, "not found");
        if (u && o.email !== u.email) return err(403, "not your order");
        const items = (o.items || []).map((it, i) => ({
          id: i + 1,
          order_id: o.id,
          product_id: it.productId,
          product_title: it.title,
          product_slug: it.slug,
          quantity: it.qty,
          unit_price_usd: it.unit,
          download_count: o.status === "paid" ? 1 : 0
        }));
        return json(200, { order: orderGetJSON(o), items });
      }
      if (seg(2) === "status" && M === "GET") {
        const o = Paw.getOrder(seg(1));
        if (!o) return err(404, "not found");
        return json(200, {
          order_id: o.id,
          status: o.status,
          payment_tx_hash: o.paidAt ? "0xMOCK" : "",
          confirmations: o.status === "paid" ? 12 : 0
        });
      }
      if (seg(2) === "payment" && M === "GET") {
        const o = Paw.getOrder(seg(1));
        if (!o) return err(404, "not found");
        return json(200, { order_id: o.id, status: o.status, total_usd: o.total });
      }
      if (seg(2) === "confirm" && M === "POST") {
        const u = needUser(init, input);
        Paw.confirmPay(seg(1), u.email);
        return json(200, { message: "payment confirmed" });
      }
      if (seg(2) === "download" && M === "GET") {
        needUser(init, input);
        return json(200, { message: "download endpoint" });
      }
    }

    if (M === "GET" && seg(0) === "payments" && seg(1) === "order" && seg(2) && !seg(3)) {
      const o = Paw.getOrder(seg(2));
      if (!o) return err(404, "not found");
      return json(200, {
        order_id: o.id,
        status: o.status,
        total_usd: o.total,
        crypto_chain: o.chain || "BSC",
        crypto_amount: String(o.crypto || 0),
        memo: String(o.id),
        payment_tx_hash: o.paidAt ? "0xMOCK" : "",
        created_at: iso(o.created)
      });
    }
    if (M === "GET" && seg(0) === "payments" && seg(3) === "status") {
      const o = Paw.getOrder(seg(2));
      if (!o) return err(404, "not found");
      return json(200, {
        order_id: o.id,
        status: o.status,
        payment_confirmations: o.status === "paid" ? 12 : 0,
        payment_tx_hash: o.paidAt ? "0xMOCK" : ""
      });
    }
    if (M === "POST" && seg(0) === "payments" && seg(3) === "confirm") {
      const u = needUser(init, input);
      Paw.confirmPay(seg(2), u.email);
      return json(200, { message: "payment confirmed", order_id: Number(seg(2)) });
    }

    if (path === "/api/v1/guest-orders" && M === "POST") {
      const b = bodyOf(init);
      const order = Paw.checkout({ guestEmail: b.email, couponCode: b.coupon_code, chain: b.crypto_chain });
      return json(201, { id: order.id, message: "guest order created" });
    }
    if (M === "GET" && path === "/api/v1/guest-orders") {
      const email = q.email || "";
      const orders = Paw.db().orders.filter((o) => o.email === email && !o.userId).map((o) => ({
        id: o.id, email: o.email, total_usd: o.total, status: o.status, created_at: iso(o.created)
      }));
      return json(200, { orders });
    }
    if (M === "GET" && seg(0) === "guest-orders" && seg(1)) {
      const o = Paw.getOrder(seg(1), q.email);
      if (!o) return err(404, "not found");
      return json(200, { id: o.id, email: o.email, total_usd: o.total, status: o.status, created_at: iso(o.created) });
    }

    // ---- community ----
    if (seg(0) === "community" || seg(0) === "posts" || seg(0) === "users" || seg(0) === "follow") {
      const rest = seg(0) === "community" ? p.slice(1) : p;
      if (M === "GET" && (path === "/api/v1/community/posts" || path === "/api/v1/community/feed" || path === "/api/v1/posts")) {
        let all = Paw.listPosts();
        if (q.user_id) all = all.filter((p) => p.userId === Number(q.user_id));
        if (q.filter === "following") {
          const u = authUser(init, input);
          if (!u) all = [];
          else {
            const ids = Paw.db().follows.filter((f) => f.followerId === u.id).map((f) => f.followingId);
            all = all.filter((p) => ids.indexOf(p.userId) >= 0);
          }
        }
        const page = Number(q.page) || 1;
        const perPage = Number(q.per_page) || 20;
        const slice = all.slice((page - 1) * perPage, page * perPage);
        return json(200, { posts: slice.map(feedPostJSON), total: all.length, page, per_page: perPage });
      }
      if (M === "GET" && rest[0] === "posts" && rest[1] && !rest[2]) {
        const post = Paw.getPost(rest[1]);
        if (!post) return err(404, "post not found");
        const u = Paw.db().users.find((x) => x.id === post.userId);
        const comments = (post.comments || []).map((c) => {
          const cu = Paw.db().users.find((x) => x.id === c.userId) || {};
          return {
            id: c.id,
            post_id: post.id,
            user_id: c.userId,
            content: c.content,
            like_count: 0,
            created_at: iso(post.created),
            updated_at: iso(post.created),
            user: {
              id: cu.id || 0,
              email: cu.email || "",
              name: cu.name || (c.author && c.author.name) || "",
              display_name: cu.name || "",
              avatar_url: cu.avatar ? Paw.media(cu.avatar) : ""
            }
          };
        });
        const viewer = authUser(init, input);
        return json(200, {
          post: Object.assign(communityPostJSON(post), { liked: !!post.liked }),
          author: authorJSON(u, !!(viewer && Paw.isFollowing(post.userId))),
          comments,
          total_comments: comments.length
        });
      }
      if (M === "POST" && (path === "/api/v1/community/posts" || path === "/api/v1/posts")) {
        const u = needUser(init, input);
        const b = bodyOf(init);
        const post = Paw.addPost(b.content);
        return json(201, { id: post.id, user_id: u.id, content: b.content, type: b.type || "post" });
      }
      if (rest[0] === "posts" && rest[2] === "like") {
        const u = needUser(init, input);
        const postId = Number(rest[1]);
        const liked = Paw.db().likes.some((l) => l.postId === postId && l.userId === u.id);
        if (M === "POST") {
          if (!liked) Paw.toggleLike(postId);
          return json(200, { message: "post liked" });
        }
        if (M === "DELETE") {
          if (liked) Paw.toggleLike(postId);
          return json(200, { message: "post unliked" });
        }
      }
      if (M === "POST" && rest[0] === "posts" && rest[2] === "comments") {
        needUser(init, input);
        const post = Paw.addComment(Number(rest[1]), bodyOf(init).content);
        const last = (post.comments || [])[post.comments.length - 1];
        return json(201, { comment: { id: last && last.id, post_id: Number(rest[1]), content: bodyOf(init).content } });
      }
      if (M === "DELETE" && rest[0] === "posts" && rest[2] === "comments") {
        needUser(init, input);
        Paw.deleteComment(Number(rest[3]));
        return json(200, { message: "comment deleted" });
      }
      if (M === "DELETE" && rest[0] === "posts" && rest[1] && !rest[2]) {
        needUser(init, input);
        Paw.deleteOwnPost(Number(rest[1]));
        return json(200, { message: "post deleted" });
      }
      if (rest[0] === "follow" && rest[1]) {
        const u = needUser(init, input);
        void u;
        if (M === "POST") {
          Paw.follow(rest[1]);
          return json(200, { message: "user followed" });
        }
        if (M === "DELETE") {
          Paw.unfollow(rest[1]);
          return json(200, { message: "user unfollowed" });
        }
      }
      if (M === "GET" && path === "/api/v1/community/suggestions") {
        return json(200, { users: Paw.suggestions().map((p) => memberJSON(Paw.db().users.find((x) => x.id === p.id))) });
      }
      if (M === "GET" && (path === "/api/v1/community/users" || path === "/api/v1/users") && !rest[1]) {
        let users = Paw.db().users.map((x) => memberJSON(x));
        if (q.q) {
          const s = String(q.q).toLowerCase();
          users = users.filter((u) => ((u.name || "") + " " + (u.bio || "") + " " + (u.email || "")).toLowerCase().includes(s));
        }
        users.sort((a, b) => (b.follower_count || 0) - (a.follower_count || 0));
        const page = Math.max(1, Number(q.page) || 1);
        const perPage = Math.min(50, Math.max(1, Number(q.per_page) || 12));
        const total = users.length;
        const slice = users.slice((page - 1) * perPage, page * perPage);
        return json(200, { users: slice, total, page, per_page: perPage });
      }
      if (M === "GET" && rest[0] === "users" && rest[2] === "followers") {
        const all = Paw.listFollowers(rest[1]).map((x) => memberJSON(x));
        const page = Math.max(1, Number(q.page) || 1);
        const perPage = Math.min(50, Math.max(1, Number(q.per_page) || 12));
        return json(200, { users: all.slice((page-1)*perPage, page*perPage), total: all.length, page, per_page: perPage });
      }
      if (M === "GET" && rest[0] === "users" && rest[2] === "following") {
        const all = Paw.listFollowing(rest[1]).map((x) => memberJSON(x));
        const page = Math.max(1, Number(q.page) || 1);
        const perPage = Math.min(50, Math.max(1, Number(q.per_page) || 12));
        return json(200, { users: all.slice((page-1)*perPage, page*perPage), total: all.length, page, per_page: perPage });
      }
      if (M === "GET" && rest[0] === "users" && rest[1] && !rest[2]) {
        const u = Paw.profile(rest[1]);
        if (!u) return err(404, "user not found");
        const full = Paw.db().users.find((x) => x.id === u.id);
        const me = authUser(init, input);
        return json(200, Object.assign(memberJSON(full), {
          wallet_address: full.wallet || "",
          post_count: u.postCount,
          created_at: "2026-01-01T00:00:00Z",
          updated_at: "2026-01-01T00:00:00Z",
          is_self: !!(me && me.id === u.id)
        }));
      }
    }

    if (M === "GET" && path === "/api/v1/profile") {
      const u = needUser(init, input);
      const pub = Paw.profile(u.id);
      return json(200, {
        user: userPublic(u, true),
        post_count: pub.postCount,
        follower_count: pub.followerCount,
        following_count: pub.followingCount
      });
    }

    // ---- reviews ----
    if (M === "POST" && path === "/api/v1/reviews") {
      const u = needUser(init, input);
      const b = bodyOf(init);
      Paw.addReview(Number(b.product_id), b.rating);
      return json(201, { review: { id: Date.now() % 100000, product_id: Number(b.product_id), rating: Number(b.rating), user_id: u.id } });
    }
    if (M === "GET" && seg(0) === "reviews" && seg(2) === "average") {
      const prod = Paw.products.find((x) => x.id === Number(seg(1))) || Paw.getProduct(seg(1));
      if (!prod) return err(404, "not found");
      const full = Paw.getProduct(prod.slug);
      const list = full.reviewList || [];
      const totalRating = list.reduce((n, r) => n + r.rating, 0);
      const dist = { "1_star": 0, "2_star": 0, "3_star": 0, "4_star": 0, "5_star": 0 };
      list.forEach((r) => { dist[r.rating + "_star"] = (dist[r.rating + "_star"] || 0) + 1; });
      return json(200, {
        product_id: prod.id,
        average_rating: full.rating || 0,
        total_reviews: list.length || full.reviews || 0,
        total_rating: totalRating,
        distribution: dist
      });
    }
    if (M === "GET" && seg(0) === "reviews" && seg(1) && !seg(2)) {
      const prod = Paw.products.find((x) => x.id === Number(seg(1))) || Paw.getProduct(seg(1));
      const full = prod ? Paw.getProduct(prod.slug) : null;
      const reviews = ((full && full.reviewList) || []).map((r, i) => ({
        id: i + 1,
        user_id: r.userId,
        product_id: r.productId,
        rating: r.rating,
        title: "",
        content: "",
        verified: true,
        is_verified_purchase: true,
        is_helpful: false,
        helpful_count: 0,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
        author: {
          id: r.user && r.user.id,
          name: (r.user && r.user.name) || "",
          email: "",
          avatar_url: (r.user && r.user.avatar) || "",
          display_name: (r.user && r.user.name) || ""
        }
      }));
      const page = Number(q.page) || 1;
      const perPage = Number(q.per_page) || 20;
      return json(200, { reviews, total: reviews.length, page, per_page: perPage });
    }

    // ---- rates / settings ----
    if (M === "GET" && path === "/api/v1/exchange-rates") {
      return json(200, { rates: Paw.rates().map(ratePublic) });
    }
    if (M === "GET" && seg(0) === "exchange-rates" && seg(1) && seg(0) !== "admin") {
      const idx = Paw.rates().findIndex((x) => x.chain === seg(1));
      if (idx < 0) return err(404, "rate not found");
      return json(200, { rate: ratePublic(Paw.rates()[idx], idx) });
    }
    if (M === "GET" && path === "/api/v1/settings") {
      const settings = Object.entries(Paw.settings()).map(([key, value]) => ({
        key, value: String(value), updated_at: "2026-01-01T00:00:00Z"
      }));
      return json(200, { settings });
    }
    if (M === "GET" && seg(0) === "settings" && seg(1) && seg(0) !== "admin") {
      const v = Paw.setting(seg(1));
      if (v == null) return err(404, "not found");
      return json(200, { key: seg(1), value: String(v) });
    }

    if (M === "POST" && path === "/api/v1/product-requests") {
      const b = bodyOf(init);
      if (!b.title) return err(400, "title is required");
      const row = Paw.addRequest(b);
      return json(201, { id: row.id, message: "request submitted" });
    }

    if (M === "GET" && path === "/api/v1/product-requests") {
      const u = authUser(init, input);
      const all = Paw.listRequests().filter((r) => !u || !r.email || r.email === u.email);
      return json(200, { requests: all });
    }

    // ---- admin ----
    if (seg(0) === "admin") {
      needAdmin(init, input);

      if (M === "PUT" && seg(1) === "users" && seg(3) === "access") {
        const b = bodyOf(init);
        const saved = Paw.setUserAccess(seg(2), b.role, b.staff_tabs);
        return json(200, Object.assign({ message: "access updated" }, saved));
      }

      if (M === "GET" && (path === "/api/v1/admin/stats" || path === "/api/v1/admin/products/stats")) {
        const s = Paw.stats();
        return json(200, {
          total_users: s.users,
          total_orders: s.orders,
          total_revenue: Math.floor(s.revenue),
          total_products: Paw.db().products.length,
          active_products: s.products,
          pinned_products: s.pinned,
          total_categories: s.categories,
          total_downloads: s.downloads,
          revenue_daily: s.revenue_daily,
          top_products: s.top_products,
          conversions: s.conversions,
          order_by_step: s.order_by_step || []
        });
      }
      if (M === "GET" && path === "/api/v1/admin/users") {
        return json(200, { users: Paw.db().users.map((u) => userPublic(u, true)) });
      }
      if (M === "GET" && path === "/api/v1/admin/categories") {
        const counts = {};
        Paw.db().products.forEach((p) => { counts[p.category] = (counts[p.category] || 0) + 1; });
        return json(200, {
          categories: Paw.listCategories(true).map((c) => Object.assign(categoryJSON(c), { product_count: counts[c.slug] || 0 }))
        });
      }
      if (M === "DELETE" && seg(1) === "users" && seg(2) && !seg(3)) {
        Paw.deleteUser(seg(2));
        return json(200, { message: "user deleted" });
      }
      if (M === "POST" && seg(1) === "users" && seg(3) === "reset-password") {
        const pw = Paw.resetPassword(seg(2));
        return json(200, { message: "password reset", new_password: pw });
      }
      if (M === "GET" && path === "/api/v1/admin/products") {
        const all = Paw.db().products.map((raw) => {
          const p = Paw.getProduct(raw.slug) || raw;
          const cat = catBySlug(p.category);
          return {
            id: p.id,
            title: p.title,
            slug: p.slug,
            description: p.description,
            category_id: cat ? cat.id : 0,
            price_usd: p.price,
            status: p.status || "active",
            stock_quantity: 0,
            max_downloads_per_user: 0,
            pinned: !!p.pinned,
            views: p.popular || 0,
            downloads: 0,
            purchase_count: p.popular || 0,
            created_at: iso(p.created),
            updated_at: iso(p.created),
            category_name: cat ? cat.name : "",
            category_slug: cat ? cat.slug : "",
            pwyw_enabled: !!p.pwyw,
            pwyw_min_price: p.pwywMin || 0,
            digital_formats: p.fileType || "",
            seo_title: p.seoTitle || "",
            seo_description: p.seoDescription || "",
            og_image: p.ogImage || ""
          };
        });
        return json(200, { products: all, total: all.length });
      }
      if (M === "POST" && path === "/api/v1/admin/products") {
        const b = bodyOf(init);
        Paw.saveProduct({
          title: b.title,
          slug: b.slug,
          description: b.description,
          price: b.price_usd,
          category: (catById(b.category_id) || {}).slug,
          status: b.status,
          pwyw: b.pwyw_enabled,
          pwywMin: b.pwyw_min_price,
          fileType: b.digital_formats || b.file_mime_type,
          pinned: b.pinned,
          seoTitle: b.seo_title,
          seoDescription: b.seo_description,
          ogImage: b.og_image
        });
        return json(201, { id: Paw.db().nextIds.product - 1, message: "product created" });
      }
      if (M === "PUT" && ((seg(0) === "products" && seg(1) === "appearance") || (seg(1) === "products" && seg(2) === "appearance"))) {
        needAdmin(init, input);
        return json(200, Paw.saveAppearance(bodyOf(init)));
      }
      if (M === "PUT" && ((seg(0) === "products" && seg(1) === "banner") || (seg(1) === "products" && seg(2) === "banner"))) {
        needAdmin(init, input);
        const ids = bodyOf(init).product_ids || [];
        Paw.setBanner(ids);
        return json(200, { message: "banner updated", product_ids: ids });
      }
      if (M === "PUT" && seg(1) === "products" && seg(2) && !seg(3)) {
        const b = bodyOf(init);
        Paw.saveProduct({
          id: Number(seg(2)),
          title: b.title,
          slug: b.slug,
          description: b.description,
          price: b.price_usd,
          category: b.category_id != null ? (catById(b.category_id) || {}).slug : undefined,
          status: b.status,
          pwyw: b.pwyw_enabled,
          pwywMin: b.pwyw_min_price,
          fileType: b.digital_formats || b.file_mime_type,
          pinned: b.pinned,
          seoTitle: b.seo_title,
          seoDescription: b.seo_description,
          ogImage: b.og_image
        });
        return json(200, { message: "product updated" });
      }
      if (M === "POST" && path === "/api/v1/admin/products/bulk") {
        const b = bodyOf(init);
        const ids = b.ids || b.product_ids || [];
        const fields = {};
        if (b.status) fields.status = b.status;
        if (b.action === "activate") fields.status = "active";
        if (b.action === "deactivate") fields.status = "draft";
        if (b.action === "delete") fields.status = "archived";
        if (b.action === "update_status" && b.status) fields.status = b.status;
        if (b.category_id != null) fields.category = (catById(b.category_id) || {}).slug;
        Paw.bulkProducts(ids.map(Number), fields);
        return json(200, { message: "bulk update applied", action: b.action || "update", rows_affected: ids.length });
      }
      if (M === "POST" && seg(1) === "products" && seg(3) === "tiers" && !seg(4)) {
        const b = bodyOf(init);
        const id = Paw.addTier(seg(2), b);
        return json(201, { id, message: "tier added" });
      }
      if (M === "PUT" && seg(1) === "products" && seg(3) === "tiers" && seg(4)) {
        Paw.updateTier(seg(2), seg(4), bodyOf(init));
        return json(200, { message: "tier updated" });
      }
      if (M === "DELETE" && seg(1) === "products" && seg(3) === "tiers" && seg(4)) {
        Paw.deleteTier(seg(2), seg(4));
        return json(200, { message: "tier deleted" });
      }
      if (M === "POST" && seg(1) === "products" && seg(3) === "images" && !seg(4)) {
        Paw.addImage(seg(2), bodyOf(init).url);
        return json(201, { id: Date.now() % 100000, message: "image added" });
      }
      if (M === "DELETE" && seg(1) === "products" && seg(3) === "images" && seg(4)) {
        Paw.deleteImage(seg(2), seg(4));
        return json(200, { message: "image deleted" });
      }
      if (M === "POST" && seg(1) === "products" && seg(3) === "generate-previews") {
        Paw.generatePreviews(seg(2));
        return json(200, { message: "previews generated" });
      }
      if (M === "POST" && seg(1) === "products" && seg(3) === "pin") {
        Paw.pinProduct(seg(2), true);
        return json(200, { message: "product pinned" });
      }
      if (M === "DELETE" && seg(1) === "products" && seg(3) === "pin") {
        Paw.pinProduct(seg(2), false);
        return json(200, { message: "product unpinned" });
      }
      if (M === "DELETE" && seg(1) === "products" && seg(2) && !seg(3)) {
        Paw.archiveProduct(seg(2));
        return json(200, { message: "product deleted" });
      }
      if (M === "GET" && path === "/api/v1/admin/order-steps") {
        return json(200, { steps: Paw.listOrderSteps() });
      }
      if (M === "POST" && path === "/api/v1/admin/order-steps") {
        const s = Paw.createOrderStep(bodyOf(init).label);
        return json(201, s);
      }
      if (M === "PUT" && path === "/api/v1/admin/order-steps") {
        return json(200, { steps: Paw.saveOrderSteps(bodyOf(init).steps || []) });
      }
      if (M === "PUT" && seg(1) === "order-steps" && seg(2) && !seg(3)) {
        Paw.renameOrderStep(seg(2), bodyOf(init).label);
        return json(200, { id: Number(seg(2)), label: bodyOf(init).label });
      }
      if (M === "DELETE" && seg(1) === "order-steps" && seg(2)) {
        Paw.deleteOrderStep(seg(2));
        return json(200, { message: "step deleted" });
      }
      if (M === "GET" && path === "/api/v1/admin/orders") {
        const steps = Paw.listOrderSteps();
        const labels = {};
        steps.forEach((s) => { labels[s.slug] = s.label; });
        let rows = Paw.db().orders.slice();
        if (q.status) rows = rows.filter((o) => o.status === q.status);
        if (q.q) {
          const needle = String(q.q).toLowerCase();
          rows = rows.filter((o) => String(o.id).indexOf(needle) >= 0 || String(o.email || "").toLowerCase().indexOf(needle) >= 0);
        }
        const orders = rows.sort((a, b) => b.id - a.id).map((o) => {
          const u = Paw.db().users.find((x) => x.id === o.userId);
          return {
            id: o.id,
            user_id: o.userId || 0,
            status: o.status,
            status_label: labels[o.status] || o.status,
            total_usd: o.total,
            total_crypto: String(o.crypto || 0),
            crypto_chain: o.chain || "BSC",
            payment_tx_hash: o.paidAt ? "0xMOCK" : "",
            payment_confirmations: o.status === "paid" ? 12 : 0,
            paid_at: o.paidAt || null,
            created_at: iso(o.created),
            updated_at: iso(o.paidAt || o.created),
            user_email: o.email,
            email: o.email,
            user_name: u ? u.name : ""
          };
        });
        const by_step = steps.map((s) => ({
          slug: s.slug, label: s.label, count: Paw.db().orders.filter((o) => o.status === s.slug).length, is_terminal: !!s.is_terminal
        }));
        return json(200, { orders, total: orders.length, by_step });
      }
      if (M === "GET" && seg(1) === "orders" && seg(2) && !seg(3)) {
        const o = Paw.getOrder(seg(2));
        if (!o) return err(404, "order not found");
        const u = Paw.db().users.find((x) => x.id === o.userId);
        const items = (o.items || []).map((it, i) => ({
          id: i + 1,
          order_id: o.id,
          product_id: it.productId,
          product_tier_id: null,
          product_title: it.title,
          product_slug: it.slug,
          quantity: it.qty,
          unit_price_usd: it.unit,
          download_count: o.status === "paid" ? 1 : 0,
          max_downloads: 0,
          created_at: iso(o.created),
          updated_at: iso(o.created)
        }));
        return json(200, {
          order: Object.assign(orderGetJSON(o), { user_id: o.userId || 0, user_email: o.email, user_name: u ? u.name : "" }),
          items
        });
      }
      if (M === "PUT" && seg(1) === "orders" && seg(3) === "status") {
        Paw.setOrderStatus(seg(2), bodyOf(init).status);
        return json(200, { message: "status updated" });
      }
      if (M === "GET" && path === "/api/v1/admin/community/posts") {
        return json(200, { posts: Paw.listPosts().map(communityPostJSON) });
      }
      if (M === "DELETE" && seg(1) === "community" && seg(2) === "posts") {
        Paw.adminDeletePost(seg(3));
        return json(200, { message: "post deleted" });
      }
      if (M === "GET" && path === "/api/v1/admin/settings") {
        return json(200, {
          settings: Object.entries(Paw.settings()).map(([key, value]) => ({
            key, value: String(value), updated_at: "2026-01-01T00:00:00Z"
          }))
        });
      }
      if (M === "PUT" && seg(1) === "settings" && seg(2)) {
        Paw.saveSetting(seg(2), bodyOf(init).value != null ? bodyOf(init).value : "");
        return json(200, { message: "setting updated" });
      }
      if (M === "GET" && path === "/api/v1/admin/referrals") {
        return json(200, {
          referrals: Paw.db().users.map((u) => {
            const kids = Paw.db().users.filter((x) => x.referredBy === u.id);
            const earnings = Paw.db().commissions.filter((c) => c.referrerId === u.id);
            return {
              id: u.id,
              user_id: u.id,
              code: u.referral,
              is_active: true,
              created_at: "2026-01-01T00:00:00Z",
              email: u.email,
              name: u.name,
              referral_count: kids.length,
              total_commission: earnings.reduce((n, c) => n + c.usd, 0)
            };
          })
        });
      }
      if (M === "POST" && path === "/api/v1/admin/export/emails") {
        const b = bodyOf(init);
        const users = Paw.exportEmails({
          from: b.from || q.from,
          to: b.to || q.to,
          category: b.category || q.category,
          product_id: b.product_id || q.product_id
        }).map((u) => ({
          email: u.email,
          name: u.name,
          created_at: iso(u.first_purchase),
          total_purchases: u.orders,
          first_purchase_date: iso(u.first_purchase),
          last_purchase_date: iso(u.last_purchase)
        }));
        if ((q.format || b.format) === "csv") {
          const header = "email,name,created_at";
          const rows = users.map((u) => [u.email, u.name, u.created_at].join(","));
          return new Response([header].concat(rows).join("\n"), {
            status: 200,
            headers: { "Content-Type": "text/csv", "Content-Disposition": "attachment; filename=users.csv" }
          });
        }
        return json(200, { users, count: users.length });
      }
      if (M === "GET" && path === "/api/v1/admin/bundles") {
        return json(200, { bundles: Paw.bundles().map(bundleJSON) });
      }
      if (M === "POST" && path === "/api/v1/admin/bundles") {
        const b = bodyOf(init);
        Paw.saveBundle({ title: b.title, slug: b.slug, description: b.description, price: b.price_usd, productIds: b.product_ids || [], status: b.status });
        return json(201, { id: Paw.db().nextIds.bundle - 1, message: "bundle created" });
      }
      if (M === "PUT" && seg(1) === "bundles" && seg(2)) {
        const b = bodyOf(init);
        Paw.saveBundle({ id: Number(seg(2)), title: b.title, slug: b.slug, description: b.description, price: b.price_usd, productIds: b.product_ids, status: b.status });
        return json(200, { message: "bundle updated" });
      }
      if (M === "DELETE" && seg(1) === "bundles" && seg(2)) {
        Paw.deleteBundle(seg(2));
        return json(200, { message: "bundle deleted" });
      }
      if (M === "GET" && path === "/api/v1/admin/coupons") {
        return json(200, { coupons: Paw.coupons().map(couponAdmin) });
      }
      if (M === "POST" && path === "/api/v1/admin/coupons") {
        const b = bodyOf(init);
        Paw.saveCoupon({
          code: b.code,
          type: b.discount_type,
          value: b.discount_value,
          min: b.min_purchase_usd,
          usageLimit: b.usage_limit,
          expires: b.expires_at,
          productId: b.product_id,
          active: b.is_active
        });
        return json(201, { id: Paw.db().nextIds.coupon - 1, message: "coupon created" });
      }
      if (M === "PUT" && seg(1) === "coupons" && seg(2)) {
        const b = bodyOf(init);
        Paw.saveCoupon({
          id: Number(seg(2)),
          code: b.code,
          type: b.discount_type,
          value: b.discount_value,
          min: b.min_purchase_usd,
          usageLimit: b.usage_limit,
          expires: b.expires_at,
          productId: b.product_id,
          active: b.is_active
        });
        return json(200, { message: "coupon updated" });
      }
      if (M === "DELETE" && seg(1) === "coupons" && seg(2)) {
        Paw.deleteCoupon(seg(2));
        return json(200, { message: "coupon deleted" });
      }
      if (M === "GET" && path === "/api/v1/admin/exchange-rates") {
        return json(200, {
          rates: Paw.rates().map((r) => ({
            chain: r.chain,
            symbol: r.symbol,
            rate_to_usd: r.rate,
            updated_at: "2026-01-01T00:00:00Z"
          }))
        });
      }
      if (M === "PUT" && seg(1) === "exchange-rates" && seg(2)) {
        const b = bodyOf(init);
        Paw.saveRate({ chain: seg(2), rate: b.rate_to_usd, symbol: b.symbol });
        return json(200, { message: "exchange rate updated" });
      }
      if (M === "DELETE" && seg(1) === "exchange-rates" && seg(2)) {
        Paw.deleteRate(seg(2));
        return json(200, { message: "exchange rate deleted" });
      }
      if (M === "GET" && path === "/api/v1/admin/guest-orders") {
        const orders = Paw.guestOrders().map((o) => ({
          id: o.id,
          email: o.email,
          total_usd: o.total,
          status: o.status,
          crypto_chain: o.chain || "BSC",
          created_at: iso(o.created)
        }));
        return json(200, { orders, total: orders.length });
      }
      if (M === "GET" && seg(1) === "guest-orders" && seg(2)) {
        const o = Paw.getOrder(seg(2));
        if (!o) return err(404, "not found");
        return json(200, { order: orderGetJSON(o), email: o.email, items: o.items || [] });
      }
      if (M === "GET" && path === "/api/v1/admin/product-requests") {
        return json(200, { requests: Paw.listRequests() });
      }
      if (M === "PUT" && seg(1) === "product-requests" && seg(2)) {
        Paw.setRequestStatus(seg(2), bodyOf(init).status || "reviewed");
        return json(200, { message: "request updated" });
      }
    }

    return err(404, "page not found");
  }

  global.fetch = async function (input, init) {
    const url = typeof input === "string" ? input : (input && input.url);
    let u;
    try { u = new URL(url, location.origin); } catch (e) { return realFetch ? realFetch(input, init) : err(500, "bad url"); }
    if (!u.pathname.startsWith("/api/v1") && u.pathname !== "/health") {
      if (realFetch) return realFetch(input, init);
      return err(404, "not found");
    }
    const method = (init && init.method) || (typeof input !== "string" && input.method) || "GET";
    try {
      return await route(method, u.pathname, u.search, init || {}, input);
    } catch (e) {
      const status = e.status || e.code || ((e.message === "unauthorized" || e.message === "auth") ? 401 : 400);
      return err(status, e.message === "auth" ? "unauthorized" : (e.message || "error"));
    }
  };
})(window);
