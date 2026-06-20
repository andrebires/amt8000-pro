# Online Control Safety Checklist

Mutating Online commands are blocked until this checklist is completed for each
command under test.

## Applies To

- arm
- disarm
- PGM activation or deactivation
- temporary zone bypass or un-bypass
- panel date/time sync
- clear fired-zone or alarm memory

## Required Before Implementation

- Confirm the exact AMT 8000 Pro command, response, and payload format from a
  packet capture or other reproducible protocol evidence.
- Choose a safe test window when arming, siren activity, monitoring dispatch,
  and automation side effects are acceptable.
- Confirm every person with access to the protected area knows the test is
  happening.
- Confirm the test account, partition, zone, or PGM target is safe to operate.
- Confirm there is a manual recovery path using the keypad, app, or official
  Intelbras tooling.

## Required During Test

- Record the command, expected result, observed result, and timestamp under
  `docs/test-runs/`.
- Do not record remote passwords, cookies, packet captures with secrets, or
  installation-sensitive raw data.
- For any state-changing command, read status after the command and verify the
  panel reached the expected state.

## Required Before Marking Done

- Add unit tests or sanitized fixtures for the command parser/encoder.
- Add UI confirmation for the command.
- Add failure output that is visible to the operator without leaking secrets.
- Update `docs/protocol/` with the proven command and response details.
