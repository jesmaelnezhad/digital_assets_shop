(function (global) {
  const root = () => "/";
  const here = global.SITE_PAGE || "";

  function navLink(href, label, page, extra, ico) {
    const on = here === page ? " is-on" : "";
    const mark = (global.PawUI && PawUI.icon && ico) ? PawUI.icon(ico) : "";
    return `<a class="${on}" href="${href}">${mark}<span>${label}</span>${extra || ""}</a>`;
  }

  async function mount() {
    if (global.PawUI && PawUI.bootTheme) PawUI.bootTheme(global.Store4bots && Store4bots.api);
    const header = document.getElementById("site-header");
    const footer = document.getElementById("site-footer");
    let count = 0;
    let user = null;
    if (global.Store4bots && Store4bots.api) {
      try { user = await Store4bots.api.auth.getProfile(); } catch (e) { user = null; }
      try {
        const cart = await Store4bots.api.cart.get();
        count = (cart.items || []).reduce((n, i) => n + (i.quantity || 0), 0);
      } catch (e) { count = 0; }
    }
    let wallet = "";
    if (user && user.profile && user.profile.wallet_address) wallet = user.profile.wallet_address;
    if (header) {
      header.className = "site-header";
      header.innerHTML = `
        <a class="brand" href="/">
          <span class="brand-mark">S</span> STORE4BOTS
        </a>
        <form class="header-search" action="/">
          ${(global.PawUI && PawUI.icon) ? PawUI.icon("search") : ""}
          <input name="search" placeholder="Search assets, kits, packs" value="${new URLSearchParams(location.search).get("search") || ""}" />
        </form>
        <button type="button" class="nav-toggle" aria-label="Menu" aria-expanded="false">${(global.PawUI && PawUI.icon) ? PawUI.icon("menu") : "☰"}</button>
        <nav class="header-nav">
          ${navLink("/", "Shop", "shop", "", "shop")}
          ${navLink("/product/bundles.html", "Bundles", "bundles", "", "bundles")}
          ${navLink("/community", "Community", "community", "", "community")}
          ${navLink("/cart", "Cart", "cart", count ? `<span class="badge">${count}</span>` : "", "cart")}
          ${user
            ? navLink("/account", (user.name || "Account").split(" ")[0], "account", "", "account") +
              `<button type="button" class="nav-auth" data-logout>${(global.PawUI && PawUI.icon) ? PawUI.icon("logout") : ""}<span>Log out</span></button>`
            : navLink("/login", "Log in", "auth", "", "login")}
          ${(user && (user.role === "admin" || user.role === "staff"))
            ? navLink("/admin", "Admin", "admin", "", "admin")
            : ""}
          <span class="wallet-dot" title="Wallet">${(global.PawUI && PawUI.icon) ? PawUI.icon("wallet") : ""} ${wallet ? "Wallet · " + PawUI.esc(wallet.slice(0, 8)) + "…" : "Wallet · none"}</span>
        </nav>
      `;
      const toggle = header.querySelector(".nav-toggle");
      if (toggle) {
        toggle.onclick = () => {
          const open = header.classList.toggle("nav-open");
          toggle.setAttribute("aria-expanded", open ? "true" : "false");
        };
      }
      const out = header.querySelector("[data-logout]");
      if (out) out.onclick = async () => {
        try { await Store4bots.api.auth.logout(); } catch (e) { /* still leave */ }
        location.href = "/";
      };
    }
    if (footer) {
      footer.className = "site-footer";
      footer.innerHTML = `
        <span>© Store4bots · buy once, download forever</span>
        <span class="footer-links">
          <a href="/">Shop</a>
          <a href="/category">Categories</a>
          <a href="/product/bundles.html">Bundles</a>
          <a href="/product/request.html">Request</a>
          <a href="/community">Community</a>
          <a href="/people">People</a>
          <a href="/wishlist">Wishlist</a>
          <a href="/recent">Recent</a>
          <a href="/referrals">Referrals</a>
          <a href="/guest">Guest order</a>
          <a href="/legal">Terms</a>
          <a href="/sitemap.xml">Sitemap</a>
        </span>`;
    }
    let bar = document.getElementById("compare-bar");
    let compared = [];
    try {
      if (global.PawUI && PawUI.lists) await PawUI.lists.refresh();
      const ids = (global.PawUI && PawUI.lists) ? [...PawUI.lists.compare] : [];
      compared = ids.length ? await PawUI.catalog.byIdsAsync(ids) : [];
    } catch (e) { compared = []; }
    if (compared.length) {
      if (!bar) { bar = document.createElement("div"); bar.id = "compare-bar"; document.body.appendChild(bar); }
      bar.className = "compare-bar";
      bar.innerHTML = `<span class="compare-bar-count">Compare ${compared.length}/4</span>` +
        compared.map((p) => `<a class="compare-bar-item" href="/product/${p.slug}">${PawUI.esc(p.title)}</a>`).join("") +
        `<a class="btn btn-accent compare-bar-open" href="/compare">Open</a>`;
    } else if (bar) bar.remove();
    if (!document.querySelector(".grain")) {
      const g = document.createElement("div");
      g.className = "grain";
      document.body.appendChild(g);
    }
    try {
      if (!document.querySelector('meta[name="description"]')) {
        const d = await Store4bots.api.settings.get("site_description");
        if (d && d.value) {
          const m = document.createElement("meta");
          m.name = "description";
          m.content = d.value;
          document.head.appendChild(m);
        }
      }
    } catch (e) { /* public settings may 404 on live */ }
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", mount);
  else mount();
  global.PawChrome = { mount };
})(window);
