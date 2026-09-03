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
              invalidates: ["getTags", "getDevices", "getDevice"],
            },
            {
              onMutations: [
                "addNamespaceMember",
                "updateNamespaceMember",
                "removeNamespaceMember",
              ],
              invalidates: [
                "getNamespaces",
                "getNamespace",
                "listNamespaceMembers",
              ],
            },
            {
              onMutations: ["approveUser"],
              invalidates: [
                "getNamespaces",
                "getNamespace",
                "listNamespaceMembers",
                "getUsers",
                "getUser",
              ],
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
              invalidates: ["getUsers", "getUser"],
            },
            {
              onMutations: ["editNamespaceAdmin", "deleteNamespaceAdmin"],
              invalidates: ["getNamespacesAdmin", "getNamespaceAdmin"],
            },
            {
              onMutations: ["createAnnouncement"],
              invalidates: ["listAnnouncementsAdmin"],
            },
            {
              onMutations: ["updateAnnouncement", "deleteAnnouncement"],
              invalidates: ["listAnnouncementsAdmin", "getAnnouncementAdmin"],
            },
            {
              onMutations: ["createServiceAccount", "deleteServiceAccount"],
              invalidates: ["listServiceAccounts", "listSshIdentities"],
            },
            {
              onMutations: ["acceptInvite"],
              invalidates: [
                "getMembershipInvitationList",
                "getNamespace",
                "getNamespaces",
              ],
            },
            {
              onMutations: ["generateInvitationLink"],
              invalidates: [
                "getNamespaceMembershipInvitationList",
                "listNamespaceMembers",
                "getNamespace",
                "getNamespaces",
              ],
            },
            {
              onMutations: ["cancelMembershipInvitation"],
              invalidates: ["getNamespaceMembershipInvitationList"],
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
              invalidates: [
                "getDevices",
                "getDevice",
                "getStatusDevices",
                "getStats",
              ],
            },
            {
              onMutations: [
                "updateDeviceStatus",
                "updateDevice",
                "pullTagFromDevice",
                "choiceDevices",
              ],
              invalidates: ["getDevices", "getDevice", "getStatusDevices"],
            },
            {
              onMutations: ["setDeviceCustomField", "deleteDeviceCustomField"],
              invalidates: ["getDevices", "getDevice"],
            },
            {
              onMutations: [
                "updateContainerStatus",
                "deleteContainer",
                "updateContainer",
                "pullTagFromContainer",
              ],
              invalidates: ["getContainers", "getContainer"],
            },
            {
              onMutations: ["clsoeSession"],
              invalidates: ["getSessions", "getSession", "getStatusDevices"],
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
              invalidates: ["getNamespaces", "getNamespace"],
            },
            {
              onMutations: [
                "createCustomer",
                "createSubscription",
                "attachPaymentMethod",
                "detachPaymentMethod",
                "setDefaultPaymentMethod",
              ],
              invalidates: ["getCustomer", "getSubscription", "getNamespace"],
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
