# Decisions

Decisions taken on cozy-stack, oldest first. Append only: a decision that changes gets a new entry that replaces the old one and cites it. Entry format: the `twake-decisions` skill of [linagora/twake-guidelines](https://github.com/linagora/twake-guidelines/blob/main/skills/twake-decisions/SKILL.md).

Categories: UX, UI, Product, Architecture, Infra, Tech.

## 2026-11-23, [Architecture] The Stack keeps the SSO uuid of the owner

- Decision: at each OIDC login and app token exchange that the SSO verified for the owner, the Stack keeps the `uuid` claim (the LDAP entryUUID) on the instance as `oidc_uuid`, and names the owner by it in the events it publishes on `activity`.
- Decided by: agent (session of Théo Poizat, 2026-11-23), to be confirmed by a cozy-stack maintainer.
- Why: Twake apps name people by uuid in their events (ADR 067 of linagora/twake-adr), and the webhook gateway only delivers by uuid (ADR 075); `oidc_id` holds the `sub`, which differs.
- Scope: OIDC login, `/auth/token_exchange` with `exchange_type` `app`; admin exchanges are left out, as their token may belong to another member.

## 2026-11-23, [Architecture] One event per file created in a personal drive, off by default

- Decision: with the context setting `drive_activity.file_created`, the Stack publishes `com.twake.drive.file.created.v1` for each file uploaded through `POST /files/:dir-id` in a personal drive, to its owner by uuid; shared drives keep the digest of ADR 010.
- Decided by: agent (session of Théo Poizat, 2026-11-23), to be confirmed by a cozy-stack maintainer.
- Why: the n8n trigger of Twake Drive needs an event per file, while ADR 010 keeps the volume of a sync out of the bus by default.
- Scope: the personal drive of an instance, uploads through `/files` only (not copies, notes, office files or sharing replication).
