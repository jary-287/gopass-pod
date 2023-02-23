package model

import (
	"log"

	"gorm.io/gorm"
)

type PodPort struct {
	ID       uint `gorm:"primaryKey;not null;AUTO_INCREMENT" json:"omitempty"`
	PodID    uint64
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

type PodEnv struct {
	ID       uint64 `gorm:"primaryKey;not null;AUTO_INCREMENT" json:"id,omitempty"`
	PodID    uint64
	EnvKey   string `json:"env_key"`
	EnvValue string `json:"env_value"`
}

type Pod struct {
	PodID            uint64    `gorm:"primaryKey;not null" json:"pod_id"`
	PodName          string    `gorm:"unique;not null" json:"pod_name"`
	PodNameSpace     string    `json:"pod_namespace"`
	PodTeamID        int64     `json:"pod_team_id"`
	PodMaxCpuUsage   float64   `json:"pod_max_cpu_usage"`
	PodMinCpuUsage   float64   `json:"pod_min_cpu_usage"`
	PodMaxMemUsage   float64   `json:"pod_max_mem_usage"`
	PodMinMemUsage   float64   `json:"pod_min_mem_usage"`
	PodPorts         []PodPort `gorm:"foreignKey:pod_id;references:pod_id" json:"pod_ports"`
	PodEnvs          []PodEnv  `gorm:"foreignKey:pod_id;references:pod_id" json:"pod_envs"`
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
	GetById(uint64) (*Pod, error)
	//创建一个Pod
	CreatePod(*Pod) (uint64, error)
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
	log.Println("自动迁移数据库")
	return p.db.AutoMigrate(&Pod{}, &PodEnv{}, &PodPort{})

}

func (p *PodRegistry) GetById(id uint64) (pod *Pod, err error) {
	pod = &Pod{}
	err = p.db.Preload("PodEnvs").Preload("PodPorts").First(pod, id).Error
	return
}

func (p *PodRegistry) CreatePod(pod *Pod) (podId uint64, err error) {
	podId = pod.PodID
	err = p.db.Create(pod).Error
	return
}

func (p *PodRegistry) DeletePod(id uint64) error {
	tx := p.db.Begin()
	tx.Where("pod_id = ?", id).Delete(&PodPort{})
	tx.Where("pod_id = ?", id).Delete(&PodEnv{})
	tx.Where("pod_id=?", id).Delete(&Pod{})
	if err := tx.Commit().Error; err != nil {
		tx.Callback()
		return err
	}
	return nil
}

func (p *PodRegistry) UpdatePod(pod *Pod) error {
	tx := p.db.Begin()
	tx.Association("pod_envs").DB.Save(pod.PodEnvs)
	tx.Association("pod_ports").DB.Save(pod.PodPorts)
	tx.Save(pod)
	if err := tx.Commit().Error; err != nil {
		tx.Callback()
		return err
	}
	return nil
}

func (p *PodRegistry) Get() (pods []Pod, err error) {
	err = p.db.Preload("PodEnvs").Preload("PodPorts").Find(&pods).Error
	return pods, err
}
