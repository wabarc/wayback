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
  document.querySelector('input[type=submit]').disabled = '';
  document.getElementById('text').disabled = '';
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
    html += '<li class="src">' + collect.src + '</li>';
    html += ' <li class="dst">';
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
  unblock();
};

var post = function () {
  "use strict";
  var http = new XMLHttpRequest(),
    params = new URLSearchParams(),
    text = document.getElementById('text').value,
    url = "/w";
  if (!text || text.length === 0) {
    return;
  }

  document.getElementById('text').disabled = 'true';
  document.querySelector('input[type=submit]').disabled = 'true';
  document.querySelector('div.form').style.backgroundColor = '#5454541f';

  http.open("POST", url, true);
  http.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  http.onreadystatechange = function () {
    if (http.readyState === 4 && http.status === 200) {
      if (http.response !== undefined && http.response) {
        var collects = JSON.parse(http.response)
        render(collects);
      } else {
        unblock();
      }
    } else {
      unblock();
    }
  };
  params.append("text", text);
  params.append("data-type", "json");
  http.send(params);
};
