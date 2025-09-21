let policy;
if (typeof window.trustedTypes !== "undefined") {
  try {
    policy = window.trustedTypes.createPolicy('tt-policy', {
      createScriptURL: (url) => url
    });
  } catch (e) {
    console.warn('Failed to create Trusted Types policy: ', e);
    policy = null;
  }
} else {
  policy = {
    createScriptURL: (url) => url
  };
}

if ("serviceWorker" in navigator) {
  const scriptURL = policy.createScriptURL('/service-worker.js');

  try {
    navigator.serviceWorker.register(scriptURL).catch(error => {
      console.error('Service Worker registration failed: ', error);
      if (error.name === 'SecurityError') {
        console.warn('Service Worker registration failed due to security restrictions');
      }
    });
  } catch (error) {
    console.error('Failed to register Service Worker: ', error);
  }
}

window.addEventListener('beforeinstallprompt', (e) => {
  // Prevent Chrome 67 and earlier from automatically showing the prompt
  e.preventDefault();

  let deferredPrompt = e;
  const addBtn = document.querySelector('.add-button');

  if (addBtn) {
    // Stash the event so it can be triggered later.
    deferredPrompt = e;
    // Update UI to notify the user they can add to home screen
    addBtn.style.display = 'block';

    addBtn.addEventListener('click', () => {
      e.preventDefault();
      addBtn.style.display = 'none';
      deferredPrompt.prompt();
      // Wait for the user to respond to the prompt
      deferredPrompt.userChoice.then(() => {
        deferredPrompt = null;
      });
    });
  }
});

const unblock = () => {
  document.querySelector('div.form').style.backgroundColor = '';
  document.querySelector('#wayback').disabled = false;
  document.querySelector('#playback').disabled = false;
  document.getElementById('text').disabled = false;
  document.getElementById('text').value = '';
};

const render = (collects) => {
  "use strict";
  if (!collects || typeof collects !== 'object') {
    return;
  }
  const archived = document.getElementById('archived');
  const fragment = document.createDocumentFragment();

  // Generate a light gray color with slight variation
  const r = Math.floor(240 + Math.random() * 16);  // 240-255
  const g = Math.floor(240 + Math.random() * 16);  // 240-255
  const b = Math.floor(240 + Math.random() * 16);  // 240-255
  const renderColor = `rgb(${r},${g},${b})`;

  collects.forEach((collect) => {
    const row = document.createElement('ul');
    row.className = 'row';
    row.style.backgroundColor = renderColor;
    row.style.color = '#333';

    const src = document.createElement('li');
    src.className = 'src';
    src.title = collect.src;
    src.textContent = collect.src;

    const dst = document.createElement('li');
    dst.className = 'dst';
    dst.title = collect.dst;

    let link;
    try {
      const url = new URL(collect.dst);
      link = document.createElement('a');
      link.href = url.href;
      link.target = 'blank';
      link.textContent = collect.dst;
    } catch (_) {
      link = document.createElement('a');
      link.href = 'javascript:;';
      link.textContent = collect.dst;
    }

    dst.appendChild(link);
    row.appendChild(src);
    row.appendChild(dst);
    fragment.appendChild(row);
  });
  archived.insertBefore(fragment, archived.firstChild);
};

const post = (url) => {
  "use strict";
  const http = new XMLHttpRequest(),
    params = new URLSearchParams(),
    text = document.getElementById('text').value;
  if (!text || text.length === 0) {
    return;
  }
  http.onload = () => {
    http.onreadystatechange = null;
  };

  document.getElementById('text').disabled = true;
  document.querySelector('#wayback').disabled = true;
  document.querySelector('#playback').disabled = true;
  document.querySelector('div.form').style.backgroundColor = '#5454541f';

  http.open("POST", url, true);
  http.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  http.onreadystatechange = () => {
    if (http.readyState === 4 && http.status === 200) {
      if (http.response !== undefined && http.response) {
        const collects = JSON.parse(http.response)
        render(collects);
      }
    }
    unblock();
  };
  params.append("text", text);
  params.append("data-type", "json");
  http.send(params);
}

window.addEventListener('submit', (e) => {
  // Prevent Chrome 67 and earlier from automatically showing the prompt
  e.preventDefault();

  const wayback = document.getElementById('wayback');
  const playback = document.getElementById('playback');
  switch (e.submitter) {
    default:
    case wayback: {
      post("/wayback");
      break
    };
    case playback: {
      post("/playback");
      break
    };
  }
});

