package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/zap"
)

var err error

/*
type victoriaMetricsAnswerStruct []struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Leaf int    `json:"is_leaf"`
}
*/

type victoriaMetricsAnswerStruct struct {
	Metrics []struct {
		Path   string `json:"path"`
		Name   string `json:"-"`
		IsLeaf int    `json:"is_leaf"`
	} `json:"metrics"`
}

type graphiteStruct struct {
	IsLeaf    bool              `json:"is_leaf"`
	Path      string            `json:"path"`
	Intervals graphiteIntervals `json:"intervals,omitempty"`
}

type graphiteIntervals [][]int

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
	modifiedURL, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Can't parse URL.", zap.Error(err))
		return
	}
	modifiedURLValues, err := url.ParseQuery(modifiedURL.RawQuery)
	modifiedURLValues.Set("format", "completer")
	modifiedURL.RawQuery = modifiedURLValues.Encode()
	victoriaMetricsRequestURL := config.VictoriaMetrics + modifiedURL.String()
	victoriaMetricsAnswer := victoriaMetricsAnswerStruct{}
	logger.Debug("finalVictoriaMetricsURL", zap.String("URL", victoriaMetricsRequestURL))
	victoriaMetricsRAWAnswer, statusCode, err := doRequest(victoriaMetricsRequestURL)
	if err != nil {
		logger.Error("doRequest failed.", zap.Error(err))
		w.WriteHeader(statusCode)
		return
	}

	err = json.Unmarshal(victoriaMetricsRAWAnswer, &victoriaMetricsAnswer)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("unmarshal response failed.", zap.Error(err), zap.String("response", string(victoriaMetricsRAWAnswer)))
		return
	}
	graphiteRStruct := makeGraphiteAnswer(&victoriaMetricsAnswer)
	// graphiteRecord, _ := msgpack.Marshal(graphiteRStruct)

	/*
		file, _ := os.OpenFile(
			"/tmp/asd",
			os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
			0644,
		)
		defer file.Close()
		msgpack.NewEncoder(file).UseJSONTag(true).Encode(graphiteRStruct)
	*/
	logger.Debug("Debug jsonResp", zap.String("victoriaMetricsAnswerJsonResp", string(victoriaMetricsRAWAnswer)))
	w.Header().Set("content-type", "application/x-msgpack")
	w.WriteHeader(statusCode)
	msgpack.NewEncoder(w).UseJSONTag(true).Encode(graphiteRStruct)
}

func doRequest(URL string) ([]byte, int, error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return []byte{}, 0, err
	}
	resp, err := client.Do(request)

	if err != nil {
		return []byte{}, resp.StatusCode, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, resp.StatusCode, err
	}
	return body, resp.StatusCode, err
}
