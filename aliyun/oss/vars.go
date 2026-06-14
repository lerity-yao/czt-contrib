package oss

// HTTP header constants.
const (
	headerAuthorization = "Authorization"
	headerContentMD5    = "Content-MD5"
	headerContentType   = "Content-Type"
	headerDate          = "Date"
	headerUserAgent     = "User-Agent"

	headerCopySource = "x-oss-copy-source"
)

const (
	authPrefix       = "OSS "
	defaultUserAgent = "Go-Aliyun-OSS-Client"
	defaultScheme    = "https"
)

// ossSubResources defines the OSS V1 sub-resources that participate in
// the CanonicalizedResource of the signature.
// Regular query parameters (e.g. prefix, marker, max-keys) are NOT included.
var ossSubResources = map[string]struct{}{
	"acl":                          {},
	"uploads":                      {},
	"location":                     {},
	"cors":                         {},
	"logging":                      {},
	"website":                      {},
	"referer":                      {},
	"lifecycle":                    {},
	"delete":                       {},
	"append":                       {},
	"tagging":                      {},
	"objectMeta":                   {},
	"uploadId":                     {},
	"img":                          {},
	"stat":                         {},
	"cache":                        {},
	"qos":                          {},
	"symlink":                      {},
	"response-content-type":        {},
	"response-content-language":    {},
	"response-expires":             {},
	"response-cache-control":       {},
	"response-content-disposition": {},
	"response-content-encoding":    {},
	"restore":                      {},
	"x-oss-process":                {},
	"live":                         {},
	"status":                       {},
	"vod":                          {},
	"startTime":                    {},
	"endTime":                      {},
	"x-oss-traffic-limit":          {},
	"objectType":                   {},
	"meta":                         {},
	"security-token":               {},
}
