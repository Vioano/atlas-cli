# Performance and limits

The measured development target is creation, membership, repository access and agent communication within 60 seconds on a reachable network. A disconnected network cannot meet that deadline. Downloading all assets is constrained by their byte size and the actual path throughput.

Initial real private project creation plus invitation: 4.581 seconds. First real QA enrollment and current Desktop binding: 11.849 seconds. A later fresh 0.2.0 project plus invitation and binding took 4.863 seconds; a fresh QA identity enrollment and binding took 11.559 seconds. These are different identities on the captain Mac, not a Windows acceptance result. Public Windows ZIP download: 3,891,061 bytes in 1.719 seconds with HTTP 200 and default TLS verification. Do not add these observations and call the sum a Windows end-to-end test.

## Actual bottlenecks removed

- Initial event synchronization inherited a 45-second long-poll default. It now explicitly uses `wait=0`. Background delivery waits 10 seconds, with a 15-second request deadline.
- The actual Linux HTTP handler showed about 41 ms per persistent request with TCP_NODELAY off and 0.16–0.28 ms with it on in an isolated A/B test. Production persistent health now measures 0.256–0.351 ms after its first request.
- Pooled pinned SSH and HTTP avoid repeated handshakes. A restricted LAN route uses hostname discovery instead of a stale WiFi address.
- Certificate-verified HTTPS provides compatibility when a transparent VPN intercepts SSH. Windows invitation commands select HTTPS. Existing identities and SSH entrances are preserved.
- Git LFS batch action URLs are translated by an authenticated loopback proxy. Private server credentials never appear in Git remote URLs. A 1 MiB payload was uploaded and independently downloaded with exact byte comparison.
- Busy Desktop delivery uses its existing-turn steer operation. Idle delivery uses official Codex queue. Raw turn/start is disabled. Actual busy bidirectional messages 689/691 and read receipts were observed in the same existing Mac turn. Version 0.2.2 also passed delivery between two distinct existing Mac chats: message 729 started the idle recipient through Codex queue, and reply 739 entered the busy sender through Desktop steer. Both were actually read and acknowledged. Windows busy/idle behavior still requires live verification.

## Latency samples

Twenty sequential health requests per scenario, including the first cold request; no failures. Measurements were not simultaneous. They describe this network and time, not all teammates or a universal SLA.

| Scenario | Median ms | Min ms | Max ms | n |
|---|---:|---:|---:|---:|
| network-lan.json | 51.900 | 9.917 | 218.840 | 20 |
| network-lan-after-nodelay.json | 8.637 | 5.877 | 214.162 | 20 |
| network-https.json | 658.730 | 585.827 | 1291.608 | 20 |
| network-https-after-nodelay.json | 601.428 | 565.597 | 1476.227 | 20 |

Raw samples are in `evidence/`. Authenticated backend manifest measurements after warmup were 0.299–0.387 ms. An events-page test measured 5.07–5.39 ms while returning historical pages; it is not an empty-event or CPU-cycle measurement.

An actual private LFS 1 MiB download measured 1.91, 2.41 and 3.20 Mbit/s across three requests including request overhead. Those observations are achieved throughput, not a measured line-capacity ceiling. The public ZIP download achieved 18.10 Mbit/s. The slower private transfer path remains a material optimization opportunity.

## Physical and hardware bounds

[The BIPM defines vacuum light speed](https://www.bipm.org/en/measurement-units/si-defining-constants) as exactly 299,792,458 m/s. For required propagation distance d, RTT is at least 2d/c. For an optical path of actual length L and effective group index n, its propagation contribution is 2nL/c. [Corning's SMF-28 Ultra specification](https://www.corning.com/content/dam/corning/media/worldwide/coc/documents/Fiber/product-information-sheets/PI-1424-AEN.pdf) gives n=1.4682 at 1550 nm. This is a fiber example, not proof of which fiber an ISP uses. At 10,000 km, vacuum RTT is at least 66.713 ms; that fiber model yields 97.948 ms before switching and queuing. This distance is illustrative: actual device cities and cable route lengths have not been established. Proxy-node detours must be included, not replaced by the endpoints' geographic distance.

For B bytes on a path with usable bottleneck C bit/s, serialization time is at least 8B/C. End-to-end throughput cannot exceed the slowest access link, relay, encryption engine, storage or CPU service stage. Observed throughput does not identify every link's capacity.

The server is a Ryzen 7 5800H with 16 logical CPUs. Reported maximum clock is 4.465261 GHz, corresponding to a 0.224 ns clock period. Its observed clock varies with power management. A critical chain of q CPU cycles needs at least q/f_max; q was not measured, and instructions are not equivalent to clock cycles. Eliminating a measured 0.3 ms backend stage can save at most 0.3 ms for that request; it cannot eliminate a 600 ms network path. The WiFi reported 866.7 Mbit/s PHY rates over 80 MHz. That is not usable Internet capacity: airtime sharing, retransmission, protocols and upstream links reduce it.

## Pareto frontier and marginal cost

The currently measured choices trade compatibility, latency and added operation: LAN is about 8.64 ms and needs local reachability; HTTPS is about 601.43 ms and works through ordinary 443 paths using existing infrastructure. Both are retained. There is no extra VPS fee for this deployment. Native supervision, cursors and keepalives provide reconnection without model polling.

A per-socket physical-interface TCP probe reached the same relay SSH banner in 474–631 ms, compared with 736–779 ms via the existing TUN route. It was not an authenticated tunnel or throughput A/B, so no production routing change or claimed speedup follows from that probe. Changing relay location or path could improve WAN performance. Its price and operating cost have not been measured.

A rational stopping calculation is H > N × delta / 3600, where H is added engineering/operation hours, N is expected operations over the decision horizon, and delta is seconds saved per operation. For example, saving 0.3 ms over 10,000 operations saves only 3 seconds; a backend rewrite is hard to justify for that narrow gain. This is a conditional decision rule, not proof that all further optimization is uneconomic. Large asset transfer and WAN routes have a different payoff.

**No global optimality claim is established.** The physical path, every capacity, alternative relay cost, independent Windows flow and relocation recovery need stronger evidence. The tested changes remove specific waste; they do not prove the network has reached its physical limit or that every further improvement costs too much.

Version 0.2.1 prefixes newly generated operation identifiers with a valid letter. Version 0.2.0 used base64url identifiers whose first character could be rejected by the server. Existing credentials stay valid, and invalid pending enrollment identifiers are repaired without changing the pending credential.

Version 0.2.2 falls back to the idle queue when Desktop explicitly rejects steering because the active turn already ended. Unknown delivery failures remain uncertain and are not blindly retried. Sending from an agent uses its current chat identity rather than another chat's last saved binding.
