Name:           mcp-server-pcdp
Version:        0.1.0
Release:        1%{?dist}
Summary:        MCP server for PCDP specification management and linting

License:        GPL-2.0-only
URL:            https://github.com/pcdp/mcp-server-pcdp
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.24
Requires:       systemd

%description
mcp-server-pcdp is an MCP (Model Context Protocol) server that provides
tools and resources for managing PCDP (Post-Coding Development Paradigm)
specifications. It supports template management, resource serving, and
specification linting.

%prep
%setup -q

%build
export CGO_ENABLED=0
go build -ldflags="-X main.serverVersion=%{version}" -o %{name} .

%install
install -D -m 0755 %{name} %{buildroot}%{_bindir}/%{name}
install -D -m 0644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service

%files
%{_bindir}/%{name}
%{_unitdir}/%{name}.service

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%changelog
* Thu Mar 26 2026 Matthias G. Eckermann <pcdp@mailbox.org> - 0.1.0-1
- Initial release
