package framework

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/memcached/pkg/controller"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) GetDatabasePod(meta metav1.ObjectMeta) (*core.Pod, error) {
	pods, err := f.kubeClient.CoreV1().Pods(meta.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, meta.Name) {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("no pod found for memcache: %s", meta.Name)
}

func (f *Framework) GetMemcacheClient(meta metav1.ObjectMeta) (*memcache.Client, error) {
	clusterIP := net.IP{192, 168, 99, 100} //minikube ip

	pod, err := f.GetDatabasePod(meta)
	if err != nil {
		return nil, err
	}

	if pod.Spec.NodeName != "minikube" {
		node, err := f.kubeClient.CoreV1().Nodes().Get(pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		for _, addr := range node.Status.Addresses {
			if addr.Type == core.NodeExternalIP {
				clusterIP = net.ParseIP(addr.Address)
				break
			}
		}
	}

	svc, err := f.kubeClient.CoreV1().Services(f.Namespace()).Get(meta.Name+"-test-svc", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	nodePort := strconv.Itoa(int(svc.Spec.Ports[0].NodePort))
	address := fmt.Sprintf(clusterIP.String() + ":" + nodePort)

	return memcache.New(address), nil
}

func (f *Framework) EventuallySetItem(meta metav1.ObjectMeta, key, value string) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			client, err := f.GetMemcacheClient(meta)
			Expect(err).NotTo(HaveOccurred())

			err = client.Set(&memcache.Item{Key: key, Value: []byte(value)})
			if err != nil {
				return false
			}
			return true
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyGetItem(meta metav1.ObjectMeta, key string) GomegaAsyncAssertion {
	return Eventually(
		func() string {
			client, err := f.GetMemcacheClient(meta)
			Expect(err).NotTo(HaveOccurred())

			item, err := client.Get(key)
			if err != nil {
				return ""
			}
			return string(item.Value)
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Invocation) EventuallyConfigSourceVolumeMounted(meta metav1.ObjectMeta) GomegaAsyncAssertion {

	return Eventually(
		func() bool {
			pod, err := f.GetDatabasePod(meta)
			if err != nil {
				return false
			}

			for _, c := range pod.Spec.Containers {
				if c.Name == api.ResourceSingularMemcached {
					for _, vm := range c.VolumeMounts {
						if vm.Name == controller.CONFIG_SOURCE_VOLUME {
							return true
						}
					}
				}
			}
			return false
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyMemcachedConfigs(meta metav1.ObjectMeta) GomegaAsyncAssertion {

	return Eventually(
		func() string {

			// TODO
			ret := make([]string, 0)
			return strings.Join(ret, " ")
		},
		time.Minute*5,
		time.Second*5,
	)
}