(function (global) {
  const root = () => global.PROTOTYPE_ROOT || "./";
  const here = global.PROTOTYPE_PAGE || "";

  function navLink(href, label, page, extra) {
    const on = here === page ? " is-on" : "";
    return `<a class="${on}" href="${root()}${href}">${label}${extra || ""}</a>`;
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
        <a class="brand" href="${root()}shop-mfe/index.html">
          <span class="brand-mark">S</span> STORE4BOTS
        </a>
        <form class="header-search" action="${root()}shop-mfe/index.html">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><path d="M20 20l-3-3"/></svg>
          <input name="search" placeholder="Search assets, kits, packs" value="${new URLSearchParams(location.search).get("search") || ""}" />
        </form>
        <button type="button" class="nav-toggle" aria-label="Menu" aria-expanded="false">☰</button>
        <nav class="header-nav">
          ${navLink("shop-mfe/index.html", "Shop", "shop")}
          ${navLink("product-mfe/bundles.html", "Bundles", "bundles")}
          ${navLink("community-mfe/index.html", "Community", "community")}
          ${navLink("account-mfe/cart.html", "Cart", "cart", count ? `<span class="badge">${count}</span>` : "")}
          ${user
            ? navLink("account-mfe/account.html", (user.name || "Account").split(" ")[0], "account") +
              `<button type="button" class="nav-auth" data-logout>${(global.PawUI && PawUI.icon) ? PawUI.icon("logout") : ""}<span>Log out</span></button>`
            : navLink("auth-mfe/login.html", "Log in", "auth")}
          ${(user && (user.role === "admin" || user.role === "staff"))
            ? navLink("admin-app/index.html", "Admin", "admin")
            : ""}
          <span class="wallet-dot" title="Wallet">${wallet ? "Wallet · " + PawUI.esc(wallet.slice(0, 8)) + "…" : "Wallet · none"}</span>
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
        try { await Store4bots.api.auth.logout(); } catch (e) {}
        location.href = root() + "shop-mfe/index.html";
      };
    }
    if (footer) {
      footer.className = "site-footer";
      footer.innerHTML = `
        <span>© Store4bots · buy once, download forever</span>
        <span class="footer-links">
          <a href="${root()}shop-mfe/index.html">Shop</a>
          <a href="${root()}shop-mfe/category.html">Categories</a>
          <a href="${root()}product-mfe/bundles.html">Bundles</a>
          <a href="${root()}product-mfe/request.html">Request</a>
          <a href="${root()}community-mfe/index.html">Community</a>
          <a href="${root()}community-mfe/people.html">People</a>
          <a href="${root()}account-mfe/wishlist.html">Wishlist</a>
          <a href="${root()}account-mfe/recent.html">Recent</a>
          <a href="${root()}account-mfe/referrals.html">Referrals</a>
          <a href="${root()}account-mfe/guest.html">Guest order</a>
          <a href="${root()}shop-mfe/legal.html">Terms</a>
          <a href="${root()}sitemap.xml">Sitemap</a>
          <button type="button" id="reset-demo">Reset demo data</button>
        </span>`;
      const btn = footer.querySelector("#reset-demo");
      if (btn) btn.onclick = () => Paw.resetDemo();
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
        compared.map((p) => `<a class="compare-bar-item" href="${root()}product-mfe/index.html?slug=${p.slug}">${PawUI.esc(p.title)}</a>`).join("") +
        `<a class="btn btn-accent compare-bar-open" href="${root()}account-mfe/compare.html">Open</a>`;
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
