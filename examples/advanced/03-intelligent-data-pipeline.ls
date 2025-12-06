# LangSpace Advanced Example: Intelligent Data Pipeline
# AI-powered ETL with anomaly detection, auto-remediation, and observability.
#
# This example demonstrates:
# - Script-first efficiency for data operations
# - Anomaly detection and pattern recognition
# - Auto-remediation with human escalation
# - Multi-database support
# - Comprehensive alerting and monitoring

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

  # Data pipeline settings
  pipeline: {
    batch_size: 10000
    parallelism: 4
    retry_attempts: 3
    anomaly_threshold: 2.5  # Standard deviations
  }

  # Database connections
  databases: {
    source: {
      type: "postgresql"
      connection: env("SOURCE_DB_URL")
    }
    warehouse: {
      type: "snowflake"
      connection: env("WAREHOUSE_DB_URL")
    }
    cache: {
      type: "redis"
      connection: env("REDIS_URL")
    }
  }

  # Alerting
  alerts: {
    slack_webhook: env("SLACK_WEBHOOK_URL")
    pagerduty_key: env("PAGERDUTY_KEY")
    email_recipients: ["data-team@company.com"]
  }
}

# ============================================================================
# DATA QUALITY RULES
# ============================================================================

file "quality-rules" {
  contents: ```
    # Data Quality Rules

    ## Completeness
    - Required fields must not be null
    - Foreign keys must reference valid records
    - Time series data must not have gaps > 1 hour

    ## Accuracy
    - Numeric values within expected ranges
    - Dates not in the future (except scheduled items)
    - Email formats validated
    - Phone numbers normalized

    ## Consistency
    - Currency codes match ISO 4217
    - Country codes match ISO 3166
    - Status values in allowed enums
    - Cross-table totals match

    ## Timeliness
    - Data arrives within SLA window
    - Timestamps in expected timezone
    - No duplicate processing

    ## Uniqueness
    - Primary keys are unique
    - Business keys deduplicated
    - Reference data normalized
  ```
}

file "remediation-playbook" {
  contents: ```
    # Data Remediation Playbook

    ## NULL Value Remediation
    - Look up from related tables
    - Use default values from schema
    - Flag for manual review if critical

    ## Format Normalization
    - Phone: E.164 format (+1234567890)
    - Dates: ISO 8601 (YYYY-MM-DDTHH:MM:SSZ)
    - Currency: Decimal with 2 places
    - Addresses: USPS/postal service standards

    ## Duplicate Handling
    - Keep record with most recent timestamp
    - Merge non-conflicting fields
    - Log merge decisions for audit

    ## Orphan Records
    - Create placeholder parent if allowed
    - Move to quarantine table
    - Alert for manual resolution

    ## Range Violations
    - Clamp to valid range with flag
    - Reject and alert if extreme
    - Log original value
  ```
}

# ============================================================================
# DATABASE SCRIPTS (Context-Efficient)
# ============================================================================

# Main extraction script - runs outside context window
script "extract-batch" {
  language: "sql"
  runtime: "postgresql"

  capabilities: [database.read]

  parameters: {
    table: string required
    batch_size: number optional 10000
    offset: number optional 0
    since: string optional "Extract records modified since"
  }

  code: ```sql
    SELECT *
    FROM {{table}}
    WHERE ($since IS NULL OR modified_at > $since::timestamp)
    ORDER BY id
    LIMIT $batch_size
    OFFSET $offset;
  ```
}

# Data profiling script
script "profile-data" {
  language: "python"
  runtime: "python3"

  capabilities: [database.read]

  parameters: {
    table: string required
    columns: array optional "Specific columns to profile"
  }

  code: ```python
    import db
    import json
    from collections import Counter
    from statistics import mean, stdev

    # Get sample data
    data = db.query(f"SELECT * FROM {table} LIMIT 100000")

    profile = {
        "table": table,
        "row_count": len(data),
        "columns": {}
    }

    columns_to_check = columns if columns else data[0].keys() if data else []

    for col in columns_to_check:
        values = [row[col] for row in data if row.get(col) is not None]
        null_count = len(data) - len(values)

        col_profile = {
            "null_count": null_count,
            "null_pct": round(null_count / len(data) * 100, 2) if data else 0,
            "distinct_count": len(set(str(v) for v in values))
        }

        # Numeric analysis
        numeric_values = [v for v in values if isinstance(v, (int, float))]
        if numeric_values:
            col_profile.update({
                "min": min(numeric_values),
                "max": max(numeric_values),
                "mean": round(mean(numeric_values), 2),
                "stdev": round(stdev(numeric_values), 2) if len(numeric_values) > 1 else 0
            })

        # Cardinality check
        if len(set(str(v) for v in values)) < 20:
            col_profile["value_distribution"] = dict(Counter(str(v) for v in values))

        profile["columns"][col] = col_profile

    print(json.dumps(profile, indent=2, default=str))
  ```
}

# Quality check script
script "check-quality" {
  language: "python"
  runtime: "python3"

  capabilities: [database.read]

  parameters: {
    table: string required
    rules: object required "Quality rules to check"
  }

  code: ```python
    import db
    import json

    rules = json.loads(rules) if isinstance(rules, str) else rules
    results = {"passed": [], "failed": [], "warnings": []}

    # NULL checks
    for field in rules.get("required_fields", []):
        null_count = db.query_scalar(
            f"SELECT COUNT(*) FROM {table} WHERE {field} IS NULL"
        )
        if null_count > 0:
            results["failed"].append({
                "rule": f"required_field:{field}",
                "count": null_count,
                "severity": "high"
            })
        else:
            results["passed"].append(f"required_field:{field}")

    # Range checks
    for field, bounds in rules.get("ranges", {}).items():
        violations = db.query_scalar(f"""
            SELECT COUNT(*) FROM {table}
            WHERE {field} < {bounds['min']} OR {field} > {bounds['max']}
        """)
        if violations > 0:
            results["failed"].append({
                "rule": f"range:{field}",
                "count": violations,
                "bounds": bounds,
                "severity": "medium"
            })

    # Uniqueness checks
    for field in rules.get("unique_fields", []):
        duplicates = db.query_scalar(f"""
            SELECT COUNT(*) FROM (
                SELECT {field}, COUNT(*) c FROM {table}
                GROUP BY {field} HAVING COUNT(*) > 1
            ) t
        """)
        if duplicates > 0:
            results["failed"].append({
                "rule": f"unique:{field}",
                "count": duplicates,
                "severity": "high"
            })

    # Foreign key checks
    for fk in rules.get("foreign_keys", []):
        orphans = db.query_scalar(f"""
            SELECT COUNT(*) FROM {table} t
            LEFT JOIN {fk['ref_table']} r ON t.{fk['field']} = r.{fk['ref_field']}
            WHERE t.{fk['field']} IS NOT NULL AND r.{fk['ref_field']} IS NULL
        """)
        if orphans > 0:
            results["failed"].append({
                "rule": f"fk:{fk['field']}",
                "count": orphans,
                "severity": "high"
            })

    # Summary
    results["summary"] = {
        "total_rules": len(results["passed"]) + len(results["failed"]),
        "passed": len(results["passed"]),
        "failed": len(results["failed"]),
        "pass_rate": round(len(results["passed"]) / max(1, len(results["passed"]) + len(results["failed"])) * 100, 2)
    }

    print(json.dumps(results, indent=2))
  ```
}

# Anomaly detection script
script "detect-anomalies" {
  language: "python"
  runtime: "python3"

  capabilities: [database.read]

  parameters: {
    table: string required
    metric_columns: array required
    time_column: string required
    threshold: number optional 2.5 "Standard deviations for anomaly"
    lookback_days: number optional 30
  }

  code: ```python
    import db
    import json
    from datetime import datetime, timedelta
    from statistics import mean, stdev

    threshold = threshold or 2.5
    lookback = lookback_days or 30
    cutoff = (datetime.now() - timedelta(days=lookback)).isoformat()

    anomalies = []

    for col in metric_columns:
        # Get historical stats
        historical = db.query(f"""
            SELECT {col} as value
            FROM {table}
            WHERE {time_column} >= '{cutoff}'
            AND {col} IS NOT NULL
        """)

        values = [r['value'] for r in historical if isinstance(r['value'], (int, float))]

        if len(values) < 10:
            continue

        avg = mean(values)
        std = stdev(values)

        if std == 0:
            continue

        # Check recent values
        recent = db.query(f"""
            SELECT id, {col} as value, {time_column} as timestamp
            FROM {table}
            WHERE {time_column} >= NOW() - INTERVAL '1 hour'
            AND {col} IS NOT NULL
        """)

        for row in recent:
            z_score = (row['value'] - avg) / std
            if abs(z_score) > threshold:
                anomalies.append({
                    "column": col,
                    "record_id": row['id'],
                    "timestamp": str(row['timestamp']),
                    "value": row['value'],
                    "expected_range": [avg - threshold * std, avg + threshold * std],
                    "z_score": round(z_score, 2),
                    "severity": "high" if abs(z_score) > 4 else "medium"
                })

    print(json.dumps({
        "anomaly_count": len(anomalies),
        "anomalies": anomalies
    }, indent=2))
  ```
}

# Batch transformation script
script "transform-batch" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    source_table: string required
    target_table: string required
    transformations: object required
    batch_size: number optional 5000
  }

  code: ```python
    import db
    import json
    from datetime import datetime

    transforms = json.loads(transformations) if isinstance(transformations, str) else transformations

    # Get batch
    records = db.query(f"SELECT * FROM {source_table} LIMIT {batch_size}")

    transformed = []
    errors = []

    for record in records:
        try:
            new_record = {}

            for target_col, transform in transforms.items():
                if transform["type"] == "copy":
                    new_record[target_col] = record.get(transform["source"])

                elif transform["type"] == "concat":
                    new_record[target_col] = " ".join(
                        str(record.get(f, "")) for f in transform["fields"]
                    ).strip()

                elif transform["type"] == "lookup":
                    lookup_val = db.query_scalar(f"""
                        SELECT {transform['return_field']}
                        FROM {transform['table']}
                        WHERE {transform['match_field']} = '{record.get(transform['source'])}'
                    """)
                    new_record[target_col] = lookup_val or transform.get("default")

                elif transform["type"] == "calculate":
                    # Simple expression evaluation
                    expr = transform["expression"]
                    for field in transform.get("fields", []):
                        expr = expr.replace(f"${field}", str(record.get(field, 0)))
                    new_record[target_col] = eval(expr)

                elif transform["type"] == "default":
                    new_record[target_col] = record.get(transform["source"]) or transform["value"]

            new_record["_source_id"] = record.get("id")
            new_record["_processed_at"] = datetime.now().isoformat()

            transformed.append(new_record)

        except Exception as e:
            errors.append({
                "source_id": record.get("id"),
                "error": str(e)
            })

    # Insert transformed records
    if transformed:
        db.bulk_insert(target_table, transformed)

    print(f"Processed: {len(transformed)} records")
    print(f"Errors: {len(errors)} records")
    if errors:
        print("Error details:", json.dumps(errors[:10], indent=2))
  ```
}

# Auto-remediation script
script "remediate-issues" {
  language: "python"
  runtime: "python3"

  capabilities: [database]

  parameters: {
    table: string required
    issues: array required "List of issues to remediate"
    audit_table: string optional "audit_log"
  }

  code: ```python
    import db
    import json
    from datetime import datetime

    issues = json.loads(issues) if isinstance(issues, str) else issues
    results = {"fixed": [], "failed": [], "escalated": []}

    for issue in issues:
        try:
            if issue["type"] == "null_field":
                # Try to fill from related data
                if issue.get("lookup_table"):
                    value = db.query_scalar(f"""
                        SELECT {issue['field']}
                        FROM {issue['lookup_table']}
                        WHERE id = (SELECT {issue['lookup_key']} FROM {table} WHERE id = {issue['record_id']})
                    """)
                    if value:
                        db.execute(f"""
                            UPDATE {table} SET {issue['field']} = '{value}'
                            WHERE id = {issue['record_id']}
                        """)
                        results["fixed"].append(issue)
                        continue

                # Use default if available
                if issue.get("default"):
                    db.execute(f"""
                        UPDATE {table} SET {issue['field']} = '{issue['default']}'
                        WHERE id = {issue['record_id']}
                    """)
                    results["fixed"].append(issue)
                else:
                    results["escalated"].append(issue)

            elif issue["type"] == "duplicate":
                # Keep most recent, mark others
                db.execute(f"""
                    UPDATE {table} SET _status = 'duplicate'
                    WHERE {issue['key_field']} = '{issue['key_value']}'
                    AND id != (
                        SELECT id FROM {table}
                        WHERE {issue['key_field']} = '{issue['key_value']}'
                        ORDER BY created_at DESC LIMIT 1
                    )
                """)
                results["fixed"].append(issue)

            elif issue["type"] == "range_violation":
                if issue.get("clamp", False):
                    db.execute(f"""
                        UPDATE {table}
                        SET {issue['field']} = LEAST(GREATEST({issue['field']}, {issue['min']}), {issue['max']})
                        WHERE id = {issue['record_id']}
                    """)
                    results["fixed"].append(issue)
                else:
                    results["escalated"].append(issue)

            elif issue["type"] == "orphan":
                # Move to quarantine
                db.execute(f"""
                    INSERT INTO quarantine_{table} SELECT * FROM {table} WHERE id = {issue['record_id']};
                    DELETE FROM {table} WHERE id = {issue['record_id']};
                """)
                results["fixed"].append(issue)

        except Exception as e:
            issue["error"] = str(e)
            results["failed"].append(issue)

    # Log to audit table
    if audit_table:
        db.insert(audit_table, {
            "timestamp": datetime.now().isoformat(),
            "action": "auto_remediation",
            "details": json.dumps(results)
        })

    print(f"Fixed: {len(results['fixed'])}")
    print(f"Failed: {len(results['failed'])}")
    print(f"Escalated: {len(results['escalated'])}")
  ```
}

# ============================================================================
# INTELLIGENT AGENTS
# ============================================================================

agent "data-analyst" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.2

  instruction: ```
    You are a data quality analyst. You interpret data profiling results
    and quality check outputs to identify issues and their severity.

    Your responsibilities:
    1. Interpret statistical profiles and identify anomalies
    2. Categorize issues by severity and business impact
    3. Recommend remediation strategies
    4. Identify patterns that suggest systemic problems
    5. Generate clear reports for stakeholders

    When analyzing data:
    - Consider business context
    - Look for correlations between issues
    - Distinguish data errors from valid edge cases
    - Prioritize issues that affect downstream systems
  ```

  scripts: [
    script("profile-data"),
    script("check-quality"),
    script("detect-anomalies")
  ]
}

agent "remediation-specialist" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.1

  instruction: file("remediation-playbook")

  scripts: [
    script("remediate-issues"),
    script("transform-batch")
  ]
}

agent "alerting-coordinator" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: ```
    You manage alerting and escalation for data pipeline issues.

    Severity levels:
    - P1 (Critical): Data loss, corruption, or complete pipeline failure
    - P2 (High): Significant data quality issues affecting reports
    - P3 (Medium): Quality issues within tolerance but trending worse
    - P4 (Low): Minor issues for tracking, no immediate action

    Escalation paths:
    - P1: Immediate PagerDuty + Slack + Email
    - P2: Slack + Email within 15 minutes
    - P3: Slack + Daily digest
    - P4: Weekly report only

    For each alert:
    1. Determine severity based on impact
    2. Select appropriate channels
    3. Compose clear, actionable message
    4. Include relevant context and links
  ```
}

# ============================================================================
# ETL PIPELINE
# ============================================================================

pipeline "data-sync" {
  # Step 1: Extract from source
  step "extract" {
    execute: script("extract-batch") {
      table: $input.source_table
      batch_size: 10000
      since: $input.last_sync
    }
  }

  # Step 2: Profile and validate
  parallel {
    step "profile" {
      execute: script("profile-data") {
        table: $input.source_table
      }
    }

    step "validate" {
      execute: script("check-quality") {
        table: $input.source_table
        rules: $input.quality_rules
      }
    }
  }

  # Step 3: Analyze quality results
  step "analyze" {
    use: agent("data-analyst")

    input: [
      step("profile").output,
      step("validate").output
    ]

    instruction: "Analyze data quality and identify issues requiring attention."
  }

  # Step 4: Handle issues
  branch step("analyze").output.has_issues {
    true => step "remediate" {
      use: agent("remediation-specialist")
      input: step("analyze").output.issues

      instruction: "Remediate the identified data quality issues."
    }
  }

  # Step 5: Check for anomalies
  step "anomaly-check" {
    execute: script("detect-anomalies") {
      table: $input.source_table
      metric_columns: $input.metric_columns
      time_column: "created_at"
      threshold: 2.5
    }
  }

  # Step 6: Transform and load
  step "transform" {
    execute: script("transform-batch") {
      source_table: $input.source_table
      target_table: $input.target_table
      transformations: $input.transform_spec
    }
  }

  # Step 7: Alert on issues
  step "alert" {
    use: agent("alerting-coordinator")

    input: [
      step("analyze").output,
      step("anomaly-check").output,
      step("remediate").output
    ]

    instruction: "Generate alerts for any issues requiring attention."
  }

  output: {
    records_processed: step("transform").output.count,
    quality_report: step("analyze").output,
    anomalies: step("anomaly-check").output,
    alerts_sent: step("alert").output
  }
}

# Real-time anomaly monitoring
pipeline "monitor-metrics" {
  step "detect" {
    execute: script("detect-anomalies") {
      table: $input.table
      metric_columns: $input.metrics
      time_column: $input.time_column
      threshold: $input.threshold
    }
  }

  step "evaluate" {
    use: agent("data-analyst")
    input: step("detect").output

    instruction: "Evaluate these anomalies and determine which require immediate action."
  }

  branch step("evaluate").output.requires_action {
    true => step "alert" {
      use: agent("alerting-coordinator")
      input: step("evaluate").output

      instruction: "Generate and send appropriate alerts."
    }
  }

  output: step("evaluate").output
}

# ============================================================================
# TRIGGERS
# ============================================================================

# Scheduled full sync
trigger "scheduled-sync" {
  event: schedule("0 2 * * *")  # 2 AM daily

  run: pipeline("data-sync") {
    input: {
      source_table: "production.events",
      target_table: "warehouse.events",
      quality_rules: file("quality-rules"),
      transform_spec: file("transform-spec.json"),
      metric_columns: ["amount", "duration", "count"],
      last_sync: cache.get("last_sync_time")
    }
  }

  on_complete: {
    cache.set("last_sync_time", now())
    slack.post(
      channel: "#data-ops",
      text: "Daily sync complete: " + output.records_processed + " records processed"
    )
  }

  on_error: {
    pagerduty.trigger(
      severity: "critical",
      summary: "Daily data sync failed",
      details: error
    )
  }
}

# Real-time monitoring
trigger "metric-monitor" {
  event: schedule("*/5 * * * *")  # Every 5 minutes

  run: pipeline("monitor-metrics") {
    input: {
      table: "production.metrics",
      metrics: ["response_time", "error_rate", "throughput"],
      time_column: "timestamp",
      threshold: 3.0
    }
  }

  on_complete: {
    if output.anomaly_count > 0 {
      slack.post(
        channel: "#alerts",
        text: "⚠️ " + output.anomaly_count + " anomalies detected",
        attachments: output.anomalies
      )
    }
  }
}

# Webhook for external triggers
trigger "external-sync" {
  event: webhook("/api/sync") {
    method: "POST"
    auth: bearer(env("WEBHOOK_SECRET"))
  }

  run: pipeline("data-sync") {
    input: webhook.body
  }

  on_complete: {
    webhook.respond(200, output)
  }
}

# ============================================================================
# CLI ENTRYPOINTS
# ============================================================================

intent "sync" {
  params: {
    source: string required "Source table"
    target: string required "Target table"
    full: bool optional false "Full sync (ignore last sync time)"
  }

  run: pipeline("data-sync") {
    input: {
      source_table: params.source,
      target_table: params.target,
      last_sync: params.full ? null : cache.get("last_sync_time")
    }
  }
}

intent "profile" {
  params: {
    table: string required "Table to profile"
  }

  execute: script("profile-data") {
    table: params.table
  }

  output: stdout
}

intent "check-quality" {
  params: {
    table: string required "Table to check"
    rules: string optional "quality-rules" "Quality rules file"
  }

  execute: script("check-quality") {
    table: params.table
    rules: file(params.rules)
  }

  output: stdout
}

intent "detect-anomalies" {
  params: {
    table: string required
    columns: array required "Metric columns to check"
    threshold: number optional 2.5
  }

  execute: script("detect-anomalies") {
    table: params.table
    metric_columns: params.columns
    time_column: "created_at"
    threshold: params.threshold
  }

  output: stdout
}
