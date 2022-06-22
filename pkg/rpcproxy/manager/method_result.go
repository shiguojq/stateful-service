package manager

type MethodResult struct {
	Message    []byte
	IsRequest  bool
	Target     string
	TargetHost string
	Callback   string
}
