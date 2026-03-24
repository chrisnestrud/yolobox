import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import sharp from 'sharp'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const docsDir = path.resolve(__dirname, '..')
const brandDir = path.join(docsDir, 'brand')
const publicDir = path.join(docsDir, 'public')

const logoReferencePath = path.join(brandDir, 'logo-dark-reference.png')

const SOCIAL_CARD = {
  width: 1200,
  height: 630,
  radius: 32,
}

function pngDimensions(buffer) {
  return {
    width: buffer.readUInt32BE(16),
    height: buffer.readUInt32BE(20),
  }
}

function imageSvg(buffer, description) {
  const { width, height } = pngDimensions(buffer)
  const encoded = buffer.toString('base64')

  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" role="img" aria-labelledby="title desc">
  <title id="title">yolobox logo</title>
  <desc id="desc">${description}</desc>
  <image width="${width}" height="${height}" href="data:image/png;base64,${encoded}"/>
</svg>`
}

async function generateLightLogo(darkLogoBuffer) {
  const { data, info } = await sharp(darkLogoBuffer)
    .ensureAlpha()
    .raw()
    .toBuffer({ resolveWithObject: true })

  for (let i = 0; i < data.length; i += 4) {
    const r = data[i]
    const g = data[i + 1]
    const b = data[i + 2]

    if (r > 220 && g > 90 && b < 40) {
      data[i] = 255
      data[i + 1] = 122
      data[i + 2] = 10
      continue
    }

    if (r > 110 && g > 35 && b < 30) {
      data[i] = 43
      data[i + 1] = 26
      data[i + 2] = 14
      continue
    }

    data[i] = 255
    data[i + 1] = 247
    data[i + 2] = 236
  }

  return sharp(data, { raw: info }).png().toBuffer()
}

function socialCardBackgroundSvg() {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${SOCIAL_CARD.width}" height="${SOCIAL_CARD.height}" viewBox="0 0 ${SOCIAL_CARD.width} ${SOCIAL_CARD.height}">
  <defs>
    <linearGradient id="card-bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="#121217"/>
      <stop offset="100%" stop-color="#08080A"/>
    </linearGradient>
    <radialGradient id="card-glow" cx="30%" cy="24%" r="62%">
      <stop offset="0%" stop-color="#FF8C00" stop-opacity="0.16"/>
      <stop offset="100%" stop-color="#FF8C00" stop-opacity="0"/>
    </radialGradient>
    <pattern id="grid" width="36" height="36" patternUnits="userSpaceOnUse">
      <path d="M36 0H0V36" fill="none" stroke="white" stroke-opacity="0.05"/>
    </pattern>
  </defs>
  <rect width="${SOCIAL_CARD.width}" height="${SOCIAL_CARD.height}" rx="${SOCIAL_CARD.radius}" fill="url(#card-bg)"/>
  <rect width="${SOCIAL_CARD.width}" height="${SOCIAL_CARD.height}" rx="${SOCIAL_CARD.radius}" fill="url(#grid)"/>
  <ellipse cx="420" cy="176" rx="520" ry="220" fill="url(#card-glow)"/>
  <rect x="24" y="24" width="${SOCIAL_CARD.width - 48}" height="${SOCIAL_CARD.height - 48}" rx="24" fill="none" stroke="white" stroke-opacity="0.08"/>
  <text x="96" y="402" fill="white" font-family="Arial, Helvetica, sans-serif" font-size="46" font-weight="800">Run AI coding agents in a sandboxed container.</text>
  <text x="96" y="462" fill="#BCBFC9" font-family="Arial, Helvetica, sans-serif" font-size="28" font-weight="500">Your home directory stays home.</text>
  <rect x="96" y="510" width="428" height="56" rx="14" fill="#14141A" stroke="white" stroke-opacity="0.09"/>
  <text x="122" y="546" fill="#FFB66E" font-family="'Courier New', monospace" font-size="21" font-weight="700">brew install finbarr/tap/yolobox</text>
  <text x="1108" y="546" text-anchor="end" fill="#8F939C" font-family="Arial, Helvetica, sans-serif" font-size="24" font-weight="700">yolobox.dev</text>
</svg>`
}

async function generateSocialCard(darkLogoBuffer) {
  const background = await sharp(Buffer.from(socialCardBackgroundSvg())).png().toBuffer()
  const logoMeta = await sharp(darkLogoBuffer).metadata()
  const targetWidth = 996
  const resizedLogo = await sharp(darkLogoBuffer)
    .resize({ width: targetWidth })
    .png()
    .toBuffer()

  const left = Math.round((SOCIAL_CARD.width - targetWidth) / 2)
  const top = 56
  const logoHeight = Math.round((logoMeta.height / logoMeta.width) * targetWidth)

  return sharp(background)
    .composite([
      {
        input: resizedLogo,
        left,
        top,
      },
    ])
    .png()
    .toBuffer()
}

async function writeFile(name, contents) {
  await fs.writeFile(path.join(publicDir, name), contents)
}

async function main() {
  await fs.mkdir(publicDir, { recursive: true })

  const darkLogoBuffer = await fs.readFile(logoReferencePath)
  const lightLogoBuffer = await generateLightLogo(darkLogoBuffer)
  const socialCardBuffer = await generateSocialCard(darkLogoBuffer)

  await writeFile('logo-dark.png', darkLogoBuffer)
  await writeFile('logo-dark.svg', imageSvg(darkLogoBuffer, 'The yolobox logo on a black background.'))
  await writeFile('logo-light.png', lightLogoBuffer)
  await writeFile('logo-light.svg', imageSvg(lightLogoBuffer, 'The yolobox logo adapted for light mode.'))
  await writeFile('social-card.png', socialCardBuffer)
  await writeFile('social-card.svg', imageSvg(socialCardBuffer, 'The yolobox social card with logo, tagline, and install command.'))
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
