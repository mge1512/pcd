Name:           pcd-lint
Version:        0.3.21
Release:        1%{?dist}
Summary:        Post-Coding Development specification linter
License:        GPL-2.0-only
Source0:        %{name}-%{version}.tar.gz
BuildRequires:  golang
BuildRequires:  pandoc

%description
pcd-lint validates Post-Coding Development (PCD) specification files
against the structural rules defined in the PCD spec schema.
It checks required sections, META fields, SPDX license identifiers,
BEHAVIOR blocks, EXAMPLES structure, INVARIANTS tags, and more.

%prep
%autosetup

%build
CGO_ENABLED=0 go build -ldflags="-s -w" -o %{name} .
pandoc %{name}.1.md -s -t man -o %{name}.1

%install
install -Dm755 %{name} %{buildroot}%{_bindir}/%{name}
install -Dm644 %{name}.1 %{buildroot}%{_mandir}/man1/%{name}.1

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_mandir}/man1/%{name}.1*

%changelog
* Tue Apr 07 2026 Matthias G. Eckermann <pcd@mailbox.org> - 0.3.21-1
- Initial package release
