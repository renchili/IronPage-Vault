# RBAC Documentation

IronPage Vault uses three discrete roles only:

- Admin
- Editor
- Reviewer

The roles are intentionally separate. Admin is not treated as an automatic super-editor.

## Role Purpose

| Role | Primary purpose |
|---|---|
| Admin | System administration and configuration |
| Editor | Legal document manipulation and finalization |
| Reviewer | Review, annotation, and review workflow participation |

## Capability Matrix

| Capability | Admin | Editor | Reviewer |
|---|---:|---:|---:|
| Login | Yes | Yes | Yes |
| View own principal | Yes | Yes | Yes |
| User creation | Yes | No | No |
| User listing | Yes | No | No |
| Configuration listing | Yes | No | No |
| Configuration update | Yes | No | No |
| Workflow status definition listing | Yes | No | No |
| Notification template listing | Yes | No | No |
| Backup job metadata creation | Yes | No | No |
| Backup job listing | Yes | No | No |
| Document listing | Yes | Yes | Yes |
| Document metadata retrieval | Yes | Yes | Yes |
| PDF file retrieval | Yes | Yes | Yes |
| PDF upload | No | Yes | No |
| Batch PDF upload | No | Yes | No |
| Version listing | Yes | Yes | Yes |
| Version rollback | No | Yes | No |
| Redaction proposal | No | Yes | No |
| Redaction confirmation | No | Yes | No |
| Bates numbering | No | Yes | No |
| Annotation creation | No | No | Yes |
| Annotation disposition update | No | No | Yes |
| Workflow transition | No | Yes | Yes |
| Finalization | No | Yes | No |
| Audit log query | Yes | Yes | Yes |
| Notification query | Own records | Own records | Own records |
| Notification read acknowledgment | Own records | Own records | Own records |

## Why Admin is not Editor

The prompt defines Admin as a system-management role. Admins manage users, dictionaries, workflow status definitions, and notification templates. Editors manipulate legal documents.

Keeping these roles separate prevents accidental document tampering by system administrators and gives acceptance tests clear denial cases.

## Enforcement Layers

RBAC is enforced in two places:

1. Route-level middleware for coarse API boundary enforcement.
2. Handler/service-level validation for sensitive business rules such as Finalized immutability.

Route-level checks alone are not enough because future routes could accidentally call shared logic. Business operations must protect themselves.

## Required Denial Cases

Acceptance must verify at least:

- Reviewer cannot upload a PDF.
- Reviewer cannot apply Bates numbering.
- Reviewer cannot confirm redaction.
- Editor cannot create users.
- Editor cannot update system configuration.
- Admin cannot upload PDF unless the design is explicitly changed.
- Any role is denied mutation after Finalized.

## Required Approval Cases

Acceptance must verify at least:

- Admin can list users.
- Admin can create a user.
- Admin can view configuration.
- Editor can upload a PDF.
- Editor can stage and confirm redaction.
- Editor can create Bates job metadata.
- Reviewer can create annotation.
- Reviewer can update annotation disposition.

## Object-level Authorization

The current prototype validates document existence and immutable state. Future production hardening should add matter/team ownership rules if the legal organization requires document-level segregation between teams.

Object-level authorization should never be replaced by frontend filtering.
