# ADR 0006: Official GitLab Go Client for API Access

**Status:** Accepted

The Go port will use GitLab REST through the official `gitlab.com/gitlab-org/api/client-go` library as the only GitLab API client. This preserves the REST-over-GraphQL direction from ADR-0002 while replacing the TypeScript-specific GitBeaker dependency with the official Go client; raw HTTP and alternative Go clients are intentionally avoided unless this decision is revisited.
