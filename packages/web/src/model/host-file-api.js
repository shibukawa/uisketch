import {
  createProjectFile,
  listProjectFiles,
  localProjectConfig,
  readProjectFile,
  renderLayoutSource,
  writeProjectFile,
} from "./local-project-api.js";

export function createHostFileApi() {
  const wailsHost = globalThis.uisketchHostFileApi || globalThis.go?.main?.DesktopApp;
  if (wailsHost) {
    return {
      config: () => ({
        enabled: true,
        initialFile: globalThis.uisketchInitialFile || "",
        host: "wails",
      }),
      listFiles: () => callHost(wailsHost, "listFiles", "ListFiles"),
      readFile: (path) => callHost(wailsHost, "readFile", "ReadFile", path),
      writeFile: ({ path, source, revision }) => callHost(wailsHost, "writeFile", "WriteFile", path, source, revision),
      createFile: ({ path, source }) => callHost(wailsHost, "createFile", "CreateFile", path, source),
      render: (source) => callHost(wailsHost, "renderSource", "RenderSource", source),
      setDirty: (dirty) => callHost(wailsHost, "setDirty", "SetDirty", dirty).catch(() => {}),
    };
  }
  return {
    config: () => ({ ...localProjectConfig(), host: "http" }),
    listFiles: listProjectFiles,
    readFile: readProjectFile,
    writeFile: writeProjectFile,
    createFile: createProjectFile,
    render: renderLayoutSource,
    setDirty: () => {},
  };
}

function callHost(host, lowerName, upperName, ...args) {
  const fn = host[lowerName] || host[upperName];
  if (!fn) return Promise.reject(new Error(`host file API missing ${upperName}`));
  return fn(...args);
}
