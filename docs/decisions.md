# Decisions

Decisions taken on cozy-stack, oldest first. Append only: a decision that changes gets a new entry that replaces the old one and cites it. Entry format: the `twake-decisions` skill of [linagora/twake-guidelines](https://github.com/linagora/twake-guidelines/blob/main/skills/twake-decisions/SKILL.md).

Categories: UX, UI, Product, Architecture, Infra, Tech.

## 2026-11-23, [Architecture] The Stack keeps the SSO uuid of the owner

- Decision: at each OIDC login and app token exchange that the SSO verified for the owner, the Stack keeps the `uuid` claim (the LDAP entryUUID) on the instance as `oidc_uuid`, and names the owner by it in the events it publishes on `activity`.
- Decided by: agent (session of Théo Poizat, 2026-11-23), to be confirmed by a cozy-stack maintainer.
- Why: Twake apps name people by uuid in their events (ADR 067 of linagora/twake-adr), and the webhook gateway only delivers by uuid (ADR 075); `oidc_id` holds the `sub`, which differs.
- Scope: OIDC login, `/auth/token_exchange` with `exchange_type` `app`; admin exchanges are left out, as their token may belong to another member.
