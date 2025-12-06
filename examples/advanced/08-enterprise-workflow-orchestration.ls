# LangSpace Advanced Example: Enterprise Workflow Orchestration
# Complex business process automation with approvals, compliance, and audit trails.
#
# This example demonstrates:
# - Human-in-the-loop workflows
# - Multi-level approval chains
# - Compliance and audit logging
# - SLA monitoring and escalation
# - Cross-system integration
# - Role-based access control
# - State machine workflows

# ============================================================================
# CONFIGURATION
# ============================================================================

config {
  default_model: "claude-sonnet-4-20250514"

  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")
    }
  }

  # Organization settings
  organization: {
    name: env("ORG_NAME")
    domain: env("ORG_DOMAIN")
    timezone: "America/New_York"
  }

  # Compliance settings
  compliance: {
    audit_retention_days: 2555  # 7 years
    require_dual_approval: true
    pii_handling: "strict"
    data_classification: ["public", "internal", "confidential", "restricted"]
  }

  # SLA settings
  sla: {
    expense_approval: "48h"
    vendor_onboarding: "5d"
    access_request: "24h"
    contract_review: "72h"
  }

  # Integration endpoints
  integrations: {
    erp: env("SAP_API_URL")
    hr: env("WORKDAY_API_URL")
    finance: env("NETSUITE_API_URL")
    legal: env("DOCUSIGN_API_URL")
    it: env("SERVICENOW_API_URL")
  }
}

# ============================================================================
# POLICIES AND PROCEDURES
# ============================================================================

file "approval-matrix" {
  contents: ```
    # Approval Matrix

    ## Expense Approvals

    | Amount Range     | Required Approvers                    | SLA  |
    |-----------------|---------------------------------------|------|
    | $0 - $500       | Direct Manager                        | 24h  |
    | $500 - $5,000   | Direct Manager + Department Head      | 48h  |
    | $5,000 - $25,000| Department Head + Finance             | 72h  |
    | $25,000 - $100k | VP + Finance Director                 | 5d   |
    | > $100,000      | C-Suite + CFO + Board (if >$500k)    | 10d  |

    ## Vendor Onboarding

    | Risk Level | Required Reviews                          | SLA  |
    |------------|------------------------------------------|------|
    | Low        | Procurement                              | 3d   |
    | Medium     | Procurement + Legal                      | 5d   |
    | High       | Procurement + Legal + Security + Finance | 10d  |
    | Critical   | All above + CISO + General Counsel       | 15d  |

    ## Access Requests

    | Access Type        | Approvers                    | Review Period |
    |-------------------|------------------------------|---------------|
    | Standard Systems  | Manager                      | Annual        |
    | Financial Systems | Manager + Finance            | Quarterly     |
    | Admin/Root        | Manager + IT Security + CISO | Monthly       |
    | Customer Data     | Manager + Legal + DPO        | Quarterly     |

    ## Escalation Policy

    - 75% SLA elapsed: Reminder to approvers
    - 90% SLA elapsed: Escalate to approver's manager
    - 100% SLA elapsed: Escalate to VP + auto-page
    - 150% SLA elapsed: Executive escalation
  ```
}

file "compliance-requirements" {
  contents: ```
    # Compliance Requirements

    ## Data Classification Handling

    ### Public
    - No restrictions on storage or transmission
    - Standard access logging

    ### Internal
    - Encrypt at rest
    - Access logging required
    - Annual access review

    ### Confidential
    - Encrypt at rest and in transit
    - Detailed access logging
    - Quarterly access review
    - Manager approval for access

    ### Restricted
    - End-to-end encryption
    - Immutable audit logs
    - Monthly access review
    - Dual approval for access
    - DLP monitoring enabled
    - No external sharing

    ## Audit Requirements

    ### All Transactions Must Log:
    - Timestamp (UTC)
    - Actor (user/system)
    - Action performed
    - Target resource
    - Outcome (success/failure)
    - IP address / session ID
    - Justification (if applicable)

    ### Retention
    - Financial records: 7 years
    - Access logs: 3 years
    - Approval chains: 7 years
    - Communication logs: 5 years

    ## Regulatory Compliance
    - SOC 2 Type II
    - GDPR (for EU data)
    - CCPA (for CA data)
    - SOX (for financial processes)
    - HIPAA (if health data)
  ```
}

file "workflow-states" {
  contents: ```
    # Standard Workflow States

    ## Request Lifecycle
    1. DRAFT - Initial creation, not submitted
    2. SUBMITTED - Awaiting initial review
    3. IN_REVIEW - Being reviewed by approvers
    4. PENDING_APPROVAL - All reviews complete, awaiting final approval
    5. APPROVED - Fully approved
    6. REJECTED - Denied (with reason)
    7. CANCELLED - Withdrawn by requester
    8. ON_HOLD - Temporarily paused (with reason)
    9. COMPLETED - Request fulfilled
    10. ARCHIVED - Closed and archived

    ## State Transitions

    DRAFT -> SUBMITTED (by requester)
    SUBMITTED -> IN_REVIEW (by system)
    IN_REVIEW -> PENDING_APPROVAL (all reviews positive)
    IN_REVIEW -> REJECTED (any review negative)
    IN_REVIEW -> ON_HOLD (more info needed)
    PENDING_APPROVAL -> APPROVED (final approver)
    PENDING_APPROVAL -> REJECTED (final approver)
    ON_HOLD -> IN_REVIEW (info provided)
    APPROVED -> COMPLETED (fulfillment confirmed)
    COMPLETED -> ARCHIVED (after retention period)
    * -> CANCELLED (by requester, before APPROVED)

    ## Audit Events per State
    - All transitions logged with actor, timestamp, reason
    - IN_REVIEW: Log each reviewer's decision
    - REJECTED/ON_HOLD: Require reason
    - APPROVED: Log all approvers in chain
  ```
}

# ============================================================================
# TOOLS
# ============================================================================

mcp "servicenow" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@company/mcp-servicenow"]
}

mcp "docusign" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@company/mcp-docusign"]
}

tool "get_user_info" {
  description: "Get user information from HR system"

  parameters: {
    user_id: string required
    include: array optional ["manager", "department", "role", "cost_center"]
  }

  handler: http {
    method: "GET"
    url: config.integrations.hr + "/users/" + params.user_id
    headers: {
      "Authorization": "Bearer " + env("WORKDAY_API_KEY")
    }
    query: { include: params.include.join(",") }
  }
}

tool "get_approval_chain" {
  description: "Determine required approvers based on request type and amount"

  parameters: {
    request_type: string required "expense, vendor, access, contract"
    amount: number optional
    risk_level: string optional
    user_id: string required "Requester's ID"
  }

  handler: builtin("approval_chain_resolver")
}

tool "create_request" {
  description: "Create a new workflow request"

  parameters: {
    type: string required
    title: string required
    description: string required
    requester: string required
    data: object required
    priority: string optional "low, normal, high, critical"
    attachments: array optional
  }

  handler: http {
    method: "POST"
    url: config.integrations.it + "/requests"
    headers: {
      "Authorization": "Bearer " + env("SERVICENOW_API_KEY"),
      "Content-Type": "application/json"
    }
    body: params
  }
}

tool "update_request_status" {
  description: "Update request status"

  parameters: {
    request_id: string required
    status: string required
    reason: string optional
    actor: string required
  }

  handler: mcp("servicenow").update_record
}

tool "send_approval_request" {
  description: "Send approval request to approver"

  parameters: {
    request_id: string required
    approver_id: string required
    approval_type: string required "review, approve, sign"
    due_date: string required
    context: object required
  }

  handler: http {
    method: "POST"
    url: config.integrations.it + "/approvals"
    body: params
  }
}

tool "get_approval_status" {
  description: "Get current approval status"

  parameters: {
    request_id: string required
  }

  handler: mcp("servicenow").get_approvals
}

tool "log_audit_event" {
  description: "Log an audit event"

  parameters: {
    event_type: string required
    request_id: string required
    actor: string required
    action: string required
    target: string required
    outcome: string required "success, failure"
    details: object optional
    classification: string optional
  }

  handler: http {
    method: "POST"
    url: env("AUDIT_LOG_URL") + "/events"
    headers: {
      "Authorization": "Bearer " + env("AUDIT_API_KEY"),
      "X-Classification": params.classification
    }
    body: {
      timestamp: now(),
      ...params
    }
  }
}

tool "send_notification" {
  description: "Send notification to user"

  parameters: {
    user_id: string required
    channel: string required "email, slack, sms, push"
    template: string required
    data: object required
    priority: string optional "low, normal, high, urgent"
  }

  handler: http {
    method: "POST"
    url: env("NOTIFICATION_SERVICE_URL") + "/send"
    body: params
  }
}

tool "check_sla" {
  description: "Check SLA status for a request"

  parameters: {
    request_id: string required
    sla_type: string required
  }

  handler: builtin("sla_checker")
}

tool "escalate" {
  description: "Escalate a request"

  parameters: {
    request_id: string required
    escalation_level: number required
    reason: string required
    escalate_to: string required
  }

  handler: http {
    method: "POST"
    url: config.integrations.it + "/escalations"
    body: params
  }
}

tool "create_contract" {
  description: "Create a contract for signature"

  parameters: {
    template_id: string required
    parties: array required
    data: object required
    expiry: string optional
  }

  handler: mcp("docusign").create_envelope
}

tool "check_compliance" {
  description: "Run compliance checks on a request"

  parameters: {
    request_id: string required
    checks: array required "Required compliance checks to run"
  }

  handler: builtin("compliance_checker")
}

# ============================================================================
# WORKFLOW SCRIPTS
# ============================================================================

# Determine approval requirements
script "resolve-approval-chain" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    request_type: string required
    amount: number optional
    risk_level: string optional
    requester_id: string required
  }

  code: ```python
    import json
    import db

    # Get requester's org chain
    user = db.query_one(f"""
        SELECT u.*, m.id as manager_id, m.name as manager_name,
               d.head_id as dept_head_id, d.name as dept_name,
               d.vp_id, d.cost_center
        FROM users u
        JOIN users m ON u.manager_id = m.id
        JOIN departments d ON u.department_id = d.id
        WHERE u.id = '{requester_id}'
    """)

    chain = []

    if request_type == "expense":
        amount = amount or 0

        # Tier 1: Direct Manager
        if amount > 0:
            chain.append({
                "role": "manager",
                "user_id": user["manager_id"],
                "name": user["manager_name"],
                "required": True,
                "order": 1
            })

        # Tier 2: Department Head (>$500)
        if amount > 500:
            chain.append({
                "role": "department_head",
                "user_id": user["dept_head_id"],
                "required": True,
                "order": 2
            })

        # Tier 3: Finance (>$5000)
        if amount > 5000:
            finance = db.query_one("SELECT id, name FROM users WHERE role = 'finance_approver'")
            chain.append({
                "role": "finance",
                "user_id": finance["id"],
                "name": finance["name"],
                "required": True,
                "order": 3
            })

        # Tier 4: VP (>$25000)
        if amount > 25000:
            chain.append({
                "role": "vp",
                "user_id": user["vp_id"],
                "required": True,
                "order": 4
            })

        # Tier 5: CFO (>$100000)
        if amount > 100000:
            cfo = db.query_one("SELECT id, name FROM users WHERE role = 'cfo'")
            chain.append({
                "role": "cfo",
                "user_id": cfo["id"],
                "name": cfo["name"],
                "required": True,
                "order": 5
            })

    elif request_type == "vendor":
        risk = risk_level or "medium"

        # Always need procurement
        proc = db.query_one("SELECT id, name FROM users WHERE role = 'procurement'")
        chain.append({
            "role": "procurement",
            "user_id": proc["id"],
            "required": True,
            "order": 1
        })

        if risk in ["medium", "high", "critical"]:
            legal = db.query_one("SELECT id, name FROM users WHERE role = 'legal'")
            chain.append({
                "role": "legal",
                "user_id": legal["id"],
                "required": True,
                "order": 2
            })

        if risk in ["high", "critical"]:
            security = db.query_one("SELECT id, name FROM users WHERE role = 'security'")
            chain.append({
                "role": "security",
                "user_id": security["id"],
                "required": True,
                "order": 3
            })

        if risk == "critical":
            ciso = db.query_one("SELECT id, name FROM users WHERE role = 'ciso'")
            chain.append({
                "role": "ciso",
                "user_id": ciso["id"],
                "required": True,
                "order": 4
            })

    elif request_type == "access":
        # Manager always required
        chain.append({
            "role": "manager",
            "user_id": user["manager_id"],
            "required": True,
            "order": 1
        })

        # Add IT Security for privileged access
        if risk_level in ["high", "critical"]:
            it_sec = db.query_one("SELECT id, name FROM users WHERE role = 'it_security'")
            chain.append({
                "role": "it_security",
                "user_id": it_sec["id"],
                "required": True,
                "order": 2
            })

    result = {
        "requester": {
            "id": user["id"],
            "department": user["dept_name"],
            "cost_center": user["cost_center"]
        },
        "approval_chain": chain,
        "parallel_approvals": False,  # Sequential by default
        "sla_hours": {
            "expense": 48,
            "vendor": 120,
            "access": 24,
            "contract": 72
        }.get(request_type, 48)
    }

    print(json.dumps(result, indent=2))
  ```
}

# Check SLA compliance
script "check-sla-status" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    request_id: string required
  }

  code: ```python
    import json
    import db
    from datetime import datetime, timedelta

    request = db.query_one(f"""
        SELECT r.*, rt.sla_hours
        FROM requests r
        JOIN request_types rt ON r.type = rt.name
        WHERE r.id = '{request_id}'
    """)

    created = datetime.fromisoformat(request["created_at"])
    sla_deadline = created + timedelta(hours=request["sla_hours"])
    now = datetime.now()

    elapsed = (now - created).total_seconds() / 3600
    remaining = (sla_deadline - now).total_seconds() / 3600
    percentage = (elapsed / request["sla_hours"]) * 100

    status = {
        "request_id": request_id,
        "created_at": request["created_at"],
        "sla_deadline": sla_deadline.isoformat(),
        "elapsed_hours": round(elapsed, 1),
        "remaining_hours": round(max(0, remaining), 1),
        "percentage_elapsed": round(percentage, 1),
        "status": "ok"
    }

    if percentage >= 150:
        status["status"] = "critical"
        status["escalation_level"] = 4
    elif percentage >= 100:
        status["status"] = "breached"
        status["escalation_level"] = 3
    elif percentage >= 90:
        status["status"] = "at_risk"
        status["escalation_level"] = 2
    elif percentage >= 75:
        status["status"] = "warning"
        status["escalation_level"] = 1
    else:
        status["escalation_level"] = 0

    # Get pending approvers
    pending = db.query(f"""
        SELECT a.*, u.name, u.email
        FROM approvals a
        JOIN users u ON a.approver_id = u.id
        WHERE a.request_id = '{request_id}'
        AND a.status = 'pending'
    """)

    status["pending_approvers"] = [
        {"id": p["approver_id"], "name": p["name"], "email": p["email"]}
        for p in pending
    ]

    print(json.dumps(status, indent=2))
  ```
}

# Audit trail generation
script "generate-audit-report" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    request_id: string required
    include_pii: bool optional false
  }

  code: ```python
    import json
    import db
    import hashlib

    def mask_pii(value, include_pii):
        if include_pii:
            return value
        if isinstance(value, str) and "@" in value:
            parts = value.split("@")
            return parts[0][:2] + "***@" + parts[1]
        return value

    # Get request details
    request = db.query_one(f"""
        SELECT r.*, u.name as requester_name, u.email as requester_email
        FROM requests r
        JOIN users u ON r.requester_id = u.id
        WHERE r.id = '{request_id}'
    """)

    # Get all audit events
    events = db.query(f"""
        SELECT ae.*, u.name as actor_name
        FROM audit_events ae
        LEFT JOIN users u ON ae.actor_id = u.id
        WHERE ae.request_id = '{request_id}'
        ORDER BY ae.timestamp ASC
    """)

    # Get all approvals
    approvals = db.query(f"""
        SELECT a.*, u.name as approver_name
        FROM approvals a
        JOIN users u ON a.approver_id = u.id
        WHERE a.request_id = '{request_id}'
        ORDER BY a.order_num ASC
    """)

    report = {
        "request": {
            "id": request["id"],
            "type": request["type"],
            "title": request["title"],
            "status": request["status"],
            "requester": mask_pii(request["requester_email"], include_pii),
            "created_at": request["created_at"],
            "completed_at": request.get("completed_at"),
            "amount": request.get("amount")
        },
        "timeline": [],
        "approvals": [],
        "compliance": {
            "sla_met": True,
            "all_approvals_obtained": True,
            "audit_complete": True
        }
    }

    for event in events:
        report["timeline"].append({
            "timestamp": event["timestamp"],
            "action": event["action"],
            "actor": mask_pii(event["actor_name"], include_pii),
            "outcome": event["outcome"],
            "details": event.get("details")
        })

    for approval in approvals:
        report["approvals"].append({
            "role": approval["role"],
            "approver": mask_pii(approval["approver_name"], include_pii),
            "status": approval["status"],
            "decided_at": approval.get("decided_at"),
            "comments": approval.get("comments")
        })

        if approval["status"] == "pending":
            report["compliance"]["all_approvals_obtained"] = False

    # Generate hash for integrity
    content = json.dumps(report, sort_keys=True)
    report["integrity_hash"] = hashlib.sha256(content.encode()).hexdigest()

    print(json.dumps(report, indent=2))
  ```
}

# ============================================================================
# WORKFLOW AGENTS
# ============================================================================

agent "request-classifier" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You classify and validate incoming workflow requests.

    For each request, determine:
    1. Request type (expense, vendor, access, contract, custom)
    2. Risk level (low, medium, high, critical)
    3. Priority (low, normal, high, critical)
    4. Data classification (public, internal, confidential, restricted)
    5. Required compliance checks

    Validation checks:
    - All required fields present
    - Amount within acceptable range
    - Proper justification provided
    - No policy violations

    Output:
    {
      "valid": true,
      "type": "expense",
      "risk_level": "medium",
      "priority": "normal",
      "classification": "internal",
      "compliance_checks": ["budget_check", "duplicate_check"],
      "issues": [],
      "recommendations": []
    }
  ```
}

agent "compliance-reviewer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.1

  instruction: file("compliance-requirements") + ```

    You review requests for compliance with organizational policies.

    Check for:
    1. Policy compliance (approval matrix, spending limits, etc.)
    2. Regulatory requirements (SOX, GDPR, etc.)
    3. Conflict of interest
    4. Segregation of duties
    5. Data handling requirements

    Flag any issues and provide recommendations.
    Be thorough but fair - don't block legitimate requests.

    Output:
    {
      "compliant": true,
      "checks_performed": ["list of checks"],
      "findings": [
        {
          "severity": "high",
          "issue": "Description",
          "regulation": "SOX",
          "recommendation": "How to resolve"
        }
      ],
      "approval_recommended": true,
      "conditions": ["Any conditions for approval"]
    }
  ```

  tools: [
    tool("check_compliance"),
  ]
}

agent "approval-coordinator" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: file("approval-matrix") + ```

    You coordinate the approval process for requests.

    Responsibilities:
    1. Determine required approvers based on request type/amount
    2. Send approval requests in the correct order
    3. Track approval status
    4. Handle escalations when SLAs are at risk
    5. Ensure proper documentation

    For each approval step:
    - Provide approvers with necessary context
    - Set appropriate deadlines
    - Follow up on pending approvals
    - Handle delegations and out-of-office
  ```

  tools: [
    tool("get_user_info"),
    tool("get_approval_chain"),
    tool("send_approval_request"),
    tool("get_approval_status"),
    tool("send_notification"),
  ]

  scripts: [
    script("resolve-approval-chain"),
    script("check-sla-status")
  ]
}

agent "sla-monitor" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You monitor SLA compliance and manage escalations.

    Monitoring actions:
    - 75% elapsed: Send reminder to pending approvers
    - 90% elapsed: Escalate to approver's manager
    - 100% elapsed (breach): Escalate to VP, page on-call
    - 150% elapsed: Executive escalation

    For each escalation:
    - Provide full context
    - List all previous attempts
    - Recommend actions
    - Log the escalation

    Goal: Prevent SLA breaches while respecting approvers' time.
  ```

  tools: [
    tool("check_sla"),
    tool("escalate"),
    tool("send_notification"),
    tool("log_audit_event"),
  ]

  scripts: [
    script("check-sla-status")
  ]
}

agent "request-fulfiller" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: ```
    You fulfill approved requests by integrating with backend systems.

    Fulfillment actions by type:
    - Expense: Create payment in finance system
    - Vendor: Create vendor record, set up contracts
    - Access: Provision access in identity systems
    - Contract: Send for signature, file when complete

    For each fulfillment:
    1. Verify all approvals are in place
    2. Execute the action
    3. Verify completion
    4. Notify requester
    5. Update audit trail
  ```

  tools: [
    tool("create_contract"),
    tool("update_request_status"),
    tool("send_notification"),
    tool("log_audit_event"),
  ]
}

# ============================================================================
# WORKFLOW PIPELINES
# ============================================================================

pipeline "process-request" {
  # Step 1: Classify and validate
  step "classify" {
    use: agent("request-classifier")
    input: $input.request
  }

  # Step 2: Check compliance
  step "compliance" {
    use: agent("compliance-reviewer")

    input: {
      request: $input.request,
      classification: step("classify").output
    }
  }

  # Step 3: Handle non-compliant requests
  branch step("compliance").output.compliant == false {
    true => step "reject" {
      tools: [
        tool("update_request_status"),
        tool("send_notification"),
        tool("log_audit_event")
      ]

      input: {
        request_id: $input.request.id,
        status: "REJECTED",
        reason: step("compliance").output.findings[0].issue,
        actor: "system"
      }
    }
  }

  # Step 4: Resolve approval chain
  step "resolve-approvers" {
    execute: script("resolve-approval-chain") {
      request_type: step("classify").output.type
      amount: $input.request.amount
      risk_level: step("classify").output.risk_level
      requester_id: $input.request.requester_id
    }
  }

  # Step 5: Create the request
  step "create" {
    tools: [tool("create_request"), tool("log_audit_event")]

    input: {
      request: $input.request,
      classification: step("classify").output,
      approval_chain: step("resolve-approvers").output.approval_chain
    }
  }

  # Step 6: Initiate approval flow
  step "start-approvals" {
    use: agent("approval-coordinator")

    input: {
      request_id: step("create").output.id,
      approval_chain: step("resolve-approvers").output.approval_chain
    }
  }

  output: {
    request_id: step("create").output.id,
    status: "IN_REVIEW",
    approval_chain: step("resolve-approvers").output.approval_chain,
    estimated_completion: step("resolve-approvers").output.sla_hours
  }
}

pipeline "handle-approval-decision" {
  # Step 1: Log the decision
  step "log" {
    tools: [tool("log_audit_event")]

    input: {
      event_type: "approval_decision",
      request_id: $input.request_id,
      actor: $input.approver_id,
      action: $input.decision,
      outcome: $input.decision == "approved" ? "success" : "failure",
      details: { comments: $input.comments }
    }
  }

  # Step 2: Update approval status
  step "update-approval" {
    tools: [tool("update_request_status")]

    input: {
      request_id: $input.request_id,
      approval_id: $input.approval_id,
      status: $input.decision,
      comments: $input.comments
    }
  }

  # Step 3: Branch on decision
  branch $input.decision {
    "approved" => step "check-next" {
      use: agent("approval-coordinator")

      input: {
        request_id: $input.request_id,
        action: "check_remaining_approvals"
      }
    }

    "rejected" => step "handle-rejection" {
      tools: [
        tool("update_request_status"),
        tool("send_notification")
      ]

      input: {
        request_id: $input.request_id,
        status: "REJECTED",
        reason: $input.comments
      }
    }

    "request_changes" => step "request-info" {
      tools: [tool("send_notification")]

      input: {
        user_id: $input.requester_id,
        template: "additional_info_needed",
        data: { comments: $input.comments }
      }
    }
  }

  # Step 4: If all approved, proceed to fulfillment
  branch step("check-next").output.all_approved == true {
    true => step "fulfill" {
      run: pipeline("fulfill-request") {
        input: { request_id: $input.request_id }
      }
    }
  }

  output: {
    request_id: $input.request_id,
    decision: $input.decision,
    next_step: $branch.output
  }
}

pipeline "fulfill-request" {
  # Step 1: Verify all approvals
  step "verify" {
    tools: [tool("get_approval_status")]
    input: { request_id: $input.request_id }
  }

  # Step 2: Execute fulfillment
  step "execute" {
    use: agent("request-fulfiller")

    input: {
      request_id: $input.request_id,
      approvals: step("verify").output
    }
  }

  # Step 3: Update status
  step "complete" {
    tools: [
      tool("update_request_status"),
      tool("log_audit_event"),
      tool("send_notification")
    ]

    input: {
      request_id: $input.request_id,
      status: "COMPLETED",
      fulfillment: step("execute").output
    }
  }

  output: step("complete").output
}

pipeline "sla-monitoring" {
  # Get all in-progress requests
  step "get-active" {
    tools: [tool("get_active_requests")]
  }

  # Check each for SLA status
  step "check-all" {
    for request in step("get-active").output {
      execute: script("check-sla-status") {
        request_id: request.id
      }
    }
  }

  # Handle escalations
  step "escalate" {
    for status in step("check-all").output {
      branch status.escalation_level > 0 {
        true => {
          use: agent("sla-monitor")
          input: status
        }
      }
    }
  }

  output: {
    checked: step("check-all").output.length,
    escalated: step("escalate").output.filter(e => e.escalated).length
  }
}

# ============================================================================
# TRIGGERS
# ============================================================================

# New request webhook
trigger "new-request" {
  event: http("/api/requests") {
    method: "POST"
  }

  run: pipeline("process-request") {
    input: { request: http.body }
  }

  on_complete: {
    http.respond(201, {
      request_id: output.request_id,
      status: output.status,
      message: "Request created and submitted for approval"
    })
  }

  on_error: {
    http.respond(400, { error: error.message })

    log_audit_event(
      event_type: "request_failed",
      action: "create_request",
      outcome: "failure",
      details: { error: error.message }
    )
  }
}

# Approval decision webhook
trigger "approval-decision" {
  event: http("/api/approvals/{approval_id}") {
    method: "POST"
  }

  run: pipeline("handle-approval-decision") {
    input: {
      request_id: http.body.request_id,
      approval_id: http.params.approval_id,
      approver_id: http.headers["X-User-ID"],
      decision: http.body.decision,
      comments: http.body.comments
    }
  }

  on_complete: {
    http.respond(200, { status: "Decision recorded" })
  }
}

# SLA monitoring cron
trigger "sla-check" {
  event: schedule("*/15 * * * *")  # Every 15 minutes

  run: pipeline("sla-monitoring")

  on_complete: {
    if output.escalated > 0 {
      slack.post(
        channel: "#workflow-alerts",
        text: "SLA Alert: " + output.escalated + " requests escalated"
      )
    }
  }
}

# Daily audit report
trigger "daily-audit" {
  event: schedule("0 6 * * *")  # Daily at 6 AM

  run: {
    # Get yesterday's completed requests
    requests: db.query("""
      SELECT id FROM requests
      WHERE completed_at >= NOW() - INTERVAL '24 hours'
    """)

    # Generate audit reports
    for request in requests {
      script("generate-audit-report") {
        request_id: request.id
        include_pii: false
      }
    }
  }

  on_complete: {
    email.send(
      to: "compliance@company.com",
      subject: "Daily Workflow Audit Report",
      body: output.summary,
      attachments: output.reports
    )
  }
}

# ============================================================================
# CLI ENTRYPOINTS
# ============================================================================

intent "submit-expense" {
  params: {
    amount: number required "Expense amount"
    description: string required "Expense description"
    category: string required "Expense category"
    receipt: string optional "Receipt file path"
  }

  run: pipeline("process-request") {
    input: {
      request: {
        type: "expense",
        title: "Expense: " + params.category,
        amount: params.amount,
        description: params.description,
        requester_id: env("USER_ID"),
        attachments: params.receipt ? [params.receipt] : []
      }
    }
  }

  output: stdout
}

intent "check-status" {
  params: {
    request_id: string required "Request ID to check"
  }

  run: script("check-sla-status") {
    request_id: params.request_id
  }

  output: stdout
}

intent "audit-report" {
  params: {
    request_id: string required "Request ID"
    include_pii: bool optional false "Include PII in report"
  }

  run: script("generate-audit-report") {
    request_id: params.request_id
    include_pii: params.include_pii
  }

  output: file("audit-{{params.request_id}}.json")
}

intent "approve" {
  params: {
    request_id: string required
    decision: string required "approved, rejected, request_changes"
    comments: string optional
  }

  run: pipeline("handle-approval-decision") {
    input: {
      request_id: params.request_id,
      approver_id: env("USER_ID"),
      decision: params.decision,
      comments: params.comments
    }
  }

  output: stdout
}

intent "pending-approvals" {
  params: {
    user_id: string optional "User ID (defaults to current user)"
  }

  run: {
    db.query("""
      SELECT r.id, r.title, r.type, r.amount, a.due_date
      FROM approvals a
      JOIN requests r ON a.request_id = r.id
      WHERE a.approver_id = '{{params.user_id || env("USER_ID")}}'
      AND a.status = 'pending'
      ORDER BY a.due_date ASC
    """)
  }

  output: stdout
}
