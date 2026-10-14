# Moving the private server and recovering connections

Public access uses a verified HTTPS hostname and outbound restricted SSH tunnels.
The laptop's LAN address is not a public endpoint. A new WiFi lease therefore
does not require replacing invitations, teammate identities, repository grants
or public URLs. The optional LAN gateway uses discovered hostnames and listens
on all IPv4 interfaces. Public HTTPS remains the fallback when LAN discovery is
unavailable. Existing SSH clients retain their original paths and permissions.

The laptop still needs working Internet access. Join the destination WiFi and
complete any captive portal first. An unknown WiFi password, a suspended laptop,
blocked outbound access or a disconnected network cannot be repaired by a
teammate's invitation key. Keep credentials and WiFi configuration private.

## Actual deployment checks

The deployed laptop has NetworkManager and mDNS active. The current WiFi profile
has autoconnect enabled. The user has lingering enabled, and the three existing
tunnel services are enabled at boot. These settings support recovery without
keeping a model running. They do not constitute a reboot or whole-WiFi-move test.

The new web and review tunnels use `ServerAliveInterval=5`,
`ServerAliveCountMax=2`, `ConnectTimeout=8` and `RestartSec=2`. They retain normal
host-key verification and restricted forwarding accounts. The old tunnel still
uses 20 seconds, three unanswered checks and a 15-second restart delay. Its
first recovery after a silent outage can take longer than one minute. The old
service was not changed or restarted by the isolated test.

Do not interpret the keepalive settings as a guaranteed ten-second detection
deadline. An isolated connection authenticated with the existing pinned host
key and restricted key was deliberately blackholed, without a TCP FIN or RST.
The first timeout was 14.455 seconds after initial authentication. Traffic was
dropped for 30.035 seconds. After restoration, systemd restarted the connection
and authentication was observed within **9.070 seconds**. It made two restarts
before restoration; the default start limiter did not stop this case. No remote
forwarding ports were opened. All three production tunnel PIDs and active states
remained unchanged, and original Atlas/Gitea health and authenticated public
HTTPS were read back successfully. The temporary test unit was removed.
See [sanitized evidence](evidence/ssh-blackhole-recovery.json).

The separate live test that killed only the new web tunnel recovered an
authenticated public manifest in **6.821 seconds**. That test covers forwarding
readiness after a process failure; the blackhole test covers a silent transport
failure. Neither proves a whole-laptop WiFi move, DHCP/DNS failure, captive
portal, suspend or reboot recovery time.

## Compatibility during recovery

Keep the server's existing credentials, host-key pins and local service ports.
Do not redeem an existing teammate's invitation on another machine. Do not run
two receivers for the same existing chat. Each original receiver retains its
cursor, delivery journal, source-envelope storage and queue ownership.

The CLI saves operation IDs before sending messages and enrollment requests.
Reconnect retries preserve those identities and saved progress. A confirmed
Desktop admission is not a read receipt. An uncertain outcome requires
reconciliation with actual chat history before another submission. Credential
revocation remains explicit; a changed WiFi address does not expire a teammate.

For an already enrolled teammate receiving another project grant, use the
existing-member access key. Their existing identity is reused; no fresh
enrollment or replacement of old project permissions is necessary.

## What the one-minute target currently establishes

One complete fresh Mac project/join/private-clone/duplex-read run was observed
within 59.686 seconds. Independent Windows Desktop acceptance is still pending.
Native Windows installer and protocol CI do not replace that acceptance.

The original TapTap receiver actually adopted single busy admission. Three
subsequent natural business messages had admission times of 53.771, 57.816 and
54.750 ms, while their server send-to-first-read-receipt times were 65.155,
61.114 and 234.908 seconds. All were actually read. The last recipient turn
contained context compaction before reading the message. The trace does not
quantify each contribution to that delay. These observations fail a strict
60-second read-receipt target even after a receipt-first policy.

Receipt-first means an agent records its real read receipt before further
analysis or coordination. It cannot acknowledge a merely queued or admitted
message. Inline full source envelopes can remove another file-read/model
round trip when they fit within the admission input limit; old journal and
coalescing rules still apply. The original owner has now deployed this further
change in the same singleton service. Seven isolated checks passed; eight old
request IDs, texts and outcomes stayed intact. Whole batches exceeding 64 KiB
or failing source validation retain file-based reading, without truncation.
The service PID changed during its managed restart. The three timing cases above
precede that upgrade; subsequent observations follow below. Deployment and the
isolated checks alone did not establish a live improvement or deadline guarantee. See
[natural message evidence](evidence/natural-legacy-receipts-v029.json).

Three later natural business messages using the new inline-original path were
actually read and acknowledged within **31.582, 27.932 and 24.156 seconds**.
Their complete originals remained in the original owner's journal. These are
observed successful cases, not a randomized causal comparison or a whole-team
percentile guarantee. No new test marker or reply loop was added. See
[the subsequent receipts](evidence/inline-natural-receipts-v029.json).

Full onboarding, transport admission, actual agent reading and performing the
requested task remain distinct measurements. A CLI cannot guarantee the model
runtime's scheduling, context-compaction or generation latency.

## Optional per-process physical egress

The protected review tunnel now uses [ssh-route-socket.py](../scripts/ssh-route-socket.py)
as its Linux SSH ProxyCommand. It discovers the current main-table default
interface on each connection, tries a socket bound to that interface, and falls
back to the original default route if connection or SSH banner probing fails.
It writes no global route, DNS or firewall state. OpenSSH still performs the
original host-key verification, identity authentication and restricted forwarding.
This is an optional server-side transport helper, not a teammate identity change.
The deployed ordinary service user could bind its socket without a new capability.
Other hosts may refuse this operation; the unchanged default route is retained.

An actual connection authenticated with the original review identity through the
helper before cutover, with no forwarding ports requested. The live review-only
cutover used the existing publisher's runtime lock, active-release transition
lock, nonce gate and fresh drain acknowledgment. All forwarded POSTs drained;
no world existed or was reset. Only this tunnel restarted, retaining its three
original reverse forwards and base unit bytes. Atlas, Gitea, other tunnels and
review backend process IDs stayed unchanged. All three relay listeners were
read back after cutover. The owner removed only its own gate and lock.
An isolated real TCP test also verified fallback when physical binding fails.
See [sanitized deployment evidence](evidence/review-physical-route.json).

The helper chooses the available interface at reconnect rather than storing a
WiFi device name. This supports another WiFi lease or an Ethernet default route;
a whole-laptop move has not been tested. The established tunnel still depends
on the existing keepalive and systemd restart behavior. Expected speedup is
limited by the earlier paired route measurement, not inferred from deployment.
Rollback removes only this task's route drop-in and reloads/restarts the same
review tunnel under the matching maintenance contract; it leaves the original
unit, forwarding account, identity and host-key pins intact.

The original browser also completed a conditional 304 and, 6.003 seconds after
its response, a 200 GET on the same HTTP/3 connection. Both reused the upstream
connection. Response headers took 629.45 and 677.115 ms; upstream header times
were 299.963 and 357.741 ms. This warm observation did not improve on the earlier
569.5 ms browser case. Separate times and transport conditions do not establish
causal regression either. The review-only change is retained for independent
server egress with the original route as fallback, without a performance claim.
It does not resolve the browser's separate long pre-send wait.
