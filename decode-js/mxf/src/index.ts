import './wasm_exec';
import { readFile } from 'fs/promises';
import { join } from 'path';

const wasmPath = join(__dirname, 'app.wasm');
const go = new Go();

export class Mxf {
  instancePromise;
  instance: any;

  constructor() {
    this.instancePromise = (async () => {
      const appWasm = await readFile(wasmPath);
      return await WebAssembly.instantiate(appWasm, go.importObject);
    })();
  }

  get initialized() {
    return this.instancePromise.then((src) => {
      go.run(src.instance);
      this.instance = src.instance;
      return this;
    });
  }

  decode(getKey: (key: string) => string, encMsg: string): Promise<string> {
    if (!this.instance) throw Error('not initialized');
    return new Promise((resolve, reject) => {
      const context = {
        resolve,
        reject,
        getKey,
      };

      const mxfDecode = (global as any).MxfDecode;
      mxfDecode(encMsg, context);
    });
  }
}
