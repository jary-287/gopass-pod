package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/jary-287/gopass-pod/model"
	"github.com/jary-287/gopass-pod/proto/pod"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IPodService interface {
	AddPod(*model.Pod) (uint64, error)
	DeletePod(uint64) error
	UpdatePod(*model.Pod) error
	FindPodById(uint64) (*model.Pod, error)
	FindAllPod() ([]model.Pod, error)
	CreateToK8s(*pod.PodInfo) error
	DeleteFromK8s(*pod.PodInfo) error
	UpdateToK8s(*pod.PodInfo) error
}

type PodService struct {
	PodRegistry model.IPod
	K8sClient   *kubernetes.Clientset
	Deployment  *v1.Deployment
}

func NewPodService(podRegistry model.IPod, client *kubernetes.Clientset) IPodService {
	return &PodService{
		PodRegistry: podRegistry,
		K8sClient:   client,
		Deployment:  &v1.Deployment{},
	}
}

// AddPod implements IPodService
func (ps *PodService) AddPod(pod *model.Pod) (uint64, error) {
	return ps.PodRegistry.CreatePod(pod)
}

// CreateToK8s implements IPodService
func (ps *PodService) CreateToK8s(pod *pod.PodInfo) error {
	ps.SetDeployment(pod)
	if _, err := ps.K8sClient.AppsV1().Deployments(pod.PodNamespace).Get(context.TODO(),
		pod.PodName, metav1.GetOptions{}); err != nil {
		ps.SetDeployment(pod)
		if _, err = ps.K8sClient.AppsV1().Deployments(pod.PodNamespace).Create(
			context.TODO(), ps.Deployment, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("pod 已经存在 podName: %s", pod.PodName))
	}
	log.Println("创建成功,", pod.PodName)
	return nil
}

// DeleteFromK8s implements IPodService
func (ps *PodService) DeleteFromK8s(pod *pod.PodInfo) error {
	if _, err := ps.K8sClient.AppsV1().Deployments(pod.PodNamespace).Get(
		context.TODO(), pod.PodName, metav1.GetOptions{},
	); err != nil {
		return fmt.Errorf("pod 不存在，请先创建,pod name:%s", pod.PodName)
	} else {
		if err = ps.K8sClient.AppsV1().Deployments(pod.PodNamespace).Delete(
			context.TODO(),
			pod.PodName,
			metav1.DeleteOptions{},
		); err != nil {
			return err
		}
	}
	log.Println("pod 删除成功，", pod.PodName)
	return nil
}

// DeletePod implements IPodService
func (ps *PodService) DeletePod(podID uint64) error {
	return ps.PodRegistry.DeletePod(podID)
}

// FindAllPod implements IPodService
func (ps *PodService) FindAllPod() ([]model.Pod, error) {
	return ps.PodRegistry.Get()
}

// FindPodById implements IPodService
func (ps *PodService) FindPodById(podID uint64) (*model.Pod, error) {
	return ps.PodRegistry.GetById(podID)
}

// UpdatePod implements IPodService
func (ps *PodService) UpdatePod(pod *model.Pod) error {
	return ps.PodRegistry.UpdatePod(pod)
}

// UpdateToK8s implements IPodService
func (ps *PodService) UpdateToK8s(info *pod.PodInfo) error {
	if _, err := ps.K8sClient.AppsV1().Deployments(info.PodNamespace).Get(
		context.TODO(), info.PodName, metav1.GetOptions{},
	); err != nil {
		return errors.New(fmt.Sprintf("pod 不存在，请先创建,pod name:%s", info.PodName))
	} else {
		ps.SetDeployment(info)
		if _, err = ps.K8sClient.AppsV1().Deployments(info.PodNamespace).Update(
			context.TODO(),
			ps.Deployment,
			metav1.UpdateOptions{},
		); err != nil {
			return err
		}
	}
	log.Println("pod 更新成功，", info.PodName)
	return nil
}

func (ps *PodService) SetDeployment(info *pod.PodInfo) {
	//ps.Deployment := &v1.Deployment{}
	ps.Deployment.TypeMeta = metav1.TypeMeta{
		Kind:       "deployment",
		APIVersion: "v1",
	}
	ps.Deployment.ObjectMeta = metav1.ObjectMeta{
		Name:      info.PodName,
		Namespace: info.PodNamespace,
		Labels: map[string]string{
			"app":    info.PodName,
			"author": "ljw",
		},
	}
	ps.Deployment.Name = info.PodName
	ps.Deployment.Spec = v1.DeploymentSpec{
		Replicas: &info.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": info.PodName,
			},
		},
		Template: v12.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      info.PodName,
				Namespace: info.PodNamespace,
				Labels: map[string]string{
					"app": info.PodName,
				},
			},
			Spec: v12.PodSpec{
				Containers: []v12.Container{
					v12.Container{
						Name:            info.PodName,
						Image:           info.Image,
						ImagePullPolicy: v12.PullPolicy(info.PodPullPolicy),
						Ports:           ps.GetContaiinerPort(info),
						Env:             ps.GetEnvs(info.PodEnvs),
						Resources:       ps.GetResource(info),
					},
				},
			},
		},
		Strategy: v1.DeploymentStrategy{},
	}
}

func (ps *PodService) GetContaiinerPort(info *pod.PodInfo) (containerPorts []v12.ContainerPort) {
	for _, port := range info.PodPorts {
		containerPorts = append(containerPorts, v12.ContainerPort{
			ContainerPort: port.Port,
			Protocol:      GetProtocol(port.Protocol),
		})
	}
	return
}

func GetProtocol(protocol string) v12.Protocol {
	switch protocol {
	case "TCP":
		return v12.ProtocolTCP
	case "UDP":
		return v12.ProtocolUDP
	case "SCTP":
		return v12.ProtocolSCTP
	default:
		return v12.ProtocolTCP
	}
}

func (ps *PodService) GetEnvs(envs []*pod.PodEnv) (containerEnvs []v12.EnvVar) {
	for _, env := range envs {
		containerEnvs = append(containerEnvs, v12.EnvVar{
			Name:  env.EnvKey,
			Value: string(env.EnvValue),
		})
	}
	return
}

//获取资源限制
func (ps *PodService) GetResource(info *pod.PodInfo) (source v12.ResourceRequirements) {
	source.Limits = v12.ResourceList{
		"cpu":    resource.MustParse(strconv.FormatFloat(float64(info.PodMaxCpuUsage), 'f', 6, 64)),
		"memory": resource.MustParse(strconv.FormatFloat(float64(info.PodMaxMemUsage), 'f', 6, 64)),
	}
	source.Requests =
		v12.ResourceList{
			"cpu":    resource.MustParse(strconv.FormatFloat(float64(info.PodMaxCpuUsage), 'f', 6, 64)),
			"memory": resource.MustParse(strconv.FormatFloat(float64(info.PodMaxMemUsage), 'f', 6, 64)),
		}
	return
}
