package poolkeeper

import (
	"encoding/json"
	"fmt"

	// appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
)

const (
	keepNodesAliveMarkerLabel = "poolkeeper/keepNodesAliveMarkerLabel"
)

// KeepNodesAlive tries to keep an amount of nodes alive during a configured time of day
type KeepNodesAlive struct {
	// NodeSelector specifies which nodes should be kept alive
	NodeSelector string `json:"nodeSelector"`

	// Amount
	Amount int `json:"amount"`
}

func (k *KeepNodesAlive) run(clientset *kubernetes.Clientset) {
	if k.Amount == 0 {
		return
	}

	nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: k.NodeSelector,
	})
	if err != nil {
		log.Errorf("unable to list nodes", err)
		return
	}
	nodesToKeepAlive := nodeList.Items
	log.Debugf("found %d potential node to keep alive", len(nodesToKeepAlive))

	podList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", keepNodesAliveMarkerLabel),
	})
	if err != nil {
		log.Errorf("unable to list pods", err)
		return
	}
	currentKeepAlivePods := podList.Items
	log.Debugf("found %d current keep-alive pods", len(currentKeepAlivePods))

	target := k.Amount
	v1 := clientset.CoreV1()
	if target > len(currentKeepAlivePods) {
		// find free node to schedule onto (sort by node age for it to be stable)
		targetNode := "someNodeName"
		v1.Pods("").Create(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "poolkeeper-keep-alive",
				Labels: map[string]string{
					keepNodesAliveMarkerLabel: "true",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					corev1.Container{
						Name:    "keepalive",
						Image:   "bash:latest",
						Command: []string{"bash", "-c", "while true; do sleep 600; done"},
					},
				},
				NodeSelector: map[string]string{
					"kubernetes.io/hostname": targetNode,
				},
			},
		})
	} else if target < len(currentKeepAlivePods) {
		// select one or more to drain (sort by node age for it to be stable)
	}

	// for _, deployment := range deploymentsToPatch {
	// 	_, err := appsv1.Deployments(pa.Namespace).Patch(deployment.Name, types.MergePatchType, []byte(patch))
	// 	if err != nil {
	// 		log.WithField("deployment", deployment.Name).WithError(err).Error("error patching deployment")
	// 		continue
	// 	}
	// }
}
