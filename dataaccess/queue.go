package dataaccess

import (
	"sync"
	"time"
	"errors"
	"container/list"
	"github.com/google/uuid"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/metrics"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)


var (
	jobQueue = list.New()
	workingQueue 	map[string]objects.Job
	queueMutex	sync.Mutex
)


func InitializeQueue () {
	slog.PrintDebug("Initializing queue")
	workingQueue = make(map[string]objects.Job)
	metrics.JobQueueCount = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Subsystem: "enrolld",
		Name:      "jobqueue_count",
		Help:      "The number of jobs waiting in the queue",
	}, GetJobQueueLength)

	metrics.WorkingQueueCount = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Subsystem: "enrolld",
		Name:      "workingqueue_count",
		Help:      "The number of jobs currently marked as being worked on",
	}, GetWorkingQueueLength)
}


func AddJob(server objects.Server) (error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	slog.PrintDebug("API call to AddJob registered")
	slog.PrintTrace("Information:", server)

	//TODO Verify Server object?

	inQueue,active,workerID := GetStatus(server.ServerID)

	if inQueue && !active {
		slog.PrintDebug("Tried to add server", server.ServerID, " that is already in the queue")
		return errors.New("The server is already in the job queue")
	}

	if inQueue && active {
		slog.PrintDebug("Tried to add server", server.ServerID, " that was already being worked on by:", workerID)
		return errors.New("The server is currently being worked on")
	}

	var job objects.Job
	job.Server = server
	job.Timestamp = time.Now()

	slog.PrintTrace("AddJob added the following server to the jobQueue:", server)
	jobQueue.PushBack(job)

	slog.PrintInfo(server.ServerID, "have been added to the enroll queue")

	return nil
}


func GetJob(workerID string) (job objects.Job, err error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	slog.PrintDebug("Queue GetJob")
	slog.PrintTrace(workerID)

	if jobQueue.Len() == 0 {
		return job, errors.New("There are no jobs to get")
	}

	element := jobQueue.Front()
	job = element.Value.(objects.Job)
	job.WorkerID = workerID

	jobID := uuid.NewString()
	job.JobID = jobID

	if _,exists := workingQueue[jobID]; exists {
		slog.PrintError("Got duplicate UUID - Retrying")
		queueMutex.Unlock()
		return GetJob(workerID)
	}

	workingQueue[jobID] = job
	jobQueue.Remove(element)
	slog.PrintDebug("The server", job.Server.ServerID, "was assigned to", workerID, ". JobID:",jobID)

	return job,nil
}


func GetStatus(serverID string) (inQueue bool, active bool, workerID string) {
	for element := jobQueue.Front(); element != nil; element = element.Next() {
		if element.Value.(objects.Job).Server.ServerID == serverID {
			return true, false, ""
		}
	}

	for _,value := range workingQueue {
		if value.Server.ServerID == serverID {
			return true, true, value.WorkerID	
		}
	}

	return  false, false, ""
}


func CompleteJob(workerID string, jobID string) (error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if _,exists := workingQueue[jobID]; !exists {
		return errors.New("Cannot complete the job " + jobID + "as it is not active. Requested by " +  workerID)
	}

	server := workingQueue[jobID].Server
	delete(workingQueue, jobID)

	return UpdateServer(server)
}


func StopJobID(jobID string) (error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if _,exists := workingQueue[jobID]; !exists {
		return errors.New("Cannot stop the job " + jobID + "as it is not active")
	}

	server := workingQueue[jobID].Server
	delete(workingQueue, jobID)

	return AddJob(server)
}

func GetRunningJobInfo(jobID string) (objects.Job) {
	return workingQueue[jobID]
}


func StopJob(job objects.Job) (error) {
	return StopJobID(job.JobID)
}


func GetWorkingQueueLength() (float64) {
	return float64(len(workingQueue))
}


func GetJobQueueLength() (float64) {
	return float64(jobQueue.Len())
}

func IsServerIDActive(serverID string) bool {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	inQueue,active,_ := GetStatus(serverID)

	if inQueue || active {
		return true
	}

	return false
}

func IsServerActive(server objects.Server) bool {
	return IsServerIDActive(server.ServerID)
}
