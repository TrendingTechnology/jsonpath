package jsonpath

import "reflect"

type syntaxChildSingleIdentifier struct {
	*syntaxBasicNode

	identifier string
}

func (i *syntaxChildSingleIdentifier) retrieve(
	root, current interface{}, container *bufferContainer) errorRuntime {

	srcMap, ok := current.(map[string]interface{})
	if !ok {
		foundType := msgTypeNull
		if current != nil {
			foundType = reflect.TypeOf(current).String()
		}
		return ErrorTypeUnmatched{
			errorBasicRuntime: i.errorRuntime,
			expectedType:      msgTypeObject,
			foundType:         foundType,
		}
	}

	return i.retrieveMapNext(root, srcMap, i.identifier, container)
}
