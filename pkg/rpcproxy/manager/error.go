package manager

import (
	"fmt"
	"reflect"
)

type ErrFieldNotFound struct {
	clsName   string
	fieldName string
}

func (e *ErrFieldNotFound) Error() string {
	return fmt.Sprintf("could not find field [%v] in class [%v]", e.fieldName, e.clsName)
}

type ErrMethodNotFound struct {
	clsName    string
	methodName string
}

func (e *ErrMethodNotFound) Error() string {
	return fmt.Sprintf("could not find field [%v] in class [%v]", e.methodName, e.clsName)
}

type ErrServiceNotFound struct {
	serviceName string
}

func (e *ErrServiceNotFound) Error() string {
	return fmt.Sprintf("could not find service [%v] in service manager", e.serviceName)
}

type ErrFieldCanNotSet struct {
	fieldName string
}

func (e *ErrFieldCanNotSet) Error() string {
	return fmt.Sprintf("field [%v] could not be set", e.fieldName)
}

type ErrReceiverKind struct {
	kind reflect.Kind
}

func (e *ErrReceiverKind) Error() string {
	return fmt.Sprintf("unexpected kind [%v], [%v] is expected", e.kind, reflect.Ptr)
}
