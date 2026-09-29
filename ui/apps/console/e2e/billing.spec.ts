import { type Page, expect, test as base } from "@playwright/test";
import { randomUUID } from "node:crypto";
import { getCustomer, getNamespace } from "@/client";
import { isCloud } from "./env";
import { signIn, dismissWizard } from "./helpers";
import {
  password,
  buildShortId,
  composeExec,
  createUser,
  createNamespace,
} from "./seed";
import { buildRequestContext, loginAs } from "./api";

const ACCEPTED_CARD = "4242424242424242";
const DECLINED_CARD = "4000000000000002";

type Owner = {
  username: string;
  email: string;
  token: string;
  tenant: string;
};

function stripeCli(...args: string[]) {
  const output = composeExec("stripe-cli", ["stripe", ...args, "--confirm"]);
  const json = output.indexOf("{");
  if (json === -1) {
    throw new Error(
      `expected JSON from stripe ${args.join(" ")}, got: ${output}`,
    );
  }
  return JSON.parse(output.slice(json)) as {
    id: string;
    deleted?: boolean;
    status?: string;
  };
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

async function readNamespace(owner: Owner) {
  const { data } = await getNamespace({
    ...buildRequestContext({ token: owner.token }),
    path: { tenant: owner.tenant },
  });
  return data;
}

const test = base.extend<{ owner: Owner }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright reads a fixture's dependencies from this destructuring, and owner has none
  owner: async ({}, provide) => {
    const user = createUser("billing");
    const tenant = randomUUID();
    createNamespace(user.username, `e2e-billing-${buildShortId()}`, tenant);
    const { token } = await loginAs(user.username, password);
    const owner = { ...user, token, tenant };

    await provide(owner);

    const customer = (await readNamespace(owner)).billing?.customer_id;
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
    const subscription = (await readNamespace(owner)).billing?.subscription?.id;
    if (!subscription) {
      throw new Error("expected a subscription after subscribing");
    }

    cancelStripeSubscription(subscription);

    await expect
      .poll(
        async () => (await readNamespace(owner)).billing?.subscription?.status,
        { timeout: 30_000 },
      )
      .toBe("canceled");
    await page.reload();
    await expect(planGroup(page).getByText("Canceled")).toBeVisible();

    await openSubscribe(page);
    await confirmSubscribe(page);

    expect((await readNamespace(owner)).billing?.subscription?.status).toBe(
      "active",
    );
  });
});
