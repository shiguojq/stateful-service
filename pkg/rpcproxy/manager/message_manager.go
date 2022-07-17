package manager

type MessageType int

const (
	AsyncReq MessageType = iota
	SyncReq
	CheckpointReq
	RestoreReq
	BlockReq
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
	Type       MessageType
	MethodName string
	Source     string
	SourceHost string
	ReplyCh    chan ReplyMsg
	Callback   string
	Args       []byte
}

type ReplyMsg struct {
	Ok    bool
	Reply []byte
}
