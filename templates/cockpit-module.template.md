# cockpit-module

## META

Deployment:   cockpit-module
Version:      0.1.0
Spec-Schema:  0.3.22
Author:       Matthias G. Eckermann <pcd@mailbox.org>
License:      GPL-2.0-only
Verification: none
Safety-Level: QM

---

## PURPOSE

A `cockpit-module` spec describes a plugin for the Cockpit web-based
server administration interface (https://cockpit-project.org).

Cockpit modules are installed under `/usr/share/cockpit/<name>/` and
consist of plain HTML, vanilla JavaScript (using the `cockpit.js` API),
and CSS. No Node.js, no webpack, no Python, no compiled backend binary
is required for modules that interact with system D-Bus services.

The `cockpit-bridge` process, which runs on the server side, proxies
D-Bus calls from the frontend JavaScript to the system bus. When the
user has "Administrative access" in Cockpit, the bridge runs with
elevated privileges, making calls to privileged D-Bus services (such as
`org.opensuse.Snapper`) work without sudo or polkit tricks.

This template targets the minimal-toolchain pattern: HTML + JS + CSS
only. Modules requiring a compiled custom bridge (e.g. PCP integration)
are out of scope for this template version.

---

## COCKPIT-VERSION

This template was derived from Cockpit documentation version as of
2026-04-20 (https://cockpit-project.org/guide/latest/packages,
https://cockpit-project.org/guide/latest/cockpit-dbus).

Primary source references:

| Reference | Purpose |
|---|---|
| `cockpit-project.org/guide/latest/packages` | Package layout, manifest.json schema, conditions, bridges |
| `cockpit-project.org/guide/latest/cockpit-dbus` | `cockpit.dbus()` API, D-Bus type mapping, superuser mode |
| `cockpit-project.org/guide/latest/cockpit-spawn` | `cockpit.spawn()` for optional CLI fallback |
| `cockpit-project.org/guide/latest/cockpit-file` | `cockpit.file()` for reading config files |

When Cockpit is updated, re-examine the `manifest.json` schema for new
fields (`conditions`, `bridges`, `parent`) and the `cockpit.dbus()` API
for new options. Bump the `COCKPIT-VERSION` block and `Version:` in META.

---

## TYPES

```
PackageName     := lowercase ASCII alphanumeric string, may contain underscore
                   // directory name under /usr/share/cockpit/
                   // e.g. "snapper", "my_tool"

MenuSection     := "dashboard" | "menu" | "tools"
                   // dashboard = Apps section
                   // menu      = System section
                   // tools     = Tools section

MenuOrder       := int
                   // lower numbers appear first
                   // System section conventions:
                   //   10 = System Information
                   //   20 = Logs
                   //   30-40 = Major subsystems
                   //   50-60 = VMs, Containers
                   //   70-100 = Accounts, Services

DBusService     := string  // e.g. "org.opensuse.Snapper"
DBusPath        := string  // e.g. "/org/opensuse/Snapper"
DBusInterface   := string  // e.g. "org.opensuse.Snapper"
DBusBus         := "system" | "session" | "user"

Condition       := PathExists | PathNotExists | Any
PathExists      := { "path-exists": AbsolutePath }
PathNotExists   := { "path-not-exists": AbsolutePath }
Any             := { "any": List<Condition> }

CockpitVersion  := string  // minimum cockpit version required
                            // e.g. "286" for path-exists condition support
```

---

## INTERFACES

```
Primary deliverables:
  <name>/manifest.json    Package manifest — declares menu placement,
                          conditions, required cockpit version, CSP
  <name>/index.html       Entry point HTML — loads cockpit.js and module JS/CSS
  <name>/<name>.js        Module logic — cockpit.dbus() calls, DOM updates
  <name>/<name>.css       Module styles — scoped to package, PatternFly optional

Installation path:
  /usr/share/cockpit/<name>/

cockpit.js API used:
  cockpit.dbus(service, options)     D-Bus client; options.bus = "system",
                                     options.superuser = "require" for root calls
  client.call(path, iface, method, args)   Make a D-Bus method call
  client.subscribe(match, handler)   Subscribe to D-Bus signals
  client.close()                     Release D-Bus connection
  cockpit.file(path, options)        Read/watch files (e.g. snapper configs)
  cockpit.spawn(argv, options)       Spawn a process (fallback only)

Man page deliverable:
  None — Cockpit modules are browser-based; no man page required.
```

---

## DEPENDENCIES

```
Runtime (must be present on the managed system):
  cockpit-bridge >= 286   // for path-exists condition support in manifest
  cockpit-ws              // web service frontend
  <target-service>        // D-Bus service the module talks to
                          // e.g. snapper (provides org.opensuse.Snapper)

Build-time:
  None                    // plain HTML/JS/CSS — no compiler, no npm, no webpack

Packaging:
  BuildArch: noarch       // no architecture-specific binaries
  Requires: cockpit-bridge
  Requires: <target-service-package>
```

---

## BEHAVIOR: manifest

Declare the package manifest that controls Cockpit integration.

```
INPUTS:
  name:           PackageName
  label:          string              human-readable menu entry
  section:        MenuSection
  order:          MenuOrder (optional)
  conditions:     List<Condition>     (optional, recommended)
  cockpit_version: CockpitVersion     (optional, recommended)
  csp_override:   string (optional)   Content-Security-Policy override

PRECONDITIONS:
  - name matches [a-z][a-z0-9_]*
  - label is non-empty
  - if conditions reference a D-Bus service executable, use path-exists
    on the daemon binary (e.g. /usr/bin/snapperd)
  - content-security-policy must not include unsafe-eval unless justified
    in spec INVARIANTS

STEPS:
  1. Create manifest.json with:
     a. "version": 0
     b. "require": {"cockpit": cockpit_version} if specified
     c. "conditions": conditions list if specified
     d. section key (dashboard/menu/tools) containing name -> {label, path, order}
     e. "content-security-policy" if CSP override needed
  2. Set path to "index.html" (default entry point)

POSTCONDITIONS:
  - manifest.json is valid JSON
  - package name equals the directory name
  - path referenced in manifest exists as a file in the package
  - conditions use path-exists on binaries or config files,
    not on D-Bus service names

ERRORS:
  - spaces in name: reject, use underscore or hyphen
  - path-exists on a D-Bus .service file: wrong — check the binary instead
```

---

## BEHAVIOR: html-entrypoint

Declare the HTML entry point that Cockpit loads in the browser frame.

```
INPUTS:
  title:    string    page title (shown in browser tab)
  name:     PackageName

PRECONDITIONS:
  - cockpit.js loaded from ../base1/cockpit.js (relative package path)
  - no inline scripts (Cockpit default CSP forbids them; use external .js file)
  - no external CDN resources (Cockpit CSP restricts connect-src to 'self')

STEPS:
  1. Create index.html with:
     a. <!DOCTYPE html> and lang attribute
     b. <meta charset="utf-8">
     c. <title> set to title
     d. <link rel="stylesheet"> for cockpit base styles if needed:
           ../base1/cockpit.css
        and for PatternFly if used:
           ../static/patternfly/patternfly.min.css
     e. <link rel="stylesheet"> for <name>.css
     f. <script src="../base1/cockpit.js"></script>
     g. <script src="<name>.js" type="module"></script>
     h. <body> with a single root element for JS to populate

POSTCONDITIONS:
  - no inline scripts or styles
  - all script/style src paths are relative
  - no absolute URLs to external resources

ERRORS:
  - inline <script>: move to external .js file
  - CDN URL in script/link src: remove; ship assets locally or use cockpit packages
```

---

## BEHAVIOR: dbus-client

Establish a D-Bus client connection to a system service.

```
INPUTS:
  service:    DBusService
  path:       DBusPath
  interface:  DBusInterface
  superuser:  bool    // true if service requires root access

PRECONDITIONS:
  - service is a known D-Bus service name (verify with d-feet or busctl)
  - superuser = true for any service running as root (e.g. snapperd)
  - client is created once per page load, not per method call
  - client.close() is called in the page unload handler

STEPS:
  1. Create D-Bus client:
       const client = cockpit.dbus(service, {
           bus: "system",
           superuser: superuser ? "require" : undefined
       });
  2. Attach error handler:
       client.addEventListener("close", handleClose);
  3. Store client reference for reuse across all method calls
  4. On page unload: client.close()

POSTCONDITIONS:
  - single client instance per D-Bus service per page
  - superuser: "require" used for all services requiring root
  - client is closed when page navigates away

ERRORS:
  - superuser: "require" fails: surface error to user with cockpit problem string
  - service not available: catch promise rejection, show "service unavailable" state
```

---

## BEHAVIOR: dbus-method-call

Make a D-Bus method call and handle the result.

```
INPUTS:
  client:     D-Bus client (from BEHAVIOR: dbus-client)
  path:       DBusPath
  interface:  DBusInterface
  method:     string
  args:       List<any>   // typed per D-Bus method signature

PRECONDITIONS:
  - client is open (not closed)
  - args match the D-Bus method signature exactly
  - promise rejection is always handled (no unhandled promise rejections)

STEPS:
  1. Call method:
       client.call(path, interface, method, args)
           .then(result => handleResult(result))
           .catch(error => handleError(error));
  2. In handleResult: unpack result[0] (first return value) or result array
  3. In handleError: display error.message to user; log to console

POSTCONDITIONS:
  - every call has both .then() and .catch() handlers
  - UI reflects loading state during pending call
  - errors are shown to user, not silently swallowed

ERRORS:
  - org.opensuse.Snapper.Error.*: surface the D-Bus error name and message
  - timeout: show "service not responding" state
```

---

## BEHAVIOR: dbus-signal

Subscribe to D-Bus signals for live UI updates.

```
INPUTS:
  client:     D-Bus client
  interface:  DBusInterface
  signal:     string      // signal name e.g. "SnapshotCreated"
  handler:    function    // (path, interface, signal, args) => void

PRECONDITIONS:
  - subscription is set up after client is created
  - subscription is removed when no longer needed (memory management)

STEPS:
  1. Subscribe:
       const subscription = client.subscribe(
           { interface: interface, member: signal },
           handler
       );
  2. In handler: update UI state based on signal args
  3. On page unload or component teardown: subscription.remove()

POSTCONDITIONS:
  - UI updates automatically when snapperd emits signals
  - no subscription leaks across page navigations

ERRORS:
  - signal arrives with unexpected args: log and ignore, do not crash
```

---

## BEHAVIOR: pipe-read

Read large D-Bus results via file descriptor pipe (for GetFilesByPipe pattern).

```
INPUTS:
  fd_channel:   cockpit channel object   // from D-Bus call returning fd
  encoding:     string                   // "utf8" for text pipe data

PRECONDITIONS:
  - used when D-Bus method returns a file descriptor (not a value)
  - Snapper's GetFilesByPipe returns fd; GetFiles is deprecated
  - fd is consumed once; channel closed after read

STEPS:
  1. Open channel from the returned fd descriptor:
       const channel = cockpit.channel({ payload: "stream", ...fd_channel });
  2. Accumulate data:
       let buffer = "";
       channel.addEventListener("message", (event, data) => { buffer += data; });
  3. On close: parse buffer line by line
       each line: split on first space -> [filename, status_int]
       decode filename: unescape "\x??" hex sequences
  4. Close channel after processing

POSTCONDITIONS:
  - all file entries from comparison are parsed
  - hex-escaped filenames are decoded to UTF-8
  - channel is closed after data is consumed

ERRORS:
  - pipe closes with error: surface to user
  - malformed line: skip and log, do not abort entire result
```

---

## INVARIANTS

```
[invariant]  No Node.js, no npm, no webpack, no build step required.
             The deliverables are plain files installable directly.
             // Rationale: supply chain simplicity; no npm dependency graph;
             //   installable via RPM without a build environment.

[invariant]  No Python backend, no compiled C/C++ backend for D-Bus modules.
             cockpit.dbus() with superuser: "require" is the privilege path.
             // Rationale: cockpit-bridge handles privilege correctly;
             //   a separate backend process adds attack surface and complexity.

[invariant]  No inline scripts or styles in HTML.
             Cockpit's default Content-Security-Policy forbids them.
             // From: cockpit packages CSP documentation.

[invariant]  No external CDN resources.
             All JS and CSS loaded from relative package paths or cockpit
             system packages (../base1/, ../static/).
             // Rationale: Cockpit CSP restricts connect-src to 'self'.
             //   Supply chain: external URLs introduce uncontrolled dependencies.

[invariant]  Every D-Bus method call has both .then() and .catch() handlers.
             // Rationale: unhandled promise rejections crash the page silently.

[invariant]  superuser: "require" is set for any D-Bus service running as root.
             // From: cockpit.dbus() API documentation, superuser option.

[invariant]  D-Bus client is created once per page load, not per method call.
             // Rationale: creating a new client per call leaks connections.

[invariant]  GetFilesByPipe is used instead of GetFiles for snapper comparisons.
             GetFiles is deprecated due to D-Bus message size limits.
             // From: snapper dbus-protocol.txt.

[invariant]  manifest.json conditions use path-exists on the daemon binary,
             not on the D-Bus .service file.
             // Rationale: .service file presence does not imply daemon is
             //   installed or functional; binary path is more reliable.

[invariant]  BuildArch: noarch in RPM spec.
             // Rationale: no compiled binaries; architecture-independent.

[observable] cockpit-bridge --packages lists the package name when installed.

[observable] Package is only visible in Cockpit UI when all conditions in
             manifest.json are satisfied.
```

---

## EXAMPLES

### EXAMPLE: minimal-manifest

```
GIVEN:
  name    = "snapper"
  label   = "Snapper"
  section = tools
  order   = 50
  condition: /usr/bin/snapperd must exist

WHEN:
  manifest.json is written

THEN:
  {
    "version": 0,
    "require": {"cockpit": "286"},
    "conditions": [
      {"path-exists": "/usr/bin/snapperd"}
    ],
    "tools": {
      "snapper": {
        "label": "Snapper",
        "path": "index.html",
        "order": 50
      }
    }
  }
```

### EXAMPLE: dbus-list-call

List snapshots for the "root" config via cockpit.dbus().

```
GIVEN:
  service   = "org.opensuse.Snapper"
  path      = "/org/opensuse/Snapper"
  interface = "org.opensuse.Snapper"
  method    = "ListSnapshots"
  args      = ["root"]
  superuser = true

WHEN:
  JavaScript calls the method

THEN:
  const client = cockpit.dbus("org.opensuse.Snapper", {
      bus: "system",
      superuser: "require"
  });

  client.call("/org/opensuse/Snapper", "org.opensuse.Snapper",
              "ListSnapshots", ["root"])
      .then(result => renderSnapshots(result[0]))
      .catch(error => showError(error.message));
```

### EXAMPLE: signal-subscription

Refresh snapshot list when snapperd signals a new snapshot.

```
GIVEN:
  signal = "SnapshotCreated"
  action = refresh the snapshot list for the affected config

WHEN:
  snapperd creates a snapshot

THEN:
  const sub = client.subscribe(
      { interface: "org.opensuse.Snapper", member: "SnapshotCreated" },
      (path, iface, signal, [configName, snapNum]) => {
          if (configName === currentConfig) loadSnapshots(currentConfig);
      }
  );
  // on page unload:
  sub.remove();
```

### EXAMPLE: audit-failure-inline-script

```
GIVEN:
  index.html contains <script>var x = 1;</script>

WHEN:
  Cockpit loads the page

THEN:
  Browser console: Content Security Policy violation
  Page functionality broken

CORRECT:
  Move all JavaScript to snapper.js and load via
  <script src="snapper.js" type="module"></script>
```

### EXAMPLE: audit-failure-external-cdn

```
GIVEN:
  index.html contains:
  <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>

WHEN:
  Cockpit loads the page

THEN:
  CSP blocks the external script load; page broken

CORRECT:
  Use only relative paths to cockpit system packages:
  <script src="../base1/cockpit.js"></script>
  Or ship the library as a local file within the package.
```

---

## MILESTONES

### MILESTONE: 0.0.0 — Scaffold
```
Scaffold: true
Constraint: required-for-release

Deliverables:
  manifest.json   valid JSON, passes: python3 -m json.tool manifest.json
  index.html      loads cockpit.js, renders static placeholder text
  <name>.js       empty module, no errors in browser console
  <name>.css      empty stylesheet

Acceptance:
  cockpit-bridge --packages   lists package name
  Browser: package appears in Cockpit menu (conditions met)
  Browser console: no JavaScript errors on page load
```

### MILESTONE: 0.1.0 — D-Bus connection + list
```
Constraint: required-for-release

Deliverables:
  D-Bus client established with superuser: "require"
  Primary list operation rendered in page (e.g. ListSnapshots)
  Error state shown when D-Bus service unavailable

Acceptance:
  Page shows snapshot/config list from live snapperd
  Disconnecting snapperd: page shows error state, does not crash
  Browser console: no unhandled promise rejections
```

### MILESTONE: 0.2.0 — Full MVP operations
```
Constraint: required-for-release

Deliverables:
  All BEHAVIOR: dbus-method-call operations implemented
  Signal subscriptions active for live refresh
  All error states handled and visible to user

Acceptance:
  All D-Bus methods called successfully against live snapperd
  Signal-driven refresh works without page reload
  RPM package installs, cockpit-bridge --packages lists it,
  package visible in Cockpit UI when snapperd installed
```

---

## EXECUTION

```
Language:       HTML5 + ES6 JavaScript (plain, no transpilation)
                CSS3 (optionally PatternFly from cockpit system packages)
Default stack:  cockpit.js API (../base1/cockpit.js)
                Optional: PatternFly CSS (../static/patternfly/)
No build step:  Files are delivered as-is; no webpack, no npm, no make required
EXECUTION:      none for the module code itself

Compile gate:   python3 -m json.tool manifest.json   (JSON validity)
                A browser load with open DevTools (console errors = fail)

RPM packaging conventions:
  Name:         cockpit-<name>
  BuildArch:    noarch
  BuildRequires: (none — no compiler needed)
  Requires:     cockpit-bridge
  Requires:     <daemon-package>
  Install:      %{_datadir}/cockpit/<name>/
  Source0:      %{name}-%{version}.tar.gz  (no vendor tarball needed)

RPM %install section:
  mkdir -p %{buildroot}%{_datadir}/cockpit/%{name}
  install -m 0644 manifest.json %{buildroot}%{_datadir}/cockpit/%{name}/
  install -m 0644 index.html    %{buildroot}%{_datadir}/cockpit/%{name}/
  install -m 0644 %{name}.js    %{buildroot}%{_datadir}/cockpit/%{name}/
  install -m 0644 %{name}.css   %{buildroot}%{_datadir}/cockpit/%{name}/

Makefile dist target:
  tar -czf cockpit-<name>-<version>.tar.gz \
      --transform 's|^|cockpit-<name>-<version>/|' \
      $(shell git ls-files)
```

---

## HINTS

```
// Cockpit-specific patterns and anti-patterns:

- cockpit.dbus() superuser: "require" means the bridge will prompt the
  user for their password if they do not yet have administrative access.
  This is the correct UX pattern — do not work around it.

- D-Bus type signatures are not always inferred automatically. If a call
  fails with a type error, pass the explicit DBus type signature as:
    client.call(path, iface, method, args, {type: "(sa{sv})"})

- cockpit.dbus() returns a Proxy or client, not a Promise.
  The .call() method returns a Promise. Do not confuse them.

- For arrays of structs (e.g. ListSnapshots returns an array of snapshot
  structs), result[0] is the array. Each element is an array of fields
  in the D-Bus struct order. Map them to named objects immediately.

- PatternFly CSS is available from Cockpit's own installed files — no npm
  required. Load it with a plain <link> tag:
    <link rel="stylesheet" href="../static/patternfly/patternfly.min.css">
  This gives full PatternFly visual consistency (design tokens, dark mode
  CSS variables, component classes) identical to what cockpit-snapshots
  compiles into its npm bundle — but loaded directly from the system.
  The CSS is installed as part of cockpit-bridge on both SLE 15 and SLE 16.
  Verify the exact path on the target distribution before writing tests;
  the location is stable across Cockpit versions but not formally guaranteed.
  PatternFly version present: PF4 on SLE 15 / Leap 15.x;
  PF5 on SLE 16 / Leap 16.x / Tumbleweed.
  For cross-version compatibility, use only CSS classes common to PF4 and PF5
  (layout, typography, button, table, alert, badge) and avoid PF5-only
  component classes in MVP.

- The "conditions" path-exists check runs at Cockpit startup, not on
  every page load. Install/remove of the target package requires a
  cockpit-bridge restart (or re-login) to take effect.

- cockpit.file() can watch a file for changes (polling). Use it to
  reflect /etc/snapper/configs/ changes in the UI without polling D-Bus.

- The pipe returned by GetFilesByPipe is a Unix fd. cockpit.channel()
  wraps it. Each line has format: "<filename> <status_integer>".
  Status bits: 1=content, 2=permissions, 4=user, 8=group, 16=xattr,
  32=acl (from snapper source). Decode \x?? escapes in filenames.

- Do not ship minified third-party JS inside the package. Cockpit's
  CSP and SUSE's supply chain requirements both require auditable source.
  Use cockpit system packages for shared libraries.
```
