package templates

import (
	"strings"
	"text/template/parse"

	"k8s.io/apimachinery/pkg/util/sets"
)

// This walks the parsed AST and extracts the `.params.XXX` names from nodes.
// The returned names are unique.
func parseTextParamNames(tree *parse.Tree) []string {
	names := listNodeFields([]parse.Node{tree.Root})

	return sets.NewString(names...).List()
}

func listNodeFields(nodes []parse.Node) []string {
	var res []string
	for _, node := range nodes {
		switch node.Type() {
		case parse.NodePipe:
			res = append(res, listNodeFieldsFromPipe(node.(*parse.PipeNode))...)
		case parse.NodeAction:
			items := listNodeFieldsFromPipe(node.(*parse.ActionNode).Pipe)
			if items[0] == "params" {
				res = append(res, strings.Join(items[1:], "."))
			}
		case parse.NodeList:
			res = append(res, listNodeFields(node.(*parse.ListNode).Nodes)...)
		case parse.NodeCommand:
			res = append(res, listNodeFields(node.(*parse.CommandNode).Args)...)
		case parse.NodeIf, parse.NodeWith, parse.NodeRange:
			items := listNodeFieldsFromBranch(node)
			if items[0] == "params" {
				res = append(res, strings.Join(items[1:], "."))
			}
		case parse.NodeField:
			res = append(res, node.(*parse.FieldNode).Ident...)
		}
	}

	return res
}

func listNodeFieldsFromBranch(node parse.Node) []string {
	var res []string
	var b parse.BranchNode
	switch node.Type() {
	case parse.NodeIf:
		b = node.(*parse.IfNode).BranchNode
	case parse.NodeWith:
		b = node.(*parse.WithNode).BranchNode
	case parse.NodeRange:
		b = node.(*parse.RangeNode).BranchNode
	default:
		return res
	}
	if b.Pipe != nil {
		res = append(res, listNodeFieldsFromPipe(b.Pipe)...)
	}
	if b.List != nil {
		res = append(res, listNodeFields(b.List.Nodes)...)
	}
	if b.ElseList != nil {
		res = append(res, listNodeFields(b.ElseList.Nodes)...)
	}
	return res
}

func listNodeFieldsFromPipe(p *parse.PipeNode) []string {
	var res []string
	for _, c := range p.Cmds {
		res = append(res, listNodeFields(c.Args)...)
	}
	return res
}
