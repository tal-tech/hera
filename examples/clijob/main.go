package main

import (
	"fmt"
	"github.com/tal-tech/hera/clijob"
	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/flagutil"
	"time"
)

func main() {
	myJobs := make(map[string]clijob.Job)
	myJobs["rJob1"] = clijob.Job{Name: "rJob1", Task: rJob1}
	myJobs["rJob2"] = clijob.Job{Name: "rJob2", Task: rJob2}

	job := clijob.NewJobServer(clijob.OptSetCmdParser(&MyParser{}))
	job.AddJobs(myJobs)

	err := job.Start()
	if err != nil {
		fmt.Println("err:", err)
	}
}

func rJob1() error {
	for i := 0; i < 10; i++ {
		fmt.Println("job1")
		time.Sleep(time.Second)
	}
	return nil
}

func rJob2() error {
	for i := 0; i < 10000; i++ {
		fmt.Println("job2")
		time.Sleep(time.Second)
	}
	return nil
}

type MyParser struct {
}

//为了演示自定义命令行参数解析器, 此处复制了一份默认的写法
func (p *MyParser) JobArgParse(jobs map[string]clijob.Job) (selectedJobs []clijob.Job, err error) {
	fmt.Println("自定义解析器")
	cmdArg := *flagutil.GetExtendedopt()
	job, ok := jobs[cmdArg]
	if !ok {
		return nil, logger.NewError("[ " + cmdArg + " ]任务未定义")
	}

	selectedJobs = make([]clijob.Job, 0, 1)
	selectedJobs = append(selectedJobs, job)

	return selectedJobs, nil
}
