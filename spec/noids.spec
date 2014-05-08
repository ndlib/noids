# debuginfo not supported with Go
%global debug_package	%{nil}
%global gopath	%{_datadir}/gocode
%global __strip	/bin/true

Name:		noids
Version:	1.0.0
Release:	1%{?dist}
Summary:	Noids identifier server

Group:		System Environment/Daemons
License:	Apache

BuildRoot:	%(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

Url:		https://github.com/dbrower/noids
Source:		https://github.com/dbrower/noids/archive/noids-master-%{version}.zip

BuildRequires:	golang >= 1.2-7
# the remainder are local packages
#BuildRequires:	godep

Provides:	noids = %{version}

ExclusiveArch:	%{ix86} x86_64 %{arm}

%description
Noids is an identity server which plays nicely with the Ruby
NOID gem and Hydra based applications. It can be either file
or database backed.

%prep
%setup -q -n noids-master
mkdir _build
pushd _build
  mkdir -p src/github.com/dbrower
  ln -s $(dirs +1 -l) src/github.com/dbrower/noids
popd

%build
export GOPATH=$(pwd)/_build:$(pwd)/Godeps/_workspace:%{gopath}
go build
pushd cmd/noid-tool
  go build
popd


%install
install -d %{buildroot}%{_bindir}
install -d %{buildroot}/opt/noids
install -d %{buildroot}/opt/noids/bin
install -d %{buildroot}/opt/noids/log
install -d %{buildroot}/opt/noids/pools
install -p -m 755 noids-master %{buildroot}/opt/noids/bin/noids
install -p -m 755 cmd/noid-tool/noid-tool %{buildroot}%{_bindir}/noid-tool
install -p -m 644 spec/noids.logrotate %{buildroot}/etc/logrotate.d/noids
install -p -m 644 spec/noids.conf %{buildroot}/etc/init/noids.conf

%files
%defattr(-,root,root,-)
/opt/noids/
%{_bindir}/noid-tool
/etc/logrotate.d/noids
/etc/init/noids.conf
#%doc README.md
#%{_mandir}/man1/docker.1.gz
#%config(noreplace) %{_sysconfdir}/noids
#%{_initddir}/docker

%changelog

