package acl

import "git.golaxy.org/core/define"

var (
	self      = define.ServicePlugin(newACL)
	Name      = self.Name
	Using     = self.Using
	Install   = self.Install
	Uninstall = self.Uninstall
)
