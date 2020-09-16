package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hydrogen18/stalecucumber"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var err error

type victoriaMetricsAnswer []struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	AllowChildren int    `json:"allowChildren"`
	Expandable    int    `json:"expandable"`
	Leaf          int    `json:"leaf"`
}

func main() {
	makeConfig()
	makeLogger()

	prometheus.MustRegister(prometheusRequestsTotalCounter)
	prometheus.MustRegister(prometheusRequestsDuration)

	gorillaMux := mux.NewRouter()
	gorillaMux.PathPrefix("/").HandlerFunc(goDo)
	gorillaMux.Use(loggerMiddleware())
	logger.Info("Starting app.")
	http.ListenAndServe(config.AppPort, gorillaMux)
}

func goDo(w http.ResponseWriter, r *http.Request) {
	victoriaMetricsRequestURL := config.VictoriaMetrics + r.URL.String()
	victoriaMetricsAnswer := victoriaMetricsAnswer{}
	victoriaMetricsRAWAnswer, err := doRequest(victoriaMetricsRequestURL)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("doRequest failed.", zap.Error(err))
		return
	}

	err = json.Unmarshal(victoriaMetricsRAWAnswer, &victoriaMetricsAnswer)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("unmarshal response failed.", zap.Error(err), zap.String("response", string(victoriaMetricsRAWAnswer)))
		return
	}
	stalecucumber.NewPickler(w).Pickle(victoriaMetricsAnswer)
}

func doRequest(URL string) ([]byte, error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return []byte{}, err
	}
	resp, err := client.Do(request)

	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}
