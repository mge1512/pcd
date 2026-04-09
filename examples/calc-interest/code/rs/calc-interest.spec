Name:           calc-interest
Version:        0.1.0
Release:        1%{?dist}
Summary:        Simple interest calculator — reads principal/rate/periods from stdin
License:        Apache-2.0
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  rust
BuildRequires:  cargo
BuildRequires:  pandoc

%description
calc-interest reads three numeric values (principal, annual rate, number of
periods) from standard input and writes the computed simple interest and total
repayment amount to standard output.

%prep
%autosetup

%build
pandoc %{name}.1.md -s -t man -o %{name}.1
RUSTFLAGS='-C target-feature=+crt-static' cargo build --release

%install
install -Dm755 target/release/%{name} %{buildroot}%{_bindir}/%{name}
install -Dm644 %{name}.1 %{buildroot}%{_mandir}/man1/%{name}.1

%files
%license LICENSE
%{_bindir}/%{name}
%{_mandir}/man1/%{name}.1*

%changelog
* Thu Apr 09 2026 Unknown <unknown@example.com> - 0.1.0-1
- Initial package
