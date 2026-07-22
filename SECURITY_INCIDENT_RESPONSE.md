# Security Incident Response

This document describes ShellHub's internal process for handling security vulnerabilities, aligned with [CRA Article 14][cra] reporting obligations to [ENISA][enisa].

## Workflow

### 1. Report Received

- Acknowledge receipt to the reporter within **2 business days**
- Open a private GitHub Security Advisory draft
- Assign a Triage Lead

### 2. Triage and Severity Assessment

- Reproduce the vulnerability
- Assess severity using [CVSS v3.1][cvss] (or later)
- Determine whether the vulnerability is actively exploited

If **not** actively exploited, proceed to step 4.

If **actively exploited**, proceed to step 3 immediately.

### 3. ENISA Reporting (Actively Exploited Only)

CRA Article 14 requires three reports to ENISA via the [single reporting platform][enisa-srp]:

| Deadline | Report | Contents |
|----------|--------|----------|
| **24 hours** | Early warning | That the vulnerability exists, whether it is actively exploited, and whether it affects other products |
| **72 hours** | Vulnerability notification | Severity assessment, impact analysis, affected versions, remediation status, and any indicators of compromise |
| **14 days** | Final report | Root cause analysis, fix details, scope of exploitation (if known), and lessons learned |

All deadlines are counted from the moment the team becomes aware of active exploitation.

### 4. Patch Development

Develop a fix following the patch SLA defined in [SECURITY.md](SECURITY.md):

| Severity | Patch Target |
|----------|-------------|
| Critical (9.0–10.0) | 48 hours |
| High (7.0–8.9) | 30 days |
| Medium (4.0–6.9) | 90 days |
| Low (0.1–3.9) | Next release |

- Develop and review the patch in a private fork or branch
- If the reporter provided a fix, review and integrate it

### 5. Advisory and Release

- Assign a CVE via GitHub's CNA integration
- Publish the GitHub Security Advisory with: description, affected versions, CVE, fixed version, and workarounds (if any)
- Release the patched version
- Credit the reporter (unless they request anonymity)
- Update the changelog

### 6. Post-Incident

- Notify downstream users and integrators
- Conduct an internal retrospective for Critical and High severity incidents
- Update this process if gaps were identified

[cra]: https://eur-lex.europa.eu/eli/reg/2024/2847
[enisa]: https://www.enisa.europa.eu/
[enisa-srp]: https://www.enisa.europa.eu/topics/vulnerability-reporting
[cvss]: https://www.first.org/cvss/v3.1/specification-document
