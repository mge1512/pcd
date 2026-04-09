#
# calc-interest.spec — RPM spec file for OBS build
#
# Spec: calc-interest v0.1.0   License: Apache-2.0
# Template: cli-tool.template.md v0.3.20
# INSTALL-METHOD: OBS (required)
# OUTPUT-FORMAT: RPM (required)
#
# Source0 references a local tarball (no network fetch — supply chain requirement).
#

Name:           calc-interest
Version:        0.1.0
Release:        1%{?dist}
Summary:        Simple interest calculator CLI tool
License:        Apache-2.0
URL:            https://github.com/example/calc-interest
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  java-17-openjdk-devel
BuildRequires:  maven
BuildRequires:  pandoc

# No runtime deps beyond the JRE (fat JAR bundles everything)
Requires:       java-17-openjdk-headless

%description
calc-interest reads principal, annual rate, and number of periods from
standard input (one value per line), computes simple interest and total
repayment amount, and writes the results to standard output.

Simple interest formula: interest = principal * rate * periods

%prep
%setup -q

%build
# Compile and package fat JAR (no network access during build)
mvn -q package -DskipTests -Dmaven.repo.local=%{_builddir}/.m2

# Generate man page
pandoc %{name}.1.md -s -t man -o %{name}.1

%install
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_javadir}
install -d %{buildroot}%{_mandir}/man1

install -m 644 target/%{name}.jar %{buildroot}%{_javadir}/%{name}.jar

# Thin wrapper script
cat > %{buildroot}%{_bindir}/%{name} <<'EOF'
#!/bin/sh
exec java -jar %{_javadir}/%{name}.jar "$@"
EOF
chmod 755 %{buildroot}%{_bindir}/%{name}

install -m 644 %{name}.1 %{buildroot}%{_mandir}/man1/%{name}.1

%files
%license LICENSE
%{_bindir}/%{name}
%{_javadir}/%{name}.jar
%{_mandir}/man1/%{name}.1*

%changelog
* Thu Apr 09 2026 Unknown <unknown@example.com> - 0.1.0-1
- Initial package
