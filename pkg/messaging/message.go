package messaging

import (
	"bytes"
	"github.com/onsi/gomega"
	"github.com/rh-messaging/shipshape/pkg/framework"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// generateMessagingFilesConfigMap creates a new config map that holds messaging
// files to be used by QE clients. It generates a 1kb, 100kb and 500kb files.
// Note: there is a threshold on the ConfigMap size of 1mb. If larger messages are
//       needed within QE Clients, we should change container init strategy when
//       defining the pod, so it downloads files during initialization (not sure if
//       that is a good idea.
func GenerateSmallMediumLargeMessagesConfigMap(framework *framework.Framework, configMapName string) *v1.ConfigMap {
	var err error
	ctx := framework.GetFirstContext()
	configMap, err := ctx.Clients.KubeClient.CoreV1().ConfigMaps(ctx.Namespace).Create(&v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name: configMapName,
		},
		Data: map[string]string{
			"small-message.txt":  GenerateMessageContent("ThisIsARepeatableMessage", 1024),
			"medium-message.txt": GenerateMessageContent("ThisIsARepeatableMessage", 1024*100),
			"large-message.txt":  GenerateMessageContent("ThisIsARepeatableMessage", 1024*500),
		},
	})
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(configMap).NotTo(gomega.BeNil())
	return configMap
}

// GenerateMessageContent helps generating a string of a
// given size based on the provided pattern. It can be used
// to generate large message content with predictable result.
func GenerateMessageContent(pattern string, size int) string {
	var buf bytes.Buffer
	patLen := len(pattern)
	times := size / patLen
	rem := size % patLen
	for i := 0; i < times; i++ {
		buf.WriteString(pattern)
	}
	if rem > 0 {
		buf.WriteString(pattern[:rem])
	}
	return buf.String()
}
