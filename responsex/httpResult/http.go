package httpResult

import (
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func buildErrFields(err error, data ErrData) []logc.LogField {
	fields := []logc.LogField{
		logc.Field(LogFType, LogApiError),
		logc.Field(LogFResult, map[string]any{"code": data.Code, "msg": data.Msg}),
		logc.Field(LogFStack, fmt.Sprintf("%+v", err)),
	}
	if data.AlertLevel != "" {
		fields = append(fields, logc.Field(LogFAlertLevel, data.AlertLevel))
	}
	if len(data.AlertData) > 0 {
		fields = append(fields, logc.Field(LogFAlertData, data.AlertData))
	}
	return fields
}

// NewHandler 创建带拦截器的 HTTP 错误处理器，可直接传给 httpx.SetErrorHandlerCtx
func NewHandler(interceptors ...ErrInterceptor) func(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
	return func(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
		traceId := trace.TraceIDFromContext(r.Context())
		spanId := trace.SpanIDFromContext(r.Context())

		if err == nil {
			logc.Infow(r.Context(), "ok",
				logc.Field(LogFType, LogApiSuccess),
				logc.Field(LogFResult, resp))
			httpx.WriteJson(w, http.StatusOK, Success(traceId, spanId, resp))
			return
		}

		data := ParseErr(err)

		// 执行拦截器链
		for _, interceptor := range interceptors {
			interceptor(r, err, &data)
		}

		logc.Errorw(r.Context(), data.Msg, buildErrFields(err, data)...)
		httpx.WriteJson(w, http.StatusOK, Error(traceId, spanId, data.Code, data.Msg))
	}
}

// HttpResult 保持兼容，无拦截器
func HttpResult(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
	defaultHandler(r, w, resp, err)
}

var defaultHandler = NewHandler()
