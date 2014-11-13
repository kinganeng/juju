// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package diskmanager

import (
	"github.com/juju/loggo"
	"github.com/juju/names"

	"github.com/juju/juju/apiserver/common"
	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/state"
)

func init() {
	common.RegisterStandardFacade("DiskManager", 0, NewDiskManagerAPI)
}

var logger = loggo.GetLogger("juju.apiserver.diskmanager")

// DiskManagerAPI provides access to the DiskManager API facade.
type DiskManagerAPI struct {
	st          stateInterface
	authorizer  common.Authorizer
	getAuthFunc common.GetAuthFunc
}

var getState = func(st *state.State) stateInterface {
	return stateShim{st}
}

// NewDiskManagerAPI creates a new client-side DiskManager API facade.
func NewDiskManagerAPI(
	st *state.State,
	resources *common.Resources,
	authorizer common.Authorizer,
) (*DiskManagerAPI, error) {

	if !authorizer.AuthMachineAgent() {
		return nil, common.ErrPerm
	}

	authEntityTag := authorizer.GetAuthTag()
	getAuthFunc := func() (common.AuthFunc, error) {
		return func(tag names.Tag) bool {
			// A machine agent can always access its own machine.
			return tag == authEntityTag
		}, nil
	}

	return &DiskManagerAPI{
		st:          getState(st),
		authorizer:  authorizer,
		getAuthFunc: getAuthFunc,
	}, nil
}

func (d *DiskManagerAPI) SetMachineBlockDevices(args params.SetMachineBlockDevices) (params.ErrorResults, error) {
	result := params.ErrorResults{
		Results: make([]params.ErrorResult, len(args.MachineBlockDevices)),
	}
	canAccess, err := d.getAuthFunc()
	if err != nil {
		return result, err
	}
	for i, arg := range args.MachineBlockDevices {
		tag, err := names.ParseMachineTag(arg.Machine)
		if err != nil {
			result.Results[i].Error = common.ServerError(common.ErrPerm)
			continue
		}
		if !canAccess(tag) {
			err = common.ErrPerm
		} else {
			err = d.st.SetMachineBlockDevices(tag.Id(), arg.BlockDevices)
		}
		result.Results[i].Error = common.ServerError(err)
	}
	return result, nil
}
