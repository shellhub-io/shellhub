import { type Page, expect, test as base } from "@playwright/test";
import { randomUUID } from "node:crypto";
import { getCustomer, getNamespace, getNamespaces } from "@/client";
import { isCloud } from "./env";
import { createUser, signIn, dismissWizard } from "./helpers";
import {
  password,
  buildShortId,
  composeExec,
  composeLogs,
  createNamespace,
  setBillingCustomer,
} from "./seed";
import { buildRequestContext, loginAs } from "./api";

const ACCEPTED_CARD = "4242424242424242";
const DECLINED_CARD = "4000000000000002";
const FREE_PLAN_MAX_DEVICES = 3;

type Owner = {
  username: string;
  email: string;
  token: string;
  tenant: string;
  namespace: string;
  customer?: string;
  testClock?: string;
};

type StripeObject = {
  id: string;
  deleted?: boolean;
  status?: string;
};

function stripeCli<T = StripeObject>(...args: string[]): T {
  const output = composeExec("stripe-cli", ["stripe", ...args, "--confirm"]);
  const json = output.indexOf("{");
  if (json === -1) {
    throw new Error(
      `expected JSON from stripe ${args.join(" ")}, got: ${output}`,
    );
  }
  return JSON.parse(output.slice(json)) as T;
}

function deleteStripeCustomer(id: string) {
  const customer = stripeCli("customers", "delete", id);
  if (!customer.deleted) {
    throw new Error(`expected Stripe to delete customer ${id}`);
  }
}

function cancelStripeSubscription(id: string) {
  const subscription = stripeCli("subscriptions", "cancel", id);
  if (subscription.status !== "canceled") {
    throw new Error(
      `expected Stripe to cancel subscription ${id}, got ${subscription.status}`,
    );
  }
}

function createTestClockCustomer(owner: Owner) {
  const clock = stripeCli(
    "test_helpers",
    "test_clocks",
    "create",
    "-d",
    `frozen_time=${Math.floor(Date.now() / 1000)}`,
  );
  owner.testClock = clock.id;
  const customer = stripeCli(
    "customers",
    "create",
    "-d",
    `test_clock=${clock.id}`,
    "-d",
    `email=${owner.email}`,
    "-d",
    `metadata[tenant_id]=${owner.tenant}`,
  );
  setBillingCustomer(owner.tenant, customer.id);
  return { clock: clock.id, customer: customer.id };
}

function scheduleStripeCancellationAtPeriodEnd(id: string) {
  const subscription = stripeCli<StripeObject & { cancel_at: number | null }>(
    "subscriptions",
    "update",
    id,
    "-d",
    "cancel_at_period_end=true",
  );
  if (!subscription.cancel_at) {
    throw new Error(`expected Stripe to schedule the end of ${id}`);
  }
  return subscription.cancel_at;
}

function advanceTestClock(id: string, to: number) {
  stripeCli(
    "test_helpers",
    "test_clocks",
    "advance",
    id,
    "-d",
    `frozen_time=${to}`,
  );
}

function finalizeDraftInvoice(subscription: string) {
  const [draft] = stripeCli<{ data: StripeObject[] }>(
    "invoices",
    "list",
    "-d",
    `subscription=${subscription}`,
    "-d",
    "status=draft",
  ).data;
  if (!draft) {
    throw new Error(`expected a draft final invoice for ${subscription}`);
  }
  stripeCli("invoices", "finalize_invoice", draft.id);
}

function readTestClockStatus(id: string) {
  return stripeCli("test_helpers", "test_clocks", "retrieve", id).status;
}

function deleteTestClock(id: string) {
  const clock = stripeCli("test_helpers", "test_clocks", "delete", id);
  if (!clock.deleted) {
    throw new Error(`expected Stripe to delete test clock ${id}`);
  }
}

function listActiveSubscriptions(customer: string) {
  return stripeCli<{ data: StripeObject[] }>(
    "subscriptions",
    "list",
    "-d",
    `customer=${customer}`,
    "-d",
    "status=active",
  ).data;
}

const hasAbandonedFinalInvoice = (subscription: string) =>
  composeLogs("server")
    .split("\n")
    .some(
      (line) =>
        line.includes(
          "subscription was canceled for a reason other than a failed payment",
        ) && line.includes(subscription),
    );

async function readNamespace(owner: Owner) {
  const { data } = await getNamespace({
    ...buildRequestContext({ token: owner.token }),
    path: { tenant: owner.tenant },
  });
  return data;
}

async function readSubscriptionStatus(owner: Owner) {
  return (await readNamespace(owner)).billing?.subscription?.status;
}

async function readTenants(owner: Owner) {
  const { token } = await loginAs(owner.username, password);
  const { data } = await getNamespaces(buildRequestContext({ token }));
  return data.map((namespace) => namespace.tenant_id);
}

const test = base.extend<{ owner: Owner }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright reads a fixture's dependencies from this destructuring, and owner has none
  owner: async ({}, provide, testInfo) => {
    const user = await createUser("billing");
    const tenant = randomUUID();
    const namespace = `e2e-billing-${buildShortId()}`;
    createNamespace(user.username, namespace, tenant);
    const { token } = await loginAs(user.username, password);
    const owner: Owner = { ...user, token, tenant, namespace };

    await provide(owner);

    if (testInfo.status !== testInfo.expectedStatus) {
      await testInfo.attach("stripe-cli-events.log", {
        body: composeLogs("stripe-cli")
          .split("\n")
          .filter((line) => line.includes("-->") || line.includes("<--"))
          .join("\n"),
        contentType: "text/plain",
      });
      await testInfo.attach("server-billing.log", {
        body: composeLogs("server")
          .split("\n")
          .filter(
            (line) =>
              line.includes("webhook-billing") || line.includes(owner.tenant),
          )
          .join("\n"),
        contentType: "text/plain",
      });
    }

    if (owner.testClock) {
      deleteTestClock(owner.testClock);
      return;
    }

    const customer =
      owner.customer ?? (await readNamespace(owner)).billing?.customer_id;
    if (customer) deleteStripeCustomer(customer);
  },
});

async function openBilling(page: Page, owner: Owner) {
  await signIn(page, owner.username, password);
  await expect(page).toHaveURL(/\/dashboard$/);
  await dismissWizard(page);
  await page.goto("/settings/billing");
}

function subscribeDialog(page: Page) {
  return page.getByRole("dialog", { name: "Subscribe to ShellHub Cloud" });
}

function planGroup(page: Page) {
  return page.getByRole("group", { name: "Plan" });
}

async function openSubscribe(page: Page) {
  await page.getByRole("button", { name: "Subscribe" }).click();
  await subscribeDialog(page).getByRole("button", { name: "Next" }).click();
}

async function saveCard(page: Page, number: string) {
  const dialog = subscribeDialog(page);
  const card = dialog.frameLocator(
    'iframe[title="Secure card payment input frame"]',
  );
  await card.locator('[name="cardnumber"]').fill(number);
  await card.locator('[name="exp-date"]').fill("12 / 34");
  await card.locator('[name="cvc"]').fill("123");
  await dialog.getByRole("button", { name: "Save card" }).click();
}

async function expectSavedCard(page: Page) {
  await expect(
    subscribeDialog(page).getByLabel("Default payment method"),
  ).toBeVisible({ timeout: 15_000 });
}

async function addCard(page: Page, number: string) {
  await subscribeDialog(page)
    .getByRole("button", { name: "Add payment method" })
    .click();
  await saveCard(page, number);
}

async function confirmSubscribe(page: Page) {
  const dialog = subscribeDialog(page);
  await expectSavedCard(page);
  await dialog.getByRole("button", { name: "Next" }).click();
  await dialog.getByRole("button", { name: "Confirm subscription" }).click();
  await expect(
    dialog.getByRole("heading", { name: "Subscription activated" }),
  ).toBeVisible({ timeout: 30_000 });
  await dialog.getByRole("button", { name: "Done" }).click();
  await expect(planGroup(page).getByText("Active")).toBeVisible();
}

async function subscribeWithNewCard(page: Page) {
  await openSubscribe(page);
  await addCard(page, ACCEPTED_CARD);
  await confirmSubscribe(page);
}

async function readSubscription(owner: Owner) {
  const { billing } = await readNamespace(owner);
  if (!billing?.customer_id || !billing.subscription?.id) {
    throw new Error("expected a subscription after subscribing");
  }
  return { customer: billing.customer_id, id: billing.subscription.id };
}

async function submitNamespaceDeletion(page: Page, owner: Owner) {
  await page.getByRole("button", { name: "Delete namespace" }).click();
  const dialog = page.getByRole("dialog", { name: "Delete namespace" });
  await dialog
    .getByLabel(`Type "${owner.namespace}" to confirm`)
    .fill(owner.namespace);
  await dialog.getByRole("button", { name: "Delete namespace" }).click();
  return dialog;
}

test.describe("Billing", () => {
  test.skip(!isCloud, "billing only exists in the cloud edition");

  test("subscribing unlocks unlimited devices and one more namespace", async ({
    page,
    owner,
  }) => {
    const before = await loginAs(owner.username, password);
    await openBilling(page, owner);

    await subscribeWithNewCard(page);

    await expect(
      planGroup(page).getByText("Premium", { exact: true }),
    ).toBeVisible();
    const namespace = await readNamespace(owner);
    expect(namespace.max_devices).toBe(-1);
    const { data: customer } = await getCustomer(
      buildRequestContext({ token: owner.token }),
    );
    expect(customer.id).toBe(namespace.billing?.customer_id);
    expect(customer.email).toBe(owner.email);
    const after = await loginAs(owner.username, password);
    expect(after.max_namespaces).toBe(before.max_namespaces + 1);
  });

  test("a declined card is refused when it is saved", async ({
    page,
    owner,
  }) => {
    await openBilling(page, owner);
    await openSubscribe(page);

    await addCard(page, DECLINED_CARD);

    const dialog = subscribeDialog(page);
    await expect(dialog.getByRole("alert")).toHaveText(
      "Your card was declined.",
    );
    await saveCard(page, ACCEPTED_CARD);
    await expectSavedCard(page);
    const { data: customer } = await getCustomer(
      buildRequestContext({ token: owner.token }),
    );
    expect(
      (customer.payment_methods ?? []).map(({ number }) => number.slice(-4)),
    ).toEqual([ACCEPTED_CARD.slice(-4)]);
  });

  test("the portal opens on Stripe", async ({ page, owner }) => {
    await openBilling(page, owner);
    await subscribeWithNewCard(page);

    const popup = page.waitForEvent("popup");
    await page.getByRole("button", { name: "Open portal" }).click();

    expect(new URL((await popup).url()).host).toBe("billing.stripe.com");
  });

  test("a canceled namespace can subscribe again", async ({ page, owner }) => {
    test.setTimeout(60_000);
    await openBilling(page, owner);
    await subscribeWithNewCard(page);
    const subscription = await readSubscription(owner);

    cancelStripeSubscription(subscription.id);

    await expect
      .poll(() => readSubscriptionStatus(owner), { timeout: 30_000 })
      .toBe("canceled");
    await page.reload();
    await expect(planGroup(page).getByText("Canceled")).toBeVisible();

    await openSubscribe(page);
    await confirmSubscribe(page);

    expect(await readSubscriptionStatus(owner)).toBe("active");
  });

  test("a namespace with an active subscription is deleted only once it is canceled", async ({
    page,
    owner,
  }) => {
    test.setTimeout(60_000);
    await openBilling(page, owner);
    await subscribeWithNewCard(page);
    const subscription = await readSubscription(owner);
    owner.customer = subscription.customer;
    await page.goto("/settings/general");

    const dialog = await submitNamespaceDeletion(page, owner);

    await expect(
      dialog.getByText("Couldn't delete the namespace."),
    ).toBeVisible();
    expect(await readTenants(owner)).toEqual([owner.tenant]);

    cancelStripeSubscription(subscription.id);
    await expect
      .poll(() => readSubscriptionStatus(owner), { timeout: 30_000 })
      .toBe("canceled");
    await page.reload();
    await submitNamespaceDeletion(page, owner);

    await expect.poll(() => readTenants(owner)).toEqual([]);
  });

  test("cancelling at the end of the period is not undone when the final invoice is paid", async ({
    page,
    owner,
  }) => {
    test.setTimeout(360_000);
    const { clock, customer } = createTestClockCustomer(owner);
    await openBilling(page, owner);
    await subscribeWithNewCard(page);
    const subscription = await readSubscription(owner);

    const periodEnd = scheduleStripeCancellationAtPeriodEnd(subscription.id);
    await expect
      .poll(() => readSubscriptionStatus(owner), { timeout: 30_000 })
      .toBe("to_cancel_at_end_of_period");

    advanceTestClock(clock, periodEnd + 60);
    await expect
      .poll(() => readTestClockStatus(clock), {
        timeout: 180_000,
        intervals: [2_000],
      })
      .toBe("ready");
    await expect
      .poll(() => readSubscriptionStatus(owner), { timeout: 30_000 })
      .toBe("canceled");
    finalizeDraftInvoice(subscription.id);
    await expect
      .poll(() => hasAbandonedFinalInvoice(subscription.id), {
        timeout: 30_000,
        message: "the server to abandon the final invoice's invoice.paid",
      })
      .toBe(true);

    const namespace = await readNamespace(owner);
    expect(namespace.billing?.subscription?.status).toBe("canceled");
    expect(namespace.max_devices).toBe(FREE_PLAN_MAX_DEVICES);
    expect(listActiveSubscriptions(customer)).toEqual([]);
  });
});
