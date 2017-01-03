package storage

import (
	"log"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/util"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/watch"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/runtime"
)

type tprStorageServiceClass struct {
	watcher *watch.Watcher
}

func newTPRStorageServiceClass(watcher *watch.Watcher) *tprStorageServiceClass {
	return &tprStorageServiceClass{watcher: watcher}
}

func (t *tprStorageServiceClass) Get(name string) (*servicecatalog.ServiceClass, error) {
	si, err := t.watcher.GetResourceClient(watch.ServiceClass, "default").Get(name)
	if err != nil {
		return nil, err
	}
	var tmp servicecatalog.ServiceClass
	err = util.TPRObjectToSCObject(si, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

func (t *tprStorageServiceClass) List() ([]*servicecatalog.ServiceClass, error) {
	l, err := t.watcher.GetResourceClient(watch.ServiceClass, "default").List(&v1.ListOptions{})
	if err != nil {
		log.Printf("Failed to list service types: %v\n", err)
		return nil, err
	}
	var lst []*servicecatalog.ServiceClass
	for _, i := range l.(*runtime.UnstructuredList).Items {
		var tmp servicecatalog.ServiceClass
		err := util.TPRObjectToSCObject(i, &tmp)
		if err != nil {
			log.Printf("Failed to convert object: %v\n", err)
			return nil, err
		}
		lst = append(lst, &tmp)
	}
	return lst, nil

}

func (t *tprStorageServiceClass) Create(sc *servicecatalog.ServiceClass) (*servicecatalog.ServiceClass, error) {
	sc.Kind = watch.ServiceClassKind
	sc.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(sc)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", sc, err)
		return nil, err
	}
	tprObj.SetName(sc.Name)
	log.Printf("Creating k8sobject:\n%v\n", tprObj)
	_, err = t.watcher.GetResourceClient(watch.ServiceClass, "default").Create(tprObj)
	if err != nil {
		return nil, err
	}
	// krancour: Ideally the instance we return is a translation of the updated
	// 3pr as read back from k8s. It doesn't seem worth going through the trouble
	// right now since 3pr storage will be removed soon. This will at least work
	// well enough in the meantime.
	return sc, nil
}
