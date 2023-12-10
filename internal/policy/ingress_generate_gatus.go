package policy

import (
	"fmt"
	"strings"

	"github.com/aumer-amr/k8s-policy-control/internal/util"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
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

type IngressGenerateGatusData struct {
	GATUSNAME       string
	GATUSGROUP      string
	GATUSHOST       string
	GATUSPATH       string
	GATUSCONDITIONS string
	GATUSDNS        string
}

type IngressGenerateGatus struct{}

func (i IngressGenerateGatus) Name() string {
	return "Ingress Generate Gatus"
}

func (i IngressGenerateGatus) Type() int {
	return PolicyTypeIngress
}

func (i IngressGenerateGatus) Validate(obj runtime.Object) (error, bool) {
	err, result := ValidateByType(PolicyTypeIngress, obj)
	if err != nil {
		return err, false
	}

	if ingress, ok := result.(*networkv1.Ingress); ok {
		if val, ok := ingress.Annotations[gatusGenerateAnnotation]; ok {
			if val == "false" {
				podCpuLog.Info("Skipping Ingress because annotation is explicitly false", "ignress", getIngressName(ingress))
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

func (i IngressGenerateGatus) Apply(obj runtime.Object, policyOperation int) error {
	err, result := ValidateByType(PolicyTypeIngress, obj)
	if err != nil {
		return err
	}

	if ingress, ok := result.(*networkv1.Ingress); ok {
		if policyOperation == PolicyOperationDelete {

		} else if policyOperation == PolicyOperationUpsert {
			generateGatusConfigMap(ingress)
		}
	} else {
		return fmt.Errorf("could not cast object to Ingress")
	}

	return nil
}

func generateGatusConfigMap(ingress *networkv1.Ingress) string {
	ingressGenerateGatusLog.Info("Generating Gatus ConfigMap", "ingress", getIngressName(ingress))

	// TODO: Make this into a template file
	configMapDataTemplate := `
---
endpoints:
  - name: {{ .GATUSNAME }}
    group: {{ .GATUSGROUP }}
    url: https://{{ .GATUSHOST }}{{ .GATUSPATH }}
    interval: 1m
    ui:
      hide-hostname: true
      hide-url: true
    {{ .GATUSCONDITIONS }}
    {{ .GATUSDNS }}`

	configMapData, err := util.RenderTemplate(configMapDataTemplate, IngressGenerateGatusData{
		GATUSNAME:       util.GetAnnotationStringValue(gatusNameAnnotation, ingress.Annotations, getIngressName(ingress)),
		GATUSGROUP:      util.GetAnnotationStringValue(gatusGroupAnnotation, ingress.Annotations, "default"),
		GATUSHOST:       util.GetAnnotationStringValue(gatusHostAnnotation, ingress.Annotations, ingress.Spec.Rules[0].Host),
		GATUSPATH:       util.GetAnnotationStringValue(gatusPathAnnotation, ingress.Annotations, ingress.Spec.Rules[0].HTTP.Paths[0].Path),
		GATUSCONDITIONS: mutateGatusConditions(util.GetAnnotationStringValue(gatusConditions, ingress.Annotations, "")),
		GATUSDNS:        mutateGatusDns(util.GetAnnotationBoolValue(gatusDns, ingress.Annotations, false)),
	})

	if err != nil {
		ingressGenerateGatusLog.Error(err, "Failed to render template", "ingress", getIngressName(ingress))
		return ""
	}

	// configMap := &corev1.ConfigMap{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      getIngressName(ingress) + "-gatus-generated",
	// 		Namespace: ingress.GetNamespace(),
	// 		Labels: map[string]string{
	// 			"policy-control.aumer.io/managed": "true",
	// 			"gatus.io/enabled":                "enabled",
	// 		},
	// 	},
	// 	Data: map[string]string{
	// 		"config.yaml": configMapData,
	// 	},
	// }

	return configMapData

	//client.Client().Create(context.Background(), configMap)
}

func mutateGatusDns(annotationValue bool) string {
	if annotationValue == false {
		return ""
	}

	// TODO: Make this into a template file
	mutatedDns := `client:
      dns-resolver: tcp://1.1.1.1:53`

	return mutatedDns
}

func mutateGatusConditions(annotationValue string) string {
	if annotationValue == "" {
		annotationValue = "[STATUS] == 200"
	}

	conditions := strings.Split(annotationValue, ",")
	mutatedConditions := "conditions: "
	for _, condition := range conditions {
		mutatedConditions += fmt.Sprintf("\n      - \"%s\",", condition)
	}

	return mutatedConditions
}

func getIngressName(ingress *networkv1.Ingress) string {
	ingressName := ingress.GetName()
	if ingressName != "" {
		return ingressName
	}
	return ingress.GetGenerateName()
}

func init() {
	RegisterPolicy(&IngressGenerateGatus{})
}
