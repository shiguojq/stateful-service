package manager

import (
	"reflect"
	"testing"
)

type TestStruct struct {
	Num      int
	Float    float32
	Bool     bool
	String   string
	IntSlice []int
	IntMap   map[int]int
}

/*func (t *TestStruct) GetNum(args MethodArgs) MethodResult {
	idReq := args.Message.(*pb.GenerateIdRequest)
	t.String = idReq.SvcName
	t.Num++
	idResp := &pb.GenerateIdResponse{
		Id: int64(t.Num),
	}
	result := MethodResult{
		Message: idResp,
	}
	return result
}*/

func register(manager StateManager, testStruct interface{}, t *testing.T) {
	if err := manager.RegisterState(testStruct); err != nil {
		t.Fatalf("register state failed, err: %v", err.Error())
	}
	for _, fieldName := range []string{"Num", "Float", "Bool", "String", "IntSlice", "IntMap"} {
		if err := manager.RegisterField("TestStruct", fieldName); err != nil {
			t.Fatalf("register field %v failed, err: %v", fieldName, err.Error())
		}
	}
}

func TestStateManager(t *testing.T) {
	manager := NewStateManager()
	testStruct := &TestStruct{
		Num:      100,
		Float:    0,
		Bool:     true,
		String:   "1111",
		IntSlice: make([]int, 0),
		IntMap:   make(map[int]int),
	}
	register(manager, testStruct, t)
	testState := manager.Checkpoint("TestStruct")
	if testStruct.Bool != testState["Bool"].Bool() ||
		int64(testStruct.Num) != testState["Num"].Int() ||
		testStruct.String != testState["String"].String() ||
		float64(testStruct.Float) != testState["Float"].Float() ||
		len(testStruct.IntSlice) != len(testState["IntSlice"].Interface().([]int)) ||
		len(testStruct.IntMap) != len(testState["IntMap"].Interface().(map[int]int)) {
		t.Fatalf("testStruct does not equal to checkpoint state")
	}
	expectStruct := &TestStruct{
		Num:      9090,
		Float:    23123,
		Bool:     false,
		String:   "11df",
		IntSlice: []int{23232, 323232, 2312},
		IntMap:   map[int]int{1: 232, 2332: 2323},
	}

	checkpoint := make(map[string]map[string]reflect.Value)
	checkpoint["TestStruct"] = map[string]reflect.Value{}
	checkpoint["TestStruct"]["Bool"] = reflect.ValueOf(expectStruct.Bool)
	checkpoint["TestStruct"]["Num"] = reflect.ValueOf(expectStruct.Num)
	checkpoint["TestStruct"]["String"] = reflect.ValueOf(expectStruct.String)
	checkpoint["TestStruct"]["IntSlice"] = reflect.ValueOf(expectStruct.IntSlice)
	checkpoint["TestStruct"]["IntMap"] = reflect.ValueOf(expectStruct.IntMap)
	checkpoint["TestStruct"]["Float"] = reflect.ValueOf(expectStruct.Float)
	manager.Restore(checkpoint)
	if testStruct.Bool != expectStruct.Bool ||
		testStruct.Num != expectStruct.Num ||
		testStruct.String != expectStruct.String ||
		testStruct.Float != expectStruct.Float {
		t.Fatalf("testStruct does not equal to checkpoint state")
	}

	for i := range testStruct.IntSlice {
		if testStruct.IntSlice[i] != expectStruct.IntSlice[i] {
			t.Fatalf("testStruct does not equal to checkpoint state")
		}
	}

	for k, v := range testStruct.IntMap {
		if expectStruct.IntMap[k] != v {
			t.Fatalf("testStruct does not equal to checkpoint state")
		}
	}
}
