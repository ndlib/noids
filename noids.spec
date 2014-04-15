# debuginfo not supported with Go
%global debug_package	%{nil}
%global gopath	%{_datadir}/gocode
%global __strip	/bin/true

Name:		noids
Version:	1
Release:	1%{?dist}
Summary:	Noids identifier server

Group:		System Environment/Daemons
License:	Apache

BuildRoot:	%(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

Url:		https://github.com/dbrower/noids
Source:		https://github.com/dbrower/noids/archive/noids-capify-%{version}.zip

BuildRequires:	golang >= 1.2-7
# the remainder are local packages
BuildRequires:	godep

Provides:	noids = %{version}

ExclusiveArch:	%{ix86} x86_64 %{arm}

%description
Noids is an identity server which plays nicely with the Ruby
NOID gem and Hydra based applications. It can be either file
or database backed.

%prep
%setup -q -n noids-capify
mkdir _build
pushd _build
  mkdir -p src/github.com/dbrower
  ln -s $(dirs +1 -l) src/github.com/dbrower/noids
popd

%build
export GOPATH=$(pwd)/_build:%{gopath}
godep go build cmd/noids/main.go


%install
install -d %{buildroot}%{_bindir}
install -p -m 755 main %{buildroot}%{_bindir}/noids

%files
%defattr(-,root,root,-)
%{_bindir}/noids
#%doc AUTHORS CHANGELOG.md CONTRIBUTING.md FIXME LICENSE MAINTAINERS NOTICE README.md 
#%doc LICENSE-vim-syntax README-vim-syntax.md
#%{_mandir}/man1/docker.1.gz
#%config(noreplace) %{_sysconfdir}/noids
#%{_initddir}/docker

%changelog

