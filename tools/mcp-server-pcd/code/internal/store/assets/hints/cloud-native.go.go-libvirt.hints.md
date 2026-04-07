
# cloud-native · Go · github.com/digitalocean/go-libvirt

## META
Version:  1.0.0
Library:  github.com/digitalocean/go-libvirt
Language: Go
Template: cloud-native

---

## Version selection

This module has no tagged releases. Use a verified pseudo-version.
**DO NOT fabricate commit hashes or timestamps.** A fabricated pseudo-version
causes `go mod tidy` to fail with `invalid version: unknown revision`.

Verification procedure before use:
```bash
git ls-remote https://github.com/digitalocean/go-libvirt.git HEAD
```

Pseudo-version format: `v0.0.0-YYYYMMDDHHMMSS-<12-char-commit-prefix>`
The timestamp must be the actual commit time in UTC, not fabricated.

Known-good version (used by containers/podman and lima-vm/lima — widely tested):
```
github.com/digitalocean/go-libvirt v0.0.0-20220804181439-8648fbde413e
```

---

## Domain struct fields (NOT method calls)

`libvirt.Domain` is a struct with fields, not an object with getter methods.
Access these as struct fields:

```go
domain.Name  // string — the domain name
domain.UUID  // [16]byte — UUID; no DomainGetUUIDString method exists
```

**UUID formatting** — convert `[16]byte` to canonical UUID string:
```go
u := domain.UUID
uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
    u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
```

---

## Methods that do NOT exist (common translator hallucinations)

| Hallucinated call | Correct approach |
|---|---|
| `DomainGetName(domain)` | Use `domain.Name` field |
| `DomainGetUUIDString(domain)` | Does not exist; use `domain.UUID [16]byte` field |
| `DomainGetUUID(domain)` | Does not exist; use `domain.UUID [16]byte` field |

---

## DomainGetInfo — returns 6 individual values (not a struct)

```go
state, maxMem, memory, nrVirtCPU, cpuTime, err :=
    conn.DomainGetInfo(domain)
// state     uint8   — domain state; map to VMObservedState via switch
// maxMem    uint64  — maximum memory in KiB
// memory    uint64  — used memory in KiB
// nrVirtCPU uint16  — number of virtual CPUs
// cpuTime   uint64  — cumulative CPU time in nanoseconds
// err       error
```

Despite `DomainGetInfoRet` existing as a struct type in the package,
the `DomainGetInfo` method returns 6 individual values, not a struct.

Domain state uint8 mapping to VMObservedState:
```go
func mapDomainState(state uint8) VMObservedState {
    switch state {
    case 1: return VMObservedStateRunning
    case 2: return VMObservedStateBlocked
    case 3: return VMObservedStatePaused
    case 4: return VMObservedStateShutdown
    case 5: return VMObservedStateShutoff
    case 6: return VMObservedStateCrashed
    case 7: return VMObservedStatePMSuspended
    default: return VMObservedStateUnknown
    }
}
```

---

## DomainUndefineFlags — type-safe flags

```go
// NOT uint32 — the library uses its own named type
flags := libvirt.DomainUndefineFlagsValues(1 | 2 | 4)
// 1 = VIR_DOMAIN_UNDEFINE_MANAGED_SAVE
// 2 = VIR_DOMAIN_UNDEFINE_SNAPSHOTS_METADATA
// 4 = VIR_DOMAIN_UNDEFINE_NVRAM
err := conn.DomainUndefineFlags(domain, flags)
```

---

## ConnectGetMaxVcpus — requires OptString, not plain string

```go
// NOT: conn.ConnectGetMaxVcpus("kvm")
maxVCPUs, err := conn.ConnectGetMaxVcpus(libvirt.OptString{"kvm"})
```

---

## ConnectGetLibVersion and ConnectGetVersion — return uint64

```go
v, err := conn.ConnectGetLibVersion()   // libvirt daemon version
q, err := conn.ConnectGetVersion()      // QEMU/hypervisor version

// Both return uint64 (NOT uint32).
// Format as "major.minor.patch":
major := v / 1000000
minor := (v % 1000000) / 1000
patch := v % 1000
versionStr := fmt.Sprintf("%d.%d.%d", major, minor, patch)
```

---

## NodeGetFreeMemory — returns single uint64 byte count

```go
freeBytes, err := conn.NodeGetFreeMemory()
freeMiB := freeBytes / 1024 / 1024
```

Note: `NodeGetInfo` is a separate call that returns 9 individual values
(model string, memory, cpus, mhz, nodes, sockets, cores, threads, err).
Use `NodeGetFreeMemory` for free memory only.

---

## Connecting via SSH tunnel to libvirt Unix socket

```go
// 1. Dial the SSH connection
sshClient, err := ssh.Dial("tcp", addr, config)

// 2. Forward the libvirt Unix socket through SSH
conn, err := sshClient.Dial("unix", "/run/libvirt/libvirt-sock")

// 3. Create libvirt client and connect
lv := libvirt.New(conn)
err = lv.Connect()

// 4. On close: disconnect libvirt first, then close SSH
lv.Disconnect()
sshClient.Close()
```

---

## DomainDefineXML — returns libvirt.Domain struct

```go
domain, err := conn.DomainDefineXML(xmlString)
// domain.Name is the domain name (string field, not a method)
// domain.UUID is the UUID ([16]byte field, not a method)
```

---

## DomainLookupByName — use for looking up existing domains

```go
domain, err := conn.DomainLookupByName(name)
// Returns libvirt.Domain struct on success
// Returns error if domain not found
```

---

## Import path

```go
import "github.com/digitalocean/go-libvirt"
```

Package alias `libvirt` is conventional.
