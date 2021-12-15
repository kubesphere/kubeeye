package funcrules

type ValidateResults struct {
	ValidateResults []ResultReceiver
}

type ResultReceiver struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Type      string   `json:"kind"`
	Message   []string `json:"message"`
	Reason    string   `json:"reason"`
}

type FuncRule interface {
	Exec() ValidateResults
}