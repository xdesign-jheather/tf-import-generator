package main

import (
	"fmt"
	"io"
	"strings"
)

func importSingle(id, to, index, condition string) string {
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

func importMulti(resource Resource, condition string) string {
	var sb strings.Builder

	sb.WriteString("import {\n")

	switch {
	case condition == "":
		sb.WriteString("  for_each = {\n")
	default:
		sb.WriteString(fmt.Sprintf("for_each = %s ? {\n", condition))
	}

	for _, instance := range resource.Instances {
		x := instance.Index()
		sb.WriteString(fmt.Sprintf(`    %s = "%s"`+"\n", x[1:len(x)-1], instance.Import(resource.Type)))
	}

	switch {
	case condition == "":
		sb.WriteString("  }\n")
	default:
		sb.WriteString("  } : {}\n")
	}

	sb.WriteString(fmt.Sprintf("  to     = %s[each.key] \n", resource.ID()))
	sb.WriteString("  id     = each.value \n")

	sb.WriteString("}\n")

	return sb.String()
}

func importBlock(w io.Writer, resource Resource, condition string) {
	switch {
	case len(resource.Instances) == 0:
	case len(resource.Instances) == 1:
		fmt.Fprintln(w, importSingle(resource.Instances[0].Import(resource.Type), resource.ID(), resource.Instances[0].Index(), condition))

		fmt.Fprintln(w)
	default:
		fmt.Fprintln(w, importMulti(resource, condition))
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
