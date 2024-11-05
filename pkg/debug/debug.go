package debug

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"os"
)

type Debug struct {
	ctx       context.Context
	clientset kubernetes.Interface
	cfg       Config
}

type Config struct {
	Namespace  string
	Context    string
	PodName    string
	DebugImage string
}

func Setup(clientset kubernetes.Interface, cfg Config) *Debug {
	return &Debug{
		ctx:       context.Background(),
		clientset: clientset,
		cfg:       cfg,
	}
}

func (d *Debug) Debug() (*v1.Pod, error) {
	pod, err := d.clientset.CoreV1().Pods(d.cfg.Namespace).Get(d.ctx, d.cfg.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Define the ephemeral container with security context
	debugContainer := &v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:            fmt.Sprintf("debug-%s", rand.IntnRange(1000, 9999)),
			Image:           d.cfg.DebugImage,
			Command:         []string{"sh"},
			Stdin:           true,
			TTY:             true,
			SecurityContext: securityContext(),
		},
	}

	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, *debugContainer)

	p, err := d.clientset.CoreV1().Pods(d.cfg.Namespace).UpdateEphemeralContainers(d.ctx, d.cfg.PodName, pod, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update pod: %w", err)
	}

	fmt.Println("Debug container added successfully, attaching to it...")
	return p, err
}

func securityContext() *v1.SecurityContext {
	secCon := &v1.SecurityContext{
		Capabilities: &v1.Capabilities{
			Drop: []v1.Capability{"ALL"},
		},
	}
	*secCon.RunAsUser = 1000
	*secCon.RunAsGroup = 1000
	*secCon.Privileged = false
	*secCon.RunAsNonRoot = true
	*secCon.ReadOnlyRootFilesystem = true
	*secCon.AllowPrivilegeEscalation = false
	return secCon
}

func (d *Debug) AttachToEphemeralContainer(config *rest.Config, containerName string) error {
	req := d.clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(d.cfg.PodName).
		Namespace(d.cfg.Namespace).
		SubResource("attach").
		Param("container", containerName).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "true")

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Set up input/output streams
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})
	if err != nil {
		return fmt.Errorf("error in stream: %w", err)
	}

	return nil
}

func (d *Debug) tidy() {
	// Delete the pod
	err := d.clientset.CoreV1().Pods(d.cfg.Namespace).Delete(d.ctx, d.cfg.PodName, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Failed to delete pod: %v\n", err)
		return
	}

	fmt.Println("Pod deleted successfully. Kubernetes will recreate it if managed by a controller.")
}
