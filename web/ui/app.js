/* driplane web UI */
(function () {
  'use strict';

  // The token arrives once in the query string and then lives in sessionStorage
  // so it does not stay visible in the address bar.
  var params = new URLSearchParams(window.location.search);
  if (params.get('token')) {
    sessionStorage.setItem('driplane-token', params.get('token'));
    history.replaceState(null, '', window.location.pathname);
  }
  var token = sessionStorage.getItem('driplane-token') || '';

  var state = { kind: null, name: null, mtime: null, dirty: false, meta: { filters: [], feeders: [], kinds: [] } };

  function api(method, path, body) {
    var opts = {
      method: method,
      headers: { 'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json' }
    };
    if (body !== undefined) {
      opts.body = JSON.stringify(body);
    }
    return fetch(path, opts).then(function (res) {
      if (res.status === 204) { return null; }
      return res.json().then(function (data) {
        if (!res.ok) { throw new Error(data.error || ('HTTP ' + res.status)); }
        return data;
      });
    });
  }

  var editor = CodeMirror.fromTextArea(document.getElementById('code'), {
    theme: 'material-darker',
    lineNumbers: true,
    mode: 'driplane-rule',
    extraKeys: {
      'Ctrl-S': function () { save(); },
      'Cmd-S': function () { save(); },
      'Ctrl-Space': 'autocomplete'
    },
    hintOptions: { hint: hints }
  });

  function hints(cm) {
    var cur = cm.getCursor();
    var line = cm.getLine(cur.line).slice(0, cur.ch);
    var word = (line.match(/[a-zA-Z_\d-]*$/) || [''])[0];
    var list = state.meta.filters;
    if (/<\s*[a-zA-Z_\d-]*$/.test(line)) { list = state.meta.feeders; }
    list = list.filter(function (n) { return n.indexOf(word) === 0; });
    return {
      list: list,
      from: CodeMirror.Pos(cur.line, cur.ch - word.length),
      to: cur
    };
  }

  function diagnostics(message, ok) {
    var el = document.getElementById('diagnostics');
    el.textContent = message || '';
    el.className = ok ? 'ok' : '';
  }

  function markDirty(dirty) {
    state.dirty = dirty;
    var active = document.querySelector('#tree li.active');
    if (active) { active.classList.toggle('dirty', dirty); }
  }

  // kindLabel returns the display name for a file kind. "js" gets special
  // casing because it is an initialism, not a word to title-case.
  function kindLabel(kind) {
    if (kind === 'js') { return 'JS'; }
    return kind.charAt(0).toUpperCase() + kind.slice(1);
  }

  // buildTreeSections (re)creates one <section> per configured kind under
  // #tree, replacing whatever was there before. Called once meta.kinds is
  // known, and whenever it could have changed. Kind names come from the
  // server's own configuration, not attacker input, but every node is still
  // built with createElement/textContent rather than innerHTML, matching the
  // rest of this file.
  function buildTreeSections(kinds) {
    var nav = document.getElementById('tree');
    var sections = (kinds || []).map(function (kind) {
      var section = document.createElement('section');
      section.dataset.kind = kind;

      var h2 = document.createElement('h2');
      h2.appendChild(document.createTextNode(kindLabel(kind) + ' '));

      var newBtn = document.createElement('button');
      newBtn.className = 'new';
      newBtn.type = 'button';
      newBtn.dataset.kind = kind;
      newBtn.setAttribute('aria-label', 'New ' + kindLabel(kind).toLowerCase());
      newBtn.textContent = '+';
      newBtn.onclick = function (event) {
        event.stopPropagation();
        create(kind);
      };
      h2.appendChild(newBtn);

      section.appendChild(h2);
      section.appendChild(document.createElement('ul'));
      return section;
    });
    nav.replaceChildren.apply(nav, sections);
  }

  function loadTree() {
    (state.meta.kinds || []).forEach(function (kind) {
      api('GET', '/api/files?kind=' + kind).then(function (files) {
        var ul = document.querySelector('nav section[data-kind="' + kind + '"] ul');
        if (!ul) { return; }
        // Rebuild the list from scratch. Nothing here is ever assigned as
        // HTML: file names are attacker-controlled (an operator can name a
        // file anything), so every node is created with createElement and
        // populated with textContent, never with innerHTML built from a
        // string.
        var items = (files || []).map(function (f) {
          var li = document.createElement('li');
          if (state.kind === kind && state.name === f.name) { li.classList.add('active'); }

          var label = document.createElement('span');
          label.className = 'name';
          label.textContent = f.name;
          label.onclick = function () { open(kind, f.name); };
          li.appendChild(label);

          var del = document.createElement('button');
          del.className = 'delete';
          del.type = 'button';
          del.textContent = '×';
          del.setAttribute('aria-label', 'Delete ' + f.name);
          del.onclick = function (event) {
            event.stopPropagation();
            remove(kind, f.name);
          };
          li.appendChild(del);

          return li;
        });
        ul.replaceChildren.apply(ul, items);
      }).catch(function (err) { diagnostics(err.message, false); });
    });
  }

  function remove(kind, name) {
    if (!window.confirm('Delete ' + name + '? This cannot be undone.')) { return; }
    api('DELETE', '/api/files/' + kind + '/' + encodeURIComponent(name)).then(function () {
      if (state.kind === kind && state.name === name) {
        // The open buffer points at a file that no longer exists: clear it
        // rather than leaving a phantom, unsavable editor state.
        state.kind = null;
        state.name = null;
        state.mtime = null;
        clearTimeout(validateTimer);
        // setValue first, matching open()'s order: it fires the 'change'
        // handler, which calls markDirty(true), so markDirty(false) must run
        // after it to be the one that actually sticks -- otherwise the next
        // file the operator opens raises a spurious "Discard the unsaved
        // changes?" prompt from a dirty flag no one can see.
        editor.setValue('');
        markDirty(false);
        document.getElementById('current').textContent = 'no file open';
        diagnostics('', true);
      }
      loadTree();
    }).catch(function (err) { diagnostics(err.message, false); });
  }

  function modeFor(kind) {
    if (kind === 'js') { return 'javascript'; }
    if (kind === 'rules') { return 'driplane-rule'; }
    return 'null';
  }

  function open(kind, name) {
    if (state.dirty && !window.confirm('Discard the unsaved changes?')) { return; }
    // Cancel any debounced validate scheduled for the file we are leaving:
    // otherwise it fires up to 500ms later, reads the *new* file's content
    // out of the (shared) editor, and paints a bogus result over whatever
    // this open() is about to display.
    clearTimeout(validateTimer);
    api('GET', '/api/files/' + kind + '/' + encodeURIComponent(name)).then(function (file) {
      state.kind = kind;
      state.name = name;
      state.mtime = file.mtime;
      editor.setOption('mode', modeFor(kind));
      editor.setValue(file.content);
      markDirty(false);
      document.getElementById('current').textContent = kind + ' / ' + name;
      diagnostics('', true);
      loadTree();
    }).catch(function (err) { diagnostics(err.message, false); });
  }

  function save() {
    if (!state.name) { diagnostics('open a file to save it', false); return; }
    api('PUT', '/api/files/' + state.kind + '/' + encodeURIComponent(state.name), {
      content: editor.getValue(),
      mtime: state.mtime
    }).then(function (res) {
      state.mtime = res.mtime;
      markDirty(false);
      diagnostics('saved', true);
      loadTree();
    }).catch(function (err) {
      diagnostics(err.message + ' — reopen the file to load the version on disk', false);
    });
  }

  function create(kind) {
    var name = window.prompt('New file name');
    if (!name) { return; }
    api('POST', '/api/files/' + kind, { name: name, content: '' }).then(function () {
      loadTree();
      open(kind, name);
    }).catch(function (err) { diagnostics(err.message, false); });
  }

  var validateTimer = null;
  editor.on('change', function () {
    markDirty(true);
    if (state.kind !== 'rules') { return; }
    // Capture which file this validate is for. open() also cancels this
    // timer on navigation, but this re-check is the guard that actually
    // matters: it covers the window between the 500ms debounce firing and
    // that fire being processed, and it stays correct even if open() is
    // ever called through a path that forgets to clear the timer.
    var name = state.name;
    clearTimeout(validateTimer);
    validateTimer = setTimeout(function () {
      if (state.kind !== 'rules' || state.name !== name) { return; }
      api('POST', '/api/validate', { kind: 'rules', source: editor.getValue() })
        .then(function (res) {
          if (state.kind !== 'rules' || state.name !== name) { return; }
          diagnostics(res.ok ? 'syntax OK' : res.error, res.ok);
        })
        .catch(function (err) {
          if (state.kind !== 'rules' || state.name !== name) { return; }
          diagnostics(err.message, false);
        });
    }, 500);
  });

  // Builds one <tr> for the rules table. rule.name and rule.file come from
  // the API (they are operator-controlled filenames/identifiers) so every
  // cell is populated with textContent through createElement, never with
  // innerHTML built from a template string.
  function ruleRow(rule) {
    var totals = (rule.nodes || []).reduce(function (acc, n) {
      return { in: acc.in + n.in, matched: acc.matched + n.matched, errors: acc.errors + n.errors };
    }, { in: 0, matched: 0, errors: 0 });

    var tr = document.createElement('tr');
    tr.title = rule.file + '\nin / matched / errors';

    var dot = document.createElement('td');
    dot.textContent = rule.running ? '●' : '○';
    tr.appendChild(dot);

    var name = document.createElement('td');
    name.textContent = rule.name;
    tr.appendChild(name);

    [totals.in, totals.matched, totals.errors].forEach(function (n) {
      var td = document.createElement('td');
      td.className = 'num';
      td.textContent = String(n);
      tr.appendChild(td);
    });

    return tr;
  }

  // Consecutive /api/status failures. Polled every 3s; two misses in a row
  // (6s) flips the state pill to "unreachable" instead of freezing forever
  // on the last good snapshot, which otherwise looks identical to "the
  // daemon is idle and nothing changed".
  var statusFailures = 0;

  function refreshStatus() {
    api('GET', '/api/status').then(function (status) {
      statusFailures = 0;
      var el = document.getElementById('state');
      el.textContent = status.state + (status.error ? ': ' + status.error : '');
      el.className = 'state ' + status.state;

      var tbody = document.querySelector('#rules tbody');
      var rows = (status.rules || []).map(ruleRow);
      tbody.replaceChildren.apply(tbody, rows);
    }).catch(function () {
      statusFailures += 1;
      if (statusFailures >= 2) {
        var el = document.getElementById('state');
        el.textContent = 'unreachable';
        el.className = 'state unreachable';
      }
      // else: transient, the next poll retries silently.
    });
  }

  function streamLogs() {
    var box = document.getElementById('logs');
    var source = new EventSource('/api/logs?token=' + encodeURIComponent(token));
    source.onmessage = function (event) {
      var line = JSON.parse(event.data);
      var div = document.createElement('div');
      div.className = line.level.toLowerCase();
      div.textContent = line.message;
      box.appendChild(div);
      while (box.childNodes.length > 500) { box.removeChild(box.firstChild); }
      box.scrollTop = box.scrollHeight;
    };
    source.onerror = function () {
      source.close();
      setTimeout(streamLogs, 3000);
    };
  }

  // /api/reload, /api/start and /api/stop only request the action: the
  // daemon rebuilds asynchronously in its own loop, so a 200 ok:true here
  // means "accepted", never "succeeded". Whether it actually worked shows up
  // in the state pill (top of the page) within about 3 seconds, once the
  // next status poll lands -- so this reports the request outcome and
  // points the operator there instead of implying success it cannot know.
  function action(path) {
    api('POST', path, {}).then(function (res) {
      diagnostics(res.ok ? path + ' requested — see the state indicator above' : res.error, res.ok);
      refreshStatus();
    }).catch(function (err) { diagnostics(err.message, false); });
  }

  document.getElementById('btn-save').onclick = save;
  document.getElementById('btn-reload').onclick = function () { action('/api/reload'); };
  document.getElementById('btn-start').onclick = function () { action('/api/start'); };
  document.getElementById('btn-stop').onclick = function () { action('/api/stop'); };
  document.getElementById('btn-test').onclick = function () {
    if (state.kind !== 'rules') { diagnostics('open a rule to test it', false); return; }
    diagnostics('running the dry-run...', true);
    api('POST', '/api/test', { name: state.name, source: editor.getValue() }).then(function (res) {
      diagnostics(res.output || res.error || 'no output', res.ok);
    }).catch(function (err) { diagnostics(err.message, false); });
  };

  // The file-tree sections depend on which kinds the server has configured
  // (/api/meta's "kinds"), so they cannot be built until that response
  // arrives; loadTree (which lists files into those sections) is chained
  // after it for the same reason.
  api('GET', '/api/meta').then(function (meta) {
    state.meta = meta;
    buildTreeSections(meta.kinds || []);
    loadTree();
  }).catch(function (err) { diagnostics(err.message, false); });
  refreshStatus();
  setInterval(refreshStatus, 3000);
  streamLogs();
})();
