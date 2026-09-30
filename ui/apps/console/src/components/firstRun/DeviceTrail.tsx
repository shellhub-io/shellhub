import { useLocation } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import FirstRunLayout from "./FirstRunLayout";
import {
  DEVICE_STEP_TITLES,
  SETUP_STEP_TITLES,
  Trail,
  TrailStep,
  type TrailStepState,
} from "./Trail";
import InstallStep from "./InstallStep";
import PairingCodeField from "./PairingCodeField";
import PairStep from "./PairStep";
import ShellStep from "./ShellStep";
import { useFirstDevice, type FirstDeviceStage } from "./useFirstDevice";
import { readFirstRunEntry } from "./entry";

const DEVICE_STAGES: FirstDeviceStage[] = ["install", "pair", "shell"];

/**
 * The first run of an account, inside its namespace: install the agent, pair the device, open a
 * shell. The steps before it (the survey and the setup form, or getting the namespace) are
 * already behind the user, and show as done; setup hands over what it did through the route
 * state.
 */
export default function DeviceTrail() {
  const location = useLocation();
  const entry = readFirstRunEntry(location.state);
  const email = useAuthStore((s) => s.email);
  const name = useAuthStore((s) => s.name);
  const flow = useFirstDevice();
  const nsName = flow.namespace;

  const stateOf = (step: FirstDeviceStage): TrailStepState => {
    const at = DEVICE_STAGES.indexOf(flow.stage);
    const index = DEVICE_STAGES.indexOf(step);
    if (index < at) return "done";
    return index === at ? "active" : "upcoming";
  };

  const namespaceStep = entry?.survey ? 2 : 1;

  return (
    <FirstRunLayout
      eyebrow={`Welcome, ${name || email || "back"}`}
      signedIn
      inConsole
    >
      <Trail>
        {entry?.survey && (
          <TrailStep
            number={1}
            title={SETUP_STEP_TITLES.survey}
            state="done"
            summary="Thanks"
          />
        )}
        <TrailStep
          number={namespaceStep}
          title={entry ? SETUP_STEP_TITLES.account : "Namespace"}
          state="done"
          summary={
            entry ? (
              <>
                Signed in as <b className="font-medium">{email}</b> · namespace{" "}
                <b className="font-medium">{nsName}</b>
              </>
            ) : (
              <b className="font-medium">{nsName}</b>
            )
          }
        />
        <TrailStep
          number={namespaceStep + 1}
          title={DEVICE_STEP_TITLES.install}
          state={stateOf("install")}
          summary={
            flow.device || flow.pairable ? (
              <>
                Agent running on{" "}
                <b className="font-medium">
                  {flow.device?.name ?? flow.pairable?.name}
                </b>
              </>
            ) : undefined
          }
        >
          <InstallStep>
            <PairingCodeField
              onSubmit={flow.submitCode}
              isPending={flow.resolving}
              error={flow.codeError}
            />
          </InstallStep>
        </TrailStep>
        <TrailStep
          number={namespaceStep + 2}
          title={DEVICE_STEP_TITLES.pair}
          state={stateOf("pair")}
          summary={
            flow.paired ? (
              <>
                Paired into <b className="font-medium">{nsName}</b>
              </>
            ) : (
              "Accepted from the link on the device"
            )
          }
        >
          {flow.pairable && (
            <PairStep
              device={flow.pairable}
              code={flow.code}
              namespace={nsName}
              isPending={flow.pairing}
              error={flow.acceptError}
              onPair={flow.pair}
              onBack={flow.backToCode}
            />
          )}
        </TrailStep>
        <TrailStep
          number={namespaceStep + 3}
          title={DEVICE_STEP_TITLES.shell}
          state={stateOf("shell")}
        >
          {flow.device && <ShellStep device={flow.device} namespace={nsName} />}
        </TrailStep>
      </Trail>
    </FirstRunLayout>
  );
}
