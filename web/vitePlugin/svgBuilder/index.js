import { readFileSync, readdirSync } from 'node:fs'

const readSvgFiles = (directories) => {
  return directories.flatMap((directory) => {
    return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
      const filePath = `${directory}${entry.name}`
      if (entry.isDirectory()) return readSvgFiles([`${filePath}/`])
      if (!entry.name.endsWith('.svg')) return []

      const pluginPrefix = directory.startsWith('./src/plugin/')
        ? `${directory.split('/')[3]}-`
        : ''
      const id = `${pluginPrefix}${entry.name.slice(0, -4)}`
      const source = readFileSync(filePath, 'utf8').replace(/[\r\n]/g, '')
      return source
        .replace(/<svg([^>]*)>/, (_, attributes) => {
          const width = attributes.match(/width="([^"]+)"/)?.[1] ?? 0
          const height = attributes.match(/height="([^"]+)"/)?.[1] ?? 0
          const cleanAttributes = attributes.replace(/\s*(width|height)="[^"]*"/g, '')
          const viewBox = /viewBox="[^"]+"/.test(attributes)
            ? ''
            : ` viewBox="0 0 ${width} ${height}"`
          return `<symbol id="${id}"${cleanAttributes}${viewBox}>`
        })
        .replace('</svg>', '</symbol>')
    })
  })
}

export const svgBuilder = (paths) => {
  const directories = typeof paths === 'string' ? [paths] : paths
  const symbols = readSvgFiles(directories)

  return {
    name: 'svg-transform',
    transformIndexHtml(html) {
      return html.replace(
        '<body>',
        `<body>\n  <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0">${symbols.join('')}</svg>`
      )
    }
  }
}
