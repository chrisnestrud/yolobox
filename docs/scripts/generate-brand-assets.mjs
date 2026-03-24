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
  padding: 72,
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
      data[i] = 28
      data[i + 1] = 19
      data[i + 2] = 12
      data[i + 3] = 255
      continue
    }

    if (r > 110 && g > 35 && b < 30) {
      data[i] = 232
      data[i + 1] = 139
      data[i + 2] = 45
      data[i + 3] = 255
      continue
    }

    data[i] = 255
    data[i + 1] = 255
    data[i + 2] = 255
    data[i + 3] = 0
  }

  return sharp(data, { raw: info }).png().toBuffer()
}

async function trimLogo(buffer) {
  return sharp(buffer)
    .ensureAlpha()
    .trim()
    .png()
    .toBuffer()
}

function socialCardBackgroundSvg() {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${SOCIAL_CARD.width}" height="${SOCIAL_CARD.height}" viewBox="0 0 ${SOCIAL_CARD.width} ${SOCIAL_CARD.height}">
  <rect width="${SOCIAL_CARD.width}" height="${SOCIAL_CARD.height}" fill="#000000"/>
</svg>`
}

async function generateSocialCard(darkLogoBuffer) {
  const background = await sharp(Buffer.from(socialCardBackgroundSvg())).png().toBuffer()
  const maxWidth = SOCIAL_CARD.width - (SOCIAL_CARD.padding * 2)
  const maxHeight = SOCIAL_CARD.height - (SOCIAL_CARD.padding * 2)
  const resizedLogo = await sharp(darkLogoBuffer)
    .resize({
      width: maxWidth,
      height: maxHeight,
      fit: 'inside',
    })
    .png()
    .toBuffer()
  const logoMeta = await sharp(resizedLogo).metadata()

  const left = Math.round((SOCIAL_CARD.width - logoMeta.width) / 2)
  const top = Math.round((SOCIAL_CARD.height - logoMeta.height) / 2)

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

  const darkLogoSourceBuffer = await fs.readFile(logoReferencePath)
  const darkLogoBuffer = await trimLogo(darkLogoSourceBuffer)
  const lightLogoBuffer = await trimLogo(await generateLightLogo(darkLogoSourceBuffer))
  const socialCardBuffer = await generateSocialCard(darkLogoBuffer)

  await writeFile('logo-dark.png', darkLogoBuffer)
  await writeFile('logo-dark.svg', imageSvg(darkLogoBuffer, 'The yolobox logo on a black background.'))
  await writeFile('logo-light.png', lightLogoBuffer)
  await writeFile('logo-light.svg', imageSvg(lightLogoBuffer, 'The yolobox logo adapted for light mode.'))
  await writeFile('social-card.png', socialCardBuffer)
  await writeFile('social-card.svg', imageSvg(socialCardBuffer, 'The yolobox social card with the dark logo centered on a black background.'))
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
