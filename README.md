# Atlas CLI

Create private Gitea projects, invite a teammate with two lines, and connect the
current Codex agents through Personal Atlas. Native binaries for macOS, Linux,
and Windows; no Python package, model call, or executable invitation bundle.

```sh
atlas project create my-project --invite teammate --purpose 'Build the game' --agent current
atlas join atlas1_PERSONAL_KEY --agent current
atlas clone captain/my-project
atlas send --text 'Ready for review'
```

The captain forwards two lines: an installation instruction and the personal
join command. The key contains scoped personal credentials and pinned endpoints.
Keep it private. Once enrolled, identity credentials do not expire. Revocation
remains controlled by the server.

Install a binary from [Releases](https://github.com/Vioano/atlas-cli/releases).
Windows: download `atlas-windows-amd64.exe` and run it from your agent's terminal.
macOS/Linux: place the matching binary in your PATH and make it executable.

An already issued invitation can be reused without rotating any credentials:

```sh
atlas join --bundle ORIGINAL.bundle.json --agent current --https https://YOUR_VERIFIED_GATEWAY
```

`atlas pack ORIGINAL.bundle.json` produces the compact invitation. Bundled code
is discarded. Existing enrolled members may use `atlas import identity.json`;
their old clients, credentials, grants, and SSH paths stay valid.

Captain setup references existing private profiles rather than copying secrets
into this repository:

```sh
atlas setup --atlas-profile owner.json --gitea-profile gitea-owner.json \
  --gateway-profile captain-routes.json --gateway-admin trusted-relay-alias \
  --server-admin trusted-server-alias --member-gateway-profile member-routes.json \
  --https https://YOUR_VERIFIED_GATEWAY
```

`project grant NAME --member EXISTING_ACTOR --role member` changes only that
project's repository and communication grant. Repeating `project create` resumes
existing resources. New enrollments and messages persist operation identifiers
before network requests so response interruptions can be reconciled.

Transport uses pinned SSH, pooled HTTP connections, optional discovered LAN
routes, and a certificate-verified HTTPS fallback. A transparent VPN can intercept
native SSH despite an empty application proxy setting; HTTPS compatibility does
not require changing the teammate's entire VPN.

`--agent current` uses `CODEX_THREAD_ID`. Desktop owner discovery binds the actual
existing chat; CLI uses `codex queue`. Each delivery has a durable journal. A queue
receipt does not claim the agent read the message: `atlas receipt --read ID
--channel CHANNEL --session THREAD` records actual reading. Desktop IPC is an
isolated version-dependent adapter. Verify a real idle turn on the installed
runtime. CLI queue support alone does not establish Desktop delivery.

The background receiver reconnects with bounded backoff and stores event cursors.
It does not spend model tokens while waiting. `ATLAS_HOME` selects a separate
private client directory. `atlas status`, `atlas read`, and `atlas benchmark`
provide actual state and timing. Git uses an authenticated loopback proxy;
personal server passwords are never stored in Git remote URLs.

Build: Go 1.26+, `go build ./cmd/atlas`. Deploy the optional restricted LAN gateway
with `scripts/install-lan-gateway.sh`. The `gateway` command can add `/atlas-api/`
and `/atlas-git/` to an existing HTTPS static-file host while preserving files,
directory-listing denial, and no-store headers.

MIT licensed. This project publishes software only, never private member keys,
invitation bundles, server passwords, or private diagnostics.
