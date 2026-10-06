import { type Locator, type Page, expect, test } from "@playwright/test";
import { type ChildProcess, execFileSync, spawn } from "node:child_process";
import sshpk from "sshpk";
import {
  createAccessPolicy,
  createPublicKey,
  createSshIdentity,
  deleteAccessPolicy,
  editNamespace,
  getDevice,
  getDevices,
  getSession,
  getSessionRecord,
  getSessions,
  listAccessPolicies,
  listSshIdentities,
  updateDeviceStatus,
} from "@/client";
import { buildRequestContext } from "./api";
import { isCommunity, isEnterprise } from "./env";
import {
  createTeam,
  createTeamWithOwnerInAnother,
  dismissWizard,
  findRow,
  mfaReason,
  signInAndOpen,
  singleNamespaceReason,
  switchNamespace,
} from "./helpers";
import { fillDigits, signInWithMFA, totp } from "./mfa";
import {
  buildShortId,
  composeExec,
  enableMFA,
  markSamlOrigin,
  mfaSecret,
  password,
  startAgent,
} from "./seed";
import {
  type AnswerOptions,
  answerSignOn,
  disableSaml,
  enableSaml,
  samlReason,
  signInWithSso,
} from "./saml";
import { addKey, buildPrivateKey, setUpVault } from "./vault";

test.use({ launchOptions: { args: ["--disable-webgl"] } });

const deviceLogin = "root";
const devicePassword = "password";
const keyPassphrase = "e2e-key-passphrase";

type Team = Awaited<ReturnType<typeof createTeam>>;

function docker(args: string[]) {
  return execFileSync("docker", args, {
    encoding: "utf-8",
    stdio: ["pipe", "pipe", "pipe"],
    timeout: 30_000,
  }).trim();
}

const agentExec = (container: string, args: string[]) =>
  docker(["exec", container, ...args]);

const agents: string[] = [];

test.afterEach(() => {
  if (agents.length) docker(["rm", "--force", ...agents.splice(0)]);
});

const ownerContext = (team: Team) =>
  buildRequestContext({ token: team.owner.token });

async function waitForPendingDevice(team: Team, name: string) {
  let uid: string | undefined;
  await expect
    .poll(
      async () => {
        const { data } = await getDevices({
          ...ownerContext(team),
          query: { status: "pending" },
        });
        uid = data.find((device) => device.name === name)?.uid;
        return uid;
      },
      { message: `${name} to ask to join ${team.namespace}`, timeout: 30_000 },
    )
    .toBeTruthy();
  if (!uid) {
    throw new Error(`expected ${name} to ask to join ${team.namespace}`);
  }
  return uid;
}

async function createDevice(team: Team) {
  const name = `e2e-device-${buildShortId()}`;
  const container = startAgent(team.tenant, name);
  agents.push(container);
  const uid = await waitForPendingDevice(team, name);
  const context = ownerContext(team);
  await updateDeviceStatus({ ...context, path: { uid, status: "accept" } });
  await expect
    .poll(
      async () => (await getDevice({ ...context, path: { uid } })).data.online,
      { message: `${name} to come online` },
    )
    .toBe(true);
  return { uid, name, container };
}

async function createDeviceAndSignIn(page: Page, team: Team) {
  const device = await createDevice(team);
  await signInAndOpen(page, team.owner.username, "/devices");
  return device;
}

async function createTeamWithKey({ passphrase = "" } = {}) {
  const team = await createTeam({ sshAccessMode: "legacy" });
  const privateKey = buildPrivateKey({ passphrase });
  const publicKey = sshpk
    .parsePrivateKey(privateKey, "pem", { passphrase })
    .toPublic()
    .toString("ssh");
  await createPublicKey({
    ...ownerContext(team),
    body: {
      name: `e2e-key-${buildShortId()}`,
      data: Buffer.from(publicKey).toString("base64"),
      username: ".*",
      filter: { hostname: ".*" },
    },
  });
  return { team, privateKey };
}

async function addVaultKey(page: Page, privateKey: string, passphrase = "") {
  await page.goto("/secure-vault");
  await setUpVault(page);
  return addKey(page, { privateKey, passphrase });
}

async function readIdentitySources(team: Team) {
  const { data } = await listSshIdentities(ownerContext(team));
  return data.map(({ source }) => source);
}

async function readSessions(team: Team) {
  const { data } = await getSessions(ownerContext(team));
  return data;
}

async function readDeviceSessions(team: Team, deviceUid: string) {
  return (await readSessions(team))
    .filter((session) => session.device_uid === deviceUid)
    .map(({ authenticated, web }) => ({ authenticated, web }));
}

async function expectAuthenticatedWebSession(team: Team, deviceUid: string) {
  expect(await readDeviceSessions(team, deviceUid)).toEqual([
    { authenticated: true, web: true },
  ]);
}

async function countActiveWebSessions(team: Team) {
  return (await readSessions(team)).filter(({ active, web }) => active && web)
    .length;
}

const namespaceTab = (page: Page, team: Team) =>
  page.getByRole("tab", { name: new RegExp(`${team.namespace}$`) });

const terminal = (page: Page, device: string) =>
  page.getByRole("region", { name: `Terminal for ${device}` });

const terminalOutput = (page: Page, device: string) =>
  terminal(page, device).locator(".xterm-rows");

async function typeInTerminal(page: Page, device: string, line: string) {
  await terminal(page, device)
    .getByRole("textbox", { name: "Terminal input" })
    .pressSequentially(`${line}\n`);
}

const buildMarker = () => `x${buildShortId()}`;

async function expectShell(page: Page, device: string) {
  const marker = buildMarker();
  await expect(terminalOutput(page, device)).toContainText(":~#", {
    timeout: 15_000,
  });
  await typeInTerminal(page, device, `echo ${marker} | tr a-z A-Z`);
  await expect(terminalOutput(page, device)).toContainText(
    marker.toUpperCase(),
    { timeout: 15_000 },
  );
  return marker;
}

async function openConnect(page: Page, device: string) {
  await page.getByRole("link", { name: "Devices", exact: true }).click();
  await findRow(page, device).getByRole("button", { name: "Connect" }).click();
  const dialog = page.getByRole("dialog", { name: "Connect" });
  await dialog.getByLabel("Login").fill(deviceLogin);
  return dialog;
}

async function openInBrowser(dialog: Locator, passphrase = "") {
  if (passphrase) {
    await dialog.getByLabel("Passphrase", { exact: true }).fill(passphrase);
  }
  await dialog.getByRole("button", { name: "Open in browser" }).click();
}

async function choosePassword(dialog: Locator) {
  await dialog.getByRole("radio", { name: "Password" }).press("Space");
  await dialog.getByLabel("Device password").fill(devicePassword);
}

async function connectWithPassword(page: Page, device: string) {
  const dialog = await openConnect(page, device);
  await choosePassword(dialog);
  await openInBrowser(dialog);
}

async function connectWithPastedKey(
  page: Page,
  device: string,
  privateKey: string,
  passphrase = "",
) {
  const dialog = await openConnect(page, device);
  await dialog.getByRole("radio", { name: "Private key" }).press("Space");
  await dialog.getByRole("textbox", { name: "Private key" }).fill(privateKey);
  await openInBrowser(dialog, passphrase);
}

async function connectWithVaultKey(
  page: Page,
  device: string,
  key: string,
  passphrase = "",
) {
  const dialog = await openConnect(page, device);
  await dialog.getByRole("radio", { name: "Private key" }).press("Space");
  await dialog.getByRole("radio", { name: "Vault" }).press("Space");
  await dialog.getByRole("combobox", { name: "Key" }).selectOption(key);
  await openInBrowser(dialog, passphrase);
}

async function connectWithBrowserKey(page: Page, device: string) {
  await openInBrowser(await openConnect(page, device));
  await page
    .getByRole("dialog", { name: "Register this browser" })
    .getByRole("button", { name: "Register and connect" })
    .click();
}

test.describe("Connection & Authentication", () => {
  test("a password opens a shell on the device", async ({ page }) => {
    const team = await createTeam({ sshAccessMode: "legacy" });
    const device = await createDeviceAndSignIn(page, team);

    await connectWithPassword(page, device.name);

    await expectShell(page, device.name);
    await expectAuthenticatedWebSession(team, device.uid);
  });

  test("a pasted private key signs the challenge", async ({ page }) => {
    const { team, privateKey } = await createTeamWithKey();
    const device = await createDeviceAndSignIn(page, team);

    await connectWithPastedKey(page, device.name, privateKey);

    await expectShell(page, device.name);
    await expectAuthenticatedWebSession(team, device.uid);
  });

  test("an encrypted private key opens with its passphrase", async ({
    page,
  }) => {
    const { team, privateKey } = await createTeamWithKey({
      passphrase: keyPassphrase,
    });
    const device = await createDeviceAndSignIn(page, team);

    await connectWithPastedKey(page, device.name, privateKey, keyPassphrase);

    await expectShell(page, device.name);
    await expectAuthenticatedWebSession(team, device.uid);
  });

  test("a vault key opens a shell", async ({ page }) => {
    const { team, privateKey } = await createTeamWithKey();
    const device = await createDeviceAndSignIn(page, team);
    const key = await addVaultKey(page, privateKey);

    await connectWithVaultKey(page, device.name, key);

    await expectShell(page, device.name);
    await expectAuthenticatedWebSession(team, device.uid);
  });

  test("an encrypted vault key asks for its passphrase", async ({ page }) => {
    const { team, privateKey } = await createTeamWithKey({
      passphrase: keyPassphrase,
    });
    const device = await createDeviceAndSignIn(page, team);
    const key = await addVaultKey(page, privateKey, keyPassphrase);

    await connectWithVaultKey(page, device.name, key, keyPassphrase);

    await expectShell(page, device.name);
    await expectAuthenticatedWebSession(team, device.uid);
  });

  test("identity mode registers a new browser key, then reuses it", async ({
    page,
  }) => {
    const team = await createTeam({ sshAccessMode: "identity" });
    const device = await createDeviceAndSignIn(page, team);

    await connectWithBrowserKey(page, device.name);

    await expectShell(page, device.name);
    await expectAuthenticatedWebSession(team, device.uid);
    expect(await readIdentitySources(team)).toEqual(["browser"]);

    await page.getByRole("button", { name: `Close ${device.name}` }).click();
    await openInBrowser(await openConnect(page, device.name));

    await expectShell(page, device.name);
    await expect(
      page.getByRole("dialog", { name: "Register this browser" }),
    ).toBeHidden();
    expect(await readIdentitySources(team)).toEqual(["browser"]);
  });
});

async function requireReauth(team: Team) {
  const context = ownerContext(team);
  const { data: policies } = await listAccessPolicies(context);
  for (const { id } of policies) {
    await deleteAccessPolicy({ ...context, path: { id } });
  }
  await createAccessPolicy({
    ...context,
    body: {
      name: `e2e-reauth-${buildShortId()}`,
      subject: { type: "all-members", value: "" },
      filter: {},
      logins: ["*"],
      action: "allow",
      require_reauth: true,
      reauth_period: 0,
    },
  });
}

async function signInAsSamlOwner(page: Page, options?: AnswerOptions) {
  const team = await createTeam({ sshAccessMode: "identity" });
  await requireReauth(team);
  markSamlOrigin(team.owner.username);
  const device = await createDevice(team);
  const requests = await answerSignOn(
    page.context(),
    { email: team.owner.email, name: team.owner.username },
    options,
  );
  await signInWithSso(page);
  await expect(page).toHaveURL(/\/dashboard$/);
  await dismissWizard(page);
  return { team, device, requests };
}

async function reauthenticateWithSso(page: Page) {
  const dialog = reauthDialog(page);
  await dialog.getByRole("button", { name: "Continue" }).click();
  await dialog.getByRole("button", { name: "Re-authenticate" }).click();
  return dialog;
}

function expireSamlRelayToken(token: string) {
  const out = composeExec("redis", [
    "valkey-cli",
    "DEL",
    `saml-stepup/${token}`,
  ]);
  if (out !== "1") {
    throw new Error(
      `expected to expire SAML relay token ${token}, got "${out}"`,
    );
  }
}

const reauthDialog = (page: Page) =>
  page.getByRole("dialog", { name: "Confirm SSH login" });

async function reauthenticateWithPassword(page: Page) {
  const dialog = reauthDialog(page);
  await dialog.getByRole("button", { name: "Continue" }).click();
  await dialog.getByLabel("Account password").fill(password);
  await dialog.getByRole("button", { name: "Re-authenticate" }).click();
}

test.describe("Re-authentication", () => {
  test("a policy asks for the password before the shell opens", async ({
    page,
  }) => {
    const team = await createTeam({ sshAccessMode: "identity" });
    await requireReauth(team);
    const device = await createDeviceAndSignIn(page, team);

    await connectWithBrowserKey(page, device.name);
    await reauthenticateWithPassword(page);

    await expectShell(page, device.name);
    await expectAuthenticatedWebSession(team, device.uid);
  });

  test("an MFA user re-authenticates with a TOTP code", async ({ page }) => {
    test.skip(isCommunity, mfaReason);
    const team = await createTeam({ sshAccessMode: "identity" });
    await requireReauth(team);
    const device = await createDevice(team);
    enableMFA(team.owner.username);
    await signInWithMFA(page, team.owner.username);

    await connectWithBrowserKey(page, device.name);
    const dialog = reauthDialog(page);
    await dialog.getByRole("button", { name: "Continue" }).click();
    await fillDigits(page, totp(mfaSecret, 1));
    await dialog.getByRole("button", { name: "Re-authenticate" }).click();

    await expectShell(page, device.name);
    await expectAuthenticatedWebSession(team, device.uid);
  });

  test("rejecting the re-authentication refuses the login", async ({
    page,
  }) => {
    const team = await createTeam({ sshAccessMode: "identity" });
    await requireReauth(team);
    const device = await createDeviceAndSignIn(page, team);

    await connectWithBrowserKey(page, device.name);
    await reauthDialog(page).getByRole("button", { name: "Reject" }).click();

    await expect(terminal(page, device.name).getByRole("alert")).toContainText(
      "Login not approved",
    );
    expect(await readDeviceSessions(team, device.uid)).toEqual([]);
  });

  test.describe("with SAML", () => {
    test.skip(!isEnterprise, samlReason);
    test.beforeEach(() => enableSaml());
    test.afterEach(disableSaml);

    test("a SAML user re-authenticates in the identity provider popup", async ({
      page,
    }) => {
      const { team, device, requests } = await signInAsSamlOwner(page);

      await connectWithBrowserKey(page, device.name);
      await reauthenticateWithSso(page);

      await expectShell(page, device.name);
      expect(
        requests.map(
          ({ document }) => document.getAttributeNode("ForceAuthn")?.value,
        ),
      ).toEqual([undefined, "true"]);
      await expectAuthenticatedWebSession(team, device.uid);
    });

    test("an expired relay token refuses the SAML re-authentication", async ({
      page,
    }) => {
      let expiredRelayState = "";
      const { team, device } = await signInAsSamlOwner(page, {
        beforeAnswer: ({ relayState }) => {
          if (!relayState || expiredRelayState) return;
          expireSamlRelayToken(relayState);
          expiredRelayState = relayState;
        },
      });

      await connectWithBrowserKey(page, device.name);
      const dialog = await reauthenticateWithSso(page);

      await expect(dialog).toContainText(
        "Re-authentication with your provider failed. Please try again.",
      );

      await dialog.getByRole("button", { name: "Re-authenticate" }).click();

      await expectShell(page, device.name);
      await expectAuthenticatedWebSession(team, device.uid);
    });
  });
});

const nativeClientKey = "/tmp/e2e-native-key";

function createNativeClientKey(container: string) {
  agentExec(container, [
    "ssh-keygen",
    "-q",
    "-t",
    "ed25519",
    "-N",
    "",
    "-f",
    nativeClientKey,
  ]);
  return agentExec(container, ["cat", `${nativeClientKey}.pub`]);
}

const nativeClients: ChildProcess[] = [];

function openNativeSSH(container: string, sshid: string) {
  const client = spawn("docker", [
    "exec",
    "-i",
    container,
    "script",
    "-qfc",
    `ssh -tt -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ${nativeClientKey} -p 2222 ${sshid}@server`,
    "/dev/null",
  ]);
  nativeClients.push(client);
  let output = "";
  client.stdout.on("data", (chunk: Buffer) => (output += chunk.toString()));
  client.stderr.on("data", (chunk: Buffer) => (output += chunk.toString()));
  return {
    read: () => output,
    write: (line: string) => client.stdin.write(`${line}\n`),
  };
}

type NativeSSH = ReturnType<typeof openNativeSSH>;

async function readApprovalPath(ssh: NativeSSH, flow: "new" | "confirm") {
  const pattern = new RegExp(`/ssh-identities/${flow}/[A-Z0-9]+`);
  await expect
    .poll(() => ssh.read(), { message: `an /ssh-identities/${flow} link` })
    .toMatch(pattern);
  const path = ssh.read().match(pattern)?.[0];
  if (!path) throw new Error(`expected an /ssh-identities/${flow} link`);
  return path;
}

async function expectNativeShell(ssh: NativeSSH) {
  const marker = buildMarker();
  ssh.write(`echo ${marker} | tr a-z A-Z`);
  await expect
    .poll(() => ssh.read(), { message: `the shell to echo ${marker}` })
    .toContain(marker.toUpperCase());
}

async function readConfirmationCode(page: Page) {
  const output = page.getByRole("status", { name: "Confirmation code" });
  await expect(output).toHaveText(/\w{4} \w{4}/);
  const code = await output.textContent();
  if (!code) throw new Error("expected a confirmation code");
  return code.replaceAll(" ", "");
}

const sshidOf = (team: Team, device: { name: string }) =>
  `${deviceLogin}@${team.namespace}.${device.name}`;

async function openNewKeyApproval(page: Page) {
  const team = await createTeam({ sshAccessMode: "identity" });
  const device = await createDeviceAndSignIn(page, team);
  createNativeClientKey(device.container);
  const ssh = openNativeSSH(device.container, sshidOf(team, device));
  await page.goto(await readApprovalPath(ssh, "new"));
  const dialog = page.getByRole("dialog", { name: "Add SSH key" });
  return { team, ssh, dialog };
}

test.describe("SSH Approval", () => {
  test.afterEach(() => {
    for (const client of nativeClients.splice(0)) client.kill();
  });

  test("approving a new key in the console lets the ssh client in", async ({
    page,
  }) => {
    const { team, ssh, dialog } = await openNewKeyApproval(page);

    await dialog.getByRole("button", { name: "Add key" }).click();
    ssh.write(await readConfirmationCode(page));

    await expectNativeShell(ssh);
    expect(await readIdentitySources(team)).toEqual(["approval"]);
  });

  test("re-authenticating in the console lets the ssh client in", async ({
    page,
  }) => {
    const team = await createTeam({ sshAccessMode: "identity" });
    await requireReauth(team);
    const device = await createDeviceAndSignIn(page, team);
    const publicKey = createNativeClientKey(device.container);
    await createSshIdentity({
      ...ownerContext(team),
      body: { name: `e2e-native-${buildShortId()}`, data: publicKey },
    });
    const ssh = openNativeSSH(device.container, sshidOf(team, device));

    await page.goto(await readApprovalPath(ssh, "confirm"));
    await reauthenticateWithPassword(page);
    ssh.write(await readConfirmationCode(page));

    await expectNativeShell(ssh);
  });

  test("rejecting a new key keeps the ssh client out", async ({ page }) => {
    const { team, ssh, dialog } = await openNewKeyApproval(page);

    await dialog.getByRole("button", { name: "Reject" }).click();
    await expect(
      page.getByText("The login won't go through. Nothing was changed."),
    ).toBeVisible();
    ssh.write("");

    await expect
      .poll(() => ssh.read(), { message: "the ssh client to be refused" })
      .toContain("This login was rejected in the console.");
    expect(await readIdentitySources(team)).toEqual([]);
  });
});

test.describe("Window Management", () => {
  test("two terminals stay connected side by side", async ({ page }) => {
    const team = await createTeam({ sshAccessMode: "legacy" });
    const first = await createDevice(team);
    const second = await createDeviceAndSignIn(page, team);

    await connectWithPassword(page, first.name);
    const firstMarker = await expectShell(page, first.name);
    await namespaceTab(page, team).click();
    await connectWithPassword(page, second.name);
    await expectShell(page, second.name);
    await page.getByRole("tab", { name: first.name }).click();

    await expect(terminalOutput(page, first.name)).toContainText(
      firstMarker.toUpperCase(),
    );
    await expectShell(page, first.name);
    expect(await countActiveWebSessions(team)).toBe(2);
  });

  test("an open terminal survives a namespace switch", async ({ page }) => {
    test.skip(isCommunity, singleNamespaceReason);
    const team = await createTeamWithOwnerInAnother({
      sshAccessMode: "legacy",
    });
    const device = await createDeviceAndSignIn(page, team);

    await connectWithPassword(page, device.name);
    const marker = await expectShell(page, device.name);
    await switchNamespace(page, team.other.namespace);
    await page.getByRole("tab", { name: device.name }).click();

    await expect(terminalOutput(page, device.name)).toContainText(
      marker.toUpperCase(),
    );
    await expectShell(page, device.name);
    expect(await readDeviceSessions(team, device.uid)).toHaveLength(1);
  });
});

async function setNamespaceRecording(team: Team, sessionRecord: boolean) {
  await editNamespace({
    ...ownerContext(team),
    path: { tenant: team.tenant },
    body: { settings: { session_record: sessionRecord } },
  });
}

async function readSessionUid(team: Team, deviceUid: string) {
  const uid = (await readSessions(team)).find(
    (session) => session.device_uid === deviceUid,
  )?.uid;
  if (!uid) throw new Error(`expected a session on ${deviceUid}`);
  return uid;
}

function readCastOutput(cast: string) {
  return cast
    .split("\n")
    .slice(1)
    .filter(Boolean)
    .map((line) => (JSON.parse(line) as [number, string, string])[2])
    .join("");
}

async function recordInBrowserAndPlayBack(page: Page, team: Team) {
  const device = await createDeviceAndSignIn(page, team);

  const dialog = await openConnect(page, device.name);
  await expect(
    dialog.getByRole("switch", {
      name: "Record this session in this browser",
    }),
  ).toBeChecked();
  await choosePassword(dialog);
  await openInBrowser(dialog);
  const marker = await expectShell(page, device.name);
  await expect(
    terminal(page, device.name).getByText("REC", { exact: true }),
  ).toBeVisible();
  await typeInTerminal(page, device.name, "exit");

  await page
    .getByRole("status")
    .filter({ hasText: "Session recorded" })
    .getByRole("button", { name: "View recordings" })
    .click();
  await findRow(page, device.name)
    .getByRole("button", { name: "Play recording" })
    .click();
  await expect(
    page.getByRole("region", { name: `Recording of ${device.name}` }),
  ).toContainText(marker.toUpperCase(), { timeout: 15_000 });
}

test.describe("Session Recording", () => {
  test("a browser recording plays back from the sessions page", async ({
    page,
  }) => {
    const team = await createTeam({ sshAccessMode: "legacy" });
    await setNamespaceRecording(team, false);

    await recordInBrowserAndPlayBack(page, team);
  });

  test("community records in the browser while namespace recording is on", async ({
    page,
  }) => {
    test.skip(!isCommunity, "only community leaves recording to the browser");
    const team = await createTeam({ sshAccessMode: "legacy" });
    await setNamespaceRecording(team, true);

    await recordInBrowserAndPlayBack(page, team);
  });

  test("the server keeps the recording after the session ends", async ({
    page,
  }) => {
    test.skip(isCommunity, "only enterprise and cloud record sessions");
    const team = await createTeam({ sshAccessMode: "legacy" });
    await setNamespaceRecording(team, true);
    const device = await createDeviceAndSignIn(page, team);

    const dialog = await openConnect(page, device.name);
    await expect(
      dialog.getByText(
        "This session will be recorded and stored on the server by namespace policy.",
      ),
    ).toBeVisible();
    await expect(
      dialog.getByRole("switch", {
        name: "Record this session in this browser",
      }),
    ).toHaveCount(0);
    await choosePassword(dialog);
    await openInBrowser(dialog);
    const marker = await expectShell(page, device.name);
    await typeInTerminal(page, device.name, "exit");

    const context = ownerContext(team);
    const uid = await readSessionUid(team, device.uid);
    await expect
      .poll(
        async () => (await getSession({ ...context, path: { uid } })).data,
        { message: `session ${uid} to close recorded` },
      )
      .toMatchObject({ active: false, recorded: true });
    const { data: record } = await getSessionRecord({
      ...context,
      path: { uid, seat: 0 },
    });
    expect(readCastOutput(record)).toContain(marker.toUpperCase());
  });
});
