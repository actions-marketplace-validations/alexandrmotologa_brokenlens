const fs = require('fs');
const path = require('path');
const { Resvg } = require('B:/workgit/auditforge/web/node_modules/@resvg/resvg-js');

function buildLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <clipPath id="squircle-clip">
      <rect x="24" y="24" width="976" height="976" rx="220" />
    </clipPath>

    <linearGradient id="cyan-glow" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#00f5ff"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <linearGradient id="beak-amber" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#fbbf24"/>
      <stop offset="100%" stop-color="#d97706"/>
    </linearGradient>

    <linearGradient id="slate-chest" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#334155"/>
      <stop offset="100%" stop-color="#0f172a"/>
    </linearGradient>

    <filter id="subtle-shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="16" stdDeviation="20" flood-color="#000000" flood-opacity="0.14" />
    </filter>
  </defs>

  <!-- Luxury White Squircle Container -->
  <rect x="24" y="24" width="976" height="976" rx="220" fill="#ffffff" stroke="#e2e8f0" stroke-width="6" />

  <g clip-path="url(#squircle-clip)">
    <g transform="translate(512, 512)" filter="url(#subtle-shadow)">

      <!-- Massive Hexagonal Architectural Gateway Frame -->
      <polygon points="
        0,-390
        338,-195
        338,195
        0,390
        -338,195
        -338,-195
      " fill="none" stroke="#0f172a" stroke-width="40" stroke-linejoin="round" />

      <!-- Inner Precision Reticle (Dashed Cyan) -->
      <polygon points="
        0,-355
        307,-177
        307,177
        0,355
        -307,177
        -307,-177
      " fill="none" stroke="#00f5ff" stroke-width="4" opacity="0.4" stroke-dasharray="16, 12" />

      <!-- Corner Alignment Crosshairs -->
      <line x1="-338" y1="-195" x2="-300" y2="-195" stroke="#00f5ff" stroke-width="3" opacity="0.6"/>
      <line x1="338" y1="-195" x2="300" y2="-195" stroke="#00f5ff" stroke-width="3" opacity="0.6"/>
      <line x1="0" y1="-390" x2="0" y2="-350" stroke="#00f5ff" stroke-width="3" opacity="0.6"/>
      <line x1="0" y1="390" x2="0" y2="350" stroke="#00f5ff" stroke-width="3" opacity="0.6"/>

      <!-- 100% Watertight Solid Base Silhouette (The Peregrine Raptor) -->
      <path d="
        M 0 -285
        L 55 -280 L 115 -250 L 180 -180 L 235 -90 L 260 20 L 245 130 L 205 210 L 140 275 L 0 325
        L -140 275 L -205 210 L -245 130 L -260 20 L -235 -90 L -180 -180 L -115 -250 L -55 -280
        Z
      " fill="#090d16" />

      <!-- Outer Wing Flanges & Shoulders -->
      <polygon points="-260,20 -180,-90 -140,40 -205,210 -245,130" fill="#1e293b" />
      <polygon points="260,20 180,-90 140,40 205,210 245,130" fill="#0f172a" />

      <!-- Mid Mantle / Upper Back -->
      <polygon points="-180,-90 -140,40 0,90 0,-120" fill="#334155" />
      <polygon points="180,-90 140,40 0,90 0,-120" fill="#1e293b" />

      <!-- Lower Body & Abdomen Plates -->
      <polygon points="-140,40 -70,180 0,260 0,90" fill="#475569" />
      <polygon points="140,40 70,180 0,260 0,90" fill="#334155" />

      <!-- Tail Shield Keel -->
      <polygon points="0,90 -70,180 0,325 70,180" fill="#1e293b" />
      <polygon points="0,90 0,325 70,180" fill="#0f172a" />

      <!-- Head Crest Crown Facets -->
      <polygon points="0,-285 55,-280 40,-200 0,-190" fill="#64748b" />
      <polygon points="0,-285 -55,-280 -40,-200 0,-190" fill="#475569" />
      <polygon points="55,-280 115,-250 85,-180 40,-200" fill="#475569" />
      <polygon points="-55,-280 -115,-250 -85,-180 -40,-200" fill="#334155" />
      <polygon points="115,-250 180,-180 130,-120 85,-180" fill="#334155" />
      <polygon points="-115,-250 -180,-180 -130,-120 -85,-180" fill="#1e293b" />

      <!-- Brow Armor Plates (Supraorbital Ridge) -->
      <polygon points="0,-190 60,-175 110,-140 50,-130 0,-140" fill="#94a3b8" />
      <polygon points="0,-190 -60,-175 -110,-140 -50,-130 0,-140" fill="#64748b" />

      <!-- Forehead Center Facet -->
      <polygon points="0,-190 40,-130 0,-85 -40,-130" fill="#cbd5e1" />

      <!-- Cheek Malar Polygonal Facets (Classic Peregrine Markings) -->
      <polygon points="110,-140 145,-70 95,-30 50,-130" fill="#1e293b" />
      <polygon points="-110,-140 -145,-70 -95,-30 -50,-130" fill="#0f172a" />
      <polygon points="95,-30 130,50 65,30 35,-40" fill="#334155" />
      <polygon points="-95,-30 -130,50 -65,30 -35,-40" fill="#1e293b" />

      <!-- Muzzle / Nares Platform -->
      <polygon points="0,-85 35,-40 0,-20 -35,-40" fill="#475569" />

      <!-- Curved Predatory Beak -->
      <polygon points="0,-20 30,-25 22,55 0,85" fill="url(#beak-amber)" />
      <polygon points="0,-20 -30,-25 -22,55 0,85" fill="#b45309" />
      <!-- Sharp Beak Hook -->
      <polygon points="0,50 18,55 0,85 -18,55" fill="#78350f" />

      <!-- Predatory Eye Sockets (Deep Recessed) -->
      <polygon points="45,-130 100,-135 110,-105 55,-100" fill="#050811" />
      <polygon points="-45,-130 -100,-135 -110,-105 -55,-100" fill="#050811" />

      <!-- High-Tech Optical Lens Eyes (Electric Cyan Iris) -->
      <polygon points="52,-124 96,-128 102,-110 60,-105" fill="url(#cyan-glow)" />
      <polygon points="-52,-124 -96,-128 -102,-110 -60,-105" fill="url(#cyan-glow)" />

      <!-- Obsidian Slit Pupils -->
      <polygon points="72,-126 80,-126 77,-107 70,-107" fill="#090d16" />
      <polygon points="-72,-126 -80,-126 -77,-107 -70,-107" fill="#090d16" />

      <!-- Optical Specular Glint Reflection -->
      <circle cx="82" cy="-122" r="3" fill="#ffffff" />
      <circle cx="-66" cy="-122" r="3" fill="#ffffff" />

      <!-- Central Inspection Lens Crosshair on Chest (Optical Auditor Core) -->
      <circle cx="0" cy="180" r="36" fill="#0b0f19" stroke="#00f5ff" stroke-width="3" />
      <circle cx="0" cy="180" r="24" fill="none" stroke="#00f5ff" stroke-width="2" opacity="0.6" stroke-dasharray="6,4" />
      <circle cx="0" cy="180" r="8" fill="#00f5ff" />
      <!-- Crosshair reticles -->
      <line x1="-48" y1="180" x2="-38" y2="180" stroke="#00f5ff" stroke-width="2.5" />
      <line x1="48" y1="180" x2="38" y2="180" stroke="#00f5ff" stroke-width="2.5" />
      <line x1="0" y1="132" x2="0" y2="142" stroke="#00f5ff" stroke-width="2.5" />
      <line x1="0" y1="228" x2="0" y2="218" stroke="#00f5ff" stroke-width="2.5" />

    </g>
  </g>
</svg>`;
}

async function renderLogo() {
  const outputDir = path.resolve(__dirname, '../docs/images');
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const svg = buildLogoSvg();
  const svgPath = path.join(outputDir, 'logo.svg');
  const pngPath = path.join(outputDir, 'logo.png');

  fs.writeFileSync(svgPath, svg, 'utf8');
  console.log('✓ Written:', svgPath);

  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: 1024 }
  });
  const pngData = resvg.render().asPng();
  fs.writeFileSync(pngPath, pngData);
  console.log('✓ Rendered 1024x1024 PNG:', pngPath);
}

renderLogo().catch(err => {
  console.error('Logo render failed:', err);
  process.exit(1);
});
