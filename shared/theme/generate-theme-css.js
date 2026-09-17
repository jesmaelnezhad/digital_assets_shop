#!/usr/bin/env node
// Generate theme.css from theme.tokens.json
const fs = require('fs');
const path = require('path');

const tokens = JSON.parse(fs.readFileSync(path.join(__dirname, 'theme.tokens.json'), 'utf8'));

let css = '/* Auto-generated from theme.tokens.json — do not edit */\n:root {\n';

// Colors
for (const [key, val] of Object.entries(tokens.color)) {
  css += `  --color-${key.replace(/([A-Z])/g, '-$1').toLowerCase()}: ${val};\n`;
}

// Fonts
css += `  --font-sans: ${tokens.font.sans};\n`;
css += `  --font-mono: ${tokens.font.mono};\n`;

// Radius
for (const [key, val] of Object.entries(tokens.radius)) {
  css += `  --radius-${key}: ${val};\n`;
}

// Spacing
for (const [key, val] of Object.entries(tokens.space)) {
  css += `  --space-${key}: ${val};\n`;
}

// Glow
for (const [key, val] of Object.entries(tokens.glow)) {
  css += `  --glow-${key}: ${val};\n`;
}

// Breakpoints (as comments since CSS custom properties can't be used in media queries)
css += '\n/* Breakpoints (use in media queries) */\n';
css += `/* --breakpoint-sm: ${tokens.breakpoint.sm}; */\n`;
css += `/* --breakpoint-md: ${tokens.breakpoint.md}; */\n`;
css += `/* --breakpoint-lg: ${tokens.breakpoint.lg}; */\n`;

css += '}\n';

fs.writeFileSync(path.join(__dirname, 'theme.css'), css);
console.log('Generated theme.css');
