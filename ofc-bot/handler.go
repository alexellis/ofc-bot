package function

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/openfaas/openfaas-cloud/sdk"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

	if len(input) > 0 {
		q, err := url.ParseQuery(string(input))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if cmd := q.Get("command"); len(cmd) > 0 {
			if cmd == "/users" {

				getPath := os.Getenv("gateway_host") + "/system/functions"
				req, _ := http.NewRequest(http.MethodGet, getPath, nil)
				secret, err := sdk.ReadSecret("basic-auth-password")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				req.SetBasicAuth("admin", secret)

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				functions, err := readFunctions(res)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				out := ""
				for _, function := range functions {
					out = out + function.Name + "\n"
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(out))
				return
			}
		}
	}

	http.Error(w, "Nothing to do", http.StatusBadRequest)
}

func readFunctions(res *http.Response) ([]function, error) {
	functions := []function{}

	if res.Body != nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		marshalErr := json.Unmarshal(body, &functions)
		if marshalErr != nil {
			return nil, marshalErr
		}
	}

	return functions, nil
}

type function struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	InvocationCount float64           `json:"invocationCount"`
	Replicas        uint64            `json:"replicas"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
}
