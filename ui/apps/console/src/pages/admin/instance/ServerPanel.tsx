import CopyButton from "@/components/common/CopyButton";
import { useServerInfo } from "@/hooks/useServerInfo";
import { Loaded, Panel, PanelError, Row } from "./Panel";

function Reported({
  label,
  value,
  isLoading,
  isError,
}: {
  label: string;
  value?: string;
  isLoading: boolean;
  isError: boolean;
}) {
  return (
    <Row label={label}>
      <code className="font-mono text-xs truncate">
        <Loaded isLoading={isLoading} isError={isError}>
          {value || "Not reported"}
        </Loaded>
      </code>
      {value && <CopyButton text={value} />}
    </Row>
  );
}

/**
 * The version the server runs and the endpoints it reports for SSH and the API. A value the
 * server does not report reads as Not reported, never as a blank or a placeholder.
 */
export default function ServerPanel() {
  const { version, sshEndpoint, apiEndpoint, isLoading, isError } =
    useServerInfo();
  const status = { isLoading, isError };

  return (
    <Panel title="Server">
      {isError && (
        <PanelError>Couldn&apos;t load server information.</PanelError>
      )}
      <Reported label="Version" value={version} {...status} />
      <Reported label="SSH" value={sshEndpoint} {...status} />
      <Reported label="API" value={apiEndpoint} {...status} />
    </Panel>
  );
}
