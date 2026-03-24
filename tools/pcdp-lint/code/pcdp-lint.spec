Name:           pcdp-lint
Version:        0.3.13
Release:        1%{?dist}
Summary:        Post-Coding Development Paradigm specification validator

License:        GPL-2.0-only
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.21
Requires:       glibc

%description
A command-line tool for validating Post-Coding Development Paradigm (PCDP) 
specification files. pcdp-lint validates PCDP specs against 13 structural 
rules including required sections, META field validation, deployment template 
resolution, and cross-section consistency checks.

%prep
%setup -q

%build
export CGO_ENABLED=0
export TEMPLATE_DIR=/usr/share/pcdp/templates/
go build -ldflags="-s -w -X main.templateDir=%{TEMPLATE_DIR}" -o %{name} .

%install
install -d %{buildroot}%{_bindir}
install -m 755 %{name} %{buildroot}%{_bindir}/

install -d %{buildroot}%{_datadir}/pcdp/templates
# Template files will be provided by pcdp-templates package

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%dir %{_datadir}/pcdp
%dir %{_datadir}/pcdp/templates

%changelog
* Mon Mar 24 2026 Build System <build@example.org> - 0.3.13-1
- Initial package for pcdp-lint
- Implements all 13 validation rules from PCDP spec schema 0.3.13
- Static binary with no runtime dependencies
- Supports strict mode and template listing