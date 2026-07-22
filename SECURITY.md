# Security Policy

ShellHub takes security seriously. We are committed to coordinated vulnerability disclosure (CVD) as required by the [EU Cyber Resilience Act (CRA)][cra] and aligned with [ISO/IEC 29147][iso-29147].

## Scope

This policy covers all components of ShellHub, including but not limited to:

- ShellHub Server (API, SSH gateway, WebSocket proxy)
- ShellHub Agent
- ShellHub CLI
- ShellHub UI
- Official container images published under `shellhubio/`
- ShellHub Cloud (the hosted offering)

Out of scope: third-party dependencies (report those upstream), social engineering, denial-of-service against production infrastructure, and findings from automated scanners without a demonstrated impact.

## Reporting a Vulnerability

Do not publicly disclose a vulnerability until we have released a fix and published an advisory. To report:

1. **Preferred:** [open a GitHub Security Advisory][new-advisory], visible to project maintainers only. Do *not* submit a normal issue or pull request.
2. **Alternative:** email [security@shellhub.io][security-mail] with a description of the vulnerability, steps to reproduce, and any supporting material (PoC, logs, screenshots).
We will confirm receipt within **2 business days**.

## Response Timelines

After triage and severity assessment (using [CVSS v3.1][cvss] or later), we target the following timelines from confirmation to patch release:

| Severity | CVSS Score | Patch Target |
|----------|------------|--------------|
| Critical | 9.0–10.0   | 48 hours     |
| High     | 7.0–8.9    | 30 days      |
| Medium   | 4.0–6.9    | 90 days      |
| Low      | 0.1–3.9    | Next release |

These are targets, not guarantees. Complex issues may take longer; if so, we will communicate an updated timeline to the reporter.

## CVE Assignment

We assign a CVE identifier for every confirmed vulnerability. CVEs are requested through [GitHub as a CNA][github-cna], which issues them as part of the Security Advisory workflow. The CVE will be included in the published advisory.

## Security Advisories

All confirmed vulnerabilities are published as [GitHub Security Advisories][advisories] once a fix is available. Each advisory includes:

- A description of the vulnerability and its impact
- Affected versions
- The CVE identifier
- Remediation steps (fixed version, workarounds if any)

Users can subscribe to advisory notifications via GitHub's *Watch* menu on the repository.

## ENISA Notification

For actively exploited vulnerabilities, we will notify the designated [ENISA][enisa] single reporting platform within **24 hours** of becoming aware of active exploitation, provide a full vulnerability notification within **72 hours**, and submit a final report within **14 days**, as required by CRA Article 14.

Our internal incident response process is documented in [SECURITY_INCIDENT_RESPONSE.md][incident-response].

## Researcher Acknowledgment

We value the work of security researchers. Unless the reporter requests anonymity, we will:

- Credit the reporter by name (or handle) in the published advisory
- Acknowledge the contribution in the release changelog

## Safe Harbor

We consider security research conducted in accordance with this policy to be authorized and will not pursue legal action against researchers who:

- Act in good faith and follow this disclosure policy
- Avoid privacy violations, data destruction, and service disruption
- Do not access or modify data belonging to other users
- Report findings promptly and do not exploit them beyond what is necessary to demonstrate the issue

## Contact

- **Security reports:** [security@shellhub.io][security-mail]
- **General questions:** [contact@shellhub.io][general-mail]

[cra]: https://eur-lex.europa.eu/eli/reg/2024/2847
[iso-29147]: https://www.iso.org/standard/72311.html
[new-advisory]: https://github.com/shellhub-io/shellhub/security/advisories/new
[security-mail]: mailto:security@shellhub.io
[general-mail]: mailto:contact@shellhub.io
[cvss]: https://www.first.org/cvss/v3.1/specification-document
[github-cna]: https://docs.github.com/en/code-security/security-advisories/working-with-global-security-advisories-from-the-github-advisory-database/about-the-github-advisory-database
[advisories]: https://github.com/shellhub-io/shellhub/security/advisories
[enisa]: https://www.enisa.europa.eu/
[incident-response]: SECURITY_INCIDENT_RESPONSE.md
