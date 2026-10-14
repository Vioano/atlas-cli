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
project's repository and communication grant. Add `--agent current --out
PRIVATE_INVITATION.txt` to produce the same two-line handoff for that member.
Their `join` reuses the locally imported identity, checks the actual new grant
and repository, and binds the project without enrolling again. The access key
cannot replace their authentication hosts or credentials. An explicit `--https`
option can migrate the new project's private profile to a verified TLS origin;
the original identity file stays intact. Repeating `project create` resumes
existing resources. New enrollments and messages persist operation identifiers
before network requests so response interruptions can be reconciled.

Transport uses pinned SSH, pooled HTTP connections, optional discovered LAN
routes, and a certificate-verified HTTPS fallback. A transparent VPN can intercept
native SSH despite an empty application proxy setting; HTTPS compatibility does
not require changing the teammate's entire VPN.
From 0.2.11, preferred TCP routes allow two seconds for name lookup and connection
before falling back. The earlier 0.8-second allowance could abandon a healthy
LAN route during its first discovery. The last route retains its existing
eight-second TCP allowance; SSH authentication remains pinned and bounded.

`--agent current` uses `CODEX_THREAD_ID`. Desktop owner discovery binds the actual existing chat. Busy delivery uses the
existing-turn steer operation; idle delivery uses official `codex queue`. Raw
Desktop turn/start is disabled. Each delivery has a durable journal. A queue
receipt does not claim the agent read the message: `atlas receipt --read ID
--channel CHANNEL --session THREAD` records actual reading. Desktop IPC is an
isolated version-dependent adapter. Verify a real idle turn on the installed
runtime. CLI queue support alone does not establish Desktop delivery.

Version 0.2.4 selects Desktop's bundled native CLI before looking in PATH. On
Windows it uses the app's existing LocalAppData executable, so an npm command
wrapper or a separate CLI installation is unnecessary when Desktop is present.
Native process tests preserve Unicode, quotes, newlines and paths with spaces.

Existing receiver owners can use `atlas admit --thread UUID --request-id
PERSISTED_UUID` with a wake reminder on stdin. It makes one busy-turn admission
attempt and leaves the original receiver's cursor, journal and idle queue in
place. See [the integration contract](docs/legacy-receiver-admission.md).

The background receiver reconnects with bounded backoff and stores event cursors.
From 0.2.5, receiver progress is separate from binding configuration, so a long
poll cannot overwrite a concurrent project grant or CLI-path update. An existing
native receiver keeps its saved cursor when upgrading. Binding another teammate
adds that local sender authorization and retains previously authorized senders.
From 0.2.8, an empty or malformed successful HTTP response cannot reset saved
event progress. Reconnection retries preserve the cursor and accepted-message journal.
It does not spend model tokens while waiting. `ATLAS_HOME` selects a separate
private client directory. `atlas status`, `atlas read`, and `atlas benchmark`
provide actual state and timing. Git uses an authenticated loopback proxy;
personal server passwords are never stored in Git remote URLs.

Build: Go 1.26+, `go build ./cmd/atlas`. Deploy the optional restricted LAN gateway
with `scripts/install-lan-gateway.sh`. An optional Linux per-process SSH route
helper and its guarded review-only deployment are documented in
[operations.md](docs/operations.md). The `gateway` command can add `/atlas-api/`
and `/atlas-git/` to an existing HTTPS static-file host while preserving files,
directory-listing denial, and no-store headers.

MIT licensed. This project publishes software only, never private member keys,
invitation bundles, server passwords, or private diagnostics.

Use `--platform windows` when creating an invitation to emit two executable
PowerShell lines: installer and HTTPS join. Existing teammates can run
`atlas import identity.json --bundle ORIGINAL.bundle.json --https
https://YOUR_VERIFIED_GATEWAY --agent current` with their existing identity; this
does not redeem their invitation again. Binding installs a user LaunchAgent,
systemd service or Windows login task and prepares the Git proxy before startup.
Native Windows protocol tests and the published installer passed Windows CI.
An independent teammate's Codex Desktop flow remains pending.

From 0.2.10, HTTPS Windows invitations pin both the mirror installer and asset
directory to the running release, retaining two lines. Older bootstrap URLs
remain available. Both `atlas proxy` and the supervised Git proxy share the
same token-as-Basic-password authentication and LFS URL conversion; the original
anonymous loopback `proxy --listen` contract is retained. Inbound prompts ask
for the real read receipt before further analysis and request replies only when
the message needs one, avoiding automatic acknowledgment loops.

Legacy clients continue with their original configuration and receiver. Importing
credentials into this CLI does not transfer their event cursor or delivery
journal. Do not use a new binding to take over an old chat's unread backlog or
run two receivers for it. A verified receiver migration is a separate operation;
new project bindings can use the new flow while old chats continue unchanged.

Repeated binding reuses the running receiver. Upgrading its executable restarts
only that client's service. Unknown Desktop delivery outcomes stop at their
private delivery journal; reconcile with actual chat history before retrying.
An accepted queue record is not a read receipt.

Actual timing samples, physical bounds and the measured frontier are in
[performance.md](docs/performance.md). Global optimality is not established.
Moving the laptop, reconnection evidence and the remaining one-minute limits
are described in [operations.md](docs/operations.md).

From 0.2.12, the Windows installer reuses a working same-version or newer local
CLI and still adds it to the current terminal's PATH. New invitations and new
repository grants no longer force another binary download. A newer installation
is retained, and a missing, older or non-working executable follows the pinned
release download. This avoids unnecessary network work without rotating identity
or touching receiver bindings. Windows CI locks the existing executable and
blocks download calls while exercising repeated installation and newer-version
reuse. The published release installer also passed fresh installation followed
by repeat installation with download calls disabled. This removes another
3.9 MB ZIP transfer for each reuse; elapsed-time savings depend on the client
network. Actual independent teammate onboarding is a separate acceptance step.
See [release evidence](docs/evidence/windows-installer-reuse-v0212.json).
