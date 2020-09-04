package global

import "reflect"

func isnil(args interface{}) bool {
	if reflect.ValueOf(args).IsNil() || (args == nil && reflect.ValueOf(args).Kind() == reflect.Ptr) {
		return true
	}
	return false
}
