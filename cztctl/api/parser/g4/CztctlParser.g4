// CztctlParser — 统一语法规则，覆盖 .cron 和 .rabbitmq 两种 DSL 文件。
//
// .cron  文件特有：@cron / @cronRetry 注解，路由为 TaskName[(Req)]
// .rabbitmq 文件特有：路由为 queue.name[(Msg)]，无 @cron / @cronRetry
//
// 二者共享：syntax / info / import / type / @server / @doc / @handler / service

grammar CztctlParser;

import CztctlLexer;

@lexer::members {
    const COMMENTS = 88
}

// ==================== Top-level ====================

api:            spec*;
spec:           syntaxLit
                | importSpec
                | infoSpec
                | typeSpec
                | serviceSpec
                ;

// ==================== syntax ====================

syntaxLit:      {match(p,"syntax")}syntaxToken=ID assign='=' version=STRING;

// ==================== import ====================

importSpec:     importLit | importBlock;
importLit:      {match(p,"import")}importToken=ID importValue;
importBlock:    {match(p,"import")}importToken=ID '(' importBlockValue+ ')';
importBlockValue: importValue;
importValue:    STRING;

// ==================== info ====================

infoSpec:       {match(p,"info")}infoToken=ID lp='(' kvLit+ rp=')';

// ==================== type ====================

typeSpec:       typeLit | typeBlock;
typeLit:        {match(p,"type")}typeToken=ID typeLitBody;
typeBlock:      {match(p,"type")}typeToken=ID lp='(' typeBlockBody* rp=')';
typeLitBody:    typeStruct | typeAlias;
typeBlockBody:  typeBlockStruct | typeBlockAlias;
typeStruct:     {checkKeyword(p)}structName=ID structToken=ID? lbrace='{' field* rbrace='}';
typeAlias:      {checkKeyword(p)}alias=ID assign='='? dataType;
typeBlockStruct:{checkKeyword(p)}structName=ID structToken=ID? lbrace='{' field* rbrace='}';
typeBlockAlias: {checkKeyword(p)}alias=ID assign='='? dataType;
field:          {isNormal(p)}? normalField | anonymousFiled;
normalField:    {checkKeyword(p)}fieldName=ID dataType tag=RAW_STRING?;
anonymousFiled: star='*'? ID;
dataType:       {isInterface(p)}ID
                | mapType
                | arrayType
                | inter='interface{}'
                | pointerType
                | typeStruct
                ;
pointerType:    star='*' {checkKeyword(p)}ID;
mapType:        {match(p,"map")}mapToken=ID lbrack='[' {checkKey(p)}key=ID rbrack=']' value=dataType;
arrayType:      lbrack='[' rbrack=']' dataType;

// ==================== service ====================

serviceSpec:    atServer? serviceApi;
atServer:       ATSERVER lp='(' kvLit+ rp=')';
serviceApi:     {match(p,"service")}serviceToken=ID serviceName lbrace='{' serviceRoute* rbrace='}';
serviceName:    (ID '-'?)+;

// ==================== service route ====================
// 统一规则：@doc? @cron? @cronRetry? @handler route
// @cron/@cronRetry 仅在 .cron 文件中有效，由语义层校验。

serviceRoute:   atDoc? atCron? atCronRetry? atHandler route;
atDoc:          ATDOC lp='('? ((kvLit+) | STRING) rp=')'?;
atHandler:      ATHANDLER ID;
atCron:         ATCRON STRING;
atCronRetry:    ATCRONRETRY INT;

// route：统一格式 routeName [(body)]
// .cron   — routeName = TaskName (单个 ID)
// .rabbitmq — routeName = queue.name (dot-separated IDs)
route:          routeName request=body?;
routeName:      ID ('.' ID)*;
body:           lp='(' ID? rp=')';

// ==================== kv ====================

kvLit:          key=ID ':' value=kvValue;
kvValue:        STRING | RAW_STRING | INT | ID ((',' | '-') ID)*;
