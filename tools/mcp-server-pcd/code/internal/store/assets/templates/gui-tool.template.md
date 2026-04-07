# gui-tool.template

## META
Deployment:   template
Version:      0.3.20
Spec-Schema:  0.3.20
Author:       Matthias G. Eckermann <pcd@mailbox.org>
License:      CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: gui-tool
EXECUTION:    none

---

## TYPES

```
Constraint := required | supported | default | forbidden

Platform := Linux | macOS | Windows | Android | iOS | embedded-linux
// Multiple values permitted — the set of declared platforms drives
// framework selection (see BEHAVIOR: resolve-framework).
// embedded-linux: no desktop window manager assumed; framebuffer or
//   DRM/KMS direct rendering; cross-compilation required.

GUIFramework := Qt6 | Tauri | Flutter
// Qt6:     C++ (or QML). Native widgets. Embedded-Linux primary.
//          Regulated/certified contexts. Digital sovereignty posture.
//          30-year track record. Qt Group (Helsinki) backing.
// Tauri:   Rust backend + web frontend (HTML/CSS/JS/Svelte/React).
//          Desktop-only (Linux/macOS/Windows). Best AI-assisted
//          development experience. Best supply chain posture (Rust).
// Flutter: Dart. Mobile-first (Android + iOS). Pixel-perfect cross-
//          platform. Required when Android or iOS is in PLATFORM set.

Language := CPP | Rust | Dart
// CPP:  Qt6 default. Also used for Qt6 + embedded-linux.
// Rust: Tauri backend. Frontend is HTML/CSS/JS — not a compiled
//       language in the PCD sense; translator generates web assets.
// Dart: Flutter. Dart VM on desktop/mobile; AOT on mobile.

OutputFormat := RPM | DEB | PKG | MSI | APK | AAB | IPA | binary
// RPM:    Linux RPM package via OBS
// DEB:    Linux DEB package via OBS
// PKG:    macOS installer package (.pkg)
// MSI:    Windows installer package
// APK:    Android application package (sideload / direct install)
// AAB:    Android App Bundle (Google Play)
// IPA:    iOS application archive (App Store / TestFlight)
// binary: raw executable or rootfs fragment (embedded-linux)

PackageFormat := RPM | DEB
// Native Linux package formats distributed via OBS.
// Preferred over OCI for GUI applications — applications are
// installed into the desktop session, not run as containers.
// OCI is off by default but permitted for edge/embedded use cases
// where GUI apps run in containers (e.g. kiosk, edge device, Wayland-in-container).
// Activate via preset: OUTPUT-FORMAT = OCI.
```

---

## BEHAVIOR: resolve-framework
Constraint: required

Given a spec declaring `Deployment: gui-tool`, derive the GUI framework
and language from the PLATFORM set declared in the spec.

INPUTS:
  platform-set: List<Platform>   // from spec META or preset
  framework-override: GUIFramework | none   // from preset only

PRECONDITIONS:
  - platform-set is non-empty
  - framework-override, if present, is compatible with platform-set
    (see STEPS for compatibility rules)

STEPS:
  1. If framework-override is set in preset:
     - If platform-set contains Android or iOS AND override is not Flutter:
       emit Error: "Framework override {override} is incompatible with
       mobile platform targets. Only Flutter supports Android and iOS."
     - If platform-set contains embedded-linux AND override is Flutter:
       emit Warning: "Flutter has limited embedded-linux support.
       Verify target hardware has sufficient resources."
     - If platform-set contains embedded-linux AND override is Tauri:
       emit Error: "Tauri does not support embedded-linux targets."
     - Otherwise: use override as resolved framework.
     On override accepted: go to step 4.

  2. Determine framework from platform-set (no override):
     - If platform-set contains Android OR iOS:
       resolved-framework = Flutter
       MECHANISM: mobile targets mandate Flutter — it is the only
       framework with production-quality Android and iOS support.
     - Else if platform-set contains ONLY embedded-linux
       OR platform-set contains embedded-linux AND Linux:
       resolved-framework = Qt6
       MECHANISM: Qt6 has dedicated embedded tooling (EGLFS, device
       creation, Yocto layer). Tauri and Flutter are not suitable.
     - Else:
       resolved-framework = Qt6 (default for desktop-only targets)

  3. Resolve language from framework:
     - Qt6   → CPP
     - Tauri → Rust
     - Flutter → Dart

  4. Emit resolved configuration:
     FRAMEWORK = resolved-framework
     LANGUAGE  = resolved-language

POSTCONDITIONS:
  - FRAMEWORK is one of Qt6, Tauri, Flutter
  - LANGUAGE is consistent with FRAMEWORK
  - Mobile targets always resolve to Flutter
  - embedded-linux targets always resolve to Qt6 (unless overridden)

ERRORS:
  - Error if framework-override is incompatible with platform-set
  - Error if Tauri is specified with embedded-linux target

---

## BEHAVIOR/INTERNAL: precedence-resolution
Constraint: required

Preset layering (systemd-style, later wins):
```
/usr/share/pcd/templates/gui-tool.template.md   ← this file
/usr/share/pcd/presets/                         ← vendor presets
/etc/pcd/presets/                               ← system admin
~/.config/pcd/presets/                          ← user
./.pcd/presets/                                 ← project-local
```

Framework override is only accepted from preset files, not from
spec META. The spec author declares PLATFORM; the preset may
override the framework choice within the compatibility rules.

STEPS:
1. Start with template defaults as the base map.
2. Merge /usr/share/pcd/presets/ values; later entries override earlier.
3. Merge /etc/pcd/presets/ values; overrides vendor defaults.
4. Merge ~/.config/pcd/presets/ values; overrides system.
5. Merge ./.pcd/presets/ values; overrides user.
6. If merged preset contains framework-override: validate compatibility
   with platform-set per BEHAVIOR: resolve-framework rules;
   on incompatibility → emit Error, halt.
7. Return merged preset map.

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | Semantic versioning. |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | PCD schema version. |
| AUTHOR | name \<email\> | required | Repeating key; multiple authors permitted. |
| LICENSE | SPDX identifier | required | Valid SPDX identifier or compound expression. |
| PLATFORM | Linux | default | Desktop Linux. RPM + DEB output required. |
| PLATFORM | macOS | supported | macOS desktop. PKG output required if declared. |
| PLATFORM | Windows | supported | Windows desktop. MSI output required if declared. |
| PLATFORM | Android | supported | Android mobile. Mandates Flutter. APK + AAB output. |
| PLATFORM | iOS | supported | iOS mobile. Mandates Flutter. IPA output. |
| PLATFORM | embedded-linux | supported | No window manager. Mandates Qt6. Cross-compile required. |
| FRAMEWORK | Qt6 | default | C++/QML. Native widgets. Default for all non-mobile targets. |
| FRAMEWORK | Tauri | supported | Rust + web frontend. Desktop-only. Preset override only. |
| FRAMEWORK | Flutter | supported | Dart. Required for Android/iOS. Permitted for desktop. |
| LANGUAGE | CPP | default | C++ with Qt6. Default when FRAMEWORK=Qt6. |
| LANGUAGE | Rust | supported | Rust backend with Tauri. Frontend assets are web (not compiled). |
| LANGUAGE | Dart | supported | Dart with Flutter. Required when FRAMEWORK=Flutter. |
| BINARY-TYPE | dynamic | default | GUI apps link Qt6/Flutter/WebKitGTK dynamically. System libraries. |
| BINARY-TYPE | static | supported | Permitted for embedded-linux Qt6 builds where rootfs space allows. |
| OUTPUT-FORMAT | RPM | required | Linux RPM via OBS. Required when PLATFORM includes Linux. |
| OUTPUT-FORMAT | DEB | required | Linux DEB via OBS. Required when PLATFORM includes Linux. |
| OUTPUT-FORMAT | PKG | supported | macOS .pkg. Required when PLATFORM includes macOS. |
| OUTPUT-FORMAT | MSI | supported | Windows MSI. Required when PLATFORM includes Windows. |
| OUTPUT-FORMAT | APK | supported | Android .apk. Required when PLATFORM includes Android. |
| OUTPUT-FORMAT | AAB | supported | Android App Bundle. Required when PLATFORM includes Android. |
| OUTPUT-FORMAT | IPA | supported | iOS .ipa. Required when PLATFORM includes iOS. |
| OUTPUT-FORMAT | binary | supported | Raw binary or rootfs fragment. Required for embedded-linux. |
| OUTPUT-FORMAT | OCI | supported | Container image. Off by default. Activate via preset for edge/kiosk deployments where GUI runs in a container (e.g. Wayland-in-container on edge device). Requires X11 or Wayland socket passthrough at runtime. |
| INSTALL-METHOD | OBS | required | Primary Linux distribution via build.opensuse.org. |
| INSTALL-METHOD | curl | forbidden | Supply chain security requirement. |
| INSTALL-METHOD | AppStore | supported | iOS App Store. Required when PLATFORM includes iOS. |
| INSTALL-METHOD | GooglePlay | supported | Google Play Store. Optional when PLATFORM includes Android. |
| SIGNAL-HANDLING | SIGTERM | required | Clean shutdown. Save state if applicable. No data loss. |
| SIGNAL-HANDLING | SIGINT | required | Clean shutdown on Ctrl-C. |
| QT6-RENDER | desktop | default | X11 or Wayland via Qt platform plugin. |
| QT6-RENDER | eglfs | supported | Embedded: direct DRM/KMS rendering, no window manager. |
| QT6-RENDER | linuxfb | supported | Embedded: Linux framebuffer fallback. |
| QT6-CROSS-COMPILE | false | default | Native build. |
| QT6-CROSS-COMPILE | true | supported | Required for embedded-linux targets. |
| FLUTTER-TARGET | desktop | default | Linux/macOS/Windows desktop when FRAMEWORK=Flutter. |
| FLUTTER-TARGET | mobile | supported | Android + iOS when FRAMEWORK=Flutter. |
| TAURI-FRONTEND | html-css-js | supported | Plain web frontend for Tauri. |
| TAURI-FRONTEND | svelte | supported | Svelte frontend (recommended for PCD — LLM-friendly). |
| TAURI-FRONTEND | react | supported | React frontend. |
| PRESET-SYSTEM | systemd-style | required | Preset layering follows systemd conventions. See whitepaper A.11. |

---

## TYPE-BINDINGS

| Spec type | LANGUAGE=CPP (Qt6) | LANGUAGE=Rust (Tauri) | LANGUAGE=Dart (Flutter) |
|---|---|---|---|
| String | QString | String | String |
| Path | QFileInfo | std::path::PathBuf | String (dart:io Path) |
| Duration | std::chrono::milliseconds | std::time::Duration | Duration |
| List\<T\> | QList\<T\> | Vec\<T\> | List\<T\> |
| Map\<K,V\> | QMap\<K,V\> | HashMap\<K,V\> | Map\<K,V\> |
| Optional\<T\> | std::optional\<T\> | Option\<T\> | T? |
| Result\<T\> | std::expected\<T,E\> | Result\<T,E\> | (throw / catch) |
| Bytes | QByteArray | Vec\<u8\> | Uint8List |
| Color | QColor | tauri::Color | Color |
| Signal | QSignal (Q_SIGNAL) | tauri::command | — |

---

## PRECONDITIONS

- This template is applied only when spec META declares `Deployment: gui-tool`
- PLATFORM must contain at least one value
- If PLATFORM contains Android or iOS, FRAMEWORK must resolve to Flutter
  (direct declaration or via framework-override compatibility check)
- If PLATFORM contains embedded-linux, FRAMEWORK must resolve to Qt6
  (direct declaration or via framework-override compatibility check)
- Cross-compilation toolchain must be declared in TOOLCHAIN-CONSTRAINTS
  when PLATFORM includes embedded-linux
- Preset files must be valid TOML

---

## POSTCONDITIONS

- FRAMEWORK is resolved and consistent with PLATFORM set
- LANGUAGE is consistent with FRAMEWORK
- All required OUTPUT-FORMATs for declared PLATFORMs are produced
- OCI output is produced only if explicitly activated via preset
- RPM and DEB are always produced when Linux is in PLATFORM set

---

## INVARIANTS

- [observable]      OCI output is produced only when preset activates OUTPUT-FORMAT=OCI
- [observable]      Mobile platform targets always resolve to Flutter
- [observable]      embedded-linux targets always resolve to Qt6
- [observable]      RPM and DEB are produced whenever Linux is in PLATFORM
- [implementation]  Framework selection logic is deterministic given PLATFORM set
- [implementation]  TYPE-BINDINGS table is applied without translator discretion

---

## EXAMPLES

EXAMPLE: linux-desktop-default
GIVEN:
  spec declares Deployment: gui-tool
  no PLATFORM or FRAMEWORK in preset
WHEN:
  resolve-framework runs
THEN:
  FRAMEWORK = Qt6
  LANGUAGE  = CPP
  OUTPUT-FORMAT includes RPM, DEB

EXAMPLE: mobile-mandates-flutter
GIVEN:
  spec declares Deployment: gui-tool
  PLATFORM = [Linux, Android, iOS]
WHEN:
  resolve-framework runs
THEN:
  FRAMEWORK = Flutter
  LANGUAGE  = Dart
  OUTPUT-FORMAT includes RPM, DEB, APK, AAB, IPA

EXAMPLE: embedded-linux-mandates-qt6
GIVEN:
  spec declares Deployment: gui-tool
  PLATFORM = [embedded-linux]
WHEN:
  resolve-framework runs
THEN:
  FRAMEWORK = Qt6
  LANGUAGE  = CPP
  QT6-RENDER options include eglfs, linuxfb
  OUTPUT-FORMAT includes binary
  RPM and DEB not required (no desktop Linux)

EXAMPLE: tauri-via-preset-desktop-only
GIVEN:
  spec declares Deployment: gui-tool
  PLATFORM = [Linux, macOS, Windows]
  preset declares framework-override = Tauri
WHEN:
  resolve-framework runs
THEN:
  FRAMEWORK = Tauri
  LANGUAGE  = Rust
  OUTPUT-FORMAT includes RPM, DEB, PKG, MSI

EXAMPLE: tauri-incompatible-with-mobile
GIVEN:
  spec declares Deployment: gui-tool
  PLATFORM = [Linux, Android]
  preset declares framework-override = Tauri
WHEN:
  resolve-framework runs
THEN:
  Error: "Framework override Tauri is incompatible with mobile platform targets"
  FRAMEWORK = unresolved
  translation does not proceed

EXAMPLE: tauri-incompatible-with-embedded
GIVEN:
  spec declares Deployment: gui-tool
  PLATFORM = [embedded-linux]
  preset declares framework-override = Tauri
WHEN:
  resolve-framework runs
THEN:
  Error: "Tauri does not support embedded-linux targets"
  FRAMEWORK = unresolved
  translation does not proceed

EXAMPLE: flutter-warning-on-embedded
GIVEN:
  spec declares Deployment: gui-tool
  PLATFORM = [embedded-linux]
  preset declares framework-override = Flutter
WHEN:
  resolve-framework runs
THEN:
  Warning: "Flutter has limited embedded-linux support"
  FRAMEWORK = Flutter (override accepted with warning)
  LANGUAGE  = Dart

EXAMPLE: resolve-framework-invalid-override-rejected
GIVEN:
  spec declares Deployment: gui-tool
  PLATFORM = [Android]
  preset declares framework-override = Tauri
WHEN:
  resolve-framework runs
THEN:
  errors contains: "Framework override Tauri is incompatible with mobile platform targets"
  FRAMEWORK is unresolved
  translation does not proceed
  exit_code = 1

---

## DELIVERABLES

The translator derives concrete filenames from the resolved FRAMEWORK,
LANGUAGE, and PLATFORM set. The following logical components apply:

| COMPONENT | Required when | Notes |
|---|---|---|
| implementation | always | Main application source. Language depends on FRAMEWORK. Qt6: `src/main.cpp` + CMakeLists.txt. Tauri: `src-tauri/` + frontend assets. Flutter: `lib/main.dart`. |
| build | always | Build system file. Qt6: CMakeLists.txt. Tauri: Cargo.toml + package.json. Flutter: pubspec.yaml. |
| rpm-package | PLATFORM includes Linux | `<n>.spec` — OBS RPM spec. |
| deb-package | PLATFORM includes Linux | `debian/` directory — standard layout. |
| macos-package | PLATFORM includes macOS | macOS .pkg descriptor. |
| windows-package | PLATFORM includes Windows | MSI descriptor or WiX config. |
| android-package | PLATFORM includes Android | `android/` directory. APK + AAB build config. |
| ios-package | PLATFORM includes iOS | `ios/` directory. Xcode project. IPA build config. |
| embedded-rootfs | PLATFORM includes embedded-linux | Cross-compilation config. Yocto layer recipe or OBS cross-build spec. |
| license | always | LICENSE file — SPDX reference only, no full text. |
| docs | always | README.md — installation, usage, platform notes. |
| man | always | `<n>.1.md`, `<n>.1` — Markdown source converted to troff via `pandoc`. Section 1. Install via RPM/DEB packaging. |
| tests | always | Independent test suite using declared test doubles. No live display required. |
| report | always | TRANSLATION_REPORT.md — always last. |

### Deliverable Content Requirements

**CMakeLists.txt (Qt6):**
- Must use `find_package(Qt6 REQUIRED COMPONENTS ...)` — do not vendor Qt
- Must set `CMAKE_AUTOMOC ON`, `CMAKE_AUTOUIC ON`, `CMAKE_AUTORCC ON`
- For embedded-linux: must include cross-compilation toolchain file reference

**Cargo.toml (Tauri):**
- Must declare `tauri` dependency with explicit version
- Do not fabricate versions — use hints file for verified version strings

**pubspec.yaml (Flutter):**
- Must declare `flutter` SDK dependency
- Do not use Google-controlled pub.dev packages without explicit declaration
  in spec DEPENDENCIES section

**RPM spec (`<n>.spec`):**
- `License:` must use SPDX identifier from spec META
- `BuildRequires:` must not include curl or any network fetch tool
- `BuildRequires:` must include `pandoc` (for man page generation)
- `%build` must include: `pandoc <n>.1.md -s -t man -o <n>.1`
- `%files` must include: `%{_mandir}/man1/<n>.1*`
- For Qt6: `BuildRequires: cmake qt6-base-devel` (and Qt module deps)
- For Flutter: `BuildRequires: flutter` (OBS Flutter package)
- For Tauri: `BuildRequires: cargo webkit2gtk-devel`

**debian/control:**
- `Build-Depends` must include `pandoc`

**debian/rules:**
- Must include man page build step: `pandoc <n>.1.md -s -t man -o <n>.1`
- Man page must be installed to `usr/share/man/man1/<n>.1`

**README.md:**
- Must document installation via OBS (zypper, apt, dnf)
- Must document supported platforms and their package formats
- Must NOT document curl-based installation
- For embedded-linux: must document cross-compilation procedure

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
It is read by pcd-lint (for template resolution validation) and by
AI translators (for code generation context).

Location in preset hierarchy:
  /usr/share/pcd/templates/gui-tool.template.md

Versioning:
  Template version is declared in META (Version: field).
  Specs reference the template by name (Deployment: gui-tool).
  Audit bundles record the template version used at generation time.
  Breaking changes increment the minor version.
  Current version: 0.3.18

Framework version guidance:
  Qt6:     Use distro-provided Qt6 packages via OBS BuildRequires.
           Do not vendor Qt. See hints/gui-tool.cpp.qt6.hints.md
           (to be written) for verified version strings per distro.
  Tauri:   See hints/gui-tool.rust.tauri.hints.md (to be written)
           for verified crate versions.
  Flutter: See hints/gui-tool.dart.flutter.hints.md (to be written)
           for verified SDK and pub.dev package versions.

---

## EXECUTION
EXECUTION: none

This template does not define a compile gate or delivery phases because
the build toolchain varies significantly by FRAMEWORK and PLATFORM:

- Qt6 + CMake: `cmake -B build && cmake --build build`
- Qt6 + embedded: cross-compilation via sysroot or Yocto SDK
- Tauri: `cargo tauri build`
- Flutter desktop: `flutter build linux` / `flutter build macos` etc.
- Flutter mobile: `flutter build apk` / `flutter build ipa`

The translator must derive the correct build commands from the resolved
FRAMEWORK and PLATFORM set, and document them in TRANSLATION_REPORT.md
under "Build gate — \<PLATFORM\>".

A component-specific prompt (`tools/<n>/spec/prompt.md`) should override
this section with concrete build commands once FRAMEWORK and PLATFORM
are known for a specific project.
