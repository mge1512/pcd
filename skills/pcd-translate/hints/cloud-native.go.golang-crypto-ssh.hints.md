
# cloud-native · Go · golang.org/x/crypto/ssh

## META
Version:  1.0.0
Library:  golang.org/x/crypto/ssh
Language: Go
Template: cloud-native

---

## Functions that do NOT exist (common translator hallucinations)

| Hallucinated call | Correct approach |
|---|---|
| `ssh.KeysEqual(a, b)` | Compare `string(a.Marshal()) == string(b.Marshal())` |
| `ssh.IsHostKeyMismatch(err)` | Inspect `err.Error()` string — see below |
| `ssh.IsAuthenticationError(err)` | Inspect `err.Error()` string — see below |

There is no typed error hierarchy in `golang.org/x/crypto/ssh`. Error
classification must be done by inspecting the error string.

---

## Key comparison — compare marshalled bytes

```go
// There is no ssh.KeysEqual function. Compare the wire format bytes:
if string(presented.Marshal()) != string(expected.Marshal()) {
    return fmt.Errorf("host key mismatch")
}
```

---

## Host key parsing from a known_hosts line

```go
// Parse a single known_hosts entry (one line from a ConfigMap value):
_, _, expectedKey, _, _, err := ssh.ParseKnownHosts([]byte(knownHostsLine))
if err != nil {
    return fmt.Errorf("failed to parse known_hosts line: %v", err)
}
```

---

## HostKeyCallback — strict checking implementation

```go
func strictHostKeyCallback(knownHostsLine string) ssh.HostKeyCallback {
    return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
        _, _, expectedKey, _, _, err := ssh.ParseKnownHosts([]byte(knownHostsLine))
        if err != nil {
            return fmt.Errorf("failed to parse known_hosts: %v", err)
        }
        if string(key.Marshal()) != string(expectedKey.Marshal()) {
            return fmt.Errorf("host key mismatch for %s", hostname)
        }
        return nil
    }
}
```

---

## SSH error classification — inspect error strings

```go
func classifySSHError(err error) *TransportError {
    errStr := err.Error()

    // Host key mismatch
    if strings.Contains(errStr, "host key mismatch") ||
       strings.Contains(errStr, "knownhosts") {
        return &TransportError{Reason: "HostKeyMismatch", Err: err}
    }

    // Authentication failure
    if strings.Contains(errStr, "unable to authenticate") ||
       strings.Contains(errStr, "no supported methods remain") {
        return &TransportError{Reason: "AuthFailed", Err: err}
    }

    // Network errors
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        return &TransportError{Reason: "NetworkUnreachable", Err: err}
    }
    if strings.Contains(errStr, "connection refused") ||
       strings.Contains(errStr, "no route to host") ||
       strings.Contains(errStr, "network is unreachable") {
        return &TransportError{Reason: "NetworkUnreachable", Err: err}
    }

    return &TransportError{Reason: "Unknown", Err: err}
}
```

---

## Private key parsing

```go
signer, err := ssh.ParsePrivateKey(pemBytes)
// pemBytes is []byte containing the PEM-encoded private key
// Returns ssh.Signer on success
// Returns error if key is malformed, password-protected, or unknown format
```

---

## SSH ClientConfig — building the connection config

```go
config := &ssh.ClientConfig{
    User: sshUser,
    Auth: []ssh.AuthMethod{
        ssh.PublicKeys(signer),
    },
    HostKeyCallback: strictHostKeyCallback(knownHostsLine),
    Timeout:         30 * time.Second,
}
```

---

## Dialing an SSH connection

```go
addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
sshClient, err := ssh.Dial("tcp", addr, config)
```

---

## Opening a tunnel through an SSH connection

```go
// Forward a Unix socket through the SSH connection:
conn, err := sshClient.Dial("unix", "/run/libvirt/libvirt-sock")
// conn implements net.Conn; pass to libvirt.New(conn)
```

---

## Import path

```go
import "golang.org/x/crypto/ssh"
```

Module: `golang.org/x/crypto`
Minimum version for this project: v0.17.0
