package manager

import (
	"reflect"
	"stateful-service/slog"
	"stateful-service/utils"
)

type StateManager interface {
	RegisterState(rcvr interface{}) error
	RegisterField(stateName string, fieldName string) error
	RegisterFields(stateName string, fieldNames ...string) error
	Checkpoint(svcName string) map[string]interface{}
	Restore(fieldValues map[string]interface{})
}

type stateManager struct {
	RegisteredStates map[string]*state
}

func NewStateManager() StateManager {
	manager := &stateManager{}
	manager.RegisteredStates = make(map[string]*state)
	return manager
}

func (m *stateManager) RegisterState(rcvr interface{}) error {
	receiver := reflect.ValueOf(rcvr)
	if receiver.Kind() != reflect.Ptr {
		slog.Errorf("[rpcProxy.manager.stateManager.RegisterState]: receiver expected kind %v, actual kind %v", reflect.Ptr, receiver.Kind())
		return &ErrReceiverKind{
			kind: receiver.Kind(),
		}
	}
	stateName := reflect.TypeOf(rcvr).Name()
	if _, ok := m.RegisteredStates[stateName]; !ok {
		m.RegisteredStates[stateName] = &state{
			name:             stateName,
			typ:              reflect.TypeOf(rcvr),
			rcvr:             receiver,
			registeredFields: utils.NewStringSet(),
		}
	}
	slog.Infof("[rpcProxy.manager.stateManager.AddState]: register %v state successfully", stateName)
	return nil
}

func (m *stateManager) RegisterField(stateName string, name string) error {
	return m.RegisteredStates[stateName].registerField(name)
}

func (m *stateManager) RegisterFields(stateName string, names ...string) error {
	return m.RegisteredStates[stateName].registerFields(names...)
}

func (m *stateManager) Checkpoint(svcName string) map[string]interface{} {
	result := make(map[string]interface{})
	result[svcName] = m.RegisteredStates[svcName].checkpoint()
	return result
}

func (m *stateManager) Restore(stateValues map[string]interface{}) {
	for stateName, stateValue := range stateValues {
		m.RegisteredStates[stateName].restore(stateValue.(map[string]reflect.Value))
	}
}

type state struct {
	name             string
	typ              reflect.Type
	rcvr             reflect.Value
	registeredFields utils.StringSet
}

func (st *state) registerField(fieldName string) error {
	field := st.rcvr.FieldByName(fieldName)
	if field.IsZero() {
		return &ErrFieldNotFound{
			clsName:   st.name,
			fieldName: fieldName,
		}
	}
	if !field.CanSet() {
		return &ErrFieldCanNotSet{
			fieldName: fieldName,
		}
	}
	st.registeredFields.Insert(fieldName)
	slog.Infof("[rpcProxy.manager.state.registerField]: state [%v] register field [%v] successfully", st.name, fieldName)
	return nil
}

func (st *state) registerFields(fieldNames ...string) error {
	for _, fieldName := range fieldNames {
		field := st.rcvr.FieldByName(fieldName)
		if field.IsZero() {
			return &ErrFieldNotFound{
				clsName:   st.name,
				fieldName: fieldName,
			}
		}
		if !field.CanSet() {
			return &ErrFieldCanNotSet{
				fieldName: fieldName,
			}
		}
	}
	st.registeredFields.Insert(fieldNames...)
	slog.Infof("[rpcProxy.manager.state.registerField]: state [%v] register field %v successfully", st.name, fieldNames)
	return nil
}

func (st *state) checkpoint() map[string]reflect.Value {
	fieldValues := make(map[string]reflect.Value)
	for _, fieldName := range st.registeredFields.List() {
		value := st.rcvr.FieldByName(fieldName)
		fieldValues[fieldName] = value
	}
	slog.Debugf("[rpcProxy.manager.state.checkpoint]: state [%v] checkpoint successfully, field values = [%v]", st.name, fieldValues)
	return fieldValues
}

func (st *state) restore(fieldValues map[string]reflect.Value) {
	for fieldName, fieldValue := range fieldValues {
		slog.Debugf("[rpcProxy.manager.state.restore]: state [%v] restore field [%v] value [%v]", st.name, fieldName, fieldValue)
		st.rcvr.FieldByName(fieldName).Set(fieldValue)
	}
}
