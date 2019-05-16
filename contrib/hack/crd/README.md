## Cookbook for CRDs POC 

Execute all commands from the cookbook in the `hack` directory.

### Bootstrap local environment for testing

Execute command:
```bash
./bin/bootstrap-testing-environment.sh
```

Under the hood this script is:
- creating minikube
- installing tiller
- installing Service Catalog
- installing and registering the UPS Broker

**Now you are ready to go!**

When you execute `svcat get classes`, then you should see:
```bash
                 NAME                  NAMESPACE         DESCRIPTION
+------------------------------------+-----------+-------------------------+
  user-provided-service                            A user provided service
  user-provided-service-single-plan                A user provided service
  user-provided-service-with-schemas               A user provided service
``` 

### Testing Scenario

Follow the [Walkthrough Scenario](../../../docs/walkthrough.md) and start with **Step 3 - Viewing ClusterServiceClasses and ClusterServicePlans**. 