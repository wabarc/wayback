if ("serviceWorker" in navigator) {
  navigator.serviceWorker.register('/service-worker.js');
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

var unblock = function (collects) {
  document.querySelector('div.form').style.backgroundColor = '';
  document.querySelector('#wayback').disabled = false;
  document.querySelector('#playback').disabled = false;
  document.getElementById('text').disabled = false;
  document.getElementById('text').value = '';
};

var render = function (collects) {
  "use strict";
  if (typeof collects !== "object") {
    return;
  }
  var archived = document.getElementById('archived');
  var html = '';
  collects.forEach(function (collect, i) {
    html += '<ul class="row">';
    html += '<li class="src" title="' + collect.src + '">' + collect.src + '</li>';
    html += ' <li class="dst" title="' + collect.dst + '">';
    try {
      new URL(collect.dst);
      html += '<a href="' + collect.dst + '" target="blank">' + collect.dst + '</a>';
    } catch (_) {
      html += '<a href="javascript:;">' + collect.dst + '</a>';
    }
    html += '</li>';
    html += '</ul>';
  })
  archived.innerHTML = html + archived.innerHTML;
};

var post = function (url) {
  "use strict";
  var http = new XMLHttpRequest(),
    params = new URLSearchParams(),
    text = document.getElementById('text').value;
  if (!text || text.length === 0) {
    return;
  }

  document.getElementById('text').disabled = true;
  document.querySelector('#wayback').disabled = true;
  document.querySelector('#playback').disabled = true;
  document.querySelector('div.form').style.backgroundColor = '#5454541f';

  http.open("POST", url, true);
  http.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  http.onreadystatechange = function () {
    if (http.readyState === 4 && http.status === 200) {
      if (http.response !== undefined && http.response) {
        var collects = JSON.parse(http.response)
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

