package bootstrap

type FuncSetter struct {
	beforeFuncs []BeforeServerStartFunc
	afterFuncs  []AfterServerStopFunc
}

func NewFuncSetter()*FuncSetter{
	return &FuncSetter{}
}

func (fs *FuncSetter) AddBeforeServerStartFunc(fns ...BeforeServerStartFunc) {
	for _, fn := range fns {
		fs.beforeFuncs = append(fs.beforeFuncs, fn)
	}
}

func (fs *FuncSetter) AddAfterServerStopFunc(fns ...AfterServerStopFunc){
	for _, fn := range fns {
		fs.afterFuncs = append(fs.afterFuncs, fn)
	}
}

func (fs *FuncSetter) RunBeforeServerStartFunc() error{
	for _, fn := range fs.beforeFuncs {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FuncSetter) RunAfterServerStopFunc(){
	for _, fn := range fs.afterFuncs {
		fn()
	}
}