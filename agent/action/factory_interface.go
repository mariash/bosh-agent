package action

type Factory interface {
	Create(actionType string, method string) (action Action, err error)
}
