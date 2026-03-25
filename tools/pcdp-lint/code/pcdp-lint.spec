Name:           pcdp-lint
Version:        0.3.13
Release:        1
Summary:        Linter and validator for Post-Coding Development Paradigm specifications

License:        GPL-2.0-only
URL:            https://github.com/pcdp/pcdp-lint

Source0:        pcdp-lint-0.3.13.tar.gz

BuildRequires:  golang >= 1.21
Requires:       (nothing)

%description
pcdp-lint is a command-line tool that validates specification files written
in the Post-Coding Development Paradigm (PCDP) format. It enforces structural
rules, semantic validation, and cross-section consistency checks.

%prep
%setup -q

%build
CGO_ENABLED=0 go build -o pcdp-lint .

%install
install -D -m 0755 pcdp-lint %{buildroot}%{_bindir}/pcdp-lint

%files
%{_bindir}/pcdp-lint

%changelog
* Wed Mar 25 2026 Matthias G. Eckermann <pcdp@mailbox.org> - 0.3.13-1
- Initial release of pcdp-lint
