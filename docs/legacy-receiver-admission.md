# Single admission for an existing receiver

`atlas admit` lets an existing locally authorized receiver use Desktop's busy-turn
admission without installing another receiver or importing an identity:

```sh
atlas admit --thread EXISTING_CHAT_UUID --request-id PERSISTED_WAKE_UUID \
  < PRIVATE_WAKE_TEXT
```

The caller must save the request UUID and pending message IDs in its existing
private wake journal **before** running the command. Input is read from stdin,
with a 64 KiB limit. `--source-thread ORIGINAL_CHAT_UUID` is optional; use it when
there is one actual source chat. A reminder combining several senders must not
invent a single source identity.

The command performs one local Desktop operation with a 20-second deadline. It
does not read Atlas credentials, poll messages, change a binding, write a cursor,
create an idle queue entry, or start a daemon. Existing cursor, source-envelope
storage, authorization, queue coalescing and read-receipt ownership remain with
the caller. Server permissions do not expand local human authorization.

| JSON status | Meaning | Original receiver action |
|---|---|---|
| `admitted` | Desktop confirmed an existing turn accepted the input. | Save this outcome under the original request ID; do not submit the same wake again. |
| `idle` | Desktop explicitly rejected admission because the chat is idle. | Continue the original idle queue/coalescing path. |
| `unavailable` | Desktop discovery failed before an input admission request was sent. | Continue the original queue path if it is available. |
| `uncertain` | An admission attempt did not return a confirmed outcome. | Preserve the pending journal and reconcile; do not blindly retry or fall back to queue. |

`admitted` and `idle` exit successfully. `unavailable` and `uncertain` return a
nonzero exit code after writing the structured result. Invalid arguments fail
before any Desktop admission. If there is no valid structured result, treat the
outcome as uncertain. `safe_for_idle_queue` is true only for the two known safe
fallback statuses. `admitted` is null for an uncertain outcome, rather than
falsely claiming the input was rejected. `agent_read` remains false in every
result. Actual agent reading requires the existing receiver's receipt process.

Do not call the adapter again on every poll while the same saved wake remains
unread. Preserve one wake identity per pending set and the already confirmed
outcome. If the pending set changes, the original receiver decides whether to
update its existing queue item or save a new wake identity. Do not remove an old
owned queue item until its replacement has been confirmed, and never delete
unrelated queue items.

Keep the saved text associated with its request ID. Do not reuse an ID for a
different reminder. If a wake was already placed in the idle queue, save a
separate busy-attempt UUID in the same original wake journal before calling this
adapter. Reusing an existing queue-creation ID for a second admission operation
requires runtime reconciliation evidence that this project has not established.

For the current TapTap receiver, the candidate integration point is in
`coalesce()`, after it saves `wake.client_user_message_id` and before its existing
`thread/queue/add`. Its original owner must integrate the outcome into that
journal and its queue ownership rules. Shipping this adapter does not prove the
old running receiver has adopted it. No new `atlas bind` or `atlas receive`
process should be used to take over the old chat's backlog.
