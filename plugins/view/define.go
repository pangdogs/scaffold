package view

import "git.golaxy.org/core/define"

var (
	self      = define.RuntimePlugin(newPropView)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
