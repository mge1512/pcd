# pcd-lint spec for OBS RPM build
# pcd-spec-sha256: 293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9
# SPDX-License-Identifier: GPL-2.0-only

Name:           pcd-lint
Version:        0.4.0
Release:        0
Summary:        Validate Post-Coding Development specification files
License:        GPL-2.0-only
Group:          Development/Tools
URL:            https://github.com/mge1512/pcd/tools/pcd-lint
Source0:        pcd-lint-%{version}.tar.gz
BuildRequires:  go >= 1.21
BuildRequires:  pandoc

%description
pcd-lint validates specification files against the structural rules
defined in the pcd-lint specification (Spec-Schema 0.4.0).

It applies RULE-01 through RULE-21 and reports diagnostics on stderr
with a summary line on stdout.

%prep
%autosetup

%build
pandoc pcd-lint.1.md -s -t man -o pcd-lint.1
CGO_ENABLED=0 go build -ldflags="-s -w" -o pcd-lint ./cmd/pcd-lint/

%install
install -Dm755 pcd-lint %{buildroot}%{_bindir}/pcd-lint
install -Dm644 pcd-lint.1 %{buildroot}%{_mandir}/man1/pcd-lint.1

%files
%license LICENSE
%doc README.md
%{_bindir}/pcd-lint
%{_mandir}/man1/pcd-lint.1*

%changelog
