# Decisions

Decisions taken on cozy-stack, oldest first. Append only: a decision that changes gets a new entry that replaces the old one and cites it. Entry format: the `twake-decisions` skill of [linagora/twake-guidelines](https://github.com/linagora/twake-guidelines/blob/main/skills/twake-decisions/SKILL.md).

Categories: UX, UI, Product, Architecture, Infra, Tech.

## 2026-11-24, [Tech] Shared roots are found through the referenced-by view

- Decision: the access resolver finds the shared roots above a document by reading the `referenced-by` view for `io.cozy.sharings` and keeping the directories at the ancestor paths, not with a Mango `$in` on `dir-by-path`.
- Decided by: agent ([#5382](https://github.com/linagora/cozy-stack/issues/5382)), to be confirmed by the cozy-stack maintainer who reviews the PR.
- Why: CouchDB cannot use `$in` as an index range, so the Mango query read every directory on each move and access check ([#5380](https://github.com/linagora/cozy-stack/issues/5380)).
- Scope: `model/sharing` (`AccessResolver`), used by file moves, shared drives, opening files, drive activity and recents.

## 2026-11-25, [Tech] Search keys join the domain and the doc id with a colon

- Decision: `Key` is `<domain>:<doc id>`, as in ADR 008, although a domain with a port and a doc id with a colon could give the same key.
- Decided by: Anton Shepilov (author), [review comment](https://github.com/linagora/cozy-stack/pull/5259#discussion_r4682719679)
- Why: a domain only has a colon with a port, which happens in dev only; a collision needs two instances of one stack that differ only by a port, plus a doc id with a colon chosen by a sharing peer.
- Scope: `model/search`, the `files_v1` and `share_scopes_v1` indices.

## 2026-11-25, [Architecture] Search scopes list the member addresses until #5395

- Decision: `active_member_principals` holds the lowercased hosts of the members who accepted a drive; the search uses them as is until it also checks that the caller holds the sharing (#5395).
- Decided by: Anton Shepilov (author), [review comment](https://github.com/linagora/cozy-stack/pull/5260#discussion_r4681654061)
- Why: the check reads the caller's sharings, so it belongs in the search query, which is a later change of the stack.
- Scope: `model/search`, the `share_scopes_v1` index.

## 2026-11-25, [Product] Search leaves out the trash

- Decision: the search does not return trashed files and folders, as ADR 008 says.
- Decided by: Anton Shepilov (author), [review comment](https://github.com/linagora/cozy-stack/pull/5261#discussion_r4695299650)
- Why: Drive shows search results without a trashed state, and the trash has its own view; showing trashed hits needs that display first.
- Scope: `model/search`, `POST /files/search`.

## 2026-11-25, [Product] Search drops the hits of an owner missing from the stack cluster

- Decision: `VisiblePaths` drops a drive hit whose owner is not in the global CouchDB, instead of returning it with an empty path as a drive owned outside of the stack cluster.
- Decided by: Anton Shepilov (author), [review comment](https://github.com/linagora/cozy-stack/pull/5266#discussion_r4631076331)
- Why: the indices only hold docs written by the instances of this stack cluster, so a missing owner is a deleted instance whose purge failed; its former members must not keep finding its file names.
- Scope: `model/search`, `POST /files/search`; drives of other stack clusters wait for the federated search.

## 2026-11-25, [Product] Search does not fold names stored in NFD yet

- Decision: the search index folds accents only in names stored composed (NFC); a name stored decomposed (NFD) keeps its marks apart and a search inside its words can miss, until #5394 is done.
- Decided by: Anton Shepilov (author), [review comment](https://github.com/linagora/cozy-stack/pull/5273#discussion_r4681187611)
- Why: the gap is older than the multilingual folding and also hits Latin names, so it gets its own change.
- Scope: `model/search`, the `cozy_files_v1` index.

## 2026-11-25, [Product] Search folds only the common two-letter letters before the n-grams

- Decision: `filename_fold` expands `ß æ œ þ` before the names are cut in n-grams; rarer letters that `asciifolding` expands to two letters (`ﬁ`, `ĳ`, `ǆ`…) are folded after, so they match a whole word or its start but not a search starting inside the word.
- Decided by: Anton Shepilov (author), [review comment](https://github.com/linagora/cozy-stack/pull/5273#discussion_r4681187921)
- Why: these letters are rare in file names and still match whole words through `name.prefix`.
- Scope: `model/search`, the `cozy_files_v1` index.

## 2026-11-25, [Product] Recipients find the file of a single-file drive by search

- Decision: the scope doc of a single-file drive lists its file (`root_file_id`); the search matches it on `resource_key` and shows it without a path. The other files of its folder stay hidden.
- Decided by: Anton Shepilov (author), [review comment](https://github.com/linagora/cozy-stack/pull/5260#discussion_r4681694076)
- Why: a recipient must find the files shared with them, and the parent folder of that file is not shared, so its folder key cannot be used.
- Scope: `model/search`, the `share_scopes_v1` and `files_v1` indices.

## 2026-11-26, [Product] Search indexes the names composed (NFC)

- Decision: the stack composes (NFC) the names before indexing them and the queries before sending them, and keeps a name stored decomposed (NFD) in `display_name`, which the hits return. Replaces "Search does not fold names stored in NFD yet".
- Decided by: Anton Shepilov (author), #5394
- Why: NFC composes every script, Hangul and kana included, where dropping the combining marks in the char filter only covers some; the ICU plugin would be a new requirement on the OpenSearch cluster; and the hits keep the name as stored.
- Scope: `model/search`, the `files_v1` index.

## 2026-11-26, [Architecture] Search uses a drive's scope only when the caller holds the drive

- Decision: the search reads the scope docs of the active drives that the caller holds as a recipient (one CouchDB query, one `_mget`), and keeps those whose owner still lists the caller. Replaces "Search scopes list the member addresses until #5395".
- Decided by: Anton Shepilov (author), [review comment](https://github.com/linagora/cozy-stack/pull/5260#discussion_r4681654061), #5395
- Why: a member can change its address without proof, so the owner's list alone does not show who may search a drive; ADR 008 describes these two steps. Measured locally, it adds about 1 ms to a search with 10 drives and 20 ms with 200: the `active` index does not cover `drive`, so CouchDB reads each sharing doc.
- Scope: `model/search`, `POST /files/search`.
