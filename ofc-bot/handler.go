package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/openfaas/openfaas-cloud/sdk"
)

const owner = "com.openfaas.cloud.git-owner"

type function struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	InvocationCount float64           `json:"invocationCount"`
	Replicas        uint64            `json:"replicas"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

	token, err := sdk.ReadSecret("token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var query *url.Values
	if len(input) > 0 {
		q, err := url.ParseQuery(string(input))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query = &q
	}

	if token != query.Get("token") {
		http.Error(w, fmt.Sprintf("Token: %s, invalid", query.Get("token")), http.StatusUnauthorized)
		return
	}

	command := query.Get("command")
	text := query.Get("text")

	os.Stderr.Write([]byte(fmt.Sprintf("debug - command: %q, text: %q\n", command, text)))

	headerWritten := processCommand(w, r, command, text)

	if !headerWritten {
		http.Error(w, "Nothing to do", http.StatusBadRequest)
	}
}

func processCommand(w http.ResponseWriter, r *http.Request, command, text string) bool {
	if len(command) > 0 {

		switch command {
		case "/metrics":
			if len(text) == 0 {
				w.Write([]byte("Please give a function name with this slash command"))
				w.WriteHeader(http.StatusOK)
				return true
			}

			res, err := queryStats(text, time.Hour*24)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return true
			}

			var body []byte
			if res.Body != nil {
				defer res.Body.Close()
				body, _ = ioutil.ReadAll(res.Body)
			}

			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return true

		case "/functions":
			res, err := doFunctionsQuery()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return true
			}

			functions, err := readFunctions(res)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return true
			}

			var username string
			if len(text) > 0 {
				username = text
			}

			out := ""
			list := makeFunctions(&functions, username)
			if len(list) > 0 {
				out = "Functions"
				if len(username) > 0 {
					out = out + " for (" + username + ")"
				}

				out = out + ":\n"

				for _, k := range list {
					out = out + "- " + k + "\n"
				}
			} else {
				out = "No functions found"
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(out))
			return true

		case "/users":
			res, err := doFunctionsQuery()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return true
			}

			functions, err := readFunctions(res)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return true
			}

			out := ""
			owners := makeOwners(&functions)
			if len(owners) > 0 {
				out = "Users:\n"
				for k := range owners {
					out = out + "- " + k + "\n"
				}
			} else {
				out = "No users found"
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(out))
			return true
		}
	}

	return false
}

func doFunctionsQuery() (*http.Response, error) {
	getPath := os.Getenv("gateway_host") + "/system/functions"
	req, _ := http.NewRequest(http.MethodGet, getPath, nil)
	secret, err := sdk.ReadSecret("basic-auth-password")
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("admin", secret)

	res, err := http.DefaultClient.Do(req)
	return res, err
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

func makeFunctions(functions *[]function, username string) []string {
	list := []string{}
	for _, function := range *functions {
		add := true

		// Filter out system functions
		owner := function.Labels[owner]
		if len(owner) > 0 {
			if len(username) > 0 {
				add = len(owner) > 0 && owner == username
			}

			if add {
				list = append(list, function.Name)
			}
		}

	}
	return list
}

func makeOwners(functions *[]function) map[string]int {
	owners := make(map[string]int)

	for _, function := range *functions {
		owner := function.Labels[owner]
		if len(owner) > 0 {
			if _, ok := owners[owner]; !ok {
				owners[owner] = 0
			}

			owners[owner] = owners[owner] + 1
		}
	}
	return owners
}

func queryStats(functionName string, window time.Duration) (*http.Response, error) {
	getPath := os.Getenv("gateway_host") + "/function/system-metrics?function=" + functionName + "&metrics_window=" + strconv.Itoa(int(window.Hours()))
	req, _ := http.NewRequest(http.MethodGet, getPath, nil)

	res, err := http.DefaultClient.Do(req)
	return res, err
}
