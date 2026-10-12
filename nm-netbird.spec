Name:           nm-netbird
Version:        0.8.1
Release:        1%{?dist}
Summary:        NetworkManager VPN plugin for NetBird
License:        MIT
URL:            https://github.com/Juju8901/nm-netbird
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.21
Requires:       NetworkManager
Requires:       netbird

%description
A NetworkManager VPN plugin that wraps NetBird, letting you toggle the
VPN from the GNOME network applet without the NetBird desktop UI.
Includes automatic browser OAuth when authentication is required.

%prep
%autosetup

%build
go build -mod=vendor -o nm-netbird-service .

%install
install -Dm755 nm-netbird-service \
    %{buildroot}%{_libexecdir}/nm-netbird-service
install -Dm644 nm-netbird-service.name \
    %{buildroot}%{_sysconfdir}/NetworkManager/VPN/nm-netbird-service.name
install -Dm644 nm-netbird.conf \
    %{buildroot}%{_sysconfdir}/NetworkManager/conf.d/nm-netbird.conf

%post
systemctl reload NetworkManager 2>/dev/null || true

%postun
if [ $1 -eq 0 ]; then
    systemctl reload NetworkManager 2>/dev/null || true
fi

%files
%{_libexecdir}/nm-netbird-service
%{_sysconfdir}/NetworkManager/VPN/nm-netbird-service.name
%config(noreplace) %{_sysconfdir}/NetworkManager/conf.d/nm-netbird.conf

%changelog
* Mon Oct 12 2026 Jeremiah Williams <jeremiahwilliams8901@gmail.com> - 0.8.1-1
- Build against Fedora 44

* Mon Oct 12 2026 Jeremiah Williams <jeremiahwilliams8901@gmail.com> - 0.8.0-1
- Initial release
