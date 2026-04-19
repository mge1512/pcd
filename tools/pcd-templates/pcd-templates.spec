Name:           pcd-templates
Version:        0.3.22
Release:        1
Summary:        Deployment templates and library hints for the Post-Coding Development
BuildArch:      noarch

License:        CC-BY-4.0
URL:            https://github.com/mge1512/pcd

Source0:        pcd-templates-%{version}.tar.gz

# No build dependencies — data package only
BuildRequires:  (nothing)

# Both tools require this package at runtime
# They are not listed here as Requires — this is a data package,
# the tools declare Requires: pcd-templates in their own spec files.

%description
pcd-templates provides the deployment templates and library hints files
for the Post-Coding Development (PCD).

Deployment templates define language defaults, packaging conventions,
and AI translation execution recipes for each supported deployment type.
Library hints files provide verified dependency versions and API shapes
for common libraries.

Both pcd-lint and mcp-server-pcd read from the installed template
and hints directories at runtime.

Templates are installed under:
  /usr/share/pcd/templates/

Hints are installed under:
  /usr/share/pcd/hints/

%prep
%setup -q

%build
# Nothing to build — data package

%install
install -d %{buildroot}%{_datadir}/pcd/templates
install -d %{buildroot}%{_datadir}/pcd/hints

# Templates
install -m 0644 templates/backend-service.template.md  %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/cli-tool.template.md         %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/cloud-native.template.md     %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/gui-tool.template.md         %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/library-c-abi.template.md    %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/mcp-server.template.md       %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/project-manifest.template.md %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/python-tool.template.md      %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/verified-library.template.md %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/spack-package.template.md    %{buildroot}%{_datadir}/pcd/templates/
install -m 0644 templates/cockpit-module.template.md   %{buildroot}%{_datadir}/pcd/templates/

# Hints
install -m 0644 hints/cloud-native.go.go-libvirt.hints.md       %{buildroot}%{_datadir}/pcd/hints/
install -m 0644 hints/cloud-native.go.golang-crypto-ssh.hints.md %{buildroot}%{_datadir}/pcd/hints/
install -m 0644 hints/mcp-server.go.mcp-go.hints.md             %{buildroot}%{_datadir}/pcd/hints/
install -m 0644 hints/cli-tool.go.milestones.hints.md            %{buildroot}%{_datadir}/pcd/hints/
install -m 0644 hints/cli-tool.rs.milestones.hints.md            %{buildroot}%{_datadir}/pcd/hints/
install -m 0644 hints/python-tool.hints.md                       %{buildroot}%{_datadir}/pcd/hints/

%files
%license LICENSE
%dir %{_datadir}/pcd
%dir %{_datadir}/pcd/templates
%dir %{_datadir}/pcd/hints

# Templates
%{_datadir}/pcd/templates/backend-service.template.md
%{_datadir}/pcd/templates/cli-tool.template.md
%{_datadir}/pcd/templates/cloud-native.template.md
%{_datadir}/pcd/templates/gui-tool.template.md
%{_datadir}/pcd/templates/library-c-abi.template.md
%{_datadir}/pcd/templates/mcp-server.template.md
%{_datadir}/pcd/templates/project-manifest.template.md
%{_datadir}/pcd/templates/python-tool.template.md
%{_datadir}/pcd/templates/verified-library.template.md
%{_datadir}/pcd/templates/spack-package.template.md
%{_datadir}/pcd/templates/cockpit-module.template.md

# Hints
%{_datadir}/pcd/hints/cloud-native.go.go-libvirt.hints.md
%{_datadir}/pcd/hints/cloud-native.go.golang-crypto-ssh.hints.md
%{_datadir}/pcd/hints/mcp-server.go.mcp-go.hints.md
%{_datadir}/pcd/hints/cli-tool.go.milestones.hints.md
%{_datadir}/pcd/hints/cli-tool.rs.milestones.hints.md
%{_datadir}/pcd/hints/python-tool.hints.md

%changelog
* Sat Apr 19 2026 Matthias G. Eckermann <pcd@mailbox.org> - 0.3.22-1
- Add spack-package deployment template (10th template)
- Add cli-tool.go.milestones.hints.md, cli-tool.rs.milestones.hints.md
- Add python-tool.hints.md
* Fri Mar 27 2026 Matthias G. Eckermann <pcd@mailbox.org> - 0.3.19-1
- Initial release: all deployment templates and library hints
