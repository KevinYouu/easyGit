# easyGit Design Philosophy

## Core Positioning

easyGit is a **Git workflow enhancement tool**, not a replacement for Git commands.

### Our Mission

**Integrate complex, hard-to-remember Git workflows to make Git usage simpler and more efficient.**

## Design Principles

### ✅ What We Should Do

#### 1. Integrate Multi-Step Workflows

Consolidate complex workflows that require multiple command combinations into a single interactive command.

**Example:**
```bash
# ❌ Traditional way - Multiple steps, error-prone
git add -A
git commit -m "feat: add feature"
git pull --rebase
git push

# ✅ easyGit way - One command, interactive guidance
easyGit push-all
# Guided: Select commit type → Input message → Auto add + commit + pull + push
```

#### 2. Simplify Hard-to-Remember Commands

Provide friendly interactive interfaces for Git commands with complex parameters that are difficult to memorize.

**Example:**
```bash
# ❌ Traditional way - Complex parameters, hard to remember
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# ✅ easyGit way - Interactive, no memorization needed
easyGit tag
# Guided: Enter version → Select prefix/suffix → Auto create and push
```

#### 3. Provide Visualization and Guidance

Offer visual feedback and operation guidance through TUI interface to lower the learning curve.

**Example:**
```bash
# ✅ Interactive rebase guidance
easyGit rebase-interactive
# Display commit list, visually select operations (pick/squash/edit)

# ✅ Conflict resolution assistant
easyGit resolve-conflicts
# Display conflicting files, guide resolution strategy selection
```

#### 4. Optimize Repetitive Work

Provide efficient tools for frequent batch operations.

**Example:**
```bash
# ✅ Batch branch cleanup
easyGit batch-cleanup
# Auto-detect merged branches, batch delete

# ✅ Batch tag management
easyGit tag-delete
# Multi-select delete, batch operations
```

### ❌ What We Should NOT Do

#### 1. Don't Duplicate Git's Simple Native Commands

If Git's native command is already simple enough, we shouldn't reimplement it.

**Counter-examples:**
```bash
# ❌ Should not implement
easyGit switch branch-name
# Reason: git switch is already simple enough

# ❌ Should not implement
easyGit stash
# Reason: git stash is already simple enough

# ❌ Should not implement
easyGit log
# Reason: git log works well already
```

#### 2. Don't Over-Simplify and Lose Flexibility

Over-simplification makes users lose understanding and control of Git.

**Counter-examples:**
```bash
# ❌ Should not implement
easyGit quick-push -m "message"
# Problem: Skips important confirmation steps, users don't know what happened
# May cause: Accidental commits, force pushes, etc.
```

#### 3. Don't Implement All Git Features

We are a supplement to Git, not a replacement. Users still need to learn Git basics.

**Correct attitude:**
- ✅ Use easyGit for complex workflows
- ✅ Use git commands for simple operations
- ✅ Combine both, complementary usage

## Feature Classification Guide

### Command Types We Should Implement

#### 🟢 Type A: Multi-Step Workflows

**Characteristics:**
- Requires 3+ command combinations
- Has fixed execution order
- Frequently used together

**Examples:**
- ✅ `push-all` - add + commit + pull + push
- ✅ `push-selected` - selective add + commit + push
- ✅ `release` - version bump + tag + push + changelog
- ✅ `tag` (with push) - tag + push tag

#### 🟡 Type B: Complex Parameter Commands

**Characteristics:**
- Many parameters, hard to remember
- Multiple usage modes
- Error-prone

**Examples:**
- ✅ `reset` - Interactive commit and mode selection
- ✅ `rebase-interactive` - Visual rebase operations
- ✅ `cherry-pick` - Visual commit selection
- ✅ `merge` - Guided branch and strategy selection

#### 🔵 Type C: Batch Operations

**Characteristics:**
- Needs repeated execution
- Has filtering criteria
- Can be optimized for batch

**Examples:**
- ✅ `tag-delete` - Batch delete tags
- ✅ `branch-delete` - Delete branch (extensible to batch)
- ✅ `batch-cleanup` - Batch cleanup merged branches

#### 🟣 Type D: Auxiliary and Enhancement

**Characteristics:**
- Provides additional value
- Improves user experience
- Doesn't change core functionality

**Examples:**
- ✅ `status` - Enhanced status with colors
- ✅ `set-language` - Language configuration
- ✅ `init` - Guided repository initialization

### Command Types We Should NOT Implement

#### ❌ Type X: Git Native Already Simple

**Characteristics:**
- Can be completed with single command
- Simple, easy-to-remember parameters
- No need for interactive guidance

**Examples:**
- ❌ `switch` → Use `git switch`
- ❌ `stash` → Use `git stash`
- ❌ `add` → Use `git add`
- ❌ `commit` → Use `git commit`
- ❌ `pull` → Use `git pull`
- ❌ `fetch` → Use `git fetch`
- ❌ `log` → Use `git log`
- ❌ `diff` → Use `git diff`

#### ❌ Type Y: Over-Simplification

**Characteristics:**
- Hides important details
- Reduces user control
- May cause problems

**Examples:**
- ❌ `quick-push` → Skips confirmation, high risk
- ❌ `auto-merge` → Auto-resolve conflicts, dangerous
- ❌ `force-push` → Too dangerous, shouldn't be simplified

#### ❌ Type Z: Edge Features

**Characteristics:**
- Extremely low usage frequency
- High maintenance cost
- Unclear benefits

**Examples:**
- ❌ `reflog` → Advanced feature, rarely used
- ❌ `filter-branch` → Complex scenarios, keep as is
- ❌ `bisect` → Debugging tool, native is sufficient

## User Experience Principles

### 1. Progressive Complexity

- Simple operations → Use Git directly
- Medium complexity → Use easyGit guidance
- High complexity → Learn Git advanced usage

### 2. Maintain Visibility

Users should clearly know what operations are being performed:
- ✅ Display each step
- ✅ Show actual Git commands being executed
- ✅ Provide confirmation mechanisms

### 3. Interruptible and Recoverable

All operations should be:
- ✅ Cancellable (Ctrl+C, q)
- ✅ Provide undo suggestions
- ✅ Don't break the working directory

### 4. Educational

Help users learn Git:
- ✅ Suggest equivalent Git commands
- ✅ Explain the purpose of each step
- ✅ Guide to official documentation

## Implementation Decision Flowchart

```
New Feature Proposal
    ↓
Is it a single Git command? ──Yes→ ❌ Don't implement, use native Git
    ↓ No
Needs 3+ steps? ──No→ Evaluate parameter complexity
    ↓ Yes           ↓
Has fixed process? ──Yes→ ✅ Implement as workflow command
    ↓ No            ↓
Hard to remember? ──No→ ❌ Don't implement
    ↓ Yes           ↓
Frequently used? ──No→ ❌ Don't implement
    ↓ Yes
✅ Implement as interactive command
```

## Command Naming Conventions

### Workflow Commands

- `push-all` - Complete push workflow
- `push-selected` - Selective push
- `rebase-interactive` - Interactive rebase
- `resolve-conflicts` - Conflict resolution

### Batch Operation Commands

- `tag-delete` - Delete tags
- `branch-delete` - Delete branches
- `batch-cleanup` - Batch cleanup

### Auxiliary Commands

- `set-language` - Set language
- `init` - Initialize repository

### Naming Principles

- ✅ Use verb-noun structure
- ✅ Descriptive and easy to understand
- ✅ Maintain consistency with Git naming
- ❌ Avoid conflicts with Git native commands

## Summary

### easyGit's Value Lies In

1. **Integration** - One-stop solution for multi-step workflows
2. **Guidance** - Interactive interface for complex commands
3. **Efficiency** - Batch operations and smart defaults
4. **Enhancement** - Better visual feedback and user experience

### easyGit Is NOT

1. ❌ A complete replacement for Git
2. ❌ A shortcut to learning Git
3. ❌ A TUI version of all Git commands

### Advice for Developers

When considering adding new features, ask yourself:

1. What pain point does this feature solve?
2. How complex is the Git native approach?
3. How much value does our solution provide?
4. Will it make users lose understanding of Git?

**Remember: We make Git better to use, not replace Git.**
