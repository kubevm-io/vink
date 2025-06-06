// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: types/types.proto

package types

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on NamespaceName with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *NamespaceName) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on NamespaceName with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in NamespaceNameMultiError, or
// nil if none found.
func (m *NamespaceName) ValidateAll() error {
	return m.validate(true)
}

func (m *NamespaceName) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Namespace

	// no validation rules for Name

	if len(errors) > 0 {
		return NamespaceNameMultiError(errors)
	}

	return nil
}

// NamespaceNameMultiError is an error wrapping multiple validation errors
// returned by NamespaceName.ValidateAll() if the designated constraints
// aren't met.
type NamespaceNameMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m NamespaceNameMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m NamespaceNameMultiError) AllErrors() []error { return m }

// NamespaceNameValidationError is the validation error returned by
// NamespaceName.Validate if the designated constraints aren't met.
type NamespaceNameValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NamespaceNameValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NamespaceNameValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NamespaceNameValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NamespaceNameValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NamespaceNameValidationError) ErrorName() string { return "NamespaceNameValidationError" }

// Error satisfies the builtin error interface
func (e NamespaceNameValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNamespaceName.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NamespaceNameValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NamespaceNameValidationError{}

// Validate checks the field values on FieldSelector with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *FieldSelector) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on FieldSelector with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in FieldSelectorMultiError, or
// nil if none found.
func (m *FieldSelector) ValidateAll() error {
	return m.validate(true)
}

func (m *FieldSelector) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for FieldPath

	// no validation rules for Operator

	// no validation rules for JsonPath

	if len(errors) > 0 {
		return FieldSelectorMultiError(errors)
	}

	return nil
}

// FieldSelectorMultiError is an error wrapping multiple validation errors
// returned by FieldSelector.ValidateAll() if the designated constraints
// aren't met.
type FieldSelectorMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m FieldSelectorMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m FieldSelectorMultiError) AllErrors() []error { return m }

// FieldSelectorValidationError is the validation error returned by
// FieldSelector.Validate if the designated constraints aren't met.
type FieldSelectorValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e FieldSelectorValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e FieldSelectorValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e FieldSelectorValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e FieldSelectorValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e FieldSelectorValidationError) ErrorName() string { return "FieldSelectorValidationError" }

// Error satisfies the builtin error interface
func (e FieldSelectorValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sFieldSelector.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = FieldSelectorValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = FieldSelectorValidationError{}

// Validate checks the field values on FieldSelectorGroup with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *FieldSelectorGroup) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on FieldSelectorGroup with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// FieldSelectorGroupMultiError, or nil if none found.
func (m *FieldSelectorGroup) ValidateAll() error {
	return m.validate(true)
}

func (m *FieldSelectorGroup) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Operator

	for idx, item := range m.GetFieldSelectors() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, FieldSelectorGroupValidationError{
						field:  fmt.Sprintf("FieldSelectors[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, FieldSelectorGroupValidationError{
						field:  fmt.Sprintf("FieldSelectors[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return FieldSelectorGroupValidationError{
					field:  fmt.Sprintf("FieldSelectors[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return FieldSelectorGroupMultiError(errors)
	}

	return nil
}

// FieldSelectorGroupMultiError is an error wrapping multiple validation errors
// returned by FieldSelectorGroup.ValidateAll() if the designated constraints
// aren't met.
type FieldSelectorGroupMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m FieldSelectorGroupMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m FieldSelectorGroupMultiError) AllErrors() []error { return m }

// FieldSelectorGroupValidationError is the validation error returned by
// FieldSelectorGroup.Validate if the designated constraints aren't met.
type FieldSelectorGroupValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e FieldSelectorGroupValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e FieldSelectorGroupValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e FieldSelectorGroupValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e FieldSelectorGroupValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e FieldSelectorGroupValidationError) ErrorName() string {
	return "FieldSelectorGroupValidationError"
}

// Error satisfies the builtin error interface
func (e FieldSelectorGroupValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sFieldSelectorGroup.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = FieldSelectorGroupValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = FieldSelectorGroupValidationError{}

// Validate checks the field values on OperatingSystem with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *OperatingSystem) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on OperatingSystem with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// OperatingSystemMultiError, or nil if none found.
func (m *OperatingSystem) ValidateAll() error {
	return m.validate(true)
}

func (m *OperatingSystem) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Name

	// no validation rules for Version

	if len(errors) > 0 {
		return OperatingSystemMultiError(errors)
	}

	return nil
}

// OperatingSystemMultiError is an error wrapping multiple validation errors
// returned by OperatingSystem.ValidateAll() if the designated constraints
// aren't met.
type OperatingSystemMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m OperatingSystemMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m OperatingSystemMultiError) AllErrors() []error { return m }

// OperatingSystemValidationError is the validation error returned by
// OperatingSystem.Validate if the designated constraints aren't met.
type OperatingSystemValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e OperatingSystemValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e OperatingSystemValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e OperatingSystemValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e OperatingSystemValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e OperatingSystemValidationError) ErrorName() string { return "OperatingSystemValidationError" }

// Error satisfies the builtin error interface
func (e OperatingSystemValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sOperatingSystem.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = OperatingSystemValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = OperatingSystemValidationError{}
