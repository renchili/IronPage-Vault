# Metadata Security Matrix

This matrix records which stored metadata is protected by encrypted storage and which fields remain plain operational metadata.

## Encrypted storage

| Area | Field | Storage behavior | API behavior |
|---|---|---|---|
| Redaction geometry | `x`, `y`, `width`, `height` | numeric columns are compatibility placeholders; ciphertext columns hold the source-of-truth values | list responses omit geometry |
| Redaction reason | `reason` | encrypted before insert | list responses omit reason |
| Annotation comment | `comment` | encrypted before insert | stored value is never request plaintext |

## Plain operational metadata

The following remain plain because they are identifiers, routing fields, workflow state, or audit control data: object IDs, document IDs, page number, workflow status, annotation type, annotation disposition, timestamps, actor IDs, backup job status, and audit action metadata.

## Runtime rule

Plain request text may exist only in local request memory for validation, mention extraction, PDF processing, or encryption. It must not be inserted as the source-of-truth value for protected metadata.

## Contract

`ci/metadata_storage_check.sh` validates the storage and API exposure rules above.
