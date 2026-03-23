lexer grammar CztctlLexer;

// ==================== Annotation Keywords ====================
ATDOC:              '@doc';
ATHANDLER:          '@handler';
ATSERVER:           '@server';
ATCRON:             '@cron';
ATCRONRETRY:        '@cronRetry';

// ==================== Special Tokens ====================
INTERFACE:          'interface{}';

// ==================== Whitespace & Comments ====================
WS:                 [ \t\r\n\u000C]+ -> channel(HIDDEN);
COMMENT:            '/*' .*? '*/' -> channel(88);
LINE_COMMENT:       '//' ~[\r\n]* -> channel(88);

// ==================== Literals ====================
STRING:             '"' (~["\\] | EscapeSequence)* '"';
RAW_STRING:         '`' (~[`])* '`';
INT:                [0-9]+;

// ==================== Identifiers ====================
ID:                 Letter LetterOrDigit*;

// ==================== Fragments ====================
fragment LetterOrDigit
    : Letter
    | [0-9]
    ;

fragment EscapeSequence
    : '\\' [btnfr"'\\]
    | '\\' ([0-3]? [0-7])? [0-7]
    | '\\' 'u'+ HexDigit HexDigit HexDigit HexDigit
    ;

fragment HexDigit
    : [0-9a-fA-F]
    ;

fragment Letter
    : [a-zA-Z$_]
    | ~[\u0000-\u007F\uD800-\uDBFF]
    | [\uD800-\uDBFF] [\uDC00-\uDFFF]
    ;
