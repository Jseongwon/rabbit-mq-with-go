# SpecKit - Specification-Driven Development

**Version**: 1.0

Start or continue a test-driven, specification-first development workflow.

## What is SpecKit?

SpecKit is a structured workflow system that transforms project ideas into production-ready code through four phases:

1. **Constitute** → Define project principles
2. **Specify** → Create detailed requirements
3. **Plan** → Design technical architecture
4. **Implement** → Build with TDD

Each phase has quality gates and human checkpoints. The specification stays as the source of truth.

## Usage

```
/speckit [project-name]
```

If no project name provided, you'll be prompted for one.

## How It Works

### First Time (New Workflow)

```
You: /speckit task-manager

SpecKit will:
1. Ask about your project idea
2. Create project constitution (principles)
3. Generate specification (requirements)
4. Wait for your approval
5. Create technical plan
6. Wait for your approval
7. Guide you through TDD implementation
```

### Resume Existing Workflow

```
You: /speckit

SpecKit will:
1. Detect existing workflow in .speckit/
2. Show current status and progress
3. Resume from where you left off
```

## Workflow Phases

### Phase 1: Constitute (5-10 minutes)

**Purpose**: Define guiding principles for the project.

**Agent**: Interactive conversation
**Output**: `.speckit/CONSTITUTION.md`
**Checkpoint**: Review principles, confirm or refine

**Questions Asked**:
- What is the project's core purpose?
- What principles should guide decisions?
- What should the project prioritize? (e.g., simplicity, performance, security)
- What are non-negotiables?

**Example Output**:
```markdown
# Task Manager Constitution

## Principles
1. User privacy is paramount
2. Simple > Feature-rich
3. Mobile-first design
4. Offline-capable
```

### Phase 2: Specify (20-40 minutes)

**Purpose**: Document detailed requirements.

**Agent**: Requirements Analyst
**Input**: Project description + Constitution
**Output**: `.speckit/SPECIFICATION.md` + quality report
**Quality Gate**: ≥85% (completeness, clarity, testability)
**Checkpoint**: Review spec, approve or request refinements

**What You'll Get**:
- 5+ functional requirements
- 2+ non-functional requirements
- 2+ user stories with acceptance criteria
- Constraints and success metrics
- Open questions

**Example**:
```markdown
## FR001: Task Creation
User can create a task with title (required, max 200 chars),
optional description (max 2000 chars), due date (ISO 8601),
and priority (low/medium/high).

## User Story 1
As a user, I want to create tasks with deadlines
so that I can track what needs to be done and when.

Acceptance Criteria:
- [ ] POST /tasks endpoint accepts title, description, due_date, priority
- [ ] Title validated (1-200 chars)
- [ ] Returns 201 with task object on success
```

### Phase 3: Plan (30-60 minutes)

**Purpose**: Design technical architecture and break into tasks.

**Agent**: Technical Architect
**Input**: Approved specification
**Output**: `.speckit/PLAN.md` + quality report
**Quality Gate**: ≥85% (completeness, actionability, feasibility)
**Checkpoint**: Review architecture and tasks, approve or request changes

**What You'll Get**:
- Architecture overview (components, tech choices)
- Task breakdown (2-8 hour tasks)
- Dependency graph
- Timeline estimate
- Test strategy

**Example**:
```markdown
## Architecture
Three-tier: React frontend, Express API, PostgreSQL database

## Tasks
T001: Initialize project structure (2h)
T002: Database schema (4h) [depends on T001]
T003: Write tests for task CRUD (3h) [depends on T002]
T004: Implement task CRUD (4h) [depends on T003]
...
```

### Phase 4: Implement (Varies)

**Purpose**: Build the project using TDD.

**Agent**: Implementation Engineer
**Input**: Approved plan
**Output**: Working code + tests + quality reports per task
**Quality Gate** (per task):
- All tests passing ✅
- Coverage ≥80% ✅
- No lint errors ✅
- Acceptance criteria met ✅

**Process**:
For each task:
1. **RED**: Write failing tests
2. **GREEN**: Implement to pass tests
3. **REFACTOR**: Clean up code
4. **VALIDATE**: Run quality checks
5. **COMMIT**: Save progress

**Progress Tracking**:
```
Phase 4: Implementation [████████░░] 80%
  ✅ T001: Initialize project (2h)
  ✅ T002: Database schema (4h)
  ✅ T003: Write CRUD tests (3h)
  ✅ T004: Implement CRUD (4h)
  ⏳ T005: Authentication (6h) <- Current
  ⏸  T006: Authorization (4h)
  ⏸  T007: Frontend (12h)
```

## Commands Reference

```bash
/speckit                    # Start new or resume existing workflow
/speckit my-project         # Start new workflow with name
```

Within workflow, you can:
```bash
/speckit status             # Show detailed progress
/speckit refine             # Refine current phase
/speckit reset              # Start over (with confirmation)
```

## File Structure

```
your-project/
├── .speckit/
│   ├── state.json                    # Workflow state
│   ├── CONSTITUTION.md               # Project principles
│   ├── SPECIFICATION.md              # Requirements
│   ├── PLAN.md                       # Technical plan
│   └── quality/
│       ├── spec-quality.json
│       ├── plan-quality.json
│       ├── T001-quality.json
│       └── ...
├── src/                              # Your code
├── tests/                            # Your tests
└── package.json
```

## Quality Gates Explained

### Why Quality Gates?

Quality gates prevent downstream problems:
- Bad specs → Wrong implementation
- Vague requirements → Unclear tasks
- Untestable criteria → No way to verify

### What Happens If Quality Gate Fails?

```
❌ Specification Quality: 72/100 (threshold: 85)

Issues:
- Too few functional requirements (3, need 5+)
- No user stories defined
- 4 requirements are too vague

Recommendations:
- Add more functional requirements
- Create user stories with acceptance criteria
- Use specific, measurable language

Next: Refine specification to address issues above
```

You can:
1. **Refine**: Update the current phase
2. **Override**: Proceed anyway (not recommended)
3. **Restart**: Start this phase over

## Example Workflow

### Starting Fresh

```
You: /speckit blog-api

SpecKit: Great! Let's build your blog API.
First, let's establish some guiding principles.

What is the core purpose of this blog API?

You: A REST API for managing blog posts with user authentication.

SpecKit: Perfect. What should this project prioritize?
- Performance
- Security
- Simplicity
- Feature-richness
- Scalability

You: Security and simplicity.

[Constitution created]

SpecKit: ✅ Constitution created
Review: .speckit/CONSTITUTION.md

Ready to create the specification?

You: Yes

[Analyst agent creates specification]

SpecKit: ✅ Specification created (Quality: 88/100)
- 8 functional requirements
- 3 non-functional requirements
- 4 user stories
Review: .speckit/SPECIFICATION.md

Approve specification? (yes/refine/restart)

You: Yes

[Architect agent creates plan]

SpecKit: ✅ Plan created (Quality: 91/100)
- Express.js API with PostgreSQL
- 15 tasks, ~40 hours
- Clear dependencies
Review: .speckit/PLAN.md

Approve plan? (yes/refine/restart)

You: Yes

[Implementation begins]

SpecKit: Starting implementation Phase

T001: Initialize Project Structure
Let's start with RED phase - write the tests...
```

### Resuming Work

```
You: /speckit

SpecKit:
═══════════════════════════════════════════
📊 SpecKit Workflow Status
═══════════════════════════════════════════
Project: Blog API
Progress: [████████░░] 73%

Phase Status:
  ✅ Constitute    (100%)
  ✅ Specify       (88/100)
  ✅ Plan          (91/100)
  ⏳ Implement     (11/15 tasks)

Current Task: T012 - User Authentication
Last Activity: 2 hours ago

Next Action: Continue implementing T012

Continue? (yes/status/refine)
```

## Best Practices

### 1. Take Time on Specifications

Better to spend 30 minutes getting requirements right than 30 hours building the wrong thing.

### 2. Review at Checkpoints

Actually review the outputs. Don't auto-approve. Catch issues early.

### 3. Trust the Quality Gates

If quality score is low, there's a real problem. Address it.

### 4. Follow TDD Strictly

RED → GREEN → REFACTOR. No shortcuts. Tests catch bugs early.

### 5. Keep Constitution Handy

When making decisions, refer back to your principles.

## Troubleshooting

**"Quality gate keeps failing"**
→ Read the specific issues listed in the quality report
→ Address each issue explicitly
→ Ask for clarification if recommendations unclear

**"Stuck in a phase"**
→ Use `/speckit refine` to update the current phase
→ Or `/speckit reset` to start the phase over

**"Task taking too long"**
→ Break it into smaller tasks in the plan
→ Update plan and regenerate task list

**"Tests failing"**
→ This is expected in RED phase
→ In GREEN phase, implement until they pass
→ Never skip failing tests

## Principles (From Constitution)

SpecKit follows its own constitution:

1. **Simplicity Over Complexity**: One command, clear workflow
2. **Test-First, Always**: TDD is not optional
3. **Specifications Are Executable**: They drive implementation
4. **Human Judgment Required**: Checkpoints for review
5. **Iterative, Not Linear**: Easy to refine and improve

## Getting Help

- Stuck? Ask: "What should I do next?"
- Confused about a phase? Ask: "Explain the specify phase"
- Quality gate failed? Ask: "How do I improve the specification quality?"

## Advanced: Refinement

At any checkpoint, you can refine:

```
You: /speckit refine

SpecKit: What would you like to refine?

You: The specification is missing API rate limiting requirements

SpecKit: [Updates specification]
✅ Specification updated (v2)

Changelog:
- Added NFR004: Rate limiting (100 req/min per user)
- Added FR009: Rate limit headers in responses

Quality: 91/100 (+3 from v1)

⚠️ Downstream Impact:
- Plan may need updating (add rate limiting task)

Regenerate plan? (yes/no)
```

---

## Summary

SpecKit transforms ideas → specifications → plans → code with:
- ✅ Quality gates at every phase
- ✅ Human review checkpoints
- ✅ TDD enforcement
- ✅ Clear, traceable artifacts
- ✅ Iterative refinement

**Start building better software:**
```
/speckit
```