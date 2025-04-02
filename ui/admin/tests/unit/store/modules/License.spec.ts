import { describe, expect, it } from "vitest";
import { store } from "../../../../src/store";

describe("License", () => {
  const license = {
    expired: false,
    about_to_expire: false,
    grace_period: false,
    id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    issued_at: -1,
    starts_at: -1,
    expires_at: -1,
    allowed_regions: [],
    customer: {
      id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      name: "ShellHub",
      email: "contato@ossystems.com.br",
      company: "O.S. Systems",
    },
    features: {
      devices: -1,
      session_recording: true,
      firewall_rules: true,
      reports: false,
      login_link: false,
    },
  };

  // const newLicense = {
  //   expired: false,
  //   about_to_expire: false,
  //   grace_period: false,
  //   id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  //   issued_at: -1,
  //   starts_at: -1,
  //   expires_at: -1,
  //   allowed_regions: [],
  //   customer: {
  //     id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  //     name: "New ShellHub",
  //     email: "newShelhub@ossystems.com.br",
  //     company: "New O.S. Systems",
  //   },
  //   features: {
  //     devices: -1,
  //     session_recording: true,
  //     firewall_rules: true,
  //     reports: false,
  //     login_link: false,
  //   },
  // };
  it("Return license default variables", () => {
    expect(store.getters["license/get"]).toEqual(undefined);
  });
  it("Verify initial state change for setLicense mutation", () => {
    store.commit("license/setLicense", { data: license });
    expect(store.getters["license/license"]).toEqual(license);
  });
  // TODO: Make the post License with the new license
});
