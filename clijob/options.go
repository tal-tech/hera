package clijob

type Options struct {

	//命令行参数解析器
	cmdParser CmdParser
}

type OptionFunc func(*Options)

func DefaultOptions() Options {
	return Options{
		cmdParser: &defaultCmdParse{},
	}
}

func OptSetCmdParser(parser CmdParser) OptionFunc {
	return func(o *Options) {
		o.cmdParser = parser
	}
}
