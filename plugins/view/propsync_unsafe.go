package view

// Deprecated: UnsafePropSync 访问属性同步内部函数
func UnsafePropSync(ps IPropSync) _UnsafePropSync {
	return _UnsafePropSync{
		IPropSync: ps,
	}
}

type _UnsafePropSync struct {
	IPropSync
}

// Sync 同步
func (ps _UnsafePropSync) Sync(revision int64, op string, args ...any) {
	ps.sync(revision, op, args...)
}
