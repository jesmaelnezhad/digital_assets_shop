#!/usr/bin/env python3
"""Copy prototype HTML into live MFE src with staging URL paths and shared assets."""
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]
PROTO = ROOT / "frontend/prototype"

HEAD_SCRIPTS = '''  <link rel="icon" href="/assets/favicon.svg" sizes="any" />
  <link rel="stylesheet" href="/assets/theme.css" />
  <script src="/env.js"></script>
'''
TAIL_SCRIPTS = '''  <script src="/assets/api.js"></script>
  <script src="/assets/ui.js"></script>
  <script src="/assets/chrome.js"></script>
'''

PATHS = [
    ("../shared/store.js", None),
    ("../shared/mock-api.js", None),
    ("../shared/api-client.js", None),
    ("../shared/ui.js", None),
    ("../shared/chrome.js", None),
    ("../shared/theme.css", None),
    ("shop-mfe/index.html", "/"),
    ("shop-mfe/category.html", "/category"),
    ("shop-mfe/legal.html", "/legal"),
    ("product-mfe/index.html?slug=", "/product/"),
    ("product-mfe/index.html", "/product/"),
    ("product-mfe/bundles.html", "/product/bundles.html"),
    ("product-mfe/bundle.html", "/product/bundle.html"),
    ("product-mfe/request.html", "/product/request.html"),
    ("community-mfe/index.html", "/community"),
    ("community-mfe/people.html", "/people"),
    ("community-mfe/post.html?id=", "/post/"),
    ("community-mfe/post.html", "/post"),
    ("community-mfe/profile.html?id=", "/profile/"),
    ("community-mfe/profile.html", "/profile"),
    ("account-mfe/account.html", "/account"),
    ("account-mfe/cart.html", "/cart"),
    ("account-mfe/wishlist.html", "/wishlist"),
    ("account-mfe/compare.html", "/compare"),
    ("account-mfe/recent.html", "/recent"),
    ("account-mfe/referrals.html", "/referrals"),
    ("account-mfe/guest.html", "/guest"),
    ("account-mfe/order.html", "/account/order.html"),
    ("auth-mfe/login.html", "/login"),
    ("auth-mfe/register.html", "/register"),
    ("admin-app/index.html", "/admin"),
    ("checkout-mfe/index.html", "/checkout"),
    ("sitemap.xml", "/sitemap.xml"),
]


def transform(html: str, page: str) -> str:
    html = re.sub(
        r'<script>window\.PROTOTYPE_ROOT="[^"]*"; window\.PROTOTYPE_PAGE="([^"]+)"; window\.__PAWRADISE_ENV__=\{[^}]*\};</script>\s*',
        r'  <script>window.SITE_PAGE="\1";</script>\n',
        html,
    )
    html = re.sub(r'<link rel="stylesheet" href="../shared/theme.css" />\s*', HEAD_SCRIPTS, html)
    html = re.sub(r'\s*<script src="../shared/store.js"></script>', "", html)
    html = re.sub(r'\s*<script src="../shared/mock-api.js"></script>', "", html)
    html = re.sub(r'\s*<script src="../shared/api-client.js"></script>', "", html)
    html = re.sub(r'\s*<script src="../shared/ui.js"></script>', "", html)
    html = re.sub(r'\s*<script src="../shared/chrome.js"></script>', "", html)
    # inject shared scripts before first inline <script> after body content, or before </body>
    if "/assets/api.js" not in html:
        html = html.replace("<script>\n", TAIL_SCRIPTS + "  <script>\n", 1)
        if "/assets/api.js" not in html:
            html = html.replace("</body>", TAIL_SCRIPTS + "</body>")
    for old, new in PATHS:
        if new is None:
            continue
        html = html.replace(old, new)
    html = html.replace('window.PROTOTYPE_ROOT="../"; ', "")
    html = html.replace("envName:\"prototype\"", "envName:\"staging\"")
    if page == "product":
        html = html.replace(
            'const slug = new URLSearchParams(location.search).get("slug") || "lunar-clay-characters";',
            '''const pathSlug = (location.pathname.match(/^\\/product\\/([^/]+)/) || [])[1];
    const slug = new URLSearchParams(location.search).get("slug") || (pathSlug && !pathSlug.includes(".") ? pathSlug : "");''',
        )
    if page == "profile":
        html = html.replace(
            "const params = new URLSearchParams(location.search);",
            "const params = new URLSearchParams(location.search);\n    if (!params.get('id')) { const m = location.pathname.match(/\\/profile\\/(\\d+)/); if (m) params.set('id', m[1]); }",
        )
    if page == "post":
        html = html.replace(
            "const id = new URLSearchParams(location.search).get(\"id\");",
            "const id = new URLSearchParams(location.search).get(\"id\") || (location.pathname.match(/\\/post\\/(\\d+)/) || [])[1];",
        )
    return html


MAP = {
    "shop-mfe/index.html": [("shop-mfe/src/index.html", "shop"), ("shop-mfe/src/home.html", "shop")],
    "shop-mfe/category.html": [("shop-mfe/src/category.html", "shop")],
    "shop-mfe/legal.html": [("shop-mfe/src/legal.html", "shop")],
    "product-mfe/index.html": [("product-mfe/src/product.html", "product"), ("product-mfe/src/index.html", "product")],
    "product-mfe/bundle.html": [("product-mfe/src/bundle.html", "bundles")],
    "product-mfe/bundles.html": [("product-mfe/src/bundles.html", "bundles")],
    "product-mfe/request.html": [("product-mfe/src/request.html", "shop")],
    "community-mfe/index.html": [("community-mfe/src/index.html", "community"), ("community-mfe/src/community.html", "community")],
    "community-mfe/people.html": [("community-mfe/src/people.html", "community")],
    "community-mfe/post.html": [("community-mfe/src/post.html", "post")],
    "community-mfe/profile.html": [("community-mfe/src/profile.html", "profile")],
    "account-mfe/account.html": [("account-mfe/src/account.html", "account"), ("account-mfe/src/index.html", "account")],
    "account-mfe/cart.html": [("account-mfe/src/cart.html", "cart")],
    "account-mfe/wishlist.html": [("account-mfe/src/wishlist.html", "account")],
    "account-mfe/compare.html": [("account-mfe/src/compare.html", "account")],
    "account-mfe/recent.html": [("account-mfe/src/recent.html", "account")],
    "account-mfe/referrals.html": [("account-mfe/src/referrals.html", "account")],
    "account-mfe/guest.html": [("account-mfe/src/guest.html", "account")],
    "account-mfe/order.html": [("account-mfe/src/order.html", "account")],
    "auth-mfe/login.html": [("auth-mfe/src/login.html", "auth"), ("auth-mfe/src/index.html", "auth")],
    "auth-mfe/register.html": [("auth-mfe/src/register.html", "auth")],
    "checkout-mfe/index.html": [("checkout-mfe/src/index.html", "checkout"), ("checkout-mfe/src/checkout.html", "checkout")],
    "admin-app/index.html": [("admin-app/src/index.html", "admin"), ("admin-app/src/admin.html", "admin")],
}


def main():
    for src, dests in MAP.items():
        raw = (PROTO / src).read_text()
        page_hint = dests[0][1]
        out = transform(raw, page_hint)
        for dest, _ in dests:
            path = ROOT / "frontend" / dest
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(out)
            print("wrote", dest)
    sitemap = """<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>/</loc></url>
  <url><loc>/category</loc></url>
  <url><loc>/product/bundles.html</loc></url>
  <url><loc>/community</loc></url>
  <url><loc>/legal</loc></url>
</urlset>
"""
    (ROOT / "frontend/shop-mfe/src/sitemap.xml").write_text(sitemap)
    robots = "User-agent: *\nAllow: /\nSitemap: /sitemap.xml\n"
    (ROOT / "frontend/shop-mfe/src/robots.txt").write_text(robots)
    print("done")


if __name__ == "__main__":
    main()
