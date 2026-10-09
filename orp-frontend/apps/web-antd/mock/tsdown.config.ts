import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { defineConfig } from 'tsdown';

const configDir = dirname(fileURLToPath(import.meta.url));

/**
 * 将 SW 入口打包为无依赖 IIFE，产物输出到 mock/sw-dist/，
 * 随后由构建脚本移动为项目根 public/sw.js。
 * 运行：cd orp-frontend/apps/web-antd && pnpm exec tsdown --config mock/tsdown.config.ts
 */
export default defineConfig({
  dts: false,
  entry: [join(configDir, 'sw-entry.ts')],
  format: 'iife',
  minify: true,
  output: join(configDir, 'sw-dist'),
  platform: 'browser',
  target: 'es2020',
});
