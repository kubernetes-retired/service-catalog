package v1beta1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertClusterServiceClassToProperties(t *testing.T) {
	cases := []struct {
		name string
		sc   *ClusterServiceClass
		json string
	}{
		{
			name: "nil object",
			json: "{}",
		},
		{
			name: "normal object",
			sc: &ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{Name: "service-class"},
				Spec: ClusterServiceClassSpec{
					CommonServiceClassSpec: CommonServiceClassSpec{
						ExternalName: "external-class-name",
						ExternalID:   "external-id",
					},
				},
			},
			json: `{"name":"service-class","spec.externalID":"external-id","spec.externalName":"external-class-name"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ConvertClusterServiceClassToProperties(tc.sc)
			if p == nil {
				t.Fatalf("Failed to create Properties object from %+v", tc.sc)
			}
			b, err := json.Marshal(p)
			if err != nil {
				t.Fatalf("Unexpected error with json marchal, %v", err)
			}
			js := string(b)
			if js != tc.json {
				t.Fatalf("Failed to create expected Properties object,\n\texpected: \t%q,\n \tgot: \t\t%q", tc.json, js)
			}
		})
	}
}

func TestConvertClusterServicePlanToProperties(t *testing.T) {
	cases := []struct {
		name string
		sp   *ClusterServicePlan
		json string
	}{
		{
			name: "nil object",
			json: "{}",
		},
		{
			name: "normal object",
			sp: &ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{Name: "service-plan"},
				Spec: ClusterServicePlanSpec{
					CommonServicePlanSpec: CommonServicePlanSpec{
						ExternalName: "external-plan-name",
						ExternalID:   "external-id",
					},
					ClusterServiceClassRef: ClusterObjectReference{
						Name: "cluster-service-class-name",
					},
				},
			},
			json: `{"name":"service-plan","spec.clusterServiceClass.name":"cluster-service-class-name","spec.externalID":"external-id","spec.externalName":"external-plan-name"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ConvertClusterServicePlanToProperties(tc.sp)
			if p == nil {
				t.Fatalf("Failed to create Properties object from %+v", tc.sp)
			}
			b, err := json.Marshal(p)
			if err != nil {
				t.Fatalf("Unexpected error with json marchal, %v", err)
			}
			js := string(b)
			if js != tc.json {
				t.Fatalf("Failed to create expected Properties object,\n\texpected: \t%q,\n \tgot: \t\t%q", tc.json, js)
			}
		})
	}
}
