package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func writeRes(w http.ResponseWriter, content interface{}) {
	resJson, err := json.Marshal(content)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resJson)
}

func writeErrRes(w http.ResponseWriter, err error) {
	errJson, err := json.Marshal(ErrRes{err.Error()})
	if err != nil {
		return
	}
	w.Header().Set("Content-type", "Application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(errJson)
}

func readReq(r *http.Request, reqBody interface{}) error {
	reqBodyJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body. %s", err.Error())
	}
	defer r.Body.Close()
	err = json.Unmarshal(reqBodyJson, reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal request body. %s", err.Error())
	}

	return nil
}

func readRes(r *http.Response, reqBody interface{}) error {
	reqBodyJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body. %s", err.Error())
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJson, reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response body. %s", err.Error())
	}

	return nil
}
