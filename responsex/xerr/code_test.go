package xerr

import (
	"errors"
	"fmt"
	"testing"
)

// ---------- Code.Msg ----------

func TestCode_Msg_KnownCode(t *testing.T) {
	tests := []struct {
		name     string
		code     Code
		expected string
	}{
		{"OK", OK, "SUCCESS"},
		{"ServerCommonError", ServerCommonError, "服务器开小差啦,稍后再来试一试"},
		{"RequestParamError", RequestParamError, "参数错误"},
		{"TokenExpireError", TokenExpireError, "token失效，请重新登陆"},
		{"TokenGenerateError", TokenGenerateError, "生成token失败"},
		{"DbError", DbError, "数据库繁忙,请稍后再试"},
		{"DbUpdateAffectedZeroError", DbUpdateAffectedZeroError, "更新数据影响行数为0"},
		{"MdCommonError", MdCommonError, "中间件错误"},
		{"PermitNoAccess", PermitNoAccess, "无权限操作"},
		{"SignParamError", SignParamError, "签名错误"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.Msg(); got != tt.expected {
				t.Errorf("Code(%d).Msg() = %q, want %q", tt.code, got, tt.expected)
			}
		})
	}
}

func TestCode_Msg_UnknownCode(t *testing.T) {
	code := Code(999999)
	got := code.Msg()
	want := messages[ServerCommonError]
	if got != want {
		t.Errorf("unknown Code.Msg() = %q, want fallback %q", got, want)
	}
}

// ---------- CodeError.Error ----------

func TestCodeError_Error(t *testing.T) {
	e := NewCodeError(ServerCommonError, "测试错误")
	got := e.Error()
	want := "ErrCode: 100001, ErrMsg: 测试错误"
	if got != want {
		t.Errorf("CodeError.Error() = %q, want %q", got, want)
	}
}

// ---------- CodeError.Is ----------

func TestCodeError_Is(t *testing.T) {
	err := NewCodeError(ServerCommonError, "test")

	if !errors.Is(err, &CodeError{}) {
		t.Error("errors.Is(CodeError, &CodeError{}) should be true")
	}
	if errors.Is(err, fmt.Errorf("some error")) {
		t.Error("errors.Is(CodeError, generic error) should be false")
	}
}

func TestCodeError_Is_Wrapped(t *testing.T) {
	inner := NewCodeError(DbError, "db fail")
	wrapped := fmt.Errorf("wrap: %w", inner)

	if !errors.Is(wrapped, &CodeError{}) {
		t.Error("errors.Is(wrapped CodeError, &CodeError{}) should be true")
	}
}

// ---------- NewCodeError ----------

func TestNewCodeError(t *testing.T) {
	e := NewCodeError(RequestParamError, "参数不合法")
	if e.code != RequestParamError {
		t.Errorf("code = %d, want %d", e.code, RequestParamError)
	}
	if e.message != "参数不合法" {
		t.Errorf("message = %q, want %q", e.message, "参数不合法")
	}
}

func TestNewCodeError_WithOptions(t *testing.T) {
	e := NewCodeError(ServerCommonError, "alert test",
		WithAlertLevel(AlertP0),
		WithAlertData("key1", "val1"),
	)
	if e.alertLevel != AlertP0 {
		t.Errorf("alertLevel = %q, want %q", e.alertLevel, AlertP0)
	}
	if v, ok := e.alertData["key1"]; !ok || v != "val1" {
		t.Errorf("alertData[key1] = %v, want %q", v, "val1")
	}
}

// ---------- NewErrCodeMsg ----------

func TestNewErrCodeMsg(t *testing.T) {
	e := NewErrCodeMsg(200001, "自定义错误")
	if e.code != Code(200001) {
		t.Errorf("code = %d, want 200001", e.code)
	}
	if e.message != "自定义错误" {
		t.Errorf("message = %q, want %q", e.message, "自定义错误")
	}
}

func TestNewErrCodeMsg_WithOptions(t *testing.T) {
	e := NewErrCodeMsg(200001, "custom", WithAlertLevel(AlertP2))
	if e.alertLevel != AlertP2 {
		t.Errorf("alertLevel = %q, want %q", e.alertLevel, AlertP2)
	}
}

// ---------- NewErrCode ----------

func TestNewErrCode_KnownCode(t *testing.T) {
	e := NewErrCode(uint32(DbError))
	if e.code != DbError {
		t.Errorf("code = %d, want %d", e.code, DbError)
	}
	if e.message != DbError.Msg() {
		t.Errorf("message = %q, want %q", e.message, DbError.Msg())
	}
}

func TestNewErrCode_UnknownCode(t *testing.T) {
	e := NewErrCode(999999)
	if e.code != Code(999999) {
		t.Errorf("code = %d, want 999999", e.code)
	}
	if e.message != messages[ServerCommonError] {
		t.Errorf("message = %q, want fallback %q", e.message, messages[ServerCommonError])
	}
}

// ---------- NewErrMsg ----------

func TestNewErrMsg(t *testing.T) {
	e := NewErrMsg("自定义消息")
	if e.code != ServerCommonError {
		t.Errorf("code = %d, want %d", e.code, ServerCommonError)
	}
	if e.message != "自定义消息" {
		t.Errorf("message = %q, want %q", e.message, "自定义消息")
	}
}

func TestNewErrMsg_WithOptions(t *testing.T) {
	e := NewErrMsg("msg", WithAlertLevel(AlertP1), WithAlertData("k", 42))
	if e.alertLevel != AlertP1 {
		t.Errorf("alertLevel = %q, want %q", e.alertLevel, AlertP1)
	}
	if v, ok := e.alertData["k"]; !ok || v != 42 {
		t.Errorf("alertData[k] = %v, want 42", v)
	}
}

// ---------- WithAlertLevel / WithAlertData ----------

func TestWithAlertLevel(t *testing.T) {
	for _, level := range []alertLevel{AlertP0, AlertP1, AlertP2, AlertP3} {
		e := NewCodeError(OK, "", WithAlertLevel(level))
		if e.alertLevel != level {
			t.Errorf("alertLevel = %q, want %q", e.alertLevel, level)
		}
	}
}

func TestWithAlertData_MultipleKeys(t *testing.T) {
	e := NewCodeError(OK, "",
		WithAlertData("a", 1),
		WithAlertData("b", "two"),
	)
	if len(e.alertData) != 2 {
		t.Fatalf("alertData len = %d, want 2", len(e.alertData))
	}
	if e.alertData["a"] != 1 {
		t.Errorf("alertData[a] = %v, want 1", e.alertData["a"])
	}
	if e.alertData["b"] != "two" {
		t.Errorf("alertData[b] = %v, want %q", e.alertData["b"], "two")
	}
}

func TestWithAlertData_NilMapInit(t *testing.T) {
	e := &CodeError{} // alertData 为 nil
	WithAlertData("key", "val")(e)
	if e.alertData == nil {
		t.Fatal("alertData should have been initialized")
	}
	if e.alertData["key"] != "val" {
		t.Errorf("alertData[key] = %v, want %q", e.alertData["key"], "val")
	}
}

// ---------- Wrapf ----------

func TestWrapf_NilError(t *testing.T) {
	if err := Wrapf(nil, "should be nil"); err != nil {
		t.Errorf("Wrapf(nil, ...) = %v, want nil", err)
	}
}

func TestWrapf_Simple(t *testing.T) {
	inner := errors.New("inner")
	wrapped := Wrapf(inner, "outer")
	want := "outer: inner"
	if wrapped.Error() != want {
		t.Errorf("Wrapf().Error() = %q, want %q", wrapped.Error(), want)
	}
}

func TestWrapf_WithFormatArgs(t *testing.T) {
	inner := errors.New("inner")
	wrapped := Wrapf(inner, "failed %s id=%d", "load", 42)
	want := "failed load id=42: inner"
	if wrapped.Error() != want {
		t.Errorf("Wrapf().Error() = %q, want %q", wrapped.Error(), want)
	}
}

func TestWrapf_Unwrap(t *testing.T) {
	inner := errors.New("base")
	wrapped := Wrapf(inner, "wrap")
	if !errors.Is(wrapped, inner) {
		t.Error("errors.Is should match inner error through unwrap")
	}
}

func TestWrapf_WithCodeError(t *testing.T) {
	inner := NewCodeError(DbError, "db fail")
	wrapped := Wrapf(inner, "service error")

	var ce *CodeError
	if !errors.As(wrapped, &ce) {
		t.Error("errors.As should find *CodeError in wrapped chain")
	}
	if ce.code != DbError {
		t.Errorf("unwrapped code = %d, want %d", ce.code, DbError)
	}
}

// ---------- errors.As 提取 CodeError ----------

func TestErrorsAs_DirectCodeError(t *testing.T) {
	err := NewCodeError(PermitNoAccess, "no access")
	var ce *CodeError
	if !errors.As(err, &ce) {
		t.Fatal("errors.As should find *CodeError")
	}
	if ce.code != PermitNoAccess {
		t.Errorf("code = %d, want %d", ce.code, PermitNoAccess)
	}
	if ce.message != "no access" {
		t.Errorf("message = %q, want %q", ce.message, "no access")
	}
}

func TestErrorsAs_WrappedCodeError(t *testing.T) {
	inner := NewCodeError(TokenExpireError, "expired")
	wrapped := fmt.Errorf("layer2: %w", inner)

	var ce *CodeError
	if !errors.As(wrapped, &ce) {
		t.Fatal("errors.As should find *CodeError through wrap")
	}
	if ce.code != TokenExpireError {
		t.Errorf("code = %d, want %d", ce.code, TokenExpireError)
	}
}

func TestErrorsAs_DoubleWrappedCodeError(t *testing.T) {
	inner := NewCodeError(SignParamError, "bad sign")
	layer1 := fmt.Errorf("layer1: %w", inner)
	layer2 := fmt.Errorf("layer2: %w", layer1)

	var ce *CodeError
	if !errors.As(layer2, &ce) {
		t.Fatal("errors.As should find *CodeError through double wrap")
	}
	if ce.code != SignParamError {
		t.Errorf("code = %d, want %d", ce.code, SignParamError)
	}
}

func TestErrorsAs_NoCodeError(t *testing.T) {
	err := errors.New("plain error")
	var ce *CodeError
	if errors.As(err, &ce) {
		t.Errorf("errors.As should not find *CodeError in plain error; ce=%v", ce)
	}
	if ce != nil {
		t.Errorf("ce = %v, want nil", ce)
	}
}

func TestErrorsAs_Nil(t *testing.T) {
	var ce *CodeError
	if errors.As(nil, &ce) {
		t.Error("errors.As(nil) should return false")
	}
}

// ---------- Wrapf 业务场景：替代 errors.Wrapf ----------
// 验证 xerr.Wrapf 可以替代 github.com/pkg/errors.Wrapf 来包装 CodeError
// 关键点：errors.As 仍能提取出 *CodeError，code/message/alertLevel 不丢失

func TestWrapf_Biz_NewErrCode(t *testing.T) {
	// 对应 Good: errors.Wrapf(xerr.NewErrCode(xerr.DbError), "user.QueryByPhone phone=%s err=%v", phone, err)
	originErr := errors.New("connection refused")
	phone := "13800138000"

	wrapped := Wrapf(NewErrCode(uint32(DbError)), "user.QueryByPhone phone=%s err=%v", phone, originErr)

	// errors.As 能提取出 CodeError
	var ce *CodeError
	if !errors.As(wrapped, &ce) {
		t.Fatal("errors.As should find *CodeError in xerr.Wrapf chain")
	}
	if ce.code != DbError {
		t.Errorf("code = %d, want %d", ce.code, DbError)
	}
	if ce.message != DbError.Msg() {
		t.Errorf("message = %q, want %q", ce.message, DbError.Msg())
	}

	// errors.As 也能穿透
	var ce2 *CodeError
	if !errors.As(wrapped, &ce2) {
		t.Error("errors.As should find *CodeError")
	}

	// errors.Is 匹配 CodeError 类型
	if !errors.Is(wrapped, &CodeError{}) {
		t.Error("errors.Is should match &CodeError{}")
	}

	// 格式化消息包含上下文
	msg := wrapped.Error()
	if !contains(msg, "user.QueryByPhone") || !contains(msg, "13800138000") {
		t.Errorf("Error() = %q, should contain context info", msg)
	}
}

func TestWrapf_Biz_NewErrMsg(t *testing.T) {
	// 对应 Good: errors.Wrapf(xerr.NewErrMsg("手机号已被注册"), "phone=%s", req.Phone)
	wrapped := Wrapf(NewErrMsg("手机号已被注册"), "phone=%s", "13900139000")

	var ce *CodeError
	if !errors.As(wrapped, &ce) {
		t.Fatal("errors.As should find *CodeError")
	}
	if ce.code != ServerCommonError {
		t.Errorf("code = %d, want %d (ServerCommonError)", ce.code, ServerCommonError)
	}
	if ce.message != "手机号已被注册" {
		t.Errorf("message = %q, want %q", ce.message, "手机号已被注册")
	}

	msg := wrapped.Error()
	if !contains(msg, "phone=13900139000") {
		t.Errorf("Error() = %q, should contain phone context", msg)
	}
}

func TestWrapf_Biz_NewErrCodeMsg(t *testing.T) {
	// 对应 Good: errors.Wrapf(xerr.NewErrCodeMsg(8888, "自定义业务错误"), "err: %v", err)
	originErr := errors.New("timeout")
	wrapped := Wrapf(NewErrCodeMsg(8888, "自定义业务错误"), "err: %v", originErr)

	var ce *CodeError
	if !errors.As(wrapped, &ce) {
		t.Fatal("errors.As should find *CodeError")
	}
	if ce.code != Code(8888) {
		t.Errorf("code = %d, want 8888", ce.code)
	}
	if ce.message != "自定义业务错误" {
		t.Errorf("message = %q, want %q", ce.message, "自定义业务错误")
	}

	msg := wrapped.Error()
	if !contains(msg, "timeout") {
		t.Errorf("Error() = %q, should contain origin error", msg)
	}
}

func TestWrapf_Biz_WithAlertLevel(t *testing.T) {
	// 带告警级别的 CodeError 经 Wrapf 包装后，errors.As 提取仍保留 alertLevel
	wrapped := Wrapf(
		NewErrCode(uint32(DbError), WithAlertLevel(AlertP1), WithAlertData("table", "users")),
		"query failed",
	)

	var ce *CodeError
	if !errors.As(wrapped, &ce) {
		t.Fatal("errors.As should find *CodeError")
	}
	if ce.alertLevel != AlertP1 {
		t.Errorf("alertLevel = %q, want %q", ce.alertLevel, AlertP1)
	}
	if ce.alertData["table"] != "users" {
		t.Errorf("alertData[table] = %v, want %q", ce.alertData["table"], "users")
	}
}

func TestWrapf_Biz_DoubleWrap_CauseFindsInnermost(t *testing.T) {
	// 多层 Wrapf 嵌套，errors.As 提取 CodeError
	inner := NewErrCode(uint32(DbError))
	layer1 := Wrapf(inner, "repo.FindOne")
	layer2 := Wrapf(layer1, "svc.GetUser")

	var ce *CodeError
	if !errors.As(layer2, &ce) {
		t.Fatal("errors.As should find *CodeError through double Wrapf")
	}
	if ce.code != DbError {
		t.Errorf("code = %d, want %d", ce.code, DbError)
	}
}

func TestWrapf_Biz_NilInnerReturnsNil(t *testing.T) {
	// Wrapf(nil, ...) 应返回 nil，不会产生空壳错误
	if err := Wrapf(nil, "something"); err != nil {
		t.Errorf("Wrapf(nil, ...) = %v, want nil", err)
	}
}

func TestWrapf_OutputDemo(t *testing.T) {
	// ✓ Good 场景 1
	err1 := Wrapf(NewErrCode(uint32(DbError)), "user.QueryByPhone phone=%s err=%v", "13800138000", errors.New("connection refused"))
	fmt.Println("=== Good 1: Wrapf(NewErrCode(DbError), ...) ===")
	fmt.Println("Error():", err1.Error())
	fmt.Println("PlusV output:")
	fmt.Printf("%+v\n", err1)
	var ce1 *CodeError
	_ = errors.As(err1, &ce1)
	fmt.Printf("errors.As: code=%d, msg=%q\n\n", ce1.code, ce1.message)

	// ✓ Good 场景 2
	err2 := Wrapf(NewErrMsg("手机号已被注册"), "phone=%s", "13900139000")
	fmt.Println("=== Good 2: Wrapf(NewErrMsg(...), ...) ===")
	fmt.Println("Error():", err2.Error())
	fmt.Println("PlusV output:")
	fmt.Printf("%+v\n", err2)
	var ce2 *CodeError
	_ = errors.As(err2, &ce2)
	fmt.Printf("errors.As: code=%d, msg=%q\n\n", ce2.code, ce2.message)

	// ✓ Good 场景 3
	err3 := Wrapf(NewErrCodeMsg(8888, "自定义业务错误"), "err: %v", errors.New("timeout"))
	fmt.Println("=== Good 3: Wrapf(NewErrCodeMsg(8888, ...), ...) ===")
	fmt.Println("Error():", err3.Error())
	fmt.Println("PlusV output:")
	fmt.Printf("%+v\n", err3)
	var ce3 *CodeError
	_ = errors.As(err3, &ce3)
	fmt.Printf("errors.As: code=%d, msg=%q\n\n", ce3.code, ce3.message)

	// ✗ Bad 对比
	errBad := errors.New("查询失败")
	fmt.Println("=== Bad: errors.New(\"查询失败\") ===")
	fmt.Println("Error():", errBad.Error())
	var ceBad *CodeError
	okBad := errors.As(errBad, &ceBad)
	fmt.Printf("errors.As: ce=%v, ok=%v\n\n", ceBad, okBad)

	// 带 AlertLevel
	err4 := Wrapf(NewErrCode(uint32(DbError), WithAlertLevel(AlertP1), WithAlertData("table", "users")), "query failed")
	fmt.Println("=== With AlertLevel ===")
	fmt.Println("Error():", err4.Error())
	var ce4 *CodeError
	_ = errors.As(err4, &ce4)
	fmt.Printf("errors.As: code=%d, alertLevel=%q, alertData=%v\n", ce4.code, ce4.alertLevel, ce4.alertData)
}

// ---------- Wrapf 堆栈能力验证 ----------
// xerr.Wrapf 通过 wrapError 捕获调用堆栈，可通过 fmt.Printf("%+v") 输出

func TestWrapf_HasStackTrace(t *testing.T) {
	wrapped := Wrapf(NewErrCode(uint32(DbError)), "query failed")

	out := fmt.Sprintf("%+v", wrapped)

	// %+v 输出应包含堆栈信息
	if !contains(out, "TestWrapf_HasStackTrace") {
		t.Errorf("%%+v output should contain stack frame, got %q", out)
	}
}

func TestWrapf_SprintfPlusV_PrintsStack(t *testing.T) {
	// 直接打印 err 时用 %+v 应包含堆栈信息
	wrapped := Wrapf(NewErrCode(uint32(DbError)), "query failed")

	out := fmt.Sprintf("%+v", wrapped)

	// 应包含错误消息
	if !contains(out, "query failed") {
		t.Errorf("%%+v output should contain 'query failed', got %q", out)
	}

	// 应包含堆栈函数名
	if !contains(out, "TestWrapf_SprintfPlusV_PrintsStack") {
		t.Errorf("%%+v output should contain stack frame function name, got %q", out)
	}

	// 应包含文件名
	if !contains(out, "code_test.go") {
		t.Errorf("%%+v output should contain source file, got %q", out)
	}

	// 普通 %v 不应包含堆栈
	normal := fmt.Sprintf("%v", wrapped)
	if contains(normal, "TestWrapf_SprintfPlusV") {
		t.Errorf("%%v output should NOT contain stack, got %q", normal)
	}
}

func TestWrapf_StackTraceContent(t *testing.T) {
	wrapped := Wrapf(NewErrCode(uint32(DbError)), "query failed")

	out := fmt.Sprintf("%+v", wrapped)

	// %+v 输出应包含源文件名
	if !contains(out, "code_test.go") {
		t.Errorf("%%+v output should contain code_test.go, got %q", out)
	}
}

func TestWrapf_DoubleWrap_StackTrace(t *testing.T) {
	// 多层 Wrapf，最外层有自己的堆栈
	inner := NewErrCode(uint32(DbError))
	layer1 := Wrapf(inner, "repo.FindOne")
	layer2 := Wrapf(layer1, "svc.GetUser")

	// 最外层 %+v 输出应包含当前测试函数
	out := fmt.Sprintf("%+v", layer2)
	if !contains(out, "TestWrapf_DoubleWrap_StackTrace") {
		t.Errorf("layer2 %%+v output should contain TestWrapf_DoubleWrap_StackTrace, got %q", out)
	}
}

func TestWrapf_NilError_NoStack(t *testing.T) {
	// nil 返回 nil，无堆栈
	if err := Wrapf(nil, "something"); err != nil {
		t.Errorf("Wrapf(nil, ...) = %v, want nil", err)
	}
}

func TestWrapf_StackTraceNotOnCodeError(t *testing.T) {
	// 原始 CodeError 本身不携带堆栈，%+v 输出无堆栈信息
	err := NewErrCode(uint32(DbError))
	out := fmt.Sprintf("%+v", err)
	if contains(out, "code_test.go") {
		t.Errorf("*CodeError PlusV should NOT contain stack trace")
	}
}

// contains 辅助函数，避免引入 strings 包
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ---------- 常量值校验 ----------

func TestPredefinedConstants(t *testing.T) {
	consts := []struct {
		name string
		code Code
		want uint32
	}{
		{"OK", OK, 0},
		{"ServerCommonError", ServerCommonError, 100001},
		{"RequestParamError", RequestParamError, 100002},
		{"TokenExpireError", TokenExpireError, 100003},
		{"TokenGenerateError", TokenGenerateError, 100004},
		{"DbError", DbError, 100005},
		{"DbUpdateAffectedZeroError", DbUpdateAffectedZeroError, 100006},
		{"MdCommonError", MdCommonError, 100007},
		{"PermitNoAccess", PermitNoAccess, 100008},
		{"SignParamError", SignParamError, 100009},
	}
	for _, tt := range consts {
		t.Run(tt.name, func(t *testing.T) {
			if uint32(tt.code) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}

func TestAlertLevelConstants(t *testing.T) {
	levels := []struct {
		name  string
		level alertLevel
		want  string
	}{
		{"AlertP0", AlertP0, "P0"},
		{"AlertP1", AlertP1, "P1"},
		{"AlertP2", AlertP2, "P2"},
		{"AlertP3", AlertP3, "P3"},
	}
	for _, tt := range levels {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.level) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.level, tt.want)
			}
		})
	}
}
