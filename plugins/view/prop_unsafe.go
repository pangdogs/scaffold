package view

// Deprecated: UnsafeProp 访问属性内部函数
func UnsafeProp(p IProp) _UnsafeProp {
	return _UnsafeProp{
		IProp: p,
	}
}

type _UnsafeProp struct {
	IProp
}

// IncrRevision 自增版本号
func (p _UnsafeProp) IncrRevision() int64 {
	return p.incrRevision()
}
