package k8s_client

import (
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	client *kubernetes.Clientset
}

func (kc *KubeClient) CreateClient(kubeConfigPath string) error {
	clusterConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Println("BuildConfigFromFlags error")
		return err
	}
	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		fmt.Println("NewForConfig error")
		return err
	}

	kc.client = clientSet

	return nil
}

func (kc *KubeClient) CreatePod(namespace string, pod *v1.Pod) error {
	if _, err := kc.client.CoreV1().Pods(namespace).Create(pod); err != nil {
		fmt.Printf("createPod failed %v: %v\n", pod, err)
		return err
	}

	return nil
}

func (kc *KubeClient) ListPods(opts metav1.ListOptions) (*v1.PodList, error) {
	pods, err := kc.client.CoreV1().Pods("default").List(opts)
	if err != nil {
		return nil, err
	}

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	return pods, nil
}

func (kc *KubeClient) GetPod(namespace, pod string) (*v1.Pod, error) {
	res, err := kc.client.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		return nil, err
	}

	return res, nil
}

func GetExamplePod() *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
				},
			},
		},
	}
}
