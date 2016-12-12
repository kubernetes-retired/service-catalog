package storage

import (
	"errors"
	"log"

	"github.com/kubernetes-incubator/service-catalog/controller/util"
	"github.com/kubernetes-incubator/service-catalog/controller/watch"
	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"

	// Need this for gcp auth
	_ "k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/runtime"
)

type tprStorage struct {
	watcher *watch.Watcher
}

// CreateTPRStorage creates an instance of Storage backed by Kubernetes
// third-party resources.
func CreateTPRStorage(w *watch.Watcher) Storage {
	return &tprStorage{
		watcher: w,
	}
}

func (s *tprStorage) ListBrokers() ([]*scmodel.ServiceBroker, error) {
	l, err := s.watcher.GetResourceClient(watch.ServiceBroker, "default").List(&v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var ret []*scmodel.ServiceBroker
	for _, i := range l.(*runtime.UnstructuredList).Items {
		var tmp scmodel.ServiceBroker
		err := util.TPRObjectToSCObject(i, &tmp)
		if err != nil {
			log.Printf("Failed to convert object: %v\n", err)
			return nil, err
		}
		ret = append(ret, &tmp)
	}
	return ret, nil
}

func (s *tprStorage) GetBroker(name string) (*scmodel.ServiceBroker, error) {
	log.Printf("Getting broker: %s\n", name)

	sb, err := s.watcher.GetResourceClient(watch.ServiceBroker, "default").Get(name)
	if err != nil {
		return nil, err
	}
	var tmp scmodel.ServiceBroker
	err = util.TPRObjectToSCObject(sb, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

func (s *tprStorage) GetInventory() (*scmodel.Catalog, error) {
	l, err := s.watcher.GetResourceClient(watch.ServiceType, "default").List(&v1.ListOptions{})
	if err != nil {
		log.Printf("Failed to list service types: %v\n", err)
		return nil, err
	}
	var catalog scmodel.Catalog
	for _, i := range l.(*runtime.UnstructuredList).Items {
		var tmp scmodel.Service
		err := util.TPRObjectToSCObject(i, &tmp)
		if err != nil {
			log.Printf("Failed to convert object: %v\n", err)
			return nil, err
		}
		catalog.Services = append(catalog.Services, &tmp)
	}
	return &catalog, nil

}

func (s *tprStorage) AddBroker(broker *scmodel.ServiceBroker, catalog *scmodel.Catalog) error {
	broker.Kind = watch.ServiceBrokerKind
	broker.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(broker)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", broker, err)
		return err
	}
	tprObj.SetName(broker.Name)
	// TODO: Are brokers always in default namespace, if not, need to tweak this.
	log.Printf("Creating Broker: %s\n", broker.Name)
	s.watcher.GetResourceClient(watch.ServiceBroker, "default").Create(tprObj)

	// Then add all the service types.
	for _, st := range catalog.Services {
		st.Kind = watch.ServiceTypeKind
		st.APIVersion = watch.FullAPIVersion
		// TODO: Investigate using Metadata.ownerReference instead
		// (or in conjunction) with this
		st.Broker = broker.Name
		tprObj, err := util.SCObjectToTPRObject(st)
		if err != nil {
			log.Printf("Failed to convert object %#v : %v", st, err)
			return err
		}
		tprObj.SetName(st.Name)
		// TODO: Are brokers always in default namespace, if not, need to tweak this.
		log.Printf("Creating Service Type: %s\n", st.Name)
		s.watcher.GetResourceClient(watch.ServiceType, "default").Create(tprObj)
	}

	return nil
}

func (s *tprStorage) UpdateBroker(broker *scmodel.ServiceBroker, catalog *scmodel.Catalog) error {
	return errors.New("Not implemented yet")
}

func (s *tprStorage) DeleteBroker(id string) error {
	return errors.New("Not implemented yet")
}

func (s *tprStorage) GetServiceClass(name string) (*scmodel.Service, error) {
	si, err := s.watcher.GetResourceClient(watch.ServiceType, "default").Get(name)
	if err != nil {
		return nil, err
	}
	var tmp scmodel.Service
	err = util.TPRObjectToSCObject(si, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil

}

func (s *tprStorage) ListServiceInstances(ns string) ([]*scmodel.ServiceInstance, error) {
	l, err := s.watcher.GetResourceClient(watch.ServiceInstance, ns).List(&v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var ret []*scmodel.ServiceInstance
	for _, i := range l.(*runtime.UnstructuredList).Items {
		var tmp scmodel.ServiceInstance
		err := util.TPRObjectToSCObject(i, &tmp)
		if err != nil {
			log.Printf("Failed to convert object: %v\n", err)
			return nil, err
		}
		ret = append(ret, &tmp)
	}
	return ret, nil
}

// GetService returns the service instance with the specified name in the specified namespace
func (s *tprStorage) GetServiceInstance(ns string, name string) (*scmodel.ServiceInstance, error) {
	si, err := s.watcher.GetResourceClient(watch.ServiceInstance, ns).Get(name)
	if err != nil {
		return nil, err
	}
	var tmp scmodel.ServiceInstance
	err = util.TPRObjectToSCObject(si, &tmp)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

func (s *tprStorage) ServiceInstanceExists(ns string, name string) bool {
	_, err := s.GetServiceInstance(ns, name)
	return err == nil
}

// AddService creates a Service Instance Data. This method is
// deprecated and should be replaced with the one below.
// TODO: Get rid of this method and rename AddServiceRaw to this one...
func (s *tprStorage) AddServiceInstance(si *scmodel.ServiceInstance) error {
	si.Kind = watch.ServiceInstanceKind
	si.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(si)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", si, err)
		return err
	}
	tprObj.SetName(si.Name)
	log.Printf("Creating k8sobject:\n%v\n", tprObj)
	_, err = s.watcher.GetResourceClient(watch.ServiceInstance, "default").Create(tprObj)
	if err != nil {
		return err
	}
	return nil
}

func (s *tprStorage) UpdateServiceInstance(si *scmodel.ServiceInstance) error {
	si.Kind = watch.ServiceInstanceKind
	si.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(si)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", si, err)
		return err
	}
	tprObj.SetName(si.Name)
	_, err = s.watcher.GetResourceClient(watch.ServiceInstance, "default").Update(tprObj)
	if err != nil {
		return err
	}
	return nil

}

func (s *tprStorage) DeleteServiceInstance(string) error {
	return errors.New("Not implemented yet")
}

// ListServiceBindings returns all the bindings.
// TODO: wire in namespaces.
func (s *tprStorage) ListServiceBindings() ([]*scmodel.ServiceBinding, error) {
	l, err := s.watcher.GetResourceClient(watch.ServiceBinding, "default").List(&v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var ret []*scmodel.ServiceBinding
	for _, i := range l.(*runtime.UnstructuredList).Items {
		var tmp scmodel.ServiceBinding
		err := util.TPRObjectToSCObject(i, &tmp)
		if err != nil {
			log.Printf("Failed to convert object: %v\n", err)
			return nil, err
		}
		ret = append(ret, &tmp)
	}
	return ret, nil
}
func (s *tprStorage) GetServiceBinding(string) (*scmodel.ServiceBinding, error) {
	return nil, errors.New("Not implemented yet")
}

func (s *tprStorage) AddServiceBinding(in *scmodel.ServiceBinding, cred *scmodel.Credential) error {
	in.Kind = watch.ServiceBindingKind
	in.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(in)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", in, err)
		return err
	}
	tprObj.SetName(in.Name)
	log.Printf("Creating binding %s:\n%v\n", in.Name, tprObj)
	_, err = s.watcher.GetResourceClient(watch.ServiceBinding, "default").Create(tprObj)
	return err

}

func (s *tprStorage) UpdateServiceBinding(in *scmodel.ServiceBinding) error {
	in.Kind = watch.ServiceBindingKind
	in.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(in)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", in, err)
		return err
	}
	tprObj.SetName(in.Name)
	log.Printf("Updating Binding %s in k8s:\n%v\n", in.Name, tprObj)
	_, err = s.watcher.GetResourceClient(watch.ServiceBinding, "default").Update(tprObj)
	return err

}

func (s *tprStorage) DeleteServiceBinding(string) error {
	return errors.New("Not implemented yet")
}
