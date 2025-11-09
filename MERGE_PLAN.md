# Askara Branch Merge Plan

**Generated:** 2025-11-09
**Current Branch:** `claude/check-branches-plan-merge-011CUwfGd2SXPkbVoymeEDDp`
**Target Branch:** `main`

---

## Executive Summary

Analysis of the Askara repository reveals **8 Claude branches** with **3 unmerged feature branches** containing new functionality that needs to be integrated into main. The current branch is already fully merged into main.

### Quick Stats
- **Total Remote Branches:** 9 (8 Claude branches + main)
- **Already Merged:** 5 branches (0 commits ahead of main)
- **Pending Merge:** 3 branches (5 total commits with conflicts)
- **Main Branch Commits:** 13 commits ahead of current branch

---

## Branch Status Overview

### ✅ Already Merged (No Action Needed)

These branches are fully integrated into main:

1. **claude/askara-next-steps-011CUtErzJPJHAFP2xpkh2Le** (0 commits ahead)
   - Status: Set as HEAD branch, fully merged

2. **claude/build-code-source-truth-011CUvAnb6ZPcYo7Axwb73in** (0 commits ahead)
   - Feature: Code Source of Truth system for enhanced code grounding
   - Merged in: commit `094f0c4`

3. **claude/check-branches-plan-merge-011CUwfGd2SXPkbVoymeEDDp** (0 commits ahead)
   - Status: Current branch, already merged into main
   - Can be safely deleted after this analysis

4. **claude/cloud-llm-plugin-plan-011CUux6skQbQWZGyLmq4RFr** (0 commits ahead)
   - Feature: Azure OpenAI, AWS Bedrock, Cohere cloud LLM providers
   - Merged in: commit `77b3778`

5. **claude/plan-option-plugin-speed-011CUusM78stK3ociGTVoiXH** (0 commits ahead)
   - Feature: Metrics framework and unified provider interface
   - Merged in: commit `fc8fe51`

---

## 🔴 Unmerged Branches (Action Required)

### Priority 1: claude/llm-provider-feature-011CUus3NZpZhqKrGMUpyaYL

**Commits Ahead:** 3
**Merge Conflicts:** YES (2 files)
**Complexity:** HIGH

#### Features
- Multi-provider LLM system (Anthropic Claude, Google Gemini, Groq)
- Analytics & metrics system with performance tracking
- Intelligent auto-selection with weighted scoring
- Three operating modes (Single, Auto, Failover)
- Analytics API endpoints

#### New Files (12)
- `llm/claude.go` (350 lines)
- `llm/gemini.go` (310 lines)
- `llm/groq.go` (156 lines)
- `llm/metrics.go` (360 lines)
- `llm/instrumented_provider.go` (280 lines)
- `llm/provider_manager.go` (320 lines)
- `vault-web-server/postapi/provider_analytics.go` (180 lines)
- `INTEGRATION_GUIDE.md` (350 lines)
- `TESTING_CHECKLIST.md` (200 lines)
- `FEATURE_SUMMARY.md`

#### Modified Files (5)
- `llm/claude.go`, `llm/gemini.go`, `llm/groq.go`
- `llm/instrumented_provider.go` ⚠️ **CONFLICT**
- `vault-web-server/main.go` ⚠️ **CONFLICT**

#### Conflicts
1. **llm/instrumented_provider.go** - Likely structural changes vs main
2. **vault-web-server/main.go** - API endpoint registration conflicts

#### Total Impact
~3,800 lines of new code + documentation

---

### Priority 2: claude/askara-document-storage-011CUuSWV26wnpGchYa2Uy4Q

**Commits Ahead:** 1
**Merge Conflicts:** YES (3 files)
**Complexity:** MEDIUM

#### Features
- Rich document metadata system
- ML-powered auto-tagging
- Automatic metadata extraction (summary, tags, language, entities)
- User-editable metadata fields
- Category classification
- Word count and reading time calculation

#### New Files (2)
- `METADATA_API.md`
- `storage/metadata.go`
- `vault-web-server/postapi/metadata_api.go`

#### Modified Files (5)
- `metadata/tagger.go`
- `storage/document.go` ⚠️ **CONFLICT**
- `vault-web-server/main.go` ⚠️ **CONFLICT**
- `vault-web-server/postapi/fileupload.go`
- `vault-web-server/postapi/handlercontext.go` ⚠️ **CONFLICT**

#### Conflicts
1. **storage/document.go** - Document struct modifications
2. **vault-web-server/main.go** - API endpoint registration
3. **vault-web-server/postapi/handlercontext.go** - Handler context changes

---

### Priority 3: claude/implement-pro-features-011CUuf6zMYvTr6A8mYFRoQ2

**Commits Ahead:** 1
**Merge Conflicts:** YES (1 file)
**Complexity:** LOW-MEDIUM

#### Features
- Sophisticated prompt forming system for document questions
- Automatic question type detection
- Optimized templates per question type
- Smart context formatting
- Dynamic temperature suggestion
- System instruction generation

#### New Files (9)
- `prompt/README.md`
- `prompt/builder.go`
- `prompt/builder_test.go`
- `prompt/detector.go`
- `prompt/detector_test.go`
- `prompt/examples_test.go`
- `prompt/templates.go`
- `prompt/types.go`

#### Modified Files (2)
- `llm/ollama.go` ⚠️ **CONFLICT**
- `llm/openai.go`

#### Conflicts
1. **llm/ollama.go** - Integration with prompt system vs main changes

---

## 📋 Recommended Merge Strategy

### Phase 1: Prepare Main Branch
```bash
# Ensure main is up to date
git checkout main
git pull origin main

# Create integration branch for safety
git checkout -b integration/merge-pending-features
```

### Phase 2: Merge in Recommended Order

The merge order is designed to minimize conflicts and ensure dependencies are handled correctly:

#### Step 1: Merge Prompt System (Lowest Complexity)
```bash
# Merge pro features (prompt system)
git merge origin/claude/implement-pro-features-011CUuf6zMYvTr6A8mYFRoQ2

# Resolve conflicts:
# - llm/ollama.go: Integrate prompt system calls with existing code
#
# Expected resolution time: 15-30 minutes
```

**Conflict Resolution Guide:**
- `llm/ollama.go`: Check if prompt optimization should be applied to Ollama calls
- Test that existing LLM functionality still works
- Run tests: `go test ./llm/... ./prompt/...`

#### Step 2: Merge Document Metadata System
```bash
# Merge document storage
git merge origin/claude/askara-document-storage-011CUuSWV26wnpGchYa2Uy4Q

# Resolve conflicts:
# - storage/document.go: Integrate metadata fields into Document struct
# - vault-web-server/main.go: Add metadata API endpoints
# - vault-web-server/postapi/handlercontext.go: Add metadata handler context
#
# Expected resolution time: 30-45 minutes
```

**Conflict Resolution Guide:**
- `storage/document.go`: Merge metadata fields into existing Document struct
- `vault-web-server/main.go`: Add metadata endpoints without duplicating existing routes
- `vault-web-server/postapi/handlercontext.go`: Integrate metadata tagger into handler context
- Test metadata extraction with sample documents
- Run tests: `go test ./storage/... ./metadata/... ./vault-web-server/...`

#### Step 3: Merge LLM Provider System (Highest Complexity)
```bash
# Merge LLM provider features
git merge origin/claude/llm-provider-feature-011CUus3NZpZhqKrGMUpyaYL

# Resolve conflicts:
# - llm/instrumented_provider.go: Integrate metrics instrumentation
# - vault-web-server/main.go: Add analytics API endpoints
#
# Expected resolution time: 45-60 minutes
```

**Conflict Resolution Guide:**
- `llm/instrumented_provider.go`: Ensure metrics wrap all providers including new ones
- `vault-web-server/main.go`: Add `/api/providers/*` endpoints
- Verify all 5 LLM providers work correctly
- Test auto-selection and failover modes
- Test analytics endpoints return valid data
- Run comprehensive tests: `go test ./llm/... ./vault-web-server/...`

### Phase 3: Integration Testing
```bash
# Full test suite
go test ./...

# Build verification
go build ./...

# Start server and test endpoints
go run vault-web-server/main.go

# Test key functionality:
# 1. Document upload with metadata extraction
# 2. LLM queries with different providers
# 3. Prompt optimization for questions
# 4. Analytics endpoints
# 5. Provider auto-selection
```

### Phase 4: Merge to Main
```bash
# If all tests pass
git checkout main
git merge integration/merge-pending-features

# Push to origin
git push origin main
```

---

## ⚠️ Conflict Details

### Common Conflict Point: vault-web-server/main.go

This file is modified by **2 branches**, requiring careful coordination:

- **Document Storage Branch**: Adds metadata API endpoints
  - `POST /api/documents/:id/metadata`
  - `GET /api/documents/:id/metadata`

- **LLM Provider Branch**: Adds analytics endpoints
  - `GET /api/providers/stats`
  - `GET /api/providers/metrics`
  - `GET /api/providers/health`

**Resolution:** Both sets of endpoints should be added without overlap.

### Dependency Chain

```
main (current state)
  |
  ├── Step 1: + Prompt System (isolated, minimal conflicts)
  |            └── Enhances LLM queries
  |
  ├── Step 2: + Document Metadata (uses LLM for auto-tagging)
  |            └── May benefit from prompt system
  |
  └── Step 3: + LLM Provider System (major LLM changes)
               └── Should work with both prompt system and metadata
```

---

## 🧪 Testing Checklist

After each merge step:

- [ ] Code compiles without errors
- [ ] All existing tests pass
- [ ] New tests added for new features pass
- [ ] Manual testing of new endpoints
- [ ] Integration testing of combined features
- [ ] Documentation is updated
- [ ] No regression in existing functionality

---

## 📊 Risk Assessment

### Low Risk
- Prompt system merge (isolated package, minimal changes to existing code)

### Medium Risk
- Document metadata merge (changes to core storage layer)

### High Risk
- LLM provider merge (significant changes to LLM system, multiple providers)

### Mitigation Strategy
1. Use integration branch (not directly on main)
2. Merge in order of increasing complexity
3. Test after each merge
4. Keep backups of working states
5. Be prepared to resolve conflicts manually
6. Have rollback plan ready

---

## 🎯 Post-Merge Cleanup

After successful merge to main:

```bash
# Delete merged feature branches (after confirming merge)
git push origin --delete claude/implement-pro-features-011CUuf6zMYvTr6A8mYFRoQ2
git push origin --delete claude/askara-document-storage-011CUuSWV26wnpGchYa2Uy4Q
git push origin --delete claude/llm-provider-feature-011CUus3NZpZhqKrGMUpyaYL

# Delete fully merged branches
git push origin --delete claude/check-branches-plan-merge-011CUwfGd2SXPkbVoymeEDDp
git push origin --delete claude/build-code-source-truth-011CUvAnb6ZPcYo7Axwb73in
git push origin --delete claude/cloud-llm-plugin-plan-011CUux6skQbQWZGyLmq4RFr
git push origin --delete claude/plan-option-plugin-speed-011CUusM78stK3ociGTVoiXH

# Update remote tracking
git remote prune origin
```

---

## 📝 Summary

### Total New Features to Merge
1. **Prompt Forming System** - Smart question optimization for RAG
2. **Document Metadata** - ML-powered auto-tagging and metadata extraction
3. **Multi-Provider LLM** - 5 providers with analytics and auto-selection

### Total Conflicts to Resolve
- **6 file conflicts** across 3 branches
- **1 shared conflict** (vault-web-server/main.go in 2 branches)
- Estimated total resolution time: **1.5 - 2.5 hours**

### Expected Outcome
A unified main branch with:
- 5 LLM providers (Ollama, OpenAI, Claude, Gemini, Groq)
- Smart prompt optimization for better answers
- Rich document metadata with auto-tagging
- Analytics and metrics for LLM usage
- ~4,000+ lines of new functionality

---

## 🚀 Next Steps

1. **Review this plan** with the team
2. **Schedule merge window** (recommend 2-3 hour block)
3. **Backup current main** before starting
4. **Execute Phase 1-4** as outlined above
5. **Deploy and monitor** the integrated system

---

**Document Version:** 1.0
**Last Updated:** 2025-11-09
**Author:** Claude Code Analysis
