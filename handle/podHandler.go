package handle

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/jary-287/gopass-pod/model"
	"github.com/jary-287/gopass-pod/proto/pod"
	"github.com/jary-287/gopass-pod/service"
)

type Podhandler struct {
	PodService service.IPodService
}

func (ph *Podhandler) AddPod(ctx context.Context, info *pod.PodInfo, rsp *pod.Response) error {
	log.Println("add pod :", info.PodName)
	podModel := &model.Pod{}
	if err := swap(info, podModel); err != nil {
		rsp.Msg = err.Error()
		return err
	}
	if err := ph.PodService.CreateToK8s(info); err != nil {
		rsp.Msg = err.Error()
		return err
	}
	if _, err := ph.PodService.AddPod(podModel); err != nil {
		rsp.Msg = err.Error()
		return err
	}
	log.Println("pod add success:", info.PodName)
	rsp.Msg = "success create pod,pod name " + info.PodName
	return nil
}

func (ph *Podhandler) DeletePod(ctx context.Context, info *pod.PodInfo, rsp *pod.Response) error {
	if err := ph.PodService.DeleteFromK8s(info); err != nil {
		rsp.Msg = err.Error()
		return err
	}
	log.Println("delete pod from k8s success:  podname", info.PodName)
	if err := ph.PodService.DeletePod(info.PodId); err != nil {
		rsp.Msg = err.Error()
		return err
	}
	log.Println("pod delete success:", info.PodName)
	rsp.Msg = "success delete pod,pod name " + info.PodName
	return nil
}

func (ph *Podhandler) UpdatePod(ctx context.Context, info *pod.PodInfo, rsp *pod.Response) error {
	if err := ph.PodService.UpdateToK8s(info); err != nil {
		rsp.Msg = err.Error()
		return err
	}
	log.Println("update pod to k8s success:", info.PodName)
	podModel := &model.Pod{}
	if err := swap(info, podModel); err != nil {
		rsp.Msg = err.Error()
		return err
	}
	if err := ph.PodService.UpdatePod(podModel); err != nil {
		rsp.Msg = err.Error()
		return err
	}
	log.Println("update pod  success:", info.PodName)
	return nil
}

// rpc FindPodById(PodId) returns (PodInfo) {}
func (ph *Podhandler) FindPodById(ctx context.Context, id *pod.PodId, info *pod.PodInfo) error {
	podModel, err := ph.PodService.FindPodById(id.Id)
	if err != nil {
		return err
	}
	if err = swap(podModel, info); err != nil {
		return err
	}
	log.Println("find pod by Id success")
	return nil
}

func (ph *Podhandler) FindAllPod(ctx context.Context, findAll *pod.FinadAll, allPod *pod.AllPod) error {
	pods, err := ph.PodService.FindAllPod()
	if err != nil {
		return errors.New("find all pod failed:" + err.Error())
	}
	if err := swap(pods, &allPod.PodInfo); err != nil {
		return errors.New("find all pod: swap failed " + err.Error())
	}
	log.Println("find all pod success")
	return nil
}

//proroto打包成json，在解到struct
func swap(source interface{}, target interface{}) error {
	data, err := json.Marshal(source)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, target)
	return err
}
