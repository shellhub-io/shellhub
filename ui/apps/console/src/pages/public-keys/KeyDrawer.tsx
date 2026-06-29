import { useState, FormEvent } from "react";
import { isSdkError } from "@/api/errors";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import {
  UserGroupIcon,
  UserIcon,
  TagIcon,
  ExclamationCircleIcon,
  ClipboardDocumentListIcon,
} from "@heroicons/react/24/outline";
import { DevicesIcon } from "@shellhub/design-system/primitives";
import {
  useCreatePublicKey,
  useUpdatePublicKey,
} from "@/hooks/usePublicKeyMutations";
import type {
  PublicKeyFilterResponse,
  PublicKeyRequest,
  PublicKeyResponse,
} from "@/client";
import { isPublicKeyValid } from "@/utils/sshKeys";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import TagsSelector from "@/components/common/fields/TagsSelector";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import KeyDataInput from "./KeyDataInput";
import { Button } from "@shellhub/design-system/primitives";

/* --- Drawer --- */
function KeyDrawer({
  open,
  editKey,
  onClose,
}: {
  open: boolean;
  editKey: PublicKeyResponse | null;
  onClose: () => void;
}) {
  const createKey = useCreatePublicKey();
  const updateKey = useUpdatePublicKey();
  const isEdit = !!editKey;

  const [name, setName] = useState("");
  const [keyData, setKeyData] = useState("");
  const [keyError, setKeyError] = useState<string | null>(null);
  const [usernameOption, setUsernameOption] = useState<"all" | "username">(
    "all",
  );
  const [username, setUsername] = useState("");
  const [filterOption, setFilterOption] = useState<"all" | "hostname" | "tags">(
    "all",
  );
  const [hostname, setHostname] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mapTagsNames = (tags: PublicKeyFilterResponse["tags"]) =>
    tags.map((t) => t.name);

  useResetOnOpen(open, () => {
    const decodedKeyData = editKey
      ? (() => {
          try {
            return atob(editKey.data);
          } catch {
            return editKey.data;
          }
        })()
      : "";
    const filterInit = editKey
      ? editKey.filter.tags && editKey.filter.tags.length > 0
        ? "tags"
        : editKey.filter.hostname && editKey.filter.hostname !== ".*"
          ? "hostname"
          : "all"
      : "all";

    setName(editKey?.name ?? "");
    setKeyData(decodedKeyData);
    setKeyError(null);
    setUsernameOption(
      editKey ? (editKey.username === ".*" ? "all" : "username") : "all",
    );
    setUsername(editKey && editKey.username !== ".*" ? editKey.username : "");
    setFilterOption(filterInit);
    setHostname(
      editKey && filterInit === "hostname"
        ? (editKey.filter.hostname ?? "")
        : "",
    );
    setSelectedTags(
      editKey && filterInit === "tags" ? mapTagsNames(editKey.filter.tags) : [],
    );
    setSubmitting(false);
    setError(null);
  });

  const handleKeyDataChange = (v: string) => {
    setKeyData(v);
    if (v && !isPublicKeyValid(v))
      setKeyError("This is not a valid public key.");
    else setKeyError(null);
  };

  const handleFileName = (filename: string) => {
    if (!name) setName(filename || "Imported Public Key");
  };

  const buildFilter = (): PublicKeyRequest["filter"] => {
    if (filterOption === "hostname" && hostname) return { hostname };
    if (filterOption === "tags" && selectedTags.length > 0)
      return { tags: selectedTags };
    return { hostname: ".*" };
  };

  const tagError =
    selectedTags.length > 3
      ? "You can select up to 3 tags"
      : filterOption === "tags" && selectedTags.length === 0
        ? "Select at least one tag"
        : undefined;

  const confirmDisabled = isEdit
    ? !name.trim() ||
      (usernameOption === "username" && !username.trim()) ||
      (filterOption === "hostname" && !hostname.trim()) ||
      (filterOption === "tags" &&
        (selectedTags.length === 0 || selectedTags.length > 3))
    : !name.trim() ||
      !keyData.trim() ||
      !!keyError ||
      (usernameOption === "username" && !username.trim()) ||
      (filterOption === "hostname" && !hostname.trim()) ||
      (filterOption === "tags" &&
        (selectedTags.length === 0 || selectedTags.length > 3));

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (confirmDisabled) return;
    setError(null);
    setSubmitting(true);
    try {
      if (isEdit && editKey) {
        await updateKey.mutateAsync({
          path: { fingerprint: editKey.fingerprint },
          body: {
            name: name.trim(),
            username: usernameOption === "all" ? ".*" : username.trim(),
            filter: buildFilter(),
          },
        });
      } else {
        await createKey.mutateAsync({
          body: {
            name: name.trim(),
            data: btoa(keyData.trim()),
            username: usernameOption === "all" ? ".*" : username.trim(),
            filter: buildFilter(),
          },
        });
      }
      onClose();
    } catch (err: unknown) {
      if (!isEdit && isSdkError(err) && err.status === 409) {
        setKeyError("This public key already exists.");
      } else {
        setError(
          err instanceof Error
            ? err.message
            : `Failed to ${isEdit ? "update" : "create"} public key`,
        );
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit Public Key" : "New Public Key"}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            type="submit"
            onClick={() => void handleSubmit()}
            disabled={submitting || confirmDisabled}
            loading={submitting}
          >
            {submitting ? "Saving..." : isEdit ? "Save Changes" : "Create Key"}
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        {/* Name */}
        <InputField
          id="public-key-name"
          label="Name"
          value={name}
          onChange={setName}
          placeholder="Name used to identify the public key"
        />

        {/* Username access */}
        <div>
          <RadioGroupField
            label="Username access"
            value={usernameOption}
            onChange={setUsernameOption}
          >
            <RadioCard
              value="all"
              icon={<UserGroupIcon className="w-4 h-4" />}
              label="Allow any user"
              description="The key will work for all usernames on the device."
            />
            <RadioCard
              value="username"
              icon={<UserIcon className="w-4 h-4" />}
              label="Restrict by username"
              description="Only allow connections matching a username pattern."
            />
          </RadioGroupField>
          {usernameOption === "username" && (
            <div className="mt-2">
              <InputField
                id="public-key-username-pattern"
                label="Username pattern"
                hideLabel
                value={username}
                onChange={setUsername}
                placeholder="e.g. root"
                variant="mono"
              />
            </div>
          )}
        </div>

        {/* Device access */}
        <div>
          <RadioGroupField
            label="Device access"
            value={filterOption}
            onChange={setFilterOption}
          >
            <RadioCard
              value="all"
              icon={<DevicesIcon className="w-4 h-4" />}
              label="All devices"
              description="The key will be accepted by any device in the namespace."
            />
            <RadioCard
              value="hostname"
              icon={<ClipboardDocumentListIcon className="w-4 h-4" />}
              label="Filter by hostname"
              description="Restrict access using a regexp pattern for hostname."
            />
            <RadioCard
              value="tags"
              icon={<TagIcon className="w-4 h-4" />}
              label="Filter by tags"
              description="Restrict access to devices matching specific tags."
            />
          </RadioGroupField>
          {filterOption === "hostname" && (
            <div className="mt-2">
              <InputField
                id="public-key-hostname-pattern"
                label="Hostname pattern"
                hideLabel
                value={hostname}
                onChange={setHostname}
                placeholder="e.g. .*"
                variant="mono"
              />
            </div>
          )}
          {filterOption === "tags" && (
            <div className="mt-2">
              <TagsSelector
                id="public-key-filter-tags"
                label="Filter by tags"
                selected={selectedTags}
                onChange={setSelectedTags}
                error={tagError}
              />
            </div>
          )}
        </div>

        {/* Public Key Data */}
        <KeyDataInput
          value={keyData}
          onChange={handleKeyDataChange}
          error={keyError || undefined}
          disabled={isEdit}
          onFileName={handleFileName}
        />

        {/* Error */}
        {error && (
          <p className="text-xs font-mono text-accent-red flex items-center gap-1.5">
            <ExclamationCircleIcon
              className="w-3.5 h-3.5 shrink-0"
              strokeWidth={2}
            />
            {error}
          </p>
        )}
      </form>
    </Drawer>
  );
}

export default KeyDrawer;
