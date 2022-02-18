package api

import (
	"net"
	"errors"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/auth"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/dataaccess"
	"github.com/Olling/Enrolld/utils/objects"
)

func prepare(w http.ResponseWriter, r *http.Request) (requestIP string, workerID string, err error) {
	requestIP, _, err = net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		slog.PrintError("Failed to get request IP")
		return "","",err
	}
	
	if !auth.CheckAccess(w,r, "enroll", objects.Server{}) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintDebug("Unauthorized call to jobqueue from", requestIP)
		return "","", errors.New("Unauthorized access")
	}

	params := mux.Vars(r)
	workerID = params["workerid"]
	
	if workerID == "" {
		http.Error(w, http.StatusText(406), 406)
		return "","", errors.New("WorkerID not found")
	}

	return requestIP, workerID, nil
}


func getJob(w http.ResponseWriter, r *http.Request) {
	requestIP,workerID, err := prepare(w,r)
	if err != nil {
		slog.PrintError("Failed to get information from request", requestIP, workerID, err)
		return		
	}

	job,err := dataaccess.GetJob(workerID)

	if !auth.CheckAccess(w,r, "enroll", job.Server ) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to jobqueue (GetJob)", job.Server.ServerID, "from", requestIP)

		dataaccess.StopJob(job)
		return
	}

	if err != nil {
		w.WriteHeader(204)
		return
	}

	json,err := utils.StructToJson(job)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		slog.PrintError("Failed to converting job to json", job, "from", requestIP)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(202)
	w.Write([]byte(json))
}


func finishJob(w http.ResponseWriter, r *http.Request) {
	requestIP, workerID, err := prepare(w,r)
	if err != nil {
		slog.PrintError("Failed to get information from request", requestIP, workerID, err)
		return		
	}

	params := mux.Vars(r)
	jobID := params["jobid"]

	job := dataaccess.GetRunningJobInfo(jobID)
	if !auth.CheckAccess(w,r, "enroll", job.Server) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to jobqueue from", requestIP)
		return
	}

	if job.WorkerID != workerID {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Worker trying to close wrong job", requestIP)
		return
	}

	err = dataaccess.CompleteJob(workerID, jobID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		slog.PrintError("Worker trying to close job that is not active (timeout?)", requestIP, err)
		return
	}

	slog.PrintInfo("The worker", job.WorkerID, "successfully enrolled the following server:", job.Server.ServerID)
	slog.PrintDebug(job)
}


func rescueJob(w http.ResponseWriter, r *http.Request) {
	requestIP, workerID, err := prepare(w,r)
	if err != nil {
		slog.PrintError("Failed to get information from request", requestIP, workerID, err)
		return		
	}

	params := mux.Vars(r)
	jobID := params["jobid"]

	job := dataaccess.GetRunningJobInfo(jobID)
	if !auth.CheckAccess(w,r, "enroll", job.Server) {
		http.Error(w, http.StatusText(401), 401)
		slog.PrintError("Unauthorized call to jobqueue from", requestIP)
		return
	}

	err = dataaccess.StopJobID(jobID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		slog.PrintError("Worker trying to rescue a job that is not active (timeout?)", requestIP, err)
		return
	}
}
