package model

import (
	"github.com/jinzhu/gorm"
)

type PodPort struct {
	ID       uint   `gorm:"primary_key,not null,AUTO_INCREMENT" json:"id"`
	PodID    string `json:"pod_id"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

type PodEnv struct {
	ID       int64  `gorm:"primary_key,not null,AUTO_INCREMENT" json:"id"`
	PodID    string `json:"pod_id"`
	EnvKey   string `json:"env_key"`
	EnvValue string `json:"env_value"`
}

type Pod struct {
	PodID            int64     `gorm:"primaryKey;AUTO_INCREMENT" json:"pod_id"`
	PodName          string    `gorm:"unique;not null" json:"pod_name"`
	PodNameSpace     string    `json:"pod_namespace"`
	PodTeamID        int64     `json:"pod_team_id"`
	PodMaxCpuUsage   float64   `json:"pod_max_cpu_usage"`
	PodMinCpuUsage   float64   `json:"pod_min_cpu_usage"`
	PodMaxMemUsage   float64   `json:"pod_max_mem_usage"`
	PodMinMemUsage   float64   `json:"pod_min_mem_usage"`
	PodPorts         []PodPort `gorm:"foreignKey:PodID" json:"pod_ports"`
	PodEnvs          []PodEnv  `gorm:"foreignKey:PodID" json:"pod_envs"`
	Image            string    `gorm:"not null" json:"image"`
	PodPullPolicy    string    `gorm:"default:'if_not_present'" json:"pod_pull_policy"`
	PodRestartPolicy string    `gorm:"default:'always'" json:"pod_restart_policy"`
	PodDeployType    string    `json:"pod_deploy_type"`
	Replicas         int32     `json:"replicas"`
}

type IPod interface {
	//初始化表
	InitTable() error
	//根据ID查找数据
	GetById(int64) (*Pod, error)
	//创建一个Pod
	CreatePod(*Pod) (int64, error)
	//删除pod
	DeletePod(uint64) error
	//更新Pod
	UpdatePod(*Pod) error
	//查找所有
	Get() ([]Pod, error)
}

func NewPodRegistry(db *gorm.DB) *PodRegistry {
	return &PodRegistry{
		db: db,
	}
}

type PodRegistry struct {
	db *gorm.DB
}

func (p *PodRegistry) InitTable() error {
	return p.db.AutoMigrate(&Pod{}, &PodEnv{}, &PodPort{}).Error

}

func (p *PodRegistry) GetById(id int64) (pod *Pod, err error) {
	pod = &Pod{}
	err = p.db.Preloads("PodEnv").Preloads("PodPort").First(pod, id).Error
	return
}

func (p *PodRegistry) CreatePod(pod *Pod) (podId int64, err error) {
	podId = pod.PodID
	err = p.db.Create(pod).Error
	return
}

func (p *PodRegistry) DeletePod(id uint64) error {
	return p.db.Where("pod_id=?", id).Delete(&Pod{}).Error
}

func (p *PodRegistry) UpdatePod(pod *Pod) error {
	return p.db.Model(&Pod{}).Update(pod).Error
}

func (p *PodRegistry) Get() (pods []Pod, err error) {
	err = p.db.Preloads("PodEnv").Preloads("PodPort").Find(&pods).Error
	return pods, err
}
