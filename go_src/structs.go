package main

import "go.wit.com/widget"

type Node struct {
	Parent   *Node
	children []*Node

	WidgetId   int // widget ID
	WidgetType widget.WidgetType
	ParentId   int // parent ID

	State widget.State

	ddStrings []string

	// // the internal plugin toolkit structure
	// // in the gtk plugin, it has gtk things like margin & border settings
	// // in the text console one, it has text console things like colors for menus & buttons
	// TK any
}
