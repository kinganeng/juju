// Copyright 2012-2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/mgo.v2/txn"

	"github.com/juju/juju/network"
)

// ErrCharmAlreadyUploaded is returned by UpdateUploadedCharm() when
// the given charm is already uploaded and marked as not pending in
// state.
type ErrCharmAlreadyUploaded struct {
	curl *charm.URL
}

func (e *ErrCharmAlreadyUploaded) Error() string {
	return fmt.Sprintf("charm %q already uploaded", e.curl)
}

// IsCharmAlreadyUploadedError returns if the given error is
// ErrCharmAlreadyUploaded.
func IsCharmAlreadyUploadedError(err interface{}) bool {
	if err == nil {
		return false
	}
	// In case of a wrapped error, check the cause first.
	value := err
	cause := errors.Cause(err.(error))
	if cause != nil {
		value = cause
	}
	_, ok := value.(*ErrCharmAlreadyUploaded)
	return ok
}

// ErrCharmRevisionAlreadyModified is returned when a pending or
// placeholder charm is no longer pending or a placeholder, signaling
// the charm is available in state with its full information.
var ErrCharmRevisionAlreadyModified = fmt.Errorf("charm revision already modified")

var ErrDead = fmt.Errorf("not found or dead")
var errNotAlive = fmt.Errorf("not found or not alive")

func onAbort(txnErr, err error) error {
	if txnErr == txn.ErrAborted ||
		errors.Cause(txnErr) == txn.ErrAborted {
		return errors.Trace(err)
	}
	return errors.Trace(txnErr)
}

// ErrProviderIDNotUnique is a standard error to indicate the value specified
// for a ProviderID field is not unique within the current model.
type ErrProviderIDNotUnique struct {
	duplicateIDs []string
}

func (e *ErrProviderIDNotUnique) Error() string {
	idList := strings.Join(e.duplicateIDs, ", ")
	return fmt.Sprintf("ProviderID(s) not unique: %s", idList)
}

// NewProviderIDNotUniqueError returns an instance of ErrProviderIDNotUnique
// initialized with the given duplicate provider IDs.
func NewProviderIDNotUniqueError(providerIDs ...network.Id) error {
	stringIDs := make([]string, len(providerIDs))
	for i, providerID := range providerIDs {
		stringIDs[i] = string(providerID)
	}
	return newProviderIDNotUniqueErrorFromStrings(stringIDs)
}

func newProviderIDNotUniqueErrorFromStrings(providerIDs []string) error {
	return &ErrProviderIDNotUnique{
		duplicateIDs: providerIDs,
	}
}

// IsProviderIDNotUniqueError returns if the given error or its cause is
// ErrProviderIDNotUnique.
func IsProviderIDNotUniqueError(err interface{}) bool {
	if err == nil {
		return false
	}
	// In case of a wrapped error, check the cause first.
	value := err
	cause := errors.Cause(err.(error))
	if cause != nil {
		value = cause
	}
	_, ok := value.(*ErrProviderIDNotUnique)
	return ok
}

// ErrParentDeviceHasChildren is a standard error to indicate a network
// link-layer device cannot be removed because other existing devices refer to
// it as their parent.
type ErrParentDeviceHasChildren struct {
	parentName  string
	numChildren int
}

func (e *ErrParentDeviceHasChildren) Error() string {
	return fmt.Sprintf("parent device %q has %d children", e.parentName, e.numChildren)
}

func newParentDeviceHasChildrenError(parentName string, numChildren int) error {
	return &ErrParentDeviceHasChildren{
		parentName:  parentName,
		numChildren: numChildren,
	}
}

// IsParentDeviceHasChildrenError returns if the given error or its cause is
// ErrParentDeviceHasChildren.
func IsParentDeviceHasChildrenError(err interface{}) bool {
	if err == nil {
		return false
	}
	// In case of a wrapped error, check the cause first.
	value := err
	cause := errors.Cause(err.(error))
	if cause != nil {
		value = cause
	}
	_, ok := value.(*ErrParentDeviceHasChildren)
	return ok
}
