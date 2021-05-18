// Incrementing OFFLINE_VERSION will kick off the install event and force
// previously cached resources to be updated from the network.
const OFFLINE_VERSION = 1;
const CACHE_NAME = "static-cache";

self.addEventListener("install", (event) => {
  const urlsToCache = [
    '/',
    '/index.js',
    '/service-worker.js',
    '/offline.html',
    '/favicon.ico',
    '/icon/favicon-16.png',
    '/icon/favicon-32.png',
    '/icon/icon-128.png',
    '/icon/icon-192.png',
    '/icon/icon-120.png',
    '/icon/icon-152.png',
    '/icon/icon-167.png',
    '/icon/icon-180.png'
  ];
  if (typeof window !== "undefined" && "caches" in window) {
    event.waitUntil(
      (async () => {
        caches.open(CACHE_NAME).then(function (cache) {
          cache.addAll(urlsToCache);
        });
      })()
    );
  }

  // Force the waiting service worker to become the active service worker.
  self.skipWaiting();
});

self.addEventListener('activate', (e) => {
  if (typeof window !== "undefined" && "caches" in window) {
    e.waitUntil(caches.keys().then((keyList) => {
      Promise.all(keyList.map((key) => {
        if (key === CACHE_NAME) { return; }
        caches.delete(key);
      }))
    })());
  }
});

self.addEventListener("fetch", (event) => {
  if (event.request.url !== '') {
    try {
      const url = new URL(event.request.url)
      if (url.pathname.match('^.*(\/w|\/healthcheck|\/version|\/metrics)$')) {
        return false
      }
    } catch (_) { }
  }
  if (typeof window !== "undefined" && "caches" in window) {
    event.respondWith(
      (async () => {
        const cache = await caches.open(CACHE_NAME);
        const cachedResponse = await cache.match(event.request);
        return cachedResponse || fetchAndCache(event.request);
      })()
    );
  }
});

function fetchAndCache(url) {
  return fetch(url)
    .then(function (response) {
      // Check if we received a valid response
      if (!response.ok) {
        throw Error(response.statusText);
      }
      return caches.open(CACHE_NAME)
        .then(function (cache) {
          if (!(url.indexOf('http') === 0)) {
            return response;
          }
          cache.put(url, response.clone());
          return response;
        });
    })
    .catch(function (error) {
      return caches.match('/offline.html');
    });
}
