# LangSpace Advanced Example: Continuous Monitoring & Incident Response
# 24/7 system monitoring with intelligent incident detection, triage, and response.
#
# This example demonstrates:
# - Event-driven monitoring triggers
# - Intelligent alert correlation
# - Automated runbook execution
# - Escalation policies
# - Post-incident analysis
# - On-call management integration

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

  # Monitoring settings
  monitoring: {
    check_interval: "1m"
    alert_cooldown: "5m"
    correlation_window: "10m"
    max_incidents_per_hour: 50
  }

  # Severity definitions
  severity: {
    critical: {
      response_time: "5m"
      escalation_after: "15m"
      channels: ["pagerduty", "slack", "email"]
    }
    high: {
      response_time: "15m"
      escalation_after: "30m"
      channels: ["slack", "email"]
    }
    medium: {
      response_time: "1h"
      escalation_after: "2h"
      channels: ["slack"]
    }
    low: {
      response_time: "4h"
      escalation_after: "8h"
      channels: ["email"]
    }
  }

  # On-call schedule
  oncall: {
    schedule_id: env("PAGERDUTY_SCHEDULE_ID")
    escalation_policy: env("PAGERDUTY_POLICY_ID")
  }
}

# ============================================================================
# RUNBOOKS AND PROCEDURES
# ============================================================================

file "runbooks/high-cpu" {
  contents: ```
    # High CPU Usage Runbook

    ## Symptoms
    - CPU usage > 80% sustained for 5+ minutes
    - Slow response times
    - Process queuing

    ## Diagnostic Steps
    1. Identify top CPU-consuming processes: `top -b -n 1 -o %CPU | head -20`
    2. Check for runaway processes: `ps aux --sort=-%cpu | head -10`
    3. Review recent deployments
    4. Check for scheduled jobs running

    ## Remediation Steps

    ### If application process:
    1. Check application logs for errors
    2. Consider graceful restart if stuck
    3. Scale horizontally if under heavy load

    ### If system process:
    1. Identify if it's expected behavior (backups, updates)
    2. Check for malware if unknown process
    3. Investigate disk I/O correlation

    ### If database:
    1. Identify slow queries
    2. Check for table locks
    3. Review recent schema changes

    ## Escalation
    - After 15 minutes unresolved: Page on-call engineer
    - After 30 minutes: Page team lead
  ```
}

file "runbooks/high-memory" {
  contents: ```
    # High Memory Usage Runbook

    ## Symptoms
    - Memory usage > 85%
    - OOM killer activity
    - Swap usage increasing

    ## Diagnostic Steps
    1. Check memory by process: `ps aux --sort=-%mem | head -10`
    2. Review memory trends: `free -m && cat /proc/meminfo`
    3. Check for memory leaks in application logs
    4. Review heap dumps if available

    ## Remediation Steps

    ### Immediate (if critical):
    1. Clear caches: `sync; echo 3 > /proc/sys/vm/drop_caches`
    2. Restart largest non-critical process
    3. Scale horizontally if possible

    ### Investigation:
    1. Compare memory profile before/after recent deploys
    2. Check for memory leak patterns
    3. Review garbage collection logs

    ## Prevention
    - Set proper memory limits in container configs
    - Enable heap profiling for applications
    - Configure swap appropriately
  ```
}

file "runbooks/high-latency" {
  contents: ```
    # High Latency Runbook

    ## Symptoms
    - P99 latency > threshold
    - Increased timeout errors
    - User complaints

    ## Diagnostic Steps
    1. Identify slow endpoints from APM
    2. Check database query times
    3. Review external dependency latencies
    4. Check network connectivity

    ## Remediation Steps

    ### Quick Wins:
    1. Enable circuit breakers for slow dependencies
    2. Increase timeout if legitimate slow operation
    3. Enable caching if not already active

    ### Investigation:
    1. Profile slow requests end-to-end
    2. Check for N+1 queries
    3. Review recent code changes

    ## Escalation
    - Customer-impacting: Immediate page
    - Internal only: 30-minute investigation window
  ```
}

file "runbooks/disk-space" {
  contents: ```
    # Disk Space Runbook

    ## Symptoms
    - Disk usage > 80%
    - Write failures
    - Log rotation failures

    ## Diagnostic Steps
    1. Identify large directories: `du -h --max-depth=2 / | sort -hr | head -20`
    2. Find large files: `find / -size +100M -type f 2>/dev/null`
    3. Check log sizes: `ls -lhS /var/log/`
    4. Review docker/container storage

    ## Remediation Steps

    ### Safe to delete:
    - Old log files (> 7 days)
    - Temp files: `/tmp/*`, `/var/tmp/*`
    - Old container images
    - Package manager cache

    ### Requires review:
    - Application data
    - Database files
    - User uploads

    ## Automation Available
    - Log rotation: `logrotate -f /etc/logrotate.conf`
    - Docker cleanup: `docker system prune -af`
    - Package cache: `apt clean` or `yum clean all`
  ```
}

file "escalation-policy" {
  contents: ```
    # Escalation Policy

    ## Level 1: On-Call Engineer
    - First responder for all alerts
    - 5 minutes to acknowledge critical
    - 15 minutes to acknowledge high

    ## Level 2: Team Lead
    - Escalated after L1 timeout
    - Complex issues requiring coordination
    - Customer-impacting incidents

    ## Level 3: Engineering Manager
    - Major outages
    - Multi-team coordination needed
    - Executive communication required

    ## Level 4: VP Engineering / CTO
    - Company-wide impact
    - Security incidents
    - Data breach potential

    ## Special Escalations
    - Security issues: Direct to Security team + L3
    - Data issues: Direct to Data team + L2
    - Payment issues: Direct to Payments team + L2
  ```
}

# ============================================================================
# TOOLS
# ============================================================================

mcp "prometheus" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@company/mcp-prometheus"]
}

mcp "kubernetes" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@company/mcp-kubernetes"]
}

tool "query_metrics" {
  description: "Query Prometheus metrics"

  parameters: {
    query: string required "PromQL query"
    start: string optional "Start time (relative: -1h, or ISO)"
    end: string optional "End time"
    step: string optional "15s" "Step duration"
  }

  handler: http {
    method: "GET"
    url: env("PROMETHEUS_URL") + "/api/v1/query_range"
    query: {
      query: params.query,
      start: params.start,
      end: params.end,
      step: params.step
    }
  }
}

tool "get_logs" {
  description: "Retrieve logs from Loki or similar"

  parameters: {
    query: string required "LogQL query"
    start: string optional "-1h"
    limit: number optional 1000
  }

  handler: http {
    method: "GET"
    url: env("LOKI_URL") + "/loki/api/v1/query_range"
    query: params
  }
}

tool "execute_command" {
  description: "Execute a command on a remote host"

  parameters: {
    host: string required "Target host"
    command: string required "Command to execute"
    timeout: number optional 30 "Timeout in seconds"
  }

  handler: shell {
    command: "ssh {{params.host}} '{{params.command}}'"
    timeout: "{{params.timeout}}s"
  }
}

tool "restart_service" {
  description: "Restart a systemd service"

  parameters: {
    host: string required
    service: string required
  }

  handler: shell {
    command: "ssh {{params.host}} 'sudo systemctl restart {{params.service}}'"
    timeout: "60s"
  }
}

tool "scale_deployment" {
  description: "Scale a Kubernetes deployment"

  parameters: {
    deployment: string required
    namespace: string optional "default"
    replicas: number required
  }

  handler: mcp("kubernetes").scale_deployment
}

tool "page_oncall" {
  description: "Page the on-call engineer"

  parameters: {
    title: string required "Incident title"
    severity: string required "critical, high, medium, low"
    details: string required "Incident details"
    runbook: string optional "Link to runbook"
  }

  handler: http {
    method: "POST"
    url: "https://events.pagerduty.com/v2/enqueue"
    headers: {
      "Content-Type": "application/json"
    }
    body: {
      routing_key: env("PAGERDUTY_ROUTING_KEY"),
      event_action: "trigger",
      payload: {
        summary: params.title,
        severity: params.severity,
        source: "langspace-monitor",
        custom_details: {
          details: params.details,
          runbook: params.runbook
        }
      }
    }
  }
}

tool "create_incident" {
  description: "Create an incident in the incident management system"

  parameters: {
    title: string required
    severity: string required
    description: string required
    affected_services: array optional
    timeline: array optional "List of timeline events"
  }

  handler: http {
    method: "POST"
    url: env("INCIDENT_API_URL") + "/incidents"
    headers: {
      "Authorization": "Bearer " + env("INCIDENT_API_KEY")
    }
    body: params
  }
}

tool "update_incident" {
  description: "Update an existing incident"

  parameters: {
    incident_id: string required
    status: string optional "investigating, identified, monitoring, resolved"
    update: string required "Status update message"
  }

  handler: http {
    method: "PATCH"
    url: env("INCIDENT_API_URL") + "/incidents/" + params.incident_id
    headers: {
      "Authorization": "Bearer " + env("INCIDENT_API_KEY")
    }
    body: params
  }
}

# ============================================================================
# MONITORING SCRIPTS
# ============================================================================

# Efficient metric aggregation
script "aggregate-metrics" {
  language: "python"
  runtime: "python3"

  capabilities: [network]

  parameters: {
    queries: object required "Map of name -> PromQL query"
    duration: string optional "5m"
  }

  code: ```python
    import json
    import urllib.request
    import os

    queries = json.loads(queries) if isinstance(queries, str) else queries
    duration = duration or "5m"

    results = {}

    for name, query in queries.items():
        try:
            url = f"{os.environ['PROMETHEUS_URL']}/api/v1/query"
            req = urllib.request.Request(
                f"{url}?query={urllib.parse.quote(query)}",
                headers={"Accept": "application/json"}
            )
            response = urllib.request.urlopen(req, timeout=10)
            data = json.loads(response.read())

            if data["status"] == "success" and data["data"]["result"]:
                result = data["data"]["result"][0]
                results[name] = {
                    "value": float(result["value"][1]),
                    "labels": result["metric"]
                }
            else:
                results[name] = {"value": None, "error": "No data"}

        except Exception as e:
            results[name] = {"value": None, "error": str(e)}

    print(json.dumps(results, indent=2))
  ```
}

# Alert correlation script
script "correlate-alerts" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    new_alert: object required "The new alert to correlate"
    window_minutes: number optional 10
  }

  code: ```python
    import db
    import json
    from datetime import datetime, timedelta

    alert = json.loads(new_alert) if isinstance(new_alert, str) else new_alert
    window = window_minutes or 10

    cutoff = (datetime.now() - timedelta(minutes=window)).isoformat()

    # Find recent related alerts
    related = db.query(f"""
        SELECT * FROM alerts
        WHERE timestamp > '{cutoff}'
        AND (
            host = '{alert.get("host", "")}'
            OR service = '{alert.get("service", "")}'
            OR alert_name = '{alert.get("alert_name", "")}'
        )
        ORDER BY timestamp DESC
    """)

    # Group by potential incident
    correlations = {
        "same_host": [],
        "same_service": [],
        "same_type": [],
        "potential_root_cause": None
    }

    for r in related:
        if r["host"] == alert.get("host"):
            correlations["same_host"].append(r)
        if r["service"] == alert.get("service"):
            correlations["same_service"].append(r)
        if r["alert_name"] == alert.get("alert_name"):
            correlations["same_type"].append(r)

    # Identify potential root cause (first alert in chain)
    all_related = set()
    for group in [correlations["same_host"], correlations["same_service"]]:
        for a in group:
            all_related.add(a["id"])

    if all_related:
        first_alert = db.query_one(f"""
            SELECT * FROM alerts
            WHERE id IN ({",".join(map(str, all_related))})
            ORDER BY timestamp ASC
            LIMIT 1
        """)
        correlations["potential_root_cause"] = first_alert

    correlations["is_correlated"] = len(all_related) > 0
    correlations["related_count"] = len(all_related)

    print(json.dumps(correlations, indent=2, default=str))
  ```
}

# Runbook execution script
script "execute-runbook-step" {
  language: "python"
  runtime: "python3"

  capabilities: [network, filesystem.read]

  parameters: {
    step: object required "Runbook step to execute"
    context: object required "Incident context"
  }

  code: ```python
    import json
    import subprocess
    import urllib.request
    import os

    step = json.loads(step) if isinstance(step, str) else step
    context = json.loads(context) if isinstance(context, str) else context

    result = {
        "step": step.get("name"),
        "success": False,
        "output": None,
        "error": None
    }

    try:
        if step["type"] == "shell":
            # Execute shell command
            cmd = step["command"]
            # Substitute context variables
            for key, value in context.items():
                cmd = cmd.replace(f"${{{key}}}", str(value))

            proc = subprocess.run(
                cmd, shell=True, capture_output=True, text=True, timeout=60
            )
            result["output"] = proc.stdout
            result["success"] = proc.returncode == 0
            if proc.returncode != 0:
                result["error"] = proc.stderr

        elif step["type"] == "api":
            req = urllib.request.Request(
                step["url"],
                data=json.dumps(step.get("body", {})).encode(),
                headers=step.get("headers", {}),
                method=step.get("method", "GET")
            )
            response = urllib.request.urlopen(req, timeout=30)
            result["output"] = response.read().decode()
            result["success"] = True

        elif step["type"] == "check":
            # Verification step
            # Would implement actual check logic
            result["output"] = "Check executed"
            result["success"] = True

    except Exception as e:
        result["error"] = str(e)

    print(json.dumps(result, indent=2))
  ```
}

# ============================================================================
# INTELLIGENT AGENTS
# ============================================================================

agent "alert-analyzer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You analyze alerts and determine their severity, impact, and urgency.

    For each alert, determine:
    1. Is this a real issue or noise/false positive?
    2. What is the severity (critical, high, medium, low)?
    3. What services/users are affected?
    4. Is this related to other ongoing incidents?
    5. What is the likely root cause?

    Consider:
    - Time of day (off-hours may reduce impact)
    - Affected service criticality
    - Number of users/requests affected
    - Rate of degradation

    Output:
    {
      "is_actionable": true,
      "severity": "high",
      "affected_services": ["api", "database"],
      "estimated_users_affected": 1000,
      "likely_cause": "Database connection pool exhausted",
      "correlation_hypothesis": "Related to deploy 2 hours ago",
      "recommended_runbook": "high-latency"
    }
  ```

  tools: [
    tool("query_metrics"),
    tool("get_logs"),
  ]

  scripts: [
    script("aggregate-metrics"),
    script("correlate-alerts")
  ]
}

agent "incident-responder" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.1

  instruction: ```
    You are an incident responder who executes runbooks and coordinates resolution.

    Your responsibilities:
    1. Execute appropriate runbook steps
    2. Monitor the impact of actions taken
    3. Communicate status updates clearly
    4. Escalate when necessary
    5. Document all actions for post-mortem

    Decision framework:
    - If automated fix is available and safe: Execute it
    - If fix requires human judgment: Escalate with recommendation
    - If unsure about impact: Gather more data first
    - Always: Keep stakeholders informed

    After each action:
    - Verify the intended effect
    - Check for unintended consequences
    - Update incident timeline
  ```

  tools: [
    tool("execute_command"),
    tool("restart_service"),
    tool("scale_deployment"),
    tool("query_metrics"),
    tool("get_logs"),
    tool("update_incident"),
  ]

  scripts: [
    script("execute-runbook-step")
  ]
}

agent "incident-coordinator" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: file("escalation-policy")

  tools: [
    tool("page_oncall"),
    tool("create_incident"),
    tool("update_incident"),
  ]
}

agent "post-mortem-analyst" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.4

  instruction: ```
    You analyze resolved incidents and generate post-mortem reports.

    Post-mortem structure:
    1. Executive Summary
    2. Impact Assessment
       - Duration
       - Users affected
       - Revenue impact (if applicable)
    3. Timeline of Events
    4. Root Cause Analysis (5 Whys)
    5. What Went Well
    6. What Went Wrong
    7. Action Items
       - Immediate fixes
       - Medium-term improvements
       - Long-term prevention

    Principles:
    - Blameless analysis
    - Focus on systems, not individuals
    - Actionable recommendations
    - Measurable improvements
  ```

  tools: [
    tool("query_metrics"),
    tool("get_logs"),
  ]
}

# ============================================================================
# INCIDENT RESPONSE PIPELINE
# ============================================================================

pipeline "handle-alert" {
  # Step 1: Analyze the alert
  step "analyze" {
    use: agent("alert-analyzer")
    input: $input.alert

    instruction: "Analyze this alert and determine severity and impact."
  }

  # Step 2: Correlate with existing incidents
  step "correlate" {
    execute: script("correlate-alerts") {
      new_alert: $input.alert
      window_minutes: 10
    }
  }

  # Step 3: Decide action path
  branch step("correlate").output.is_correlated {
    true => step "add-to-incident" {
      tools: [tool("update_incident")]

      input: {
        incident_id: step("correlate").output.potential_root_cause.incident_id,
        update: "New related alert: " + $input.alert.summary
      }
    }

    false => step "create-incident" {
      use: agent("incident-coordinator")

      input: step("analyze").output

      instruction: "Create a new incident based on this analysis."
    }
  }

  # Step 4: Execute response (if actionable)
  branch step("analyze").output.is_actionable {
    true => step "respond" {
      use: agent("incident-responder")

      input: {
        analysis: step("analyze").output,
        runbook: file("runbooks/" + step("analyze").output.recommended_runbook)
      }

      instruction: "Execute the appropriate runbook to resolve this incident."
    }
  }

  # Step 5: Verify resolution
  step "verify" {
    execute: script("aggregate-metrics") {
      queries: {
        "error_rate": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m]))",
        "latency_p99": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))"
      }
    }
  }

  # Step 6: Update status
  step "update-status" {
    use: agent("incident-coordinator")

    input: {
      incident: $branch.output,
      verification: step("verify").output,
      response_actions: step("respond").output
    }

    instruction: "Update the incident status based on our response and verification."
  }

  output: {
    incident: $branch.output,
    analysis: step("analyze").output,
    response: step("respond").output,
    verification: step("verify").output
  }
}

# Post-incident analysis pipeline
pipeline "post-mortem" {
  step "gather-data" {
    parallel {
      step "metrics" {
        tools: [tool("query_metrics")]
        input: {
          query: "Various metrics from incident window"
        }
      }

      step "logs" {
        tools: [tool("get_logs")]
        input: {
          query: "Logs from affected services"
        }
      }

      step "timeline" {
        # Get incident timeline from incident system
        tools: [tool("get_incident_timeline")]
        input: $input.incident_id
      }
    }
  }

  step "analyze" {
    use: agent("post-mortem-analyst")

    input: {
      incident: $input,
      metrics: step("gather-data").metrics.output,
      logs: step("gather-data").logs.output,
      timeline: step("gather-data").timeline.output
    }
  }

  step "generate-report" {
    use: agent("post-mortem-analyst")
    input: step("analyze").output

    instruction: "Generate a complete post-mortem document."
  }

  step "create-action-items" {
    input: step("analyze").output.action_items

    # Create tickets for action items
    tools: [tool("create_ticket")]
  }

  output: {
    report: step("generate-report").output,
    action_items: step("create-action-items").output
  }
}

# ============================================================================
# TRIGGERS
# ============================================================================

# Prometheus AlertManager webhook
trigger "alertmanager-webhook" {
  event: http("/webhook/alertmanager") {
    method: "POST"
  }

  # Filter resolved alerts (only handle firing)
  filter: http.body.status == "firing"

  run: pipeline("handle-alert") {
    input: {
      alert: {
        name: http.body.alerts[0].labels.alertname,
        severity: http.body.alerts[0].labels.severity,
        summary: http.body.alerts[0].annotations.summary,
        host: http.body.alerts[0].labels.instance,
        service: http.body.alerts[0].labels.service,
        timestamp: http.body.alerts[0].startsAt
      }
    }
  }

  on_complete: {
    http.respond(200, { status: "processed" })
  }
}

# Scheduled health checks
trigger "health-check" {
  event: schedule("*/5 * * * *")  # Every 5 minutes

  run: {
    # Aggregate key metrics
    metrics: script("aggregate-metrics") {
      queries: {
        "cpu": "avg(rate(node_cpu_seconds_total{mode!='idle'}[5m])) * 100",
        "memory": "avg(node_memory_MemUsed_bytes / node_memory_MemTotal_bytes) * 100",
        "disk": "avg(node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100",
        "error_rate": "sum(rate(http_requests_total{status=~'5..'}[5m])) / sum(rate(http_requests_total[5m])) * 100"
      }
    }

    # Check thresholds
    checks: [
      { name: "cpu", value: metrics.cpu.value, threshold: 80 },
      { name: "memory", value: metrics.memory.value, threshold: 85 },
      { name: "disk", value: 100 - metrics.disk.value, threshold: 80 },
      { name: "error_rate", value: metrics.error_rate.value, threshold: 5 }
    ]

    # Alert if any threshold exceeded
    for check in checks {
      if check.value > check.threshold {
        pipeline("handle-alert") {
          input: {
            alert: {
              name: "high_" + check.name,
              severity: check.value > check.threshold * 1.2 ? "critical" : "high",
              summary: check.name + " is at " + check.value + "%"
            }
          }
        }
      }
    }
  }
}

# Incident resolution trigger
trigger "incident-resolved" {
  event: webhook("/webhook/incident-resolved") {
    method: "POST"
  }

  run: pipeline("post-mortem") {
    input: http.body
  }

  on_complete: {
    slack.post(
      channel: "#incidents",
      text: "Post-mortem ready for incident " + http.body.incident_id,
      attachments: [{ text: output.report.summary }]
    )
  }
}

# ============================================================================
# CLI ENTRYPOINTS
# ============================================================================

intent "check-health" {
  run: script("aggregate-metrics") {
    queries: {
      "cpu": "avg(rate(node_cpu_seconds_total{mode!='idle'}[5m])) * 100",
      "memory": "avg(node_memory_MemUsed_bytes / node_memory_MemTotal_bytes) * 100",
      "disk": "avg(1 - node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100",
      "error_rate": "sum(rate(http_requests_total{status=~'5..'}[5m])) / sum(rate(http_requests_total[5m])) * 100"
    }
  }

  output: stdout
}

intent "analyze-alert" {
  params: {
    alert_name: string required "Name of the alert"
    severity: string optional "high"
  }

  use: agent("alert-analyzer")
  input: params

  output: stdout
}

intent "run-runbook" {
  params: {
    runbook: string required "Runbook name"
    host: string required "Target host"
    dry_run: bool optional true "Dry run mode"
  }

  use: agent("incident-responder")

  input: {
    runbook: file("runbooks/" + params.runbook),
    context: { host: params.host },
    dry_run: params.dry_run
  }

  output: stdout
}

intent "generate-postmortem" {
  params: {
    incident_id: string required "Incident ID"
  }

  run: pipeline("post-mortem") {
    input: { incident_id: params.incident_id }
  }

  output: file("postmortems/{{params.incident_id}}.md")
}

intent "page" {
  params: {
    message: string required "Alert message"
    severity: string optional "high"
  }

  tools: [tool("page_oncall")]

  input: {
    title: params.message,
    severity: params.severity,
    details: "Manual page from CLI"
  }

  output: stdout
}
