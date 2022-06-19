package manager

const (
	AsyncReq = iota
	SyncReq
	CheckpointReq
	RestoreReq
)

type MessageManager interface {
	AddMessage(msg ReqMsg)
	GetMessage(id int64) ReqMsg
	DeleteMessage(id int64)
}

type messageManager struct {
	messages map[int64]ReqMsg
}

var m *messageManager

func init() {
	m = &messageManager{}
	m.messages = make(map[int64]ReqMsg)
}

func NewMessageManager() MessageManager {
	return m
}

func (m *messageManager) AddMessage(msg ReqMsg) {
	m.messages[msg.Id] = msg
}

func (m *messageManager) GetMessage(id int64) ReqMsg {
	return m.messages[id]
}

func (m *messageManager) DeleteMessage(id int64) {
	delete(m.messages, id)
}

type ReqMsg struct {
	Id         int64
	Type       int
	MethodName string
	Args       []byte
	Source     string
	ReplyCh    chan ReplyMsg
	Callback   string
}

type ReplyMsg struct {
	Ok    bool
	Reply []byte
}
