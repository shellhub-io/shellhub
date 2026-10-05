// Home page (info.description) for the customer docs. The full spec keeps its
// own internal-facing description; this one is injected only into the filtered
// build.
const CUSTOMER_DESCRIPTION = `Programmatic access to your namespace: devices, sessions, SSH public keys,
tags, firewall rules, and more.

## Base URL

Endpoints are served under \`/api\` on your ShellHub server. On ShellHub Cloud
the base URL is \`https://cloud.shellhub.io/api\`.

## Authentication

Send your API key in the \`X-API-KEY\` header:

\`\`\`
curl https://cloud.shellhub.io/api/devices -H "X-API-KEY: <your-key>"
\`\`\`

An API key belongs to a single namespace and is not tied to a user. Create one
in the console under **Namespace → API Keys**. Every endpoint operates within
the key's namespace, so the namespace scope is implicit throughout this
reference.

## Pagination

List endpoints accept \`page\` and \`per_page\` query parameters and return the
total item count in the \`X-Total-Count\` response header.

## Errors

Errors use standard HTTP status codes. \`401\` means the API key is missing or
invalid; \`403\` means the key's role does not allow the operation.`;

const HTTP_METHODS = [
  'get',
  'put',
  'post',
  'delete',
  'options',
  'head',
  'patch',
  'trace',
];

const INTERNAL_PREFIXES = ['/admin', '/internal'];

// Edition and audience markers used as tags across the specs. They are not
// resources, so they are stripped from the customer docs to leave a clean
// resource-based grouping (devices, sessions, rules, ...).
const NON_RESOURCE_TAGS = new Set([
  'community',
  'cloud',
  'enterprise',
  'internal',
  'external',
]);

function operationAcceptsApiKey(operation) {
  const security = operation.security;

  return (
    Array.isArray(security) &&
    security.some(
      (requirement) =>
        requirement &&
        Object.prototype.hasOwnProperty.call(requirement, 'api-key'),
    )
  );
}

function isCustomerOperation(operation) {
  return operation['x-internal'] !== true && operationAcceptsApiKey(operation);
}

function hasAnyOperation(pathItem) {
  return HTTP_METHODS.some((method) => pathItem[method] !== undefined);
}

function DropNonCustomer() {
  return {
    PathItem: {
      enter(pathItem, ctx) {
        const path = ctx.key;
        const isInternalPath =
          typeof path === 'string' &&
          INTERNAL_PREFIXES.some((prefix) => path.startsWith(prefix));

        for (const method of HTTP_METHODS) {
          const operation = pathItem[method];

          if (operation === undefined) {
            continue;
          }

          if (isInternalPath || !isCustomerOperation(operation)) {
            delete pathItem[method];

            continue;
          }

          if (Array.isArray(operation.tags)) {
            operation.tags = operation.tags.filter(
              (tag) => !NON_RESOURCE_TAGS.has(tag),
            );
          }

          if (Array.isArray(operation.security)) {
            operation.security = operation.security.filter(
              (requirement) =>
                requirement &&
                Object.prototype.hasOwnProperty.call(requirement, 'api-key'),
            );
          }
        }
      },
    },
    Root: {
      leave(root) {
        if (!root.paths) {
          return;
        }

        for (const path of Object.keys(root.paths)) {
          if (!hasAnyOperation(root.paths[path])) {
            delete root.paths[path];
          }
        }

        if (Array.isArray(root.tags)) {
          const usedTags = new Set();

          for (const pathItem of Object.values(root.paths)) {
            for (const method of HTTP_METHODS) {
              const operation = pathItem[method];

              if (operation && Array.isArray(operation.tags)) {
                operation.tags.forEach((tag) => usedTags.add(tag));
              }
            }
          }

          const seenTags = new Set();
          root.tags = root.tags.filter((tag) => {
            if (!usedTags.has(tag.name) || seenTags.has(tag.name)) {
              return false;
            }

            seenTags.add(tag.name);

            return true;
          });
        }

        if (root.components && root.components.securitySchemes) {
          delete root.components.securitySchemes.jwt;
        }

        if (root.info) {
          root.info.description = CUSTOMER_DESCRIPTION;
        }
      },
    },
  };
}

function customerFilterPlugin() {
  return {
    id: 'customer-filter',
    decorators: {
      oas3: {
        'drop-non-customer': DropNonCustomer,
      },
    },
  };
}

module.exports = customerFilterPlugin;
