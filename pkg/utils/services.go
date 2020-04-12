package otmachinery

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	redisv1alpha1 "redis-operator/pkg/apis/redis/v1alpha1"
	"strconv"
)

const (
	redisPort = 6379
)

// ServiceInterface is the interface to pass service information accross methods
type ServiceInterface struct {
	ExistingService      *corev1.Service
	NewServiceDefinition *corev1.Service
	ServiceType          string
}

// ServiceArguments is the interface to pass service information
type ServiceArguments struct {
	CR          *redisv1alpha1.Redis
	Labels      map[string]string
	PortNumber  int32
	Role        string
	ServiceName string
	ClusterIP   string
}

// GenerateServiceDef generate service definition
func GenerateServiceDef(sa ServiceArguments) *corev1.Service {
	var redisExporterPort int32 = 9121
	service := &corev1.Service{
		TypeMeta:   GenerateMetaInformation("Service", "core/v1"),
		ObjectMeta: GenerateObjectMetaInformation(sa.ServiceName, sa.CR.Namespace, sa.Labels, GenerateServiceAnots()),
		Spec: corev1.ServiceSpec{
			ClusterIP: sa.ClusterIP,
			Selector:  sa.Labels,
			Ports: []corev1.ServicePort{
				{
					Name:       sa.CR.ObjectMeta.Name + "-" + sa.Role,
					Port:       sa.PortNumber,
					TargetPort: intstr.FromInt(int(sa.PortNumber)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	if sa.CR.Spec.RedisExporter != false {
		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Name:       "redis-exporter",
			Port:       redisExporterPort,
			TargetPort: intstr.FromInt(int(redisExporterPort)),
			Protocol:   corev1.ProtocolTCP,
		})
	}
	AddOwnerRefToObject(service, AsOwner(sa.CR))
	return service
}

// CreateMasterHeadlessService creates master headless service
func CreateMasterHeadlessService(cr *redisv1alpha1.Redis) {
	labels := map[string]string{
		"app":  cr.ObjectMeta.Name + "-master",
		"role": "master",
	}
	serviceDefinition := GenerateServiceDef(ServiceArguments{
		CR:          cr,
		Labels:      labels,
		PortNumber:  int32(redisPort),
		Role:        "master",
		ServiceName: cr.ObjectMeta.Name + "-master",
		ClusterIP:   "None",
	})
	serviceBody, err := GenerateK8sClient().CoreV1().Services(cr.Namespace).Get(cr.ObjectMeta.Name+"-master", metav1.GetOptions{})
	service := ServiceInterface{
		ExistingService:      serviceBody,
		NewServiceDefinition: serviceDefinition,
		ServiceType:          "master",
	}
	CompareAndCreateService(cr, service, err)
}

// CreateMasterService creates different services for master
func CreateMasterService(cr *redisv1alpha1.Redis) {
	masterReplicas := cr.Spec.Size
	for serviceCount := 0; serviceCount <= int(*masterReplicas)-1; serviceCount++ {
		labels := map[string]string{
			"app":                                cr.ObjectMeta.Name + "-master",
			"role":                               "master",
			"statefulset.kubernetes.io/pod-name": cr.ObjectMeta.Name + "-master-" + strconv.Itoa(serviceCount),
		}
		serviceDefinition := GenerateServiceDef(ServiceArguments{
			CR:          cr,
			Labels:      labels,
			PortNumber:  int32(redisPort),
			Role:        "master",
			ServiceName: cr.ObjectMeta.Name + "-master-" + strconv.Itoa(serviceCount),
			ClusterIP:   "None",
		})
		serviceBody, err := GenerateK8sClient().CoreV1().Services(cr.Namespace).Get(cr.ObjectMeta.Name+"-master-"+strconv.Itoa(serviceCount), metav1.GetOptions{})
		service := ServiceInterface{
			ExistingService:      serviceBody,
			NewServiceDefinition: serviceDefinition,
			ServiceType:          "master",
		}
		CompareAndCreateService(cr, service, err)
	}
}

// CreateSlaveHeadlessService creates slave headless service
func CreateSlaveHeadlessService(cr *redisv1alpha1.Redis) {
	labels := map[string]string{
		"app":  cr.ObjectMeta.Name + "-slave",
		"role": "slave",
	}
	serviceDefinition := GenerateServiceDef(ServiceArguments{
		CR:          cr,
		Labels:      labels,
		PortNumber:  int32(redisPort),
		Role:        "slave",
		ServiceName: cr.ObjectMeta.Name + "-slave",
		ClusterIP:   "None",
	})
	serviceBody, err := GenerateK8sClient().CoreV1().Services(cr.Namespace).Get(cr.ObjectMeta.Name+"-slave", metav1.GetOptions{})
	service := ServiceInterface{
		ExistingService:      serviceBody,
		NewServiceDefinition: serviceDefinition,
		ServiceType:          "slave",
	}
	CompareAndCreateService(cr, service, err)
}

// CreateSlaveService creates different services for slave
func CreateSlaveService(cr *redisv1alpha1.Redis) {
	slaveReplicas := cr.Spec.Size
	for serviceCount := 0; serviceCount <= int(*slaveReplicas)-1; serviceCount++ {
		labels := map[string]string{
			"app":                                cr.ObjectMeta.Name + "-slave",
			"role":                               "slave",
			"statefulset.kubernetes.io/pod-name": cr.ObjectMeta.Name + "-slave-" + strconv.Itoa(serviceCount),
		}
		serviceDefinition := GenerateServiceDef(ServiceArguments{
			CR:          cr,
			Labels:      labels,
			PortNumber:  int32(redisPort),
			Role:        "slave",
			ServiceName: cr.ObjectMeta.Name + "-slave-" + strconv.Itoa(serviceCount),
			ClusterIP:   "None",
		})
		serviceBody, err := GenerateK8sClient().CoreV1().Services(cr.Namespace).Get(cr.ObjectMeta.Name+"-slave-"+strconv.Itoa(serviceCount), metav1.GetOptions{})
		service := ServiceInterface{
			ExistingService:      serviceBody,
			NewServiceDefinition: serviceDefinition,
			ServiceType:          "slave",
		}
		CompareAndCreateService(cr, service, err)
	}
}

// CreateStandaloneService creates redis standalone service
func CreateStandaloneService(cr *redisv1alpha1.Redis) {
	labels := map[string]string{
		"app":  cr.ObjectMeta.Name + "-" + "standalone",
		"role": "standalone",
	}
	serviceDefinition := GenerateServiceDef(ServiceArguments{
		CR:          cr,
		Labels:      labels,
		PortNumber:  int32(redisPort),
		Role:        "standalone",
		ServiceName: cr.ObjectMeta.Name,
		ClusterIP:   "None",
	})
	serviceBody, err := GenerateK8sClient().CoreV1().Services(cr.Namespace).Get(cr.ObjectMeta.Name, metav1.GetOptions{})

	service := ServiceInterface{
		ExistingService:      serviceBody,
		NewServiceDefinition: serviceDefinition,
		ServiceType:          "standalone",
	}
	CompareAndCreateService(cr, service, err)
}

// CompareAndCreateService compares and creates service
func CompareAndCreateService(cr *redisv1alpha1.Redis, service ServiceInterface, err error) {
	reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.ObjectMeta.Name)

	if err != nil {
		reqLogger.Info("Creating redis service", "Redis.Name", cr.ObjectMeta.Name+"-"+service.ServiceType, "Service.Type", service.ServiceType)
		GenerateK8sClient().CoreV1().Services(cr.Namespace).Create(service.NewServiceDefinition)
	} else if service.ExistingService != service.NewServiceDefinition {
		reqLogger.Info("Reconciling redis service", "Redis.Name", cr.ObjectMeta.Name+"-"+service.ServiceType, "Service.Type", service.ServiceType)
		GenerateK8sClient().CoreV1().Services(cr.Namespace).Update(service.NewServiceDefinition)
	} else {
		reqLogger.Info("Redis service is in sync", "Redis.Name", cr.ObjectMeta.Name+"-"+service.ServiceType, "Service.Type", service.ServiceType)
	}
}
