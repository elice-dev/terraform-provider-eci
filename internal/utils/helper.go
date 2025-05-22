package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringOrNull[T fmt.Stringer](ptr *T) types.String {
	if ptr == nil {
		return types.StringNull()
	}

	return types.StringValue((*ptr).String())
}

func StringValOrNull(ptr *string) types.String {
	if ptr == nil {
		return types.StringNull()
	}
	return types.StringValue(*ptr)
}
