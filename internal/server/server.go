// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2021: Nordix Foundation
//   %%
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//   ========================LICENSE_END===================================
//

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/dmaapmediatorproducer/internal/jobs"
)

const HealthCheckPath = "/health_check"
const AddJobPath = "/info_job"
const jobIdToken = "infoJobId"
const deleteJobPath = AddJobPath + "/{" + jobIdToken + "}"
const logLevelToken = "level"
const logAdminPath = "/admin/log"

type ErrorInfo struct {
	// A URI reference that identifies the problem type.
	Type string `json:"type" swaggertype:"string"`
	// A short, human-readable summary of the problem type.
	Title string `json:"title" swaggertype:"string"`
	// The HTTP status code generated by the origin server for this occurrence of the problem.
	Status int `json:"status" swaggertype:"integer" example:"400"`
	// A human-readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail" swaggertype:"string" example:"Info job type not found"`
	// A URI reference that identifies the specific occurrence of the problem.
	Instance string `json:"instance" swaggertype:"string"`
} // @name ErrorInfo

type ProducerCallbackHandler struct {
	jobsManager jobs.JobsManager
}

func NewProducerCallbackHandler(jm jobs.JobsManager) *ProducerCallbackHandler {
	return &ProducerCallbackHandler{
		jobsManager: jm,
	}
}

func NewRouter(jm jobs.JobsManager, hcf func(http.ResponseWriter, *http.Request)) *mux.Router {
	callbackHandler := NewProducerCallbackHandler(jm)
	r := mux.NewRouter()
	r.HandleFunc(HealthCheckPath, hcf).Methods(http.MethodGet).Name("health_check")
	r.HandleFunc(AddJobPath, callbackHandler.addInfoJobHandler).Methods(http.MethodPost).Name("add")
	r.HandleFunc(deleteJobPath, callbackHandler.deleteInfoJobHandler).Methods(http.MethodDelete).Name("delete")
	r.HandleFunc(logAdminPath, callbackHandler.setLogLevel).Methods(http.MethodPut).Name("setLogLevel")
	r.NotFoundHandler = &notFoundHandler{}
	r.MethodNotAllowedHandler = &methodNotAllowedHandler{}
	return r
}

// @Summary      Add info job
// @Description  Callback for ICS to add an info job
// @Tags         Data producer (callbacks)
// @Accept       json
// @Param        user  body  jobs.JobInfo  true  "Info job data"
// @Success      200
// @Failure      400  {object}  ErrorInfo     "Problem as defined in https://tools.ietf.org/html/rfc7807"
// @Header       400  {string}  Content-Type  "application/problem+json"
// @Router       /info_job [post]
func (h *ProducerCallbackHandler) addInfoJobHandler(w http.ResponseWriter, r *http.Request) {
	b, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		returnError(fmt.Sprintf("Unable to read body due to: %v", readErr), w)
		return
	}
	jobInfo := jobs.JobInfo{}
	if unmarshalErr := json.Unmarshal(b, &jobInfo); unmarshalErr != nil {
		returnError(fmt.Sprintf("Invalid json body. Cause: %v", unmarshalErr), w)
		return
	}
	if err := h.jobsManager.AddJobFromRESTCall(jobInfo); err != nil {
		returnError(fmt.Sprintf("Invalid job info. Cause: %v", err), w)
		return
	}
}

// @Summary      Delete info job
// @Description  Callback for ICS to delete an info job
// @Tags         Data producer (callbacks)
// @Param        infoJobId  path  string  true  "Info job ID"
// @Success      200
// @Router       /info_job/{infoJobId} [delete]
func (h *ProducerCallbackHandler) deleteInfoJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars[jobIdToken]
	if !ok {
		returnError("Must provide infoJobId.", w)
		return
	}

	h.jobsManager.DeleteJobFromRESTCall(id)
}

// @Summary      Set log level
// @Description  Set the log level of the producer.
// @Tags         Admin
// @Param        level  query  string  false  "string enums"  Enums(Error, Warn, Info, Debug)
// @Success      200
// @Failure      400  {object}  ErrorInfo     "Problem as defined in https://tools.ietf.org/html/rfc7807"
// @Header       400  {string}  Content-Type  "application/problem+json"
// @Router       /admin/log [put]
func (h *ProducerCallbackHandler) setLogLevel(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	logLevelStr := query.Get(logLevelToken)
	if loglevel, err := log.ParseLevel(logLevelStr); err == nil {
		log.SetLevel(loglevel)
	} else {
		returnError(fmt.Sprintf("Invalid log level: %v. Log level will not be changed!", logLevelStr), w)
		return
	}
}

type notFoundHandler struct{}

func (h *notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 not found.", http.StatusNotFound)
}

type methodNotAllowedHandler struct{}

func (h *methodNotAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
}

func returnError(msg string, w http.ResponseWriter) {
	errInfo := ErrorInfo{
		Status: http.StatusBadRequest,
		Detail: msg,
	}
	w.Header().Add("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errInfo)
}
