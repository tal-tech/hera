package main

import (
	"fmt"
	"github.com/tal-tech/hera/clijob"
	"time"
)

var jobFuncs map[string]clijob.Job

func GetFuncs() map[string]clijob.Job {
	return jobFuncs
}

func init() {
	jobFuncs = make(map[string]clijob.Job)
	initFuncMap()
}

func initFuncMap() {
	//example
	jobFuncs["testJob"] = clijob.Job{
		Name: "testJob",
		Task: func() error {
			for i := 0; i < 10; i++ {
				fmt.Println("testJob")
				time.Sleep(time.Second)
			}
			return nil
		},
	}

	jobFuncs["testJob2"] = clijob.Job{
		Name: "testJob2",
		Task: func() error {
			for i := 0; i < 10; i++ {
				fmt.Println("testJob2")
				time.Sleep(time.Second)
			}
			return nil
		},
	}
}
