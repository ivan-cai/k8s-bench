package main

import (
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/ivan-cai/k8s-bench/common"
	"github.com/ivan-cai/k8s-bench/k8s_client"
	"github.com/ivan-cai/k8s-bench/utils"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	VERSION = "0.0.0"
)

// some flag is repeatable, e.g. headers
type arrayFlags []string

type InputParameters struct {
	requestNum     int
	concurrencyNum int
	certFile       string
	postFile       string
	contentType    string
	headers        arrayFlags
	kubeConfigFile string
}

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func validateParameters(param *InputParameters, invalidHeadersNum int) bool {

	if param.requestNum < param.concurrencyNum {
		fmt.Println("Cannot use concurrency level greater than total number of requests")
		return false
	}

	if param.certFile != "" {
		if err := utils.PathExist(param.certFile); err != nil {
			fmt.Printf("Could not open certificate file(%v): (%v)", param.certFile, err)
			return false
		}
	}

	if param.postFile != "" {
		if err := utils.PathExist(param.postFile); err != nil {
			fmt.Printf("Could not open post file(%v): (%v)", param.certFile, err)
			return false
		}
	}

	if param.kubeConfigFile != "" {
		if err := utils.PathExist(param.kubeConfigFile); err != nil {
			fmt.Printf("Could not open kube config file(%v): (%v)", param.kubeConfigFile, err)
			return false
		}
	}

	if invalidHeadersNum != 0 {
		fmt.Println("There are some invalid headers")
		return false
	}

	return true
}

func main() {
	// add help
	var h bool
	flag.BoolVar(&h, "h", false, "this help")

	// version
	var version bool
	flag.BoolVar(&version, "version", false, "print k8s-bench version and exit")

	var parameters InputParameters
	// get command line parameters
	flag.IntVar(&parameters.requestNum, "n", 1, "Number of requests to perform")
	flag.IntVar(&parameters.concurrencyNum, "c", 1, "Number of multiple requests to make at a time")
	flag.StringVar(&parameters.certFile, "E", "", "When connecting to an SSL website, use the provided client certificate in PEM format to authenticate with the server.")
	flag.StringVar(&parameters.postFile, "p", "", "File containing data to POST. If you want to do benchmark for k8s with default example pod, you can set 'default'. If you want to do benchmark for something but other than k8s, remember to also set -T.")
	flag.StringVar(&parameters.contentType, "T", "application/json", "Content-type header to use for POST/PUT data.")
	flag.StringVar(&parameters.kubeConfigFile, "K", "", "File for building a Clientset which can communicate with kubernetes cluster.")
	flag.Var(&parameters.headers, "H", "Add Arbitrary header line, eg. 'Accept-Encoding: gzip' Inserted after all normal header lines. (repeatable)")

	flag.Parse()

	if h || len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	if version {
		fmt.Printf("k8s-bench version is %v\n", VERSION)
		return
	}

	invalidHeaderNum := 0
	headers := make(map[string]string)
	for _, oneHeader := range parameters.headers {
		h := strings.Split(oneHeader, ":")
		if len(h) != 2 {
			invalidHeaderNum += 1
			continue
		}

		headers[h[0]] = h[1]
	}

	if !validateParameters(&parameters, invalidHeaderNum) {
		return
	}

	request := os.Args[len(os.Args)-1]
	if !strings.HasPrefix(request, "http") {
		fmt.Println("Invalid request")
		return
	}

	isK8sRequest := false
	if parameters.kubeConfigFile != "" {
		isK8sRequest = true
	}

	isPostRequest := false
	var postData []byte
	if parameters.postFile != "" {
		isPostRequest = true

		podByte, err := ioutil.ReadFile(parameters.postFile)
		if err != nil {
			fmt.Printf("There is no postFile: %v\n", err)
			return
		}

		postData = podByte
	}

	var wg sync.WaitGroup
	mean := parameters.requestNum / parameters.concurrencyNum
	remainder := parameters.requestNum % parameters.concurrencyNum
	routineNum := 0
	start := time.Now().UnixNano()
	for routineNum < parameters.concurrencyNum {
		wg.Add(1)
		reqNumPerThread := mean
		if routineNum < remainder {
			reqNumPerThread = mean + 1
		}

		// kubernetes create pod benchmark
		if isPostRequest && isK8sRequest {
			pod := k8s_client.GetExamplePod()
			if parameters.postFile != "default" {
				if err := yaml.Unmarshal(postData, pod); err != nil {
					fmt.Printf("yaml Unmarshal to Pod failed: %v\n", err)
					return
				}
			}

			go func(pod *corev1.Pod, num int) {
				defer wg.Done()
				if err := k8s_client.BatchCreatePodHandler("default", parameters.kubeConfigFile, num, pod); err != nil {
					return
				}
			}(pod, reqNumPerThread)
		}

		routineNum++
	}
	wg.Wait()

	end := time.Now().UnixNano()
	spentMillisecond := (end - start) / 1000000

	totalRequests := common.SuccessNum + common.FailNum
	fmt.Printf("Request: %v\n", request)
	fmt.Printf("Concurrency Level: %v\n", parameters.concurrencyNum)
	fmt.Printf("Time taken for tests: %v seconds\n", float64(spentMillisecond)/float64(1000))
	fmt.Printf("Complete requests: %v\n", totalRequests)
	fmt.Printf("Failed requests: %v\n", common.FailNum)
	fmt.Printf("Requests per second: %v\n", totalRequests*1000/spentMillisecond)
}
