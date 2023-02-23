package main

import (
	"flag"
	"log"
	"path"
	"time"

	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/go-micro/plugins/v3/registry/consul"
	"github.com/go-micro/plugins/v3/wrapper/breaker/hystrix"
	limiter "github.com/go-micro/plugins/v3/wrapper/ratelimiter/uber"
	opentracing2 "github.com/go-micro/plugins/v3/wrapper/trace/opentracing"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jary-287/gopass-common/common"
	"github.com/jary-287/gopass-pod/handle"
	"github.com/jary-287/gopass-pod/model"
	"github.com/jary-287/gopass-pod/proto/pod"
	"github.com/jary-287/gopass-pod/service"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	// 注册中心
	consulHost string = "192.168.0.19"
	consulPort int64  = 8500
	//链路追踪
	tracerHost string = "localhost"
	tracerPort int64  = 9092
	// 熔断
	hystrixPort int64 = 9091
	// 监控
	prometheusPort int64 = 9093
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", path.Join(home, ".kube", "config"), "kubeconfig 位置")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig 位置")
	}
	flag.Parse()
	//创建config实例
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	//注册中心
	consulRegister := consul.NewRegistry(func(o *registry.Options) {
		o.Addrs = []string{"192.168.0.19:8500"}
		o.Timeout = 20 * time.Second

	})
	t, io, err := common.NewTracer("service.pod", "192.168.0.102:9333")
	if err != nil {
		log.Fatal(err)
	}

	defer io.Close()
	// 创建pod服务
	serv := micro.NewService(
		micro.Name("service.pod"),
		micro.Version("latest"),
		//注册中心
		micro.Registry(consulRegister),
		//链路追踪
		micro.WrapHandler(opentracing2.NewHandlerWrapper(t)),
		micro.WrapClient(opentracing2.NewClientWrapper(t)),
		//熔断
		micro.WrapClient(hystrix.NewClientWrapper()),
		micro.WrapHandler(limiter.NewHandlerWrapper(1000)),
	)
	serv.Init()
	// 初始化数据表
	err = model.GetDB()
	if err != nil {
		log.Fatal("数据库初始化失败", err)
	}
	if err := model.NewPodRegistry(model.Db).InitTable(); err != nil {
		log.Fatal(err)
	}

	//注册句柄
	podService := service.NewPodService(model.NewPodRegistry(model.Db), client)
	pod.RegisterPodHandler(serv.Server(), &handle.Podhandler{PodService: podService})

	if err := serv.Run(); err != nil {
		log.Fatal(err)
	}
}
