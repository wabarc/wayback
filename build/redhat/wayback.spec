# Copyright 2020 Wayback Archiver. All rights reserved.
# Use of this source code is governed by the GNU GPL v3
# license that can be found in the LICENSE file.
#
%undefine _disable_source_fetch

Name:    wayback
Version: %{_wayback_version}
Release: 1.0
Summary: Easy and fast to wayback webpage.
URL: https://github.com/wabarc/wayback
License: GNU General Public License v3.0
Source0: wayback
Source1: wayback.service
Source2: wayback.1
Source3: LICENSE
Source4: CHANGELOG.md
BuildRoot: %{_topdir}/BUILD/%{name}-%{version}-%{release}
BuildArch: x86_64
Requires(pre): shadow-utils

%{?systemd_requires}
BuildRequires: systemd

%description
%{summary}

%install
mkdir -p %{buildroot}%{_bindir}
install -p -m 755 %{SOURCE0} %{buildroot}%{_bindir}/wayback
install -D -m 644 %{SOURCE1} %{buildroot}%{_unitdir}/wayback.service
install -D -m 644 %{SOURCE2} %{buildroot}%{_mandir}/man1/wayback.1
install -D -m 644 %{SOURCE3} %{buildroot}%{_docdir}/wayback/LICENSE
install -D -m 644 %{SOURCE4} %{buildroot}%{_docdir}/wayback/CHANGELOG.md

%files
%defattr(755,root,root)
%{_bindir}/wayback
%{_docdir}/wayback
%defattr(644,root,root)
%{_unitdir}/wayback.service
%{_mandir}/man1/wayback.1*
%{_docdir}/wayback/*
%defattr(600,root,root)

%pre
getent group wayback >/dev/null || groupadd -r wayback
getent passwd wayback >/dev/null || \
    useradd -r -g wayback -d /dev/null -s /sbin/nologin \
    -c "Wayback Daemon" wayback
exit 0

%post
%systemd_post wayback.service

%preun
%systemd_preun wayback.service

%postun
%systemd_postun_with_restart wayback.service

%changelog
