const INDEX_KEY = "uisketch.saved.index.v1";
const DOCUMENT_PREFIX = "uisketch.saved.";
const DOCUMENT_SUFFIX = ".v1";
const DRAFT_KEY = "uisketch.draft.current.v1";

export function loadSavedIndex(storage = localStorage) {
  return readJSON(storage, INDEX_KEY, []);
}

export function loadSavedDocument(id, storage = localStorage) {
  return readJSON(storage, documentKey(id), null);
}

export function saveNamedDocument({ id = crypto.randomUUID(), name, source }, storage = localStorage) {
  const now = new Date().toISOString();
  const existing = loadSavedDocument(id, storage);
  const record = {
    id,
    name,
    source,
    version: 1,
    createdAt: existing?.createdAt || now,
    updatedAt: now,
  };
  storage.setItem(documentKey(id), JSON.stringify(record));
  const index = loadSavedIndex(storage).filter((item) => item.id !== id);
  index.unshift({ id, name, createdAt: record.createdAt, updatedAt: record.updatedAt });
  storage.setItem(INDEX_KEY, JSON.stringify(index));
  return record;
}

export function renameSavedDocument(id, name, storage = localStorage) {
  const record = loadSavedDocument(id, storage);
  if (!record) throw new Error("Saved document not found");
  return saveNamedDocument({ ...record, name }, storage);
}

export function deleteSavedDocument(id, storage = localStorage) {
  storage.removeItem(documentKey(id));
  storage.setItem(INDEX_KEY, JSON.stringify(loadSavedIndex(storage).filter((item) => item.id !== id)));
}

export function findSavedDocumentByName(name, storage = localStorage) {
  return loadSavedIndex(storage).find((item) => item.name === name) || null;
}

export function saveDraft(draft, storage = localStorage) {
  storage.setItem(DRAFT_KEY, JSON.stringify({ ...draft, version: 1, updatedAt: new Date().toISOString() }));
}

export function loadDraft(storage = localStorage) {
  return readJSON(storage, DRAFT_KEY, null);
}

export function clearDraft(storage = localStorage) {
  storage.removeItem(DRAFT_KEY);
}

function documentKey(id) {
  return `${DOCUMENT_PREFIX}${id}${DOCUMENT_SUFFIX}`;
}

function readJSON(storage, key, fallback) {
  try {
    const raw = storage.getItem(key);
    return raw ? JSON.parse(raw) : fallback;
  } catch {
    return fallback;
  }
}
