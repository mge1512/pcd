Name:           pcd-skills
Version:        0.4.12
Release:        1
Summary:        Agent skills for the Post-Coding Development
BuildArch:      noarch

License:        CC-BY-4.0
URL:            https://github.com/mge1512/pcd

Source0:        pcd-skills-%{version}.tar.gz

# No build dependencies: data package only.
# The skills find pcd-lint and pcd-slice on PATH at session start and
# report when they are absent; they do not require them at install time.
Recommends:     pcd-lint
Recommends:     pcd-slice

%description
pcd-skills provides the agent skills of the Post-Coding Development
(PCD). Each skill is one directory with a SKILL.md at its root. The
directory also contains the deployment templates, hints files and rule
documents the skill reads, a guard script, and a manifest of every file.

An agent harness reads a skill from the installed directory or from a
copy of it in its own skills location. The guard script of each skill,
scripts/check-tools.sh, verifies the copy against the manifest and
reports which PCD tools are available.

Skills are installed under:
  /usr/share/pcd/skills/<name>/

The version of this package is the PCD framework version it was built
from; each skill has its own version in its SKILL.md front matter.

%prep
%setup -q

%build
# Nothing to build: data package

%install
install -d %{buildroot}%{_datadir}/pcd/skills
cp -R skills/. %{buildroot}%{_datadir}/pcd/skills/
find %{buildroot}%{_datadir}/pcd/skills -type f -name '*.sh' -exec chmod 0755 {} +
find %{buildroot}%{_datadir}/pcd/skills -type f ! -name '*.sh' -exec chmod 0644 {} +

%check
# Every installed skill must verify against its manifest.
for s in %{buildroot}%{_datadir}/pcd/skills/*/; do
    (cd "$s" && sha256sum -c --quiet --strict MANIFEST.sha256)
done

%files
%license LICENSE
%dir %{_datadir}/pcd
%{_datadir}/pcd/skills/

%changelog
* Wed Nov 04 2026 Matthias G. Eckermann <pcd@mailbox.org> - 0.4.12-1
- Initial release: pcd-translate skill with derived templates, hints,
  rule documents, guard script and manifest
