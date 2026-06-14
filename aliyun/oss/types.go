package oss

import (
	"encoding/xml"
	"fmt"
)

// ListBucketResult is the XML response of the GetBucket (ListObjects) operation.
type ListBucketResult struct {
	XMLName        xml.Name       `xml:"ListBucketResult"`
	Name           string         `xml:"Name"`
	Prefix         string         `xml:"Prefix"`
	Marker         string         `xml:"Marker"`
	NextMarker     string         `xml:"NextMarker"`
	MaxKeys        int            `xml:"MaxKeys"`
	Delimiter      string         `xml:"Delimiter"`
	IsTruncated    bool           `xml:"IsTruncated"`
	EncodingType   string         `xml:"EncodingType"`
	Contents       []ObjectInfo   `xml:"Contents"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes"`
}

// ObjectInfo describes a single object returned by ListObjects.
type ObjectInfo struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Type         string `xml:"Type"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

// CommonPrefix represents a common prefix returned when delimiter is used.
type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// ObjectMeta holds object metadata returned by HeadObject.
type ObjectMeta struct {
	Size         int64
	ContentType  string
	ETag         string
	LastModified string
	Metadata     map[string]string // x-oss-meta-* custom metadata
}

// ServiceError represents an error returned by the OSS server.
type ServiceError struct {
	XMLName    xml.Name `xml:"Error"`
	StatusCode int      `xml:"-"` // HTTP status code, not part of XML
	Code       string   `xml:"Code"`
	Message    string   `xml:"Message"`
	RequestId  string   `xml:"RequestId"`
	HostId     string   `xml:"HostId"`
	Bucket     string   `xml:"Bucket"`
	Key        string   `xml:"Key"`
}

func (e *ServiceError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("oss: %s — %s (key: %s, requestId: %s)",
			e.Code, e.Message, e.Key, e.RequestId)
	}
	return fmt.Sprintf("oss: %s — %s (requestId: %s)", e.Code, e.Message, e.RequestId)
}

// CopyObjectResult is the XML response of the CopyObject operation.
type CopyObjectResult struct {
	XMLName      xml.Name `xml:"CopyObjectResult"`
	ETag         string   `xml:"ETag"`
	LastModified string   `xml:"LastModified"`
}
