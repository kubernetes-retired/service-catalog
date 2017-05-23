#debuginfo not supported with Go
%global debug_package %{nil}
# modifying the Go binaries breaks the DWARF debugging
%global __os_install_post %{_rpmconfigdir}/brp-compress

%global project github.com/kubernetes-incubator
%global repo service-catalog
%global import_path %{project}/%{repo}

# update to match the upstream tagged version
%global commit 9d18e67399edb37a1fa9a254f6fa4f7a128539bc
%global shortcommit %(c=%{commit}; echo ${c:0:7})

Name:           service-catalog
Version:        0.0.1
Release:        1.git%{shortcommit}%{?dist}
Summary:        Service Catalog for Kubernetes
License:        ASL 2.0
URL:            https://%{import_path}

Source0:        https://%{import_path}/archive/%{commit}/%{repo}-%{shortcommit}.tar.gz
BuildRequires:  golang

%description
%{summary}

# If go_arches not defined fall through to implicit golang archs
%if 0%{?go_arches:1}
ExclusiveArch:  %{go_arches}
%else
ExclusiveArch:  x86_64 aarch64 ppc64le s390x
%endif

%prep
%setup -q -n %{name}-%{commit}

%build
# GOPATH hackery
export GOPATH=$(pwd):%{gopath}
mkdir -p src/github.com/kubernetes-incubator
ln -s ../../../ src/github.com/kubernetes-incubator/service-catalog

# prevent code regen hackery
mkdir bin
pushd bin
touch defaulter-gen deepcopy-gen conversion-gen client-gen lister-gen informer-gen openapi-gen
popd
touch .generate_files .init .generate_exes

# build binaries
VERSION=%{version}-git%{shortcommit} NO_DOCKER=1 make bin/apiserver
VERSION=%{version}-git%{shortcommit} NO_DOCKER=1 make bin/controller-manager

%install
install -d %{buildroot}%{_bindir}
install -p -m 755 bin/apiserver %{buildroot}%{_bindir}/%{name}-apiserver
install -p -m 755 bin/controller-manager %{buildroot}%{_bindir}/%{name}-controller-manager

%files
%doc README.md
%license LICENSE
%{_bindir}/%{name}-apiserver
%{_bindir}/%{name}-controller-manager

%changelog
* Thu Apr 06 2017 Seth Jennings <sjenning@redhat.com> 0.0.1-1
- Initial spec file
- Upstream release v0.0.1
