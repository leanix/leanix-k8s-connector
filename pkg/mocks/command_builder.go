// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	models "github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
	workloadsmodels "github.com/leanix/leanix-k8s-connector/pkg/iris/workloads/models"
	mock "github.com/stretchr/testify/mock"
)

// CommandBuilder is an autogenerated mock type for the CommandBuilder type
type CommandBuilder struct {
	mock.Mock
}

type CommandBuilder_Expecter struct {
	mock *mock.Mock
}

func (_m *CommandBuilder) EXPECT() *CommandBuilder_Expecter {
	return &CommandBuilder_Expecter{mock: &_m.Mock}
}

// Body provides a mock function with given fields: body
func (_m *CommandBuilder) Body(body models.CommandBody) workloadsmodels.CommandBuilder {
	ret := _m.Called(body)

	var r0 workloadsmodels.CommandBuilder
	if rf, ok := ret.Get(0).(func(models.CommandBody) workloadsmodels.CommandBuilder); ok {
		r0 = rf(body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(workloadsmodels.CommandBuilder)
		}
	}

	return r0
}

// CommandBuilder_Body_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Body'
type CommandBuilder_Body_Call struct {
	*mock.Call
}

// Body is a helper method to define mock.On call
//   - body models.CommandBody
func (_e *CommandBuilder_Expecter) Body(body interface{}) *CommandBuilder_Body_Call {
	return &CommandBuilder_Body_Call{Call: _e.mock.On("Body", body)}
}

func (_c *CommandBuilder_Body_Call) Run(run func(body models.CommandBody)) *CommandBuilder_Body_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(models.CommandBody))
	})
	return _c
}

func (_c *CommandBuilder_Body_Call) Return(_a0 workloadsmodels.CommandBuilder) *CommandBuilder_Body_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CommandBuilder_Body_Call) RunAndReturn(run func(models.CommandBody) workloadsmodels.CommandBuilder) *CommandBuilder_Body_Call {
	_c.Call.Return(run)
	return _c
}

// Build provides a mock function with given fields:
func (_m *CommandBuilder) Build() models.CommandEvent {
	ret := _m.Called()

	var r0 models.CommandEvent
	if rf, ok := ret.Get(0).(func() models.CommandEvent); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(models.CommandEvent)
	}

	return r0
}

// CommandBuilder_Build_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Build'
type CommandBuilder_Build_Call struct {
	*mock.Call
}

// Build is a helper method to define mock.On call
func (_e *CommandBuilder_Expecter) Build() *CommandBuilder_Build_Call {
	return &CommandBuilder_Build_Call{Call: _e.mock.On("Build")}
}

func (_c *CommandBuilder_Build_Call) Run(run func()) *CommandBuilder_Build_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *CommandBuilder_Build_Call) Return(_a0 models.CommandEvent) *CommandBuilder_Build_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CommandBuilder_Build_Call) RunAndReturn(run func() models.CommandEvent) *CommandBuilder_Build_Call {
	_c.Call.Return(run)
	return _c
}

// Header provides a mock function with given fields: header
func (_m *CommandBuilder) Header(header models.CommandProperties) workloadsmodels.CommandBuilder {
	ret := _m.Called(header)

	var r0 workloadsmodels.CommandBuilder
	if rf, ok := ret.Get(0).(func(models.CommandProperties) workloadsmodels.CommandBuilder); ok {
		r0 = rf(header)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(workloadsmodels.CommandBuilder)
		}
	}

	return r0
}

// CommandBuilder_Header_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Header'
type CommandBuilder_Header_Call struct {
	*mock.Call
}

// Header is a helper method to define mock.On call
//   - header models.CommandProperties
func (_e *CommandBuilder_Expecter) Header(header interface{}) *CommandBuilder_Header_Call {
	return &CommandBuilder_Header_Call{Call: _e.mock.On("Header", header)}
}

func (_c *CommandBuilder_Header_Call) Run(run func(header models.CommandProperties)) *CommandBuilder_Header_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(models.CommandProperties))
	})
	return _c
}

func (_c *CommandBuilder_Header_Call) Return(_a0 workloadsmodels.CommandBuilder) *CommandBuilder_Header_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CommandBuilder_Header_Call) RunAndReturn(run func(models.CommandProperties) workloadsmodels.CommandBuilder) *CommandBuilder_Header_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewCommandBuilder interface {
	mock.TestingT
	Cleanup(func())
}

// NewCommandBuilder creates a new instance of CommandBuilder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCommandBuilder(t mockConstructorTestingTNewCommandBuilder) *CommandBuilder {
	mock := &CommandBuilder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
