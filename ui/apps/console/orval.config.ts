import { defineConfig } from "orval";

const input = process.env.OPENAPI_SPEC_PATH;
if (!input) {
  throw new Error(
    "OPENAPI_SPEC_PATH is not set; run `npm run generate -w @shellhub/console`.",
  );
}

export default defineConfig({
  shellhub: {
    input: {
      target: input,
    },
    output: {
      target: "./src/client/api.ts",
      schemas: "./src/client/model",
      client: "react-query",
      httpClient: "fetch",
      mode: "single",
      clean: true,
      override: {
        mutator: {
          path: "./src/api/customInstance.ts",
          name: "customInstance",
        },
        fetch: {
          includeHttpResponseReturnType: false,
        },
        query: {
          signal: true,
          useInvalidate: true,
          mutationInvalidates: [
            {
              onMutations: ["apiKeyCreate", "apiKeyUpdate", "apiKeyDelete"],
              invalidates: ["apiKeyList"],
            },
            {
              onMutations: [
                "createPublicKey",
                "updatePublicKey",
                "deletePublicKey",
              ],
              invalidates: ["getPublicKeys"],
            },
            {
              onMutations: [
                "createFirewallRule",
                "updateFirewallRule",
                "deleteFirewallRule",
              ],
              invalidates: ["getFirewallRules"],
            },
            {
              onMutations: [
                "createAccessPolicy",
                "updateAccessPolicy",
                "deleteAccessPolicy",
              ],
              invalidates: ["listAccessPolicies"],
            },
            {
              onMutations: ["createWebEndpoint", "deleteWebEndpoint"],
              invalidates: ["listWebEndpoints"],
            },
            {
              onMutations: ["installKeyCreate", "installKeyUpdate"],
              invalidates: ["installKeyList"],
            },
            {
              onMutations: ["createTag", "updateTag", "deleteTag"],
              invalidates: ["getTags", "getDevices"],
            },
            {
              onMutations: [
                "addNamespaceMember",
                "updateNamespaceMember",
                "removeNamespaceMember",
              ],
              invalidates: ["getNamespaces"],
            },
            {
              onMutations: ["approveUser"],
              invalidates: ["getNamespaces", "getUsers"],
            },
            {
              onMutations: ["createUserAdmin"],
              invalidates: ["getUsers"],
            },
            {
              onMutations: [
                "adminUpdateUser",
                "adminDeleteUser",
                "adminResetUserPassword",
              ],
              invalidates: ["getUsers"],
            },
            {
              onMutations: ["editNamespaceAdmin", "deleteNamespaceAdmin"],
              invalidates: ["getNamespacesAdmin"],
            },
            {
              onMutations: ["createAnnouncement"],
              invalidates: ["listAnnouncementsAdmin"],
            },
            {
              onMutations: ["updateAnnouncement", "deleteAnnouncement"],
              invalidates: ["listAnnouncementsAdmin"],
            },
            {
              onMutations: ["createServiceAccount", "deleteServiceAccount"],
              invalidates: ["listServiceAccounts", "listSshIdentities"],
            },
            {
              onMutations: ["acceptInvite"],
              invalidates: ["getMembershipInvitationList", "getNamespaces"],
            },
            {
              onMutations: ["generateInvitationLink"],
              invalidates: ["getNamespaces"],
            },
            {
              onMutations: ["cancelMembershipInvitation"],
              invalidates: ["getNamespaces"],
            },
            {
              onMutations: ["sendLicense"],
              invalidates: ["getLicense"],
            },
            {
              onMutations: [
                "acceptDevice",
                "deleteDevice",
                "acceptDevicePairing",
              ],
              invalidates: ["getDevices", "getStatusDevices", "getStats"],
            },
            {
              onMutations: [
                "updateDeviceStatus",
                "updateDevice",
                "pullTagFromDevice",
                "choiceDevices",
              ],
              invalidates: ["getDevices", "getStatusDevices"],
            },
            {
              onMutations: ["setDeviceCustomField", "deleteDeviceCustomField"],
              invalidates: ["getDevices"],
            },
            {
              onMutations: [
                "updateContainerStatus",
                "deleteContainer",
                "updateContainer",
                "pullTagFromContainer",
              ],
              invalidates: ["getContainers"],
            },
            {
              onMutations: ["clsoeSession"],
              invalidates: ["getSessions", "getStatusDevices"],
            },
            {
              onMutations: ["deleteSessionRecord"],
              invalidates: ["getSessions"],
            },
            {
              onMutations: [
                "confirmSshApproval",
                "createSshIdentity",
                "renameSshIdentity",
                "deleteSshIdentity",
              ],
              invalidates: ["listSshIdentities"],
            },
            {
              onMutations: ["editNamespace", "setSshAccessMode"],
              invalidates: ["getNamespaces"],
            },
            {
              onMutations: ["createInstanceAPIKey", "deleteInstanceAPIKey"],
              invalidates: ["listInstanceAPIKeys"],
            },
            {
              onMutations: [
                "createCustomer",
                "createSubscription",
                "attachPaymentMethod",
                "detachPaymentMethod",
                "setDefaultPaymentMethod",
              ],
              invalidates: ["getCustomer", "getSubscription"],
            },
          ],
        },
      },
      mock: {
        generators: [{ type: "msw" }, { type: "faker" }],
      },
    },
  },
});
