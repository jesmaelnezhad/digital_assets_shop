(function (global) {
  const root = () => "/";
  const here = global.SITE_PAGE || "";

  function navLink(href, label, page, extra) {
    const on = here === page ? " is-on" : "";
    return `<a class="${on}" href="${href}">${label}${extra || ""}</a>`;
  }

  async function mount() {
    const header = document.getElementById("site-header");
    const footer = document.getElementById("site-footer");
    let count = 0;
    let user = null;
    if (global.Pawradise && Pawradise.api) {
      try { user = await Pawradise.api.auth.getProfile(); } catch (e) { user = null; }
      try {
        const cart = await Pawradise.api.cart.get();
        count = (cart.items || []).reduce((n, i) => n + (i.quantity || 0), 0);
      } catch (e) { count = 0; }
    }
    let wallet = "";
    if (user && user.profile && user.profile.wallet_address) wallet = user.profile.wallet_address;
    if (header) {
      header.className = "site-header";
      header.innerHTML = `
        <a class="brand" href="/">
          <span class="brand-mark">P</span> PAWRADISE
        </a>
        <form class="header-search" action="/">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><path d="M20 20l-3-3"/></svg>
          <input name="search" placeholder="Search assets, kits, packs" value="${new URLSearchParams(location.search).get("search") || ""}" />
        </form>
        <nav class="header-nav">
          ${navLink("/", "Shop", "shop")}
          ${navLink("/product/bundles.html", "Bundles", "bundles")}
          ${navLink("/community", "Community", "community")}
          ${navLink("/cart", "Cart", "cart", count ? `<span class="badge">${count}</span>` : "")}
          ${user
            ? navLink("/account", (user.name || "Account").split(" ")[0], "account")
            : navLink("/login", "Log in", "auth")}
          ${navLink("/admin", "Admin", "admin")}
          <span class="wallet-dot" title="Wallet">${wallet ? "Wallet · " + PawUI.esc(wallet.slice(0, 8)) + "…" : "Wallet · none"}</span>
        </nav>
      `;
    }
    if (footer) {
      footer.className = "site-footer";
      footer.innerHTML = `
        <span>© Pawradise · buy once, download forever</span>
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
      const c = await Pawradise.api.compare.list();
      await PawUI.catalog.load();
      compared = (c.products || []).map((x) => PawUI.catalog.byId[x.product_id]).filter(Boolean);
    } catch (e) { compared = []; }
    if (compared.length) {
      if (!bar) { bar = document.createElement("div"); bar.id = "compare-bar"; document.body.appendChild(bar); }
      bar.className = "compare-bar";
      bar.innerHTML = `<span>Compare ${compared.length}/4</span>` +
        compared.map((p) => `<a href="/product/${p.slug}">${PawUI.esc(p.title)}</a>`).join("") +
        `<a class="btn btn-accent" href="/compare">Open</a>`;
    } else if (bar) bar.remove();
    if (!document.querySelector(".grain")) {
      const g = document.createElement("div");
      g.className = "grain";
      document.body.appendChild(g);
    }
    try {
      if (!document.querySelector('meta[name="description"]')) {
        const d = await Pawradise.api.settings.get("site_description");
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
