import assert from 'node:assert/strict'
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import test from 'node:test'
import { svgBuilder } from './index.js'

test('injects SVG symbols without adding remote scripts', () => {
  const directory = `${mkdtempSync(join(tmpdir(), 'svg-builder-'))}/`
  try {
    writeFileSync(`${directory}sample.svg`, '<svg width="20" height="10"><path /></svg>')
    const plugin = svgBuilder(directory)
    const html = plugin.transformIndexHtml('<html><body></body></html>')

    assert.equal(plugin.name, 'svg-transform')
    assert.match(html, /<symbol id="sample" viewBox="0 0 20 10">/)
    assert.doesNotMatch(html, /gin-vue-admin|createElement|setTimeout/)
  } finally {
    rmSync(directory, { recursive: true })
  }
})
