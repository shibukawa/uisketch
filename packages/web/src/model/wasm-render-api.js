let wasmExecPromise = null;

export async function renderWithWasm(source) {
  await ensureWasmExec();
  const go = new globalThis.Go();
  const response = await fetch(`${import.meta.env.BASE_URL}uisketch.wasm`);
  const wasm = await instantiate(response, go.importObject);
  globalThis.uisketchWasmInput = source;
  delete globalThis.uisketchWasmOutput;
  await go.run(wasm.instance);
  const output = globalThis.uisketchWasmOutput;
  delete globalThis.uisketchWasmInput;
  delete globalThis.uisketchWasmOutput;
  if (!output) throw new Error("uisketch wasm renderer returned no output");
  return JSON.parse(output);
}

async function ensureWasmExec() {
  if (globalThis.Go) return;
  if (!wasmExecPromise) wasmExecPromise = loadScript(`${import.meta.env.BASE_URL}wasm_exec.js`);
  await wasmExecPromise;
}

async function instantiate(response, importObject) {
  try {
    return await WebAssembly.instantiateStreaming(Promise.resolve(response.clone()), importObject);
  } catch {
    return WebAssembly.instantiate(await response.arrayBuffer(), importObject);
  }
}

function loadScript(src) {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = src;
    script.onload = resolve;
    script.onerror = () => reject(new Error(`failed to load ${src}`));
    document.head.appendChild(script);
  });
}
