const nodeCrypto = require('crypto');

if ((globalThis as any).crypto) {
  delete (globalThis as any).crypto;
}

Object.defineProperty(globalThis, 'crypto', {
  value: {
    getRandomValues: (arr: Uint8Array) => nodeCrypto.randomBytes(arr.length),
  },
});
