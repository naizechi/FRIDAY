import { spawnSync } from 'node:child_process'
import { existsSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

// Vitest suite lookup breaks on Windows when cwd uses a lowercase drive letter
// (c:\...) while module URLs use C:/ — getCurrentSuite() returns undefined and
// describe() crashes on .config. Normalize before spawning Vitest.
const webRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const cwd =
  process.platform === 'win32'
    ? webRoot.replace(/^([a-z]):/, (_, d) => `${d.toUpperCase()}:`)
    : webRoot

const vitestCli = path.join(cwd, 'node_modules', 'vitest', 'vitest.mjs')
if (!existsSync(vitestCli)) {
  console.error(`vitest CLI not found at ${vitestCli}`)
  process.exit(1)
}

const result = spawnSync(process.execPath, [vitestCli, ...process.argv.slice(2)], {
  cwd,
  stdio: 'inherit',
  env: process.env,
})

process.exit(result.status ?? 1)
