export function localProjectConfig() {
  const params = new URLSearchParams(globalThis.location?.search || "");
  return {
    enabled: params.get("mode") === "local",
    initialFile: params.get("file") || "",
  };
}

export async function listProjectFiles() {
  return request("/api/files");
}

export async function readProjectFile(path) {
  return request(`/api/file?path=${encodeURIComponent(path)}`);
}

export async function writeProjectFile({ path, source, revision }) {
  return request("/api/file", {
    method: "PUT",
    body: JSON.stringify({ path, source, revision }),
  });
}

export async function createProjectFile({ path, source }) {
  return request("/api/file", {
    method: "POST",
    body: JSON.stringify({ path, source }),
  });
}

async function request(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      "content-type": "application/json",
      ...(options.headers || {}),
    },
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.error || `${response.status} ${response.statusText}`);
  return payload;
}
