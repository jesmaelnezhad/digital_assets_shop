// Pure coupon calculation logic — mirrors server-side rules
// Server is always authority; this is for live UI preview only

export function applyCoupon(subtotal, coupon) {
    if (!coupon || !coupon.type || !coupon.value) return subtotal;
    
    let discount = 0;
    if (coupon.type === 'percentage') {
        discount = subtotal * (coupon.value / 100);
    } else if (coupon.type === 'fixed') {
        discount = coupon.value;
    }
    
    // Never go negative
    return Math.max(0, subtotal - discount);
}

export function calculateBundleSavings(bundlePrice, itemPrices) {
    const total = itemPrices.reduce((sum, price) => sum + price, 0);
    return Math.max(0, total - bundlePrice);
}

export function formatPrice(amount, currency = 'USD') {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency,
        minimumFractionDigits: 2,
    }).format(amount);
}

export function formatCrypto(amount, chain = 'BSC') {
    const symbols = { BSC: 'BNB', ETH: 'ETH', BTC: 'BTC', MATIC: 'MATIC' };
    return `${amount.toFixed(6)} ${symbols[chain] || chain}`;
}

export function calculateCryptoTotal(usdAmount, rateToUsd) {
    if (!rateToUsd || rateToUsd <= 0) return 0;
    return usdAmount / rateToUsd;
}
