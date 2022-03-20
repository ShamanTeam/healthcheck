package healthcheck

import (
	"encoding/json"
	"net/http"
)

type HealthCheck struct {
	Version  string
	AppName  string
	checkers []Checker
}

type Checker func() CheckResult

type CheckResult struct {
	Service string `json:"service"`
	Status  bool   `json:"status"`
}

func (h *HealthCheck) AddChecker(c Checker) {
	h.checkers = append(h.checkers, c)
}

func (h *HealthCheck) Check() ([]byte, bool, error) {
	var checks []CheckResult

	for _, checker := range h.checkers {
		checks = append(checks, checker())
	}

	checkResults := map[string]interface{}{
		"version":     h.Version,
		"application": h.AppName,
		"checks":      checks,
	}

	problem := h.checkProblem(checks)

	result, err := json.Marshal(checkResults)

	if err != nil {
		return nil, problem, nil
	}

	return result, problem, err
}

func (h *HealthCheck) checkProblem(checks []CheckResult) bool {
	for _, check := range checks {
		if check.Status == false {
			return true
		}
	}

	return false
}

func Handler(check *HealthCheck) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		result, problem, err := check.Check()

		if problem || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err != nil {
			w.Write(result)
		} else {
			w.Write([]byte(err.Error()))
		}
	})
}
