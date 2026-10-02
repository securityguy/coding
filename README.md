# coding

Coding standards and templates for software projects, written to be followed
by developers and AI coding agents alike.

- `standards/` — standing instructions for all development projects. Follow
  them wherever they apply. See `standards/index.md` for what each directory
  covers.
- `templates/` — files to copy into projects as needed. See
  `templates/index.md`.

## Using it

1. Clone the repository to a stable path, for example `~/.coding`.
2. Point your agent at `standards/`. For Claude Code, link the directory to
   `~/.claude/standards` and add a line to `~/.claude/CLAUDE.md` such as
   "Follow `~/.claude/standards`, starting with
   `~/.claude/standards/index.md`." Other agents take the path in their own
   instructions file.
3. Copy files from `templates/` into projects as needed.

## Licence

Everything here is released under CC0 1.0 (see `LICENSE`). Take it, change it,
use it; no attribution is required.

PRs are welcome.
