// Syntax mode for the driplane rule DSL. It mirrors the lexer defined in
// core/rule_parser.go.
CodeMirror.defineMode('driplane-rule', function () {
  return {
    startState: function () {
      return {};
    },
    token: function (stream) {
      if (stream.sol() && stream.match(/^#import\s+"[^"]*"\s*$/)) {
        return 'meta';
      }
      if (stream.sol() && stream.peek() === '#') {
        stream.skipToEnd();
        return 'comment';
      }
      if (stream.eatSpace()) {
        return null;
      }
      // Highlighting limitation, not a grammar one: the real lexer
      // (core/rule_parser.go's ruleLexer, and its [^"] string class) allows
      // a quoted string to span multiple lines. CodeMirror tokenizes one
      // line at a time, so a string left open at end-of-line is not, and
      // cannot be, carried over and highlighted as a string on the next
      // line -- it will parse correctly, it just will not look highlighted.
      if (stream.match(/"(?:\\.|[^"\\])*"/) || stream.match(/'(?:\\.|[^'\\])*'/)) {
        return 'string';
      }
      if (stream.match(/\d+(?:\.\d+)?/)) {
        return 'number';
      }
      if (stream.match(/@[a-zA-Z][a-zA-Z_\d-]*/)) {
        return 'variable-2';
      }
      if (stream.match(/=>/)) {
        return 'operator';
      }
      if (stream.match(/[a-zA-Z][a-zA-Z_\d-]*(?=\s*\()/)) {
        return 'keyword';
      }
      if (stream.match(/[a-zA-Z][a-zA-Z_\d-]*(?=\s*[:>])/)) {
        return 'tag';
      }
      if (stream.match(/[a-zA-Z][a-zA-Z_\d-]*(?=\s*=[^>])/)) {
        return 'attribute';
      }
      if (stream.match(/[|,;:=<>()!]/)) {
        return 'operator';
      }
      if (stream.match(/[a-zA-Z][a-zA-Z_\d-]*/)) {
        return 'def';
      }
      stream.next();
      return null;
    }
  };
});
