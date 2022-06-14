package manager

type MessageManager interface {
	AddMessage(msg ReqMsg)
	GetMessage(id string) ReqMsg
	DeleteMessage(id string)
}

type messageManager struct {
	messages map[string]ReqMsg
}

var m *messageManager

func init() {
	m = &messageManager{}
	m.messages = make(map[string]ReqMsg)
}

func NewMessageManager() MessageManager {
	return m
}

func (m *messageManager) AddMessage(msg ReqMsg) {
	m.messages[msg.Id] = msg
}

func (m *messageManager) GetMessage(id string) ReqMsg {
	return m.messages[id]
}

func (m *messageManager) DeleteMessage(id string) {
	delete(m.messages, id)
}

type ReqMsg struct {
	MethodName string
	Id         string
	Args       []byte
	ReplyCh    chan ReplyMsg
}

type ReplyMsg struct {
	Ok    bool
	Reply []byte
}
