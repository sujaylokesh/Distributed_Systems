package mapreduce

import "container/list"
import "fmt"




type WorkerInfo struct {
	address string
	// You can add definitions here.

}

// Clean up all workers by sending a Shutdown RPC to each one of them Collect
// the number of jobs each work has performed.
func (mr *MapReduce) KillWorkers() *list.List {
	l := list.New()
	for _, w := range mr.Workers {
		DPrintf("DoWork: shutdown %s\n", w.address)
		args := &ShutdownArgs{}
		var reply ShutdownReply
		ok := call(w.address, "Worker.Shutdown", args, &reply)
		if ok == false {
			fmt.Printf("DoWork: RPC %s shutdown error\n", w.address)
		} else {
			l.PushBack(reply.Njobs)
		}
	}
	return l
}

func (mr *MapReduce) RunMaster() *list.List {
	mapc := make(chan interface{})
	reducec := make(chan interface{})
	job_allotment := func(jobnumber int, operation string, num int){
		for {
			worker := <-mr.registerChannel
			args := DoJobArgs{
				File: mr.file,
				Operation: JobType(operation),
				JobNumber: jobnumber,
				NumOtherPhase: num,
			   }
			   reply := DoJobReply{}
			   rpc := call(worker, "Worker.DoJob", args, &reply)
			   if operation == Map{
					if rpc { 
						mapc<-args
						mr.registerChannel <- worker
						break
						}
				}else{
					if rpc {
						reducec<-args
						mr.registerChannel <- worker
						break
					}
				}
			}
		}
	// Your code here
	for i := 0; i < mr.nMap;{
		go job_allotment(i,Map,mr.nReduce)
		<-mapc
		i++
	}
	close(mapc)

	for i := 0; i < mr.nReduce;{
		go job_allotment(i,Reduce,mr.nMap)
		<-reducec
		i++
	}
	close(reducec)
	return mr.KillWorkers()
}