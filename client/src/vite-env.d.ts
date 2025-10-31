/// <reference types="vite/client" />
/// <reference types="vitest/globals" />

// globalオブジェクトの型定義をテスト環境で利用可能にする
declare global {
  var global: typeof globalThis;
}
