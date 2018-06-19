# Bug Report 

**What happened**:

**What you expected to happen**:

**How to reproduce it (as minimally and precisely as possible)**:

**Anything else we need to know?**:

**Environment**:
- Kubernetes version (use `kubectl version`):
- service-catalog version:
- Cloud provider or hardware configuration:
- Do you have api aggregation enabled? 
  - Do you see the configmap in kube-system? 
  - Does it have all the necessary fields?
    - `kubectl get cm -n kube-system extension-apiserver-authentication -o yaml` and look for `requestheader-XXX` fields
- Install tools:
  - Did you use helm? What were the helm arguments? Did you `--set` any extra values?
- Are you trying to use ALPHA features? Did you enable them?
