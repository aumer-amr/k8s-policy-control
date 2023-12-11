package policy

import (
	"context"
	"fmt"
	"strings"

	"github.com/aumer-amr/k8s-policy-control/internal/util"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	gatusGenerateAnnotation = "policy-control.aumer.io/gatus-generate"
	gatusNameAnnotation     = "policy-control.aumer.io/gatus-name"
	gatusGroupAnnotation    = "policy-control.aumer.io/gatus-group"
	gatusHostAnnotation     = "policy-control.aumer.io/gatus-host"
	gatusPathAnnotation     = "policy-control.aumer.io/gatus-path"
	gatusConditions         = "policy-control.aumer.io/gatus-conditions"
	gatusDns                = "policy-control.aumer.io/gatus-dns"
	ingressGenerateGatusLog = ctrl.Log.WithName("ingress_generate_gatus")
)

type GatusConfigMap struct {
	Endpoints []GatusEndpoint `yaml:"endpoints"`
}

type GatusEndpoint struct {
	Name       string          `yaml:"name"`
	Group      string          `yaml:"group"`
	Url        string          `yaml:"url"`
	Interval   string          `yaml:"interval"`
	Ui         GatusUi         `yaml:"ui"`
	Conditions []string        `yaml:"conditions,omitempty"`
	Dns        *GatusDnsClient `yaml:"client,omitempty"`
}

type GatusUi struct {
	HideHostname bool `yaml:"hideHostname"`
	HideUrl      bool `yaml:"hideUrl"`
}

type GatusDnsClient struct {
	DnsResolver string `yaml:"dnsResolver"`
}

type IngressGenerateGatus struct {
	Manager ctrl.Manager
}

func (i IngressGenerateGatus) Name() string {
	return "Ingress Generate Gatus"
}

func (i IngressGenerateGatus) Type() int {
	return PolicyTypeIngress
}

func (i IngressGenerateGatus) Validate(obj runtime.Object, mgr ctrl.Manager) (error, bool) {
	i.Manager = mgr

	if ingress, ok := obj.(*networkingv1.Ingress); ok {
		if val, ok := ingress.Annotations[gatusGenerateAnnotation]; ok {
			if val == "false" {
				ingressGenerateGatusLog.Info("Skipping Ingress because annotation is explicitly false", "ingress", getIngressName(ingress))
				i.Handle(ingress, true)
				return nil, false
			} else if val == "true" {
				return nil, true
			}
		}
	} else {
		return fmt.Errorf("could not cast object to Pod"), false
	}

	return nil, false
}

func (i IngressGenerateGatus) Apply(obj runtime.Object, mgr ctrl.Manager) error {
	i.Manager = mgr

	ingress, ok := obj.(*networkingv1.Ingress)
	if !ok {
		ingressGenerateGatusLog.Error(fmt.Errorf("could not cast object to Ingress"), "error casting object to Ingress")
		return nil
	}

	err := i.Handle(ingress, false)
	return err
}

func (i IngressGenerateGatus) Handle(ingress *networkingv1.Ingress, fromValidate bool) error {
	configMapList := corev1.ConfigMapList{}
	err := i.Manager.GetClient().List(context.Background(), &configMapList, client.MatchingLabels{
		"app.kubernetes.io/managed-by":       "policy-control.aumer.io",
		"policy-control.aumer.io/parent-uid": string(ingress.ObjectMeta.UID),
	})
	if err != nil {
		return err
	}

	// Delete configmap on ingress deletion
	if ingress.ObjectMeta.DeletionTimestamp.IsZero() == false || fromValidate == true {
		if len(configMapList.Items) == 1 {
			configMap := configMapList.Items[0]
			i.Delete(ingress, configMap)
		}
		return nil
	}

	// Create configmap if it doesn't exist
	if len(configMapList.Items) == 0 && fromValidate == false {
		i.Create(ingress)
		return nil
	}

	// Update configmap if it exists
	if len(configMapList.Items) == 1 && fromValidate == false {
		configMap := configMapList.Items[0]
		i.Update(ingress, configMap)
		return nil
	}

	return nil
}

func (i IngressGenerateGatus) Update(ingress *networkingv1.Ingress, configMap corev1.ConfigMap) {

	configMap.Data["config.yaml"] = generateGatusConfigMapData(ingress)
	i.Manager.GetClient().Update(context.Background(), &configMap)
}

func (i IngressGenerateGatus) Delete(ingress *networkingv1.Ingress, configMap corev1.ConfigMap) {
	i.Manager.GetClient().Delete(context.Background(), &configMap)
}

func (i IngressGenerateGatus) Create(ingress *networkingv1.Ingress) {
	ingressGenerateGatusLog.Info("Creating Gatus ConfigMap", "ingress", getIngressName(ingress))

	configMap := &corev1.ConfigMap{
		ObjectMeta: generateGatusConfigMapMetadata(ingress),
		Data: map[string]string{
			"config.yaml": generateGatusConfigMapData(ingress),
		},
	}

	i.Manager.GetClient().Create(context.Background(), configMap)
}

func generateGatusConfigMapMetadata(ingress *networkingv1.Ingress) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      getIngressName(ingress) + "-gatus-generated",
		Namespace: ingress.GetNamespace(),
		Labels: map[string]string{
			"app.kubernetes.io/managed-by":       "policy-control.aumer.io",
			"gatus.io/enabled":                   "enabled",
			"policy-control.aumer.io/parent-uid": string(ingress.ObjectMeta.UID),
		},
	}
}

func generateGatusConfigMapData(ingress *networkingv1.Ingress) string {
	url := util.GetAnnotationStringValue(gatusHostAnnotation, ingress.Annotations, ingress.Spec.Rules[0].Host) + util.GetAnnotationStringValue(gatusPathAnnotation, ingress.Annotations, ingress.Spec.Rules[0].HTTP.Paths[0].Path)

	configMapData := &GatusConfigMap{
		Endpoints: []GatusEndpoint{
			{
				Name:       util.GetAnnotationStringValue(gatusNameAnnotation, ingress.Annotations, getIngressName(ingress)),
				Group:      util.GetAnnotationStringValue(gatusGroupAnnotation, ingress.Annotations, "default"),
				Url:        url,
				Interval:   "1m",
				Ui:         GatusUi{HideHostname: true, HideUrl: true},
				Conditions: mutateGatusConditions(util.GetAnnotationStringValue(gatusConditions, ingress.Annotations, "")),
				Dns:        mutateGatusDns(util.GetAnnotationBoolValue(gatusDns, ingress.Annotations, false)),
			},
		},
	}

	outputYaml, err := yaml.Marshal(configMapData)
	if err != nil {
		ingressGenerateGatusLog.Error(err, "error marshalling config map data")
	}

	return string(outputYaml)
}

func mutateGatusDns(annotationValue bool) *GatusDnsClient {
	if annotationValue == false {
		return nil
	}

	return &GatusDnsClient{
		DnsResolver: "tcp://1.1.1.1:53",
	}
}

func mutateGatusConditions(annotationValue string) []string {
	if annotationValue == "" {
		return []string{"[STATUS] == 200"}
	}

	return strings.Split(annotationValue, ",")
}

func getIngressName(ingress *networkingv1.Ingress) string {
	ingressName := ingress.GetName()
	if ingressName != "" {
		return ingressName
	}
	return ingress.GetGenerateName()
}

func init() {
	RegisterPolicy(&IngressGenerateGatus{})
}
