package storage

import (
	"errors"
	"log"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/util"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/watch"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/runtime"
)

type tprStorageBroker struct {
	watcher *watch.Watcher
}

func newTPRStorageBroker(watcher *watch.Watcher) *tprStorageBroker {
	return &tprStorageBroker{watcher: watcher}
}

func (t *tprStorageBroker) List() ([]*servicecatalog.Broker, error) {
	l, err := t.watcher.GetResourceClient(watch.ServiceBroker, "default").List(&v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var ret []*servicecatalog.Broker
	for _, i := range l.(*runtime.UnstructuredList).Items {
		var tmp servicecatalog.Broker
		err := util.TPRObjectToSCObject(i, &tmp)
		if err != nil {
			log.Printf("Failed to convert object: %v\n", err)
			return nil, err
		}
		ret = append(ret, &tmp)
	}
	return ret, nil
}

func (t *tprStorageBroker) Get(name string) (*servicecatalog.Broker, error) {
	log.Printf("Getting broker: %s\n", name)

	sb, err := t.watcher.GetResourceClient(watch.ServiceBroker, "default").Get(name)
	if err != nil {
		return nil, err
	}
	var tmp servicecatalog.Broker
	err = util.TPRObjectToSCObject(sb, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

func (t *tprStorageBroker) Create(broker *servicecatalog.Broker) (*servicecatalog.Broker, error) {
	broker.Kind = watch.ServiceBrokerKind
	broker.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(broker)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", broker, err)
		return nil, err
	}
	tprObj.SetName(broker.Name)
	// TODO: Are brokers always in default namespace, if not, need to tweak this.
	log.Printf("Creating Broker: %s\n", broker.Name)
	t.watcher.GetResourceClient(watch.ServiceBroker, "default").Create(tprObj)

	// krancour: Ideally the broker we return is a translation of the updated 3pr
	// as read back from k8s. It doesn't seem worth going through the trouble
	// right now since 3pr storage will be removed soon. This will at least work
	// well enough in the meantime.
	return broker, nil
}

func (t *tprStorageBroker) Update(broker *servicecatalog.Broker) (*servicecatalog.Broker, error) {
	return nil, errors.New("Not implemented yet")
}

func (t *tprStorageBroker) Delete(id string) error {
	return errors.New("Not implemented yet")
}
