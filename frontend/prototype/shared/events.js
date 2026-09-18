// Store4bots frontend event SDK. Allowed names: product_view, checkout_click.
(function (global) {
  const KEY = "store4bots_sid";
  const ENV = global.__STORE4BOTS_ENV__ || { apiBase: "/api/v1" };
  const ALLOWED = { product_view: true, checkout_click: true };

  function sid() {
    try {
      let s = localStorage.getItem(KEY);
      if (!s) {
        s = (global.crypto && crypto.randomUUID) ? crypto.randomUUID() : "s" + Date.now() + Math.random().toString(16).slice(2);
        localStorage.setItem(KEY, s);
      }
      return s;
    } catch (e) {
      return "anon";
    }
  }

  function track(name, properties) {
    if (!ALLOWED[name]) return Promise.resolve();
    const body = JSON.stringify({
      name: name,
      session_id: sid(),
      path: (global.location && location.pathname) || "",
      occurred_at: new Date().toISOString(),
      properties: properties || {}
    });
    const url = (ENV.apiBase || "/api/v1") + "/events";
    try {
      if (global.navigator && navigator.sendBeacon) {
        const blob = new Blob([body], { type: "application/json" });
        if (navigator.sendBeacon(url, blob)) return Promise.resolve();
      }
    } catch (e) { /* fall through */ }
    return fetch(url, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: body,
      keepalive: true
    }).then(function (res) { return res.ok ? res.json().catch(function () { return {}; }) : null; }).catch(function () { return null; });
  }

  global.PawEvents = { track: track, sid: sid, names: Object.keys(ALLOWED) };
})(window);
