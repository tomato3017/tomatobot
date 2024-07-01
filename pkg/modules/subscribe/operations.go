package subscribe

type operations string

var (
	OperationCreate operations = "create"
	OperationList   operations = "list"
	OperationUnsub  operations = "unsubscribe"
)
