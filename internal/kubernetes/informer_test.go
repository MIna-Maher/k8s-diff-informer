package kubernetes

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"testing"
)

type syncResultsFactory struct {
	dynamicinformer.DynamicSharedInformerFactory
	results map[schema.GroupVersionResource]bool
}

func (f syncResultsFactory) WaitForCacheSync(_ <-chan struct{}) map[schema.GroupVersionResource]bool {
	return f.results
}

func TestReadinessRequiresAllCaches(t *testing.T) {
	pods := schema.GroupVersionResource{Version: "v1", Resource: "pods"}
	services := schema.GroupVersionResource{Version: "v1", Resource: "services"}
	for _, tc := range []struct {
		name                string
		results             map[schema.GroupVersionResource]bool
		canceled, wantReady bool
	}{
		{name: "no informers"},
		{name: "partial synchronization", results: map[schema.GroupVersionResource]bool{pods: true, services: false}},
		{name: "all synchronized", results: map[schema.GroupVersionResource]bool{pods: true, services: true}, wantReady: true},
		{name: "canceled startup", results: map[schema.GroupVersionResource]bool{pods: true}, canceled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			im := &InformerManager{factory: syncResultsFactory{results: tc.results}}
			if im.HasSynced() {
				t.Fatal("ready before synchronization")
			}
			stop := make(chan struct{})
			if tc.canceled {
				close(stop)
			}
			err := im.waitForCacheSync(stop)
			if (err == nil) != tc.wantReady || im.HasSynced() != tc.wantReady {
				t.Fatalf("ready=%v, err=%v, want ready=%v", im.HasSynced(), err, tc.wantReady)
			}
		})
	}
}
