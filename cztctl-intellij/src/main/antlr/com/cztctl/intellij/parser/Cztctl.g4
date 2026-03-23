// Cztctl — 统一语法规则，覆盖 .cron 和 .rabbitmq 两种 DSL 文件。
// Combined grammar（IntelliJ 插件用），去除 Go 语义谓词，关键字提升为 lexer token。

grammar Cztctl;

// ==================== Top-level ====================

api:            spec* EOF;
spec:           syntaxLit
                | importSpec
                | infoSpec
                | typeSpec
                | serviceSpec
                ;

// ==================== syntax ====================

syntaxLit:      SYNTAX ASSIGN STRING;

// ==================== import ====================

importSpec:     importLit | importBlock;
importLit:      IMPORT importValue;
importBlock:    IMPORT LPAREN importBlockValue+ RPAREN;
importBlockValue: importValue;
importValue:    STRING;

// ==================== info ====================

infoSpec:       INFO LPAREN kvLit+ RPAREN;

// ==================== type ====================

typeSpec:       typeLit | typeBlock;
typeLit:        TYPE typeLitBody;
typeBlock:      TYPE LPAREN typeBlockBody* RPAREN;
typeLitBody:    typeStruct | typeAlias;
typeBlockBody:  typeBlockStruct | typeBlockAlias;
typeStruct:     structName=identifier STRUCT? LBRACE field* RBRACE;
typeAlias:      alias=identifier ASSIGN? dataType;
typeBlockStruct: structName=identifier STRUCT? LBRACE field* RBRACE;
typeBlockAlias: alias=identifier ASSIGN? dataType;
field:          normalField | anonymousField;
normalField:    fieldName=identifier dataType tag=RAW_STRING?;
anonymousField: STAR? identifier;
dataType:       identifier
                | mapType
                | arrayType
                | INTERFACE
                | pointerType
                | typeStruct
                ;
pointerType:    STAR identifier;
mapType:        MAP LBRACK key=identifier RBRACK value=dataType;
arrayType:      LBRACK RBRACK dataType;

// ==================== service ====================

serviceSpec:    atServer? serviceApi;
atServer:       ATSERVER LPAREN kvLit+ RPAREN;
serviceApi:     SERVICE serviceName LBRACE serviceRoute* RBRACE;
serviceName:    (identifier DASH?)+;

// ==================== service route ====================

serviceRoute:   atDoc? atCron? atCronRetry? atHandler route;
atDoc:          ATDOC LPAREN? ((kvLit+) | STRING) RPAREN?;
atHandler:      ATHANDLER identifier;
atCron:         ATCRON STRING;
atCronRetry:    ATCRONRETRY INT;

// ==================== route ====================

route:          routeName request=body?;
routeName:      identifier (DOT identifier)*;
body:           LPAREN identifier? RPAREN;

// ==================== kv ====================

kvLit:          key=identifier COLON value=kvValue;
kvValue:        STRING | RAW_STRING | INT | identifier ((COMMA | DASH) identifier)*;

// ==================== identifier ====================

identifier:     ID | SYNTAX | IMPORT | INFO | TYPE | SERVICE | MAP | STRUCT;

// ====================================================================
// ==================== LEXER RULES ====================================
// ====================================================================

// ==================== Annotation Keywords ====================
ATDOC:              '@doc';
ATHANDLER:          '@handler';
ATSERVER:           '@server';
ATCRON:             '@cron';
ATCRONRETRY:        '@cronRetry';

// ==================== Language Keywords ====================
SYNTAX:             'syntax';
IMPORT:             'import';
INFO:               'info';
TYPE:               'type';
SERVICE:            'service';
MAP:                'map';
STRUCT:             'struct';

// ==================== Special Tokens ====================
INTERFACE:          'interface{}';

// ==================== Symbols ====================
LPAREN:             '(';
RPAREN:             ')';
LBRACE:             '{';
RBRACE:             '}';
LBRACK:             '[';
RBRACK:             ']';
ASSIGN:             '=';
COLON:              ':';
COMMA:              ',';
DOT:                '.';
STAR:               '*';
DASH:               '-';

// ==================== Whitespace & Comments ====================
WS:                 [ \t\r\n\u000C]+ -> channel(HIDDEN);
COMMENT:            '/*' .*? '*/' -> channel(HIDDEN);
LINE_COMMENT:       '//' ~[\r\n]* -> channel(HIDDEN);

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
