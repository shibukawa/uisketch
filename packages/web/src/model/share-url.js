export function encodeSharePayload(source) {
  const json = JSON.stringify({ v: 1, source });
  return base64UrlEncode(utf8ToBinary(json));
}

export function decodeSharePayload(encoded) {
  const payload = JSON.parse(binaryToUtf8(base64UrlDecode(encoded)));
  if (payload?.v !== 1 || typeof payload.source !== "string") {
    throw new Error("Unsupported share payload");
  }
  return payload.source;
}

export function createShareUrl(currentHref, source) {
  const url = new URL(currentHref);
  url.searchParams.delete("mode");
  url.searchParams.delete("file");
  url.hash = `s=${encodeSharePayload(source)}`;
  return url.toString();
}

function base64UrlEncode(binary) {
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function base64UrlDecode(value) {
  const padded = value.replace(/-/g, "+").replace(/_/g, "/").padEnd(Math.ceil(value.length / 4) * 4, "=");
  return atob(padded);
}

function utf8ToBinary(value) {
  return unescape(encodeURIComponent(value));
}

function binaryToUtf8(value) {
  return decodeURIComponent(escape(value));
}
