# Copyright header template

Every source file begins with this comment block. It is written for languages
with C-style block comments (Go, C, C++, Java, JavaScript, Rust and so on);
adapt the comment markers for anything else and keep the text.

```
/******************************************************************************
 * Copyright (c) [YEAR] [COMPANY]                                             *
 * Please see LICENSE file for details.                                       *
 ******************************************************************************/
```

## Use

1. Replace `[COMPANY]` with the copyright holder as it appears in the
   project's `LICENSE` file.
2. Replace `[YEAR]` with the year the file was created. Once a file spans more
   than one year, use a range such as `2025-2026`.
3. Re-pad the line with spaces so the closing `*` stays aligned with the line
   below it.
4. Put the block at the very top of the file, before the package or module
   declaration.

The second line always says to see the `LICENSE` file; do not put licence text
in the header itself.
