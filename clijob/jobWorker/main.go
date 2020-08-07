package main

import (
	"fmt"
	"github.com/tal-tech/hera/clijob"
)

func main() {

	job := clijob.NewJobServer()
	job.AddJobs(GetFuncs())

	err := job.Start()
	if err != nil {
		fmt.Println("err:", err)
	}
}
