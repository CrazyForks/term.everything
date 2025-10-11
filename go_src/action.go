package main

import (
	"os"

	"go.wit.com/lib/protobuf/guipb"
	"go.wit.com/log"
	"go.wit.com/widget"
)

// everything from the application goes through here
func (me *TermEverything) doAction(a widget.Action) {
	if a.ActionType == widget.ToolkitInit {
		log.Log(TERM_EVERYTHING, "term.doAction() trapped ToolkitInit finally!")
		a.WidgetType = widget.Root
		n := addNode(&a)
		me.Add(n)
		log.Log(TERM_EVERYTHING, "term.doAction() init() me.treeRoot")
		if me.ToolkitInit == nil {
			log.Log(TERM_EVERYTHING, "term.doAction() ToolkitInit() was called before plugin had a chance to initialize")
			log.Log(TERM_EVERYTHING, "term.doAction() TODO: fix channel to pause")
			return
		}
		log.Log(TERM_EVERYTHING, "tree.doAction() doing ToolkitInit()")
		me.ToolkitInit()
		return
	}
	if a.WidgetPB != nil {
		log.Log(TERM_EVERYTHING_WARN, "tree: got a WidgetPB")
		widgetpb := new(guipb.Widgets)
		err := widgetpb.Unmarshal(a.WidgetPB)
		if err != nil {
			log.Log(TERM_EVERYTHING_WARN, "WidgetPB unmarshal err", err)
			return
		}
		log.Log(TERM_EVERYTHING_WARN, "tree: unmarshal worked!")
		var wind *Node
		newa := new(widget.Action)
		newa.WidgetType = widget.Window
		newa.WidgetId = -234
		newa.ParentId = -1234
		newa.State.Enable = true
		wind = addNode(newa)
		if wind == nil {
			log.Log(TERM_EVERYTHING_WARN, "tree: addNode() failed to add win")
			return
		}
		wind.State.ProgName = "WinPB"
		wind.State.Label = "WinPB"
		me.Add(wind)
		me.doWidgetsPB(widgetpb.Tree)
		me.ToolkitClose()
		os.Exit(0)
		// me.doTable(a)
		return
	}
	if a.TablePB != nil {
		log.Log(TERM_EVERYTHING, "tree: got a TablePB")
		me.doTable(a)
		return
	}
	if a.WidgetId == 0 {
		if treeRoot == nil {
			log.Log(TERM_EVERYTHING, "tree.doAction() yes, treeRoot is nil. add here")
		}
	}
	n := treeRoot.FindWidgetId(a.WidgetId)
	switch a.ActionType {
	case widget.Add:
		if n == nil {
			n := me.AddNode(&a)
			me.Add(n)
			return
		}
		if a.WidgetId == 0 {
			// this is ok. This is the binary tree base and it's already initialized. This happens on startup
			return
		}
		// this shouldn't really happen. It's good to print a warning so the plugin code can be debugged
		log.Log(TERM_EVERYTHING_WARN, "attempting to re-add widget", a.WidgetId, a.WidgetType, a.ActionType)
		return
	}
	if n == nil {
		// log.Log(TERM_EVERYTHING_WARN, "tree.FindWidgetId() n == nil", a.WidgetId, a.WidgetType, a.ActionType)
		// log.Log(TERM_EVERYTHING_WARN, "tree.FindWidgetId() n == nil", a.State.CurrentS)
		// log.Log(TERM_EVERYTHING_WARN, "tree.FindWidgetId() n == nil. A bug in your application?")
		log.Log(TERM_EVERYTHING_WARN, "tree.doAction() bug in gui. trying to do action", a.ActionType, "before widget init() wId =", a.WidgetId)
		if a.WidgetId == 0 {
			log.Log(TERM_EVERYTHING_WARN, "tree.doAction() bug in gui. on wId zero. is treeRoot nil?")
			if treeRoot == nil {
				log.Log(TERM_EVERYTHING_WARN, "tree.doAction() yes, treeRoot is nil")
			}
		}
		return
	}

	switch a.ActionType {
	case widget.SetText:
		log.Log(TERM_EVERYTHING, "tree.SetText() a.State.CurrentS =", a.State.CurrentS)
		log.Log(TERM_EVERYTHING, "tree.SetText() a.State.DefaultS =", a.State.DefaultS)
		log.Log(TERM_EVERYTHING, "tree.SetText() a.State.NewString =", a.State.NewString)
		switch n.WidgetType {
		case widget.Dropdown:
			me.SetText(n, a.State.NewString)
		case widget.Combobox:
			me.SetText(n, a.State.NewString)
		case widget.Textbox:
			me.SetText(n, a.State.NewString)
		case widget.Window:
			me.SetTitle(n, a.State.Label)
		default:
			// buttons, checkboxes, groups, etc
			me.SetLabel(n, a.State.Label)
		}
	case widget.AddText:
		switch n.WidgetType {
		case widget.Dropdown:
			n.ddStrings = append(n.ddStrings, a.State.NewString)
			me.AddText(n, a.State.NewString)
		case widget.Combobox:
			n.ddStrings = append(n.ddStrings, a.State.NewString)
			me.AddText(n, a.State.NewString)
		default:
			log.Log(TERM_EVERYTHING_WARN, "AddText() not supported on widget", n.WidgetType, n.String())
		}
	case widget.Checked:
		switch n.WidgetType {
		case widget.Checkbox:
			if me.SetChecked == nil {
				log.Log(TERM_EVERYTHING_WARN, "SetChecked() == nil in toolkit", me.PluginName)
			} else {
				me.SetChecked(n, a.State.Checked)
			}
		default:
			log.Log(TERM_EVERYTHING_WARN, "SetChecked() not supported on widget", n.WidgetType, n.String())
		}
	case widget.Show:
		if n.WidgetType == widget.Table {
			t, err := loadTable(&a)
			if err != nil {
				log.Info("unmarshal data failed", err)
				return
			}
			if t == nil {
				log.Info("unmarshal data failed table == nil")
			} else {
				me.ShowTable(nil)
			}
		} else {
			n.State.Hidden = false
			me.Show(n)
		}
	case widget.Hide:
		n.State.Hidden = true
		me.Hide(n)
		log.Info("tree: doing hide here on", a.WidgetId, n.WidgetType)
	case widget.Enable:
		n.State.Enable = true
		me.Enable(n)
	case widget.Disable:
		n.State.Enable = false
		me.Disable(n)
	case widget.Delete:
		if me.Hide == nil {
			log.Info("toolkit doesn't know how to Hide() widgets")
		} else {
			me.Hide(n)
		}
		me.DeleteNode(n)
		// now remove the child from the parent
	case widget.ToolkitClose:
		log.Info("tree.ToolkitClose()")
		me.ToolkitClose()
	default:
		log.Log(TERM_EVERYTHING_WARN, "tree.Action() unknown action", a.ActionType, "on wId", a.WidgetId)
		// me.NodeAction(n, a.ActionType)
	}
}

func loadTable(a *widget.Action) (*guipb.Tables, error) {
	var t *guipb.Tables
	err := t.Unmarshal(a.TablePB)
	/*
		test := NewRepos()
		if test.Uuid != all.Uuid {
			log.Log(WARN, "uuids do not match", test.Uuid, all.Uuid)
			deleteProtobufFile(cfgname)
		}
		if test.Version != all.Version {
			log.Log(WARN, "versions do not match", test.Version, all.Version)
			deleteProtobufFile(cfgname)
		}
	*/
	// log.Log(INFO, cfgname, "protobuf versions and uuid match", all.Uuid, all.Version)
	return t, err
}
