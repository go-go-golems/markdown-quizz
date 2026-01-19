# Add REST-backed CLI commands

This is the document workspace for ticket PMQ-3.

## Structure

- **design/**: Design documents and architecture notes
- **reference/**: Reference documentation and API contracts
- **playbooks/**: Operational playbooks and procedures
- **scripts/**: Utility scripts and automation
- **sources/**: External sources and imported documents
- **various/**: Scratch or meeting notes, working notes
- **archive/**: Optional space for deprecated or reference-only artifacts

## Getting Started

Use docmgr commands to manage this workspace:

- Add documents: `docmgr doc add --ticket PMQ-3 --doc-type design-doc --title "My Design"`
- Import sources: `docmgr import file --ticket PMQ-3 --file /path/to/doc.md`
- Update metadata: `docmgr meta update --ticket PMQ-3 --field Status --value review`
