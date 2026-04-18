const API_BASE = '/api/v1';
const TOKEN_KEY = 'brandhunt_token';

let onErrorCallback = null;
let onUnauthorizedCallback = null;

export function setErrorHandler(fn) {
  onErrorCallback = fn;
}

export function setUnauthorizedHandler(fn) {
  onUnauthorizedCallback = fn;
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token) {
  if (token) localStorage.setItem(TOKEN_KEY, token);
  else localStorage.removeItem(TOKEN_KEY);
}

function buildUrl(path, params) {
  const url = new URL(API_BASE + path, window.location.origin);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null || value === '') continue;
      if (Array.isArray(value)) {
        value.forEach((v) => {
          if (v !== undefined && v !== null && v !== '') url.searchParams.append(key, v);
        });
      } else {
        url.searchParams.set(key, value);
      }
    }
  }
  return url.pathname + url.search;
}

async function request(method, path, { params, body, auth = false, silentError = false } = {}) {
  const headers = { Accept: 'application/json' };
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  if (auth) {
    const token = getToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }

  const url = buildUrl(path, params);
  let response;
  try {
    response = await fetch(url, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch (networkErr) {
    if (!silentError && onErrorCallback) onErrorCallback({ network: true });
    const err = new Error('Ошибка сети');
    err.network = true;
    throw err;
  }

  if (response.status === 204) return null;

  const contentType = response.headers.get('content-type') || '';
  const data = contentType.includes('application/json') ? await response.json() : null;

  if (!response.ok) {
    if (response.status === 401 && auth && onUnauthorizedCallback) {
      onUnauthorizedCallback();
    }
    if (!silentError && onErrorCallback) {
      onErrorCallback({ status: response.status, data });
    }
    const err = new Error((data && data.message) || `HTTP ${response.status}`);
    err.status = response.status;
    err.data = data;
    throw err;
  }

  return data;
}

export const api = {
  get: (path, opts) => request('GET', path, opts),
  post: (path, body, opts) => request('POST', path, { ...opts, body }),
  put: (path, body, opts) => request('PUT', path, { ...opts, body }),
  patch: (path, body, opts) => request('PATCH', path, { ...opts, body }),
  delete: (path, opts) => request('DELETE', path, opts),
};
