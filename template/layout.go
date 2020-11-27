package template

const html = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="theme-color" content="#f3f3f3">
  <meta name="msapplication-navbutton-color" content="#f3f3f3">
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="#f3f3f3">
  <title>Wayback Archiver</title>
  <link rel="icon" href="data:,">
  <style>
    :root {
      --c-light-text: #333;
      --c-light-textarea: #222;
      --c-light-background: #f7f7f7;
      --c-light-form-background: #fff;

      --c-dark-text: #fcfcfc;
      --c-dark-textarea: #b4c9d4;
      --c-dark-background: #222222ed;
      --c-dark-form-background: #3e3e3e;
    }

    html, body {
      margin: 0;
      padding: 0;
      height: 100%;
      min-height: 100%;
    }

    html {
      background-color: var(--c-light-background);
      -webkit-transition: color 300ms, background-color 300ms;
      -o-transition: color 300ms, background-color 300ms;
      transition: color 300ms, background-color 300ms;
    }

    @media (prefers-color-scheme: light) {
      html {
        background-color: var(--c-light-background);
        color: var(--c-light-text);
      }
    }

    @media (prefers-color-scheme: dark) {
      html {
        background-color: var(--c-dark-background);
        color: var(--c-dark-text);
      }
    }

    body {
      font: 100% / 1.5 "Open Sans", Helvetica, Arial, sans-serif;
      font-size: 1rem;
      -webkit-tap-highlight-color: transparent;
    }

    ::-webkit-scrollbar {
      background: transparent;
    }

    a {
      border: 1px blue;
      color: #999;
      width: 100%;
      float: left;
      text-decoration: none;
      text-overflow: ellipsis;
      overflow: hidden;
      -webkit-transition: background-color 0.15s ease 0s;
      -o-transition: background-color 0.15s ease 0s;
      transition: background-color 0.15s ease 0s;
    }

    a:-webkit-any-link {
      text-decoration: none;
      text-overflow: ellipsis;
      overflow: hidden;
      width: 100%;
      float: left;
    }

    a:hover {
      color: #524d4b;
    }

    ::-moz-selection {
      background-color: #cfcfcfb4;
    }

    ::selection {
      background-color: #cfcfcfb4;
    }

    @media (prefers-color-scheme: dark) {
      ::-moz-selection {
        background-color: #cfcfcf30;
      }

      ::selection {
        background-color: #cfcfcf30;
      }
    }

    .wrapper {
      position: absolute;
      margin: auto;
      width: 100%;
      max-width: 630px;
      bottom: 0;
      padding: 2rem 0;
      left: 0;
      right: 0;
      z-index: 12;
    }

    @media only screen and (max-width: 1024px) {
      .wrapper {
        padding: 1.5%;
        max-width: 100%;
      }
    }

    .form {
      border-radius: 4px;
      display: block;
      position: relative;
      background-color: var(--c-light-form-background);
      border: 1px solid #00000026;
      -webkit-box-shadow: 0 2px 3px #0000000f;
      box-shadow: 0 2px 3px #0000000f;
      padding-right: 1em;
      margin: 0;
      padding: 0;
    }

    @media (prefers-color-scheme: dark) {
      .form {
        background-color: var(--c-dark-form-background);
        color: var(--c-dark-text);
      }
    }

    textarea {
      visibility: visible;
      font-family: "Proxima Nova", "Helvetica Neue", "Helvetica", "Segoe UI", "Nimbus Sans L", "Liberation Sans", "Open Sans", FreeSans, Arial, sans-serif;
      color: var(--c-light-textarea);
      -webkit-appearance: none;
      -moz-appearance: none;
      appearance: none;
      font-size: 1.1em;
      font-weight: normal;
      display: block;
      background: none;
      outline: none;
      border: none;
      width: 100%;
      height: 8em;
      padding: .5em;
      z-index: 1;
      resize: none;
      text-align: justify;
      justify-content: center;
      box-sizing: border-box;
      -moz-box-sizing: border-box;
      -webkit-box-sizing: border-box;
      -webkit-transition: all .5s;
      -o-transition: all .5s;
      transition: all .5s;
      -webkit-user-select: none;
      -moz-user-select: none;
      -ms-user-select: none;
      user-select: none;
    }

    @media only screen and (max-width: 1024px) {
      textarea {
        height: 17em;
        font-size: 1.5em;
        -webkit-text-fill-color: initial;
        line-height: 30px;
        border-right-width: 24px;
        padding: 20px;
      }

      textarea:focus {
        outline: none !important;
        border-color: #6a98c9;
        -webkit-box-shadow: 0 0 10px #6a98c9;
        box-shadow: 0 0 10px #6a98c9;
      }

      textarea::-webkit-input-placeholder {
        font-size: 1.5em;
      }

      textarea::-moz-placeholder {
        font-size: 1.5em;
      }

      textarea:-ms-input-placeholder {
        font-size: 1.5em;
      }

      textarea::-ms-input-placeholder {
        font-size: 1.5em;
      }

      textarea::placeholder {
        font-size: 1.5em;
      }
    }

    @media (prefers-color-scheme: dark) {
      textarea {
        color: var(--c-dark-textarea);
      }
    }

    input[type="submit"] {
      visibility: visible;
      -webkit-appearance: none;
      -moz-appearance: none;
      appearance: none;
      font: normal normal normal 20px/1 FontAwesome;
      font-style: normal;
      font-weight: normal !important;
      font-variant: normal;
      text-rendering: auto;
      text-transform: none;
      text-decoration: none !important;
      cursor: pointer;
      background: transparent;
      text-align: center;
      border: none;
      position: absolute;
      bottom: .2em;
      right: .3em;
      left: auto;
      margin: auto;
      z-index: 2;
      outline: none;
      font-size: 1.25em;
      min-height: 1.8em;
      margin-top: -1px;
      margin-bottom: -1px;
      margin-right: -3px;
      line-height: 1.5;
      height: 1.8em;
      width: 1.8em;
      border-radius: 50%;
      display: inline-block;
      color: #0a0a08;
      background-color: #dee3e79c;
      background-position: 50% 50%;
      background-repeat: no-repeat;
      -webkit-font-smoothing: subpixel-antialiased;
      -webkit-transition: all .5s;
      -o-transition: all .5s;
      transition: all .5s;
      -webkit-user-select: none;
      -moz-user-select: none;
      -ms-user-select: none;
      user-select: none;
    }

    input[type="submit"]:hover {
      -webkit-transition-duration: 0.4s;
      -o-transition-duration: 0.4s;
      transition-duration: 0.4s;
      background-color: #dee3e7;
    }

    @media only screen and (max-width: 1024px) {
      input[type="submit"] {
        bottom: -1.6em;
        right: 50%;
        height: 4em;
        width: 4em;
        font-size: 1.45em;
        -webkit-transform: translate(2em);
        -ms-transform: translate(2em);
        transform: translate(2em);
      }
    }

    .archived {
      position: relative;
      overflow: hidden;
      margin: 0.5em;
      border-bottom: 1px solid #00000040;
    }

    @media only screen and (max-width: 1024px) {
      .archived {
        font-size: 2em;
        padding: 3.5px;
        bottom: .86rem;
      }
    }

    ul.row {
      width: 100%;
      padding: 0;
      display: -webkit-box;
      display: -webkit-flex;
      display: -moz-box;
      display: -ms-flexbox;
      display: flex;
      list-style: none;
      list-style-type: none;
      -webkit-box-flex: 0;
      -moz-box-flex: 0;
      flex-direction: row;
      margin: 10px 0;
      text-align: left;
      white-space: nowrap;
      -webkit-appearance: none;
      -moz-appearance: none;
      appearance: none;
      -webkit-transition: all .5s;
      -o-transition: all .5s;
      transition: all .5s;
      -webkit-user-select: none;
      -moz-user-select: none;
      -ms-user-select: none;
      user-select: none;
    }

    @media only screen and (max-width: 1024px) {
      ul.row {
        margin: 15px 0;
      }
    }

    .src {
      z-index: 1;
      margin: 0;
      text-align: left;
      width: 35%;
      margin-right: .85rem;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    @media (prefers-color-scheme: dark) {
      .src {
        color: #ccc;
      }
    }

    .dst {
      flex: auto;
      -webkit-box-flex: 0;
      -moz-box-flex: 0;
      -webkit-flex: 0 1 65%;
      -ms-flex: 0 1 65%;
      width: 65%;
      overflow: hidden;
      text-overflow: ellipsis;
      display: inline-block;
    }
  </style>
</head>
<body>
  <div class="wrapper">
    <div class="archived" id="archived">
      {{- range $i, $collect := . -}}
      <ul class="row">
        <li class="src">{{ $collect.Src }}</li>
        <li class="dst"><a href="{{ $collect.Dst }}" target="blank">{{ $collect.Dst }}</a></li>
      </ul>
      {{end}}
    </div>
    <div class="form">
      <form action="/w" method="post" onsubmit="post(); return false;">
        <textarea id="text" name="text" value="" autocapitalize="off" autocorrect="off" spellcheck="false"
          placeholder="e.g. https://www.eff.org https://www.wikipedia.org" autofocus></textarea>
        <input type="submit" value="W">
      </form>
    </div>
  </div>

  <script>
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

      document.querySelector('div.form').style.backgroundColor = '';
      document.querySelector('input[type=submit]').disabled = '';
      document.getElementById('text').disabled = '';
      document.getElementById('text').value = '';
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
          }
        }
      };
      params.append("text", text);
      params.append("data-type", "json");
      http.send(params);
    };
  </script>
</body>
</html>
`
