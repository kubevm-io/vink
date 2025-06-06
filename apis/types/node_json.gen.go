// Code generated by protoc-gen-jsonshim. DO NOT EDIT.
package types

import (
	bytes "bytes"
	jsonpb "github.com/golang/protobuf/jsonpb"
)

// MarshalJSON is a custom marshaler for NodeResourceMetrics
func (this *NodeResourceMetrics) MarshalJSON() ([]byte, error) {
	str, err := NodeMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for NodeResourceMetrics
func (this *NodeResourceMetrics) UnmarshalJSON(b []byte) error {
	return NodeUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for NodeCephStorage
func (this *NodeCephStorage) MarshalJSON() ([]byte, error) {
	str, err := NodeMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for NodeCephStorage
func (this *NodeCephStorage) UnmarshalJSON(b []byte) error {
	return NodeUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

var (
	NodeMarshaler   = &jsonpb.Marshaler{}
	NodeUnmarshaler = &jsonpb.Unmarshaler{AllowUnknownFields: true}
)
