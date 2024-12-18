package main

import (
	"fmt"
	"io"
)

func importString(id, to, index, condition string) string {
	switch {
	case condition == "":
		return fmt.Sprintf("import {\n"+
			"  to = %s%s\n"+
			"  id = \"%s\"\n"+
			"}", to, index, id)
	default:
		return fmt.Sprintf("import {\n"+
			"  for_each = %s ? [1] : []\n"+
			"  to       = %s%s\n"+
			"  id       = \"%s\"\n"+
			"}", condition, to, index, id)
	}
}

func importBlock(w io.Writer, resource Resource, condition string) {
	for _, instance := range resource.Instances {
		fmt.Fprintln(w, importString(instance.Import(resource.Type), resource.ID(), instance.Index(), condition))

		fmt.Fprintln(w)
	}
}

func removedString(from string) string {
	return fmt.Sprintf("removed {\n"+
		"  from = %s\n"+
		"  lifecycle {\n"+
		"    destroy = false\n"+
		"  }\n"+
		"}", from)
}

func removedBlock(w io.Writer, resource Resource) {
	fmt.Fprintln(w, removedString(resource.ID()))

	fmt.Fprintln(w)
}
