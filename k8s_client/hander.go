package k8s_client

import (
	"fmt"
	"github.com/ivan-cai/k8s-bench/common"
	"k8s.io/api/core/v1"
	"math/rand"
	"sync/atomic"
	"time"
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func BatchCreatePodHandler(namespace string, kubeConfigFile string, taskNum int, originalPod *v1.Pod) error {
	cli := KubeClient{}
	if err := cli.CreateClient(kubeConfigFile); err != nil {
		fmt.Printf("k8s client create failed: %v\n", err)
		return err
	}

	pod := originalPod
	count := 0
	for count < taskNum {
		pod.Name = pod.Name + "_" + RandStringRunes(8)

		if err := cli.CreatePod(namespace, pod); err != nil {
			atomic.AddInt64(&common.FailNum, 1)
			continue
		}
		atomic.AddInt64(&common.SuccessNum, 1)
	}

	return nil
}
